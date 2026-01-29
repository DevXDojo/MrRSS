package feed

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/mmcdole/gofeed"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"MrRSS/internal/version"
)

// EmailFetcher handles fetching and parsing newsletter emails
type EmailFetcher struct {
	db     *database.DB
	parser *gofeed.Parser
}

// NewEmailFetcher creates a new email fetcher
func NewEmailFetcher(db *database.DB) *EmailFetcher {
	return &EmailFetcher{
		db:     db,
		parser: gofeed.NewParser(),
	}
}

// FetchEmails fetches new emails from IMAP and converts them to feed items
func (ef *EmailFetcher) FetchEmails(ctx context.Context, feed *models.Feed) ([]*gofeed.Item, error) {
	if feed.EmailIMAPServer == "" || feed.EmailUsername == "" || feed.EmailPassword == "" {
		return nil, fmt.Errorf("IMAP credentials not configured")
	}

	// Connect to IMAP server
	c, err := ef.connectToIMAP(feed)
	if err != nil {
		return nil, fmt.Errorf("IMAP connection failed: %w", err)
	}
	defer c.Logout()

	// Select mailbox
	_, err = c.Select(feed.EmailFolder, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select mailbox %s: %w", feed.EmailFolder, err)
	}

	// Determine UID range for fetching
	fromUID := uint32(1)
	if feed.EmailLastUID > 0 {
		fromUID = uint32(feed.EmailLastUID + 1)
	}

	// Search for emails newer than last processed UID
	criteria := imap.NewSearchCriteria()
	seqset := new(imap.SeqSet)
	seqset.AddRange(fromUID, ^uint32(0)) // Use ^uint32(0) as max UID
	criteria.Uid = seqset
	criteria.Since = time.Now().AddDate(0, -1, 0) // Last 1 month

	uids, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("IMAP search failed: %w", err)
	}

	if len(uids) == 0 {
		return nil, nil
	}

	// Fetch emails in batches
	batchSize := 50
	items := make([]*gofeed.Item, 0, len(uids))
	maxUID := feed.EmailLastUID

	for i := 0; i < len(uids); i += batchSize {
		end := i + batchSize
		if end > len(uids) {
			end = len(uids)
		}
		batchUIDs := uids[i:end]

		// Track max UID in this batch
		if int(batchUIDs[len(batchUIDs)-1]) > maxUID {
			maxUID = int(batchUIDs[len(batchUIDs)-1])
		}

		batchItems, err := ef.fetchEmailBatch(c, batchUIDs)
		if err != nil {
			return nil, err
		}
		items = append(items, batchItems...)
	}

	// Update last UID if we processed new emails
	if maxUID > feed.EmailLastUID {
		if err := ef.db.UpdateFeedEmailLastUID(feed.ID, maxUID); err != nil {
			return items, fmt.Errorf("failed to update last UID: %w", err)
		}
	}

	return items, nil
}

// connectToIMAP establishes a connection to the IMAP server
func (ef *EmailFetcher) connectToIMAP(feed *models.Feed) (*client.Client, error) {
	server := fmt.Sprintf("%s:%d", feed.EmailIMAPServer, feed.EmailIMAPPort)

	// Create TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         feed.EmailIMAPServer,
	}

	// Connect with TLS
	c, err := client.DialTLS(server, tlsConfig)
	if err != nil {
		// Fallback to non-TLS if TLS fails
		c, err = client.Dial(server)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
		}
	}

	// Send ID command before login (RFC 2971)
	// This is required by some email providers like NetEase (163, 126)
	if err := ef.sendIMAPID(c); err != nil {
		// Log the error but continue - ID command is optional for most servers
		// Some servers may not support the ID extension
	}

	// Login
	if err := c.Login(feed.EmailUsername, feed.EmailPassword); err != nil {
		c.Logout()
		return nil, fmt.Errorf("IMAP authentication failed: %w", err)
	}

	return c, nil
}

// sendIMAPID sends the IMAP ID command to identify the client
// This is required by some email providers (e.g., NetEase 163/126) as per RFC 2971
func (ef *EmailFetcher) sendIMAPID(c *client.Client) error {
	idClient := id.NewClient(c)

	// Check if server supports ID extension
	supported, err := idClient.SupportID()
	if err != nil || !supported {
		// Server doesn't support ID extension, skip
		return nil
	}

	// Send client identification
	clientID := id.ID{
		id.FieldName:    "MrRSS",
		id.FieldVersion: version.Version,
		id.FieldVendor:  "MrRSS",
		id.FieldOS:      runtime.GOOS,
	}

	_, err = idClient.ID(clientID)
	return err
}

// fetchEmailBatch fetches and parses a batch of emails
func (ef *EmailFetcher) fetchEmailBatch(c *client.Client, uids []uint32) ([]*gofeed.Item, error) {
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	// Fetch the envelope and body
	messages := make(chan *imap.Message, len(uids))
	err := c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody}, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	items := make([]*gofeed.Item, 0, len(uids))

	for msg := range messages {
		if msg == nil {
			break
		}

		item, err := ef.parseEmailToItem(msg)
		if err != nil {
			// Skip invalid emails but continue processing others
			continue
		}

		if item != nil {
			items = append(items, item)
		}
	}

	return items, nil
}

// parseEmailToItem converts an IMAP message to a gofeed Item
func (ef *EmailFetcher) parseEmailToItem(msg *imap.Message) (*gofeed.Item, error) {
	item := &gofeed.Item{
		Title:     msg.Envelope.Subject,
		Link:      fmt.Sprintf("email://%d", msg.Uid),
		GUID:      fmt.Sprintf("email-%d", msg.Uid),
		Published: msg.Envelope.Date.Format(time.RFC1123),
	}

	// Extract sender as author if available
	if len(msg.Envelope.From) > 0 {
		sender := msg.Envelope.From[0]
		item.Author = &gofeed.Person{
			Name:  sender.PersonalName,
			Email: sender.Address(),
		}
		// If no subject, use sender's name as title
		if item.Title == "" && sender.PersonalName != "" {
			item.Title = fmt.Sprintf("Email from %s", sender.PersonalName)
		} else if item.Title == "" && sender.Address() != "" {
			item.Title = fmt.Sprintf("Email from %s", sender.Address())
		}
	}

	// Extract email body
	if item.Description, _ = ef.extractEmailBody(msg); item.Description == "" {
		// Fallback if no body found
		item.Description = "(No content available)"
	}

	// Clean HTML description
	item.Description = cleanEmailContent(item.Description)

	return item, nil
}

// extractEmailBody extracts the text/HTML content from an email message
func (ef *EmailFetcher) extractEmailBody(msg *imap.Message) (string, error) {
	// Try to get the body section from the message
	for _, r := range msg.Body {
		// r is already an imap.Literal (io.Reader)
		data, err := io.ReadAll(r)
		if err != nil {
			continue
		}

		content := string(data)
		if strings.TrimSpace(content) != "" {
			return content, nil
		}
	}

	return "", fmt.Errorf("no body content found")
}

// cleanEmailContent removes unnecessary elements from email HTML
func cleanEmailContent(html string) string {
	// Remove common email tracking elements and clean up HTML
	cleaner := strings.NewReplacer(
		// Remove tracking pixels
		`<img src="https://`, `<img data-tracking="true" src="https://`,
		// Remove unwanted tags
		`<style>`, `<style data-remove="true">`,
	)

	html = cleaner.Replace(html)

	// Additional cleanup can be added here
	return strings.TrimSpace(html)
}
