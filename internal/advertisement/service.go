package advertisement

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const SourceURL = "https://raw.githubusercontent.com/DevXDojo/MrRSS/main/ads.json"

type Service struct {
	fallback  []byte
	sourceURL string
	gate      chan struct{}
	loaded    bool
	ads       []Ad
	nextCheck time.Time
}

func New(fallback []byte) *Service {
	return &Service{fallback: fallback, sourceURL: SourceURL, gate: make(chan struct{}, 1)}
}

// Get checks GitHub at most hourly. Valid remote JSON is cached intact,
// including fields this version does not understand. No reader data is sent.
func (s *Service) Get(ctx context.Context, path string, client *http.Client) []Ad {
	select {
	case s.gate <- struct{}{}:
		defer func() { <-s.gate }()
	case <-ctx.Done():
		return []Ad{}
	}
	if !s.loaded {
		s.ads, _ = Parse(s.fallback)
		if data, err := readBoundedFile(path); err == nil {
			if cached, err := Parse(data); err == nil {
				s.ads = cached
			}
		}
		s.loaded = true
	}
	if client != nil && !time.Now().Before(s.nextCheck) {
		if data, err := s.fetch(ctx, client); err == nil {
			if ads, err := Parse(data); err == nil {
				s.ads = ads
				if err := writeCache(path, data); err != nil {
					log.Printf("Could not persist advertisement cache: %v", err)
				}
			}
		}
		if ctx.Err() == nil {
			s.nextCheck = time.Now().Add(time.Hour)
		}
	}
	if s.ads == nil {
		return []Ad{}
	}
	return s.ads
}

func (s *Service) fetch(ctx context.Context, client *http.Client) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.sourceURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, os.ErrNotExist
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxDocumentSize+1))
}

func readBoundedFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, maxDocumentSize+1))
}

func writeCache(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".ads-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
