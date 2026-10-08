// Package advertisement reads extensible, data-only partner promotions.
package advertisement

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const maxDocumentSize = 1 << 20

// Text supports both plain strings and locale maps. Unknown/non-string locale
// values are ignored so future metadata cannot break older clients.
type Text map[string]string

func (t *Text) UnmarshalJSON(data []byte) error {
	var plain string
	if json.Unmarshal(data, &plain) == nil {
		if len(plain) > 8000 {
			return fmt.Errorf("advertisement text exceeds size limit")
		}
		*t = Text{"en": plain}
		return nil
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	*t = Text{}
	for key, value := range values {
		var text string
		if len(key) <= 32 && json.Unmarshal(value, &text) == nil && len(text) <= 8000 {
			(*t)[key] = text
		}
	}
	return nil
}

type Block struct {
	Type  string `json:"type"`
	Text  Text   `json:"text,omitempty"`
	Items []Text `json:"items,omitempty"`
}

type Ad struct {
	ID          string  `json:"id"`
	Title       Text    `json:"title"`
	Description Text    `json:"description"`
	URL         string  `json:"url"`
	CouponCode  string  `json:"coupon_code,omitempty"`
	Blocks      []Block `json:"blocks,omitempty"`
}

// Parse projects supported fields without rejecting unknown versions, fields,
// layouts or blocks. Invalid entries do not suppress other advertisements.
func Parse(data []byte) ([]Ad, error) {
	if len(data) > maxDocumentSize {
		return nil, fmt.Errorf("advertisement document exceeds size limit")
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	var entries []json.RawMessage
	if raw, ok := document["ads"]; ok {
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("advertisement document missing ads")
	}
	ads := make([]Ad, 0)
	seen := make(map[string]bool)
	for _, entry := range entries {
		if len(ads) == 32 {
			break
		}
		ad, ok := parseAd(entry)
		if ok && !seen[ad.ID] {
			ads = append(ads, ad)
			seen[ad.ID] = true
		}
	}
	return ads, nil
}

func parseAd(raw json.RawMessage) (Ad, bool) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return Ad{}, false
	}
	// Future rich layouts may supply a simple fallback for older versions.
	var fallback map[string]json.RawMessage
	_ = json.Unmarshal(fields["fallback"], &fallback)
	read := func(key string, target interface{}) {
		if value, ok := fields[key]; ok && string(value) != "null" && json.Unmarshal(value, target) == nil {
			return
		}
		_ = json.Unmarshal(fallback[key], target)
	}
	var ad Ad
	read("id", &ad.ID)
	read("title", &ad.Title)
	read("description", &ad.Description)
	read("url", &ad.URL)
	read("coupon_code", &ad.CouponCode)
	if !hasText(ad.Title) {
		_ = json.Unmarshal(fallback["title"], &ad.Title)
	}
	if !hasText(ad.Description) {
		_ = json.Unmarshal(fallback["description"], &ad.Description)
	}
	u, err := url.Parse(ad.URL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || len(ad.URL) > 4096 || strings.ContainsAny(ad.URL, "\r\n\t") {
		return Ad{}, false
	}
	if strings.TrimSpace(ad.ID) == "" || len(ad.ID) > 128 || len(ad.CouponCode) > 128 || !hasText(ad.Title) {
		return Ad{}, false
	}
	var blocks []json.RawMessage
	_ = json.Unmarshal(fields["blocks"], &blocks)
	for _, rawBlock := range blocks {
		if len(ad.Blocks) == 12 {
			break
		}
		var block Block
		if json.Unmarshal(rawBlock, &block) != nil {
			continue
		}
		switch block.Type {
		case "text", "callout":
			if hasText(block.Text) {
				ad.Blocks = append(ad.Blocks, block)
			}
		case "list":
			if len(block.Items) > 8 {
				block.Items = block.Items[:8]
			}
			ad.Blocks = append(ad.Blocks, block)
		}
	}
	return ad, true
}

func hasText(text Text) bool {
	for _, value := range text {
		if strings.TrimSpace(value) != "" && len(value) <= 8000 {
			return true
		}
	}
	return false
}
