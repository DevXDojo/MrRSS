package settings

import (
	"MrRSS/internal/adfilter"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdFilterConfigRoundTripAndConflict(t *testing.T) {
	h := setupHandlerWithDB(t)
	defer h.DB.Close()
	get := httptest.NewRecorder()
	HandleAdFilterConfig(h, get, httptest.NewRequest("GET", "/", nil))
	var config adfilter.Config
	if get.Code != 200 || json.Unmarshal(get.Body.Bytes(), &config) != nil || !config.Enabled || config.AIEnabled {
		t.Fatalf("default: %s", get.Body)
	}
	config.Allowlist = "example.org"
	config.AIEnabled = true
	config.AIMode = "review"
	raw, _ := json.Marshal(config)
	put := httptest.NewRecorder()
	HandleAdFilterConfig(h, put, httptest.NewRequest("PUT", "/", bytes.NewReader(raw)))
	if put.Code != 200 {
		t.Fatal(put.Body)
	}
	stale := httptest.NewRecorder()
	HandleAdFilterConfig(h, stale, httptest.NewRequest("PUT", "/", bytes.NewReader(raw)))
	if stale.Code != 409 {
		t.Fatalf("stale update accepted: %d", stale.Code)
	}
	generic := httptest.NewRecorder()
	HandleSettings(h, generic, httptest.NewRequest("POST", "/", strings.NewReader(`{"ad_filter_config":"{}","theme":"dark"}`)))
	get = httptest.NewRecorder()
	HandleAdFilterConfig(h, get, httptest.NewRequest("GET", "/", nil))
	if json.Unmarshal(get.Body.Bytes(), &config) != nil || config.Revision != 1 || config.Allowlist != "example.org" || !config.AIEnabled {
		t.Fatalf("overwritten by generic settings: %s", get.Body)
	}
}

func TestAdFilterValidationAndSafePreview(t *testing.T) {
	h := setupHandlerWithDB(t)
	defer h.DB.Close()
	for _, body := range []string{`{"enabled":true,"rules":"*##body"}`, `{"enabled":true,"unknown":1}`, `{} {}`, `{"enabled":true,"ai_mode":"bad"}`} {
		w := httptest.NewRecorder()
		HandleAdFilterConfig(h, w, httptest.NewRequest("PUT", "/", strings.NewReader(body)))
		if w.Code != 400 {
			t.Errorf("accepted %s: %d", body, w.Code)
		}
	}
	input := AdFilterPreviewRequest{URL: "https://example.org/story", Options: adfilter.Options{Enabled: true, Rules: "example.org##.promotion"}, HTML: `<p>Editorial remains.</p><div class="promotion">Sale</div><img src="photo.jpg" onerror="alert(1)"><script>alert(1)</script>`}
	raw, _ := json.Marshal(input)
	w := httptest.NewRecorder()
	HandleAdFilterPreview(h, w, httptest.NewRequest("POST", "/", bytes.NewReader(raw)))
	var result adfilter.Content
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil {
		t.Fatal(w.Body)
	}
	if strings.Contains(result.Content, "Sale") || !strings.Contains(result.OriginalContent, "Sale") || strings.Contains(result.OriginalContent, "onerror") || strings.Contains(result.OriginalContent, "<script") || !strings.Contains(result.Content, "Editorial remains") {
		t.Fatalf("unsafe or irreversible: %+v", result)
	}
}
