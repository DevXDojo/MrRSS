package source

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/mmcdole/gofeed"
)

// Manager manages different feed sources and provides a unified interface.
type Manager struct {
	rss    *RSSSource
	script *ScriptSource
	xpath  *XPathSource
	email  *EmailSource

	mu sync.RWMutex
}

// NewManager creates a new source manager.
func NewManager(scriptsDir string) *Manager {
	return &Manager{
		rss:    NewRSSSource(),
		script: NewScriptSource(scriptsDir),
		xpath:  NewXPathSource(),
		email:  NewEmailSource(),
	}
}

// GetSource returns the appropriate source for the given type.
func (m *Manager) GetSource(sourceType Type) (Source, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch sourceType {
	case TypeRSS:
		return m.rss, nil
	case TypeScript:
		return m.script, nil
	case TypeXPath:
		return m.xpath, nil
	case TypeEmail:
		return m.email, nil
	default:
		return nil, fmt.Errorf("unknown source type: %s", sourceType)
	}
}

// Fetch fetches content using the appropriate source based on configuration.
func (m *Manager) Fetch(ctx context.Context, config *Config) (*gofeed.Feed, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	sourceType := m.detectSourceType(config)
	source, err := m.GetSource(sourceType)
	if err != nil {
		return nil, err
	}

	return source.Fetch(ctx, config)
}

// detectSourceType determines the source type from configuration.
func (m *Manager) detectSourceType(config *Config) Type {
	// Explicit type takes precedence
	if config.SourceType != "" {
		return config.SourceType
	}

	// Auto-detect based on configuration fields
	if config.ScriptPath != "" {
		return TypeScript
	}
	if config.EmailIMAPServer != "" {
		return TypeEmail
	}
	if config.XPathItemSelector != "" {
		return TypeXPath
	}

	// Default to RSS
	return TypeRSS
}

// SetHTTPClient sets the HTTP client for RSS and XPath sources.
func (m *Manager) SetHTTPClient(client *http.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rss.SetHTTPClient(client)
	m.xpath.SetHTTPClient(client)
}

// Validate validates the configuration for the appropriate source.
func (m *Manager) Validate(config *Config) error {
	if config == nil {
		return errors.New("config is nil")
	}

	sourceType := m.detectSourceType(config)
	source, err := m.GetSource(sourceType)
	if err != nil {
		return err
	}

	return source.Validate(config)
}

// ConfigFromFeedURL creates a simple RSS config from a URL.
func ConfigFromFeedURL(url string) *Config {
	return &Config{
		URL:        url,
		SourceType: TypeRSS,
	}
}

// ConfigFromScript creates a script config.
func ConfigFromScript(scriptPath string) *Config {
	return &Config{
		ScriptPath: scriptPath,
		SourceType: TypeScript,
	}
}

// ConfigFromXPath creates an XPath config.
func ConfigFromXPath(url, itemSelector string) *Config {
	return &Config{
		URL:               url,
		XPathItemSelector: itemSelector,
		SourceType:        TypeXPath,
	}
}

// ConfigFromEmail creates an email config.
func ConfigFromEmail(server string, port int, username, password, folder string) *Config {
	return &Config{
		EmailIMAPServer: server,
		EmailIMAPPort:   port,
		EmailUsername:   username,
		EmailPassword:   password,
		EmailFolder:     folder,
		SourceType:      TypeEmail,
	}
}
