// Package source provides a unified interface for different feed data sources.
package source

import (
	"context"
	"time"

	"github.com/mmcdole/gofeed"
)

// Type represents the type of feed source.
type Type string

const (
	TypeRSS    Type = "rss"    // Standard RSS/Atom feed via HTTP
	TypeScript Type = "script" // Custom script that outputs RSS
	TypeXPath  Type = "xpath"  // HTML scraping with XPath selectors
	TypeEmail  Type = "email"  // Email/IMAP as feed source
)

// Source is the interface that all feed sources must implement.
type Source interface {
	// Type returns the source type identifier.
	Type() Type

	// Fetch retrieves the feed content from the source.
	// Returns the parsed feed or an error if fetching fails.
	Fetch(ctx context.Context, config *Config) (*gofeed.Feed, error)

	// Validate checks if the configuration is valid for this source.
	// Returns nil if valid, or an error describing the validation failure.
	Validate(config *Config) error
}

// Config holds the configuration for fetching a feed.
type Config struct {
	// Common fields
	URL        string        // Feed URL (for RSS, XPath sources)
	Timeout    time.Duration // Request timeout
	SourceType Type          // Explicit source type (optional, auto-detected if empty)

	// Script source fields
	ScriptPath string // Path to the script file (relative to scripts dir)

	// XPath source fields
	XPathItemSelector        string // CSS/XPath selector for items container
	XPathTitleSelector       string // CSS selector for feed title
	XPathDescSelector        string // CSS selector for feed description
	XPathItemTitleSelector   string // CSS selector for item title
	XPathItemLinkSelector    string // CSS selector for item link
	XPathItemContentSelector string // CSS selector for item content
	XPathItemDateSelector    string // CSS selector for item date

	// Email source fields
	EmailIMAPServer string // IMAP server address
	EmailIMAPPort   int    // IMAP server port (default: 993)
	EmailUsername   string // IMAP username
	EmailPassword   string // IMAP password
	EmailFolder     string // IMAP folder to fetch from (default: INBOX)
	EmailLastUID    int    // Last processed email UID
	EmailAddress    string // Newsletter sender filter

	// Network configuration
	ProxyURL  string            // HTTP proxy URL
	Headers   map[string]string // Custom HTTP headers
	UserAgent string            // Custom User-Agent string

	// Authentication
	BasicAuthUser     string // HTTP Basic Auth username
	BasicAuthPassword string // HTTP Basic Auth password
}

// Result contains the fetch result with metadata.
type Result struct {
	Feed      *gofeed.Feed  // The parsed feed
	FetchedAt time.Time     // When the feed was fetched
	Duration  time.Duration // How long the fetch took
	Source    Type          // Which source type was used
}
