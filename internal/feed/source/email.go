package source

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/mmcdole/gofeed"

	"MrRSS/internal/version"
)

// EmailSource fetches newsletter emails via IMAP.
type EmailSource struct{}

// NewEmailSource creates a new email source.
func NewEmailSource() *EmailSource {
	return &EmailSource{}
}

// Type returns the source type identifier.
func (e *EmailSource) Type() Type {
	return TypeEmail
}

// Validate checks if the configuration is valid for email source.
func (e *EmailSource) Validate(config *Config) error {
	if config == nil {
		return errors.New("config is nil")
	}
	if config.EmailIMAPServer == "" {
		return errors.New("IMAP server is required for email source")
	}
	if config.EmailUsername == "" {
		return errors.New("email username is required for email source")
	}
	if config.EmailPassword == "" {
		return errors.New("email password is required for email source")
	}
	if config.EmailIMAPPort == 0 {
		config.EmailIMAPPort = 993 // Default IMAP SSL port
	}
	if config.EmailFolder == "" {
		config.EmailFolder = "INBOX"
	}
	return nil
}

// Fetch retrieves emails from IMAP server and converts them to feed format.
func (e *EmailSource) Fetch(ctx context.Context, config *Config) (*gofeed.Feed, error) {
	if err := e.Validate(config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Connect to IMAP server
	c, err := e.connectToIMAP(config)
	if err != nil {
		return nil, fmt.Errorf("IMAP connection failed: %w", err)
	}
	defer c.Logout()

	// Select mailbox
	_, err = c.Select(config.EmailFolder, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select mailbox %s: %w", config.EmailFolder, err)
	}

	// Determine UID range for fetching
	fromUID := uint32(1)
	if config.EmailLastUID > 0 {
		fromUID = uint32(config.EmailLastUID + 1)
	}

	// Search for emails
	criteria := imap.NewSearchCriteria()
	seqset := new(imap.SeqSet)
	seqset.AddRange(fromUID, ^uint32(0))
	criteria.Uid = seqset
	criteria.Since = time.Now().AddDate(0, -1, 0) // Last month
	if sender := strings.TrimSpace(config.EmailAddress); sender != "" {
		criteria.Header.Add("From", sender)
	}

	uids, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("IMAP search failed: %w", err)
	}

	feed := &gofeed.Feed{
		Title:       fmt.Sprintf("Email: %s", config.EmailUsername),
		Description: fmt.Sprintf("Emails from %s/%s", config.EmailIMAPServer, config.EmailFolder),
		Items:       []*gofeed.Item{},
	}

	if len(uids) == 0 {
		return feed, nil
	}

	// Fetch emails in batches
	batchSize := 50
	for i := 0; i < len(uids); i += batchSize {
		end := i + batchSize
		if end > len(uids) {
			end = len(uids)
		}
		batchUIDs := uids[i:end]

		items, err := e.fetchEmailBatch(c, batchUIDs, config.EmailAddress)
		if err != nil {
			return nil, err
		}
		feed.Items = append(feed.Items, items...)
	}

	return feed, nil
}

// connectToIMAP establishes a connection to the IMAP server.
func (e *EmailSource) connectToIMAP(config *Config) (*client.Client, error) {
	server := fmt.Sprintf("%s:%d", config.EmailIMAPServer, config.EmailIMAPPort)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         config.EmailIMAPServer,
	}

	// Connect with TLS
	c, err := client.DialTLS(server, tlsConfig)
	if err != nil {
		// Fallback to non-TLS
		c, err = client.Dial(server)
		if err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
	}

	// Send ID command (RFC 2971) - required by some providers
	_ = e.sendIMAPID(c)

	// Login
	if err := c.Login(config.EmailUsername, config.EmailPassword); err != nil {
		c.Logout()
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return c, nil
}

// sendIMAPID sends the IMAP ID command.
func (e *EmailSource) sendIMAPID(c *client.Client) error {
	idClient := id.NewClient(c)

	supported, err := idClient.SupportID()
	if err != nil || !supported {
		return nil
	}

	clientID := id.ID{
		id.FieldName:    "MrRSS",
		id.FieldVersion: version.Version,
		id.FieldVendor:  "MrRSS",
		id.FieldOS:      runtime.GOOS,
	}

	_, err = idClient.ID(clientID)
	return err
}

// fetchEmailBatch fetches and parses a batch of emails.
func (e *EmailSource) fetchEmailBatch(c *client.Client, uids []uint32, senderFilter string) ([]*gofeed.Item, error) {
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	messages := make(chan *imap.Message, len(uids))
	err := c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody}, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	items := make([]*gofeed.Item, 0, len(uids))
	for msg := range messages {
		if msg == nil {
			continue
		}
		if !emailMessageMatchesSenderFilter(msg, senderFilter) {
			continue
		}
		if item := e.parseEmailToItem(msg); item != nil {
			items = append(items, item)
		}
	}

	return items, nil
}

func emailMessageMatchesSenderFilter(msg *imap.Message, senderFilter string) bool {
	filter := strings.ToLower(strings.TrimSpace(senderFilter))
	if filter == "" {
		return true
	}
	if msg == nil || msg.Envelope == nil || len(msg.Envelope.From) == 0 {
		return false
	}

	for _, sender := range msg.Envelope.From {
		if sender == nil {
			continue
		}
		values := []string{
			sender.Address(),
			sender.MailboxName,
			sender.HostName,
			sender.PersonalName,
		}
		for _, value := range values {
			if strings.Contains(strings.ToLower(strings.TrimSpace(value)), filter) {
				return true
			}
		}
	}

	return false
}

// parseEmailToItem converts an IMAP message to a gofeed Item.
func (e *EmailSource) parseEmailToItem(msg *imap.Message) *gofeed.Item {
	item := &gofeed.Item{
		Title:     msg.Envelope.Subject,
		Link:      fmt.Sprintf("email://%d", msg.Uid),
		GUID:      fmt.Sprintf("email-%d", msg.Uid),
		Published: msg.Envelope.Date.Format(time.RFC1123),
	}

	// Set author from sender
	if len(msg.Envelope.From) > 0 {
		sender := msg.Envelope.From[0]
		item.Author = &gofeed.Person{
			Name:  sender.PersonalName,
			Email: sender.Address(),
		}
		if item.Title == "" {
			if sender.PersonalName != "" {
				item.Title = fmt.Sprintf("Email from %s", sender.PersonalName)
			} else {
				item.Title = fmt.Sprintf("Email from %s", sender.Address())
			}
		}
	}

	// Extract body
	item.Description = e.extractEmailBody(msg)
	if item.Description == "" {
		item.Description = "(No content available)"
	}

	return item
}

// extractEmailBody extracts text/HTML content from the message.
func (e *EmailSource) extractEmailBody(msg *imap.Message) string {
	for _, r := range msg.Body {
		data, err := io.ReadAll(r)
		if err != nil {
			continue
		}
		content := string(data)
		if strings.TrimSpace(content) != "" {
			return e.cleanEmailContent(content)
		}
	}
	return ""
}

// cleanEmailContent removes tracking elements from email HTML.
func (e *EmailSource) cleanEmailContent(html string) string {
	cleaner := strings.NewReplacer(
		`<img src="https://`, `<img data-tracking="true" src="https://`,
		`<style>`, `<style data-remove="true">`,
	)
	return strings.TrimSpace(cleaner.Replace(html))
}
