package models

import "time"

type Feed struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	URL                string    `json:"url"`
	Link               string    `json:"link"` // Website homepage link
	Description        string    `json:"description"`
	Category           string    `json:"category"`
	ImageURL           string    `json:"image_url"` // New field
	Position           int       `json:"position"`  // Position within category for custom ordering
	LastUpdated        time.Time `json:"last_updated"`
	LastError          string    `json:"last_error,omitempty"`  // Track last fetch error
	DiscoveryCompleted bool      `json:"discovery_completed"`   // Track if discovery has been run
	ScriptPath         string    `json:"script_path,omitempty"` // Path to custom script for fetching feed
	HideFromTimeline   bool      `json:"hide_from_timeline"`    // Hide articles from timeline views
	ProxyURL           string    `json:"proxy_url,omitempty"`   // Custom proxy URL for this feed (overrides global)
	ProxyEnabled       bool      `json:"proxy_enabled"`         // Whether to use proxy for this feed
	RefreshInterval    int       `json:"refresh_interval"`      // Custom refresh interval in minutes (0 = use global, -1 = intelligent, -2 = never, >0 = custom minutes)
	IsImageMode        bool      `json:"is_image_mode"`         // Whether this feed is for image gallery mode
	// XPath support for HTML/XML scraping
	Type                string `json:"type"`                   // "HTML+XPath" or "XML+XPath"
	XPathItem           string `json:"xpath_item"`             // XPath to extract feed items
	XPathItemTitle      string `json:"xpath_item_title"`       // XPath to extract item title
	XPathItemContent    string `json:"xpath_item_content"`     // XPath to extract item content
	XPathItemUri        string `json:"xpath_item_uri"`         // XPath to extract item URI
	XPathItemAuthor     string `json:"xpath_item_author"`      // XPath to extract item author
	XPathItemTimestamp  string `json:"xpath_item_timestamp"`   // XPath to extract item timestamp
	XPathItemTimeFormat string `json:"xpath_item_time_format"` // Time format for parsing timestamp
	XPathItemThumbnail  string `json:"xpath_item_thumbnail"`   // XPath to extract item thumbnail
	XPathItemCategories string `json:"xpath_item_categories"`  // XPath to extract item categories
	XPathItemUid        string `json:"xpath_item_uid"`         // XPath to extract item unique ID
	ArticleViewMode     string `json:"article_view_mode"`      // Article view mode override ('global', 'webpage', 'rendered')
	AutoExpandContent   string `json:"auto_expand_content"`    // Auto expand content mode ('global', 'enabled', 'disabled')
	// Email/Newsletter support
	EmailAddress    string `json:"email_address,omitempty"`     // Email address for newsletter subscriptions
	EmailIMAPServer string `json:"email_imap_server,omitempty"` // IMAP server address
	EmailIMAPPort   int    `json:"email_imap_port"`             // IMAP server port (default 993)
	EmailUsername   string `json:"email_username,omitempty"`    // IMAP username
	EmailPassword   string `json:"email_password,omitempty"`    // IMAP password (encrypted)
	EmailFolder     string `json:"email_folder"`                // IMAP folder to monitor (default INBOX)
	EmailLastUID    int    `json:"email_last_uid"`              // Last processed email UID for incremental updates
	// FreshRSS integration
	IsFreshRSSSource bool   `json:"is_freshrss_source"` // Whether this feed is from FreshRSS sync
	FreshRSSStreamID string `json:"freshrss_stream_id"` // FreshRSS stream ID (e.g., "feed/http://...")
	// Statistics
	LatestArticleTime *time.Time `json:"latest_article_time,omitempty"` // Latest article publish time
	ArticlesPerMonth  float64    `json:"articles_per_month,omitempty"`  // Average articles per month (last 90 days / 3)
	LastUpdateStatus  string     `json:"last_update_status,omitempty"`  // Last update status ("success" or "failed")
	// Tags (populated by API handlers)
	Tags []Tag `json:"tags,omitempty"` // Tags assigned to this feed
}

type Article struct {
	ID                    int64     `json:"id"`
	FeedID                int64     `json:"feed_id"`
	Title                 string    `json:"title"`
	URL                   string    `json:"url"`
	ImageURL              string    `json:"image_url"`
	AudioURL              string    `json:"audio_url"`
	VideoURL              string    `json:"video_url"` // YouTube video URL for embedded player
	PublishedAt           time.Time `json:"published_at"`
	HasValidPublishedTime bool      `json:"-"` // Internal field, not serialized
	IsRead                bool      `json:"is_read"`
	IsFavorite            bool      `json:"is_favorite"`
	IsHidden              bool      `json:"is_hidden"`
	IsReadLater           bool      `json:"is_read_later"`
	FeedTitle             string    `json:"feed_title,omitempty"` // Joined field
	Author                string    `json:"author,omitempty"`     // Article author
	TranslatedTitle       string    `json:"translated_title"`
	Summary               string    `json:"summary"`          // Cached AI-generated summary
	UniqueID              string    `json:"unique_id"`        // Unique identifier for deduplication (title+feed_id+published_date)
	FreshRSSItemID        string    `json:"freshrss_item_id"` // FreshRSS/Google Reader item ID for API operations
}

// SavedFilter represents a user-saved article filter
type SavedFilter struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Conditions string    `json:"conditions"` // JSON string of FilterCondition[]
	Position   int       `json:"position"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Tag represents a user-defined tag for organizing feeds
type Tag struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Position int    `json:"position"`
}
