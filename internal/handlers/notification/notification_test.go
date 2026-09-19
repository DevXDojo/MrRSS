package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	settingshandlers "MrRSS/internal/handlers/settings"
	push "MrRSS/internal/notification"
)

func handler(t *testing.T) *core.Handler {
	t.Helper()
	db, err := database.NewDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return core.NewHandler(db, nil, nil, nil)
}

func TestConfigAPIAndGenericAutosaveIsolation(t *testing.T) {
	h := handler(t)
	c := push.DefaultConfig()
	c.Channels = []push.Channel{{ID: "telegram", Name: "Personal", Provider: "telegram", Enabled: true, Token: "123:secret-token", ChatID: "1"}}
	body, _ := json.Marshal(c)
	w := httptest.NewRecorder()
	HandleConfig(h, w, httptest.NewRequest(http.MethodPut, "/api/notifications/config", bytes.NewReader(body)))
	if w.Code != 200 || strings.Contains(w.Body.String(), "secret-token") {
		t.Fatal(w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	HandleConfig(h, w, httptest.NewRequest(http.MethodPut, "/api/notifications/config", bytes.NewReader(body)))
	if w.Code != 409 {
		t.Fatal("missing revision conflict", w.Code)
	}
	w = httptest.NewRecorder()
	settingshandlers.HandleSettings(h, w, httptest.NewRequest(http.MethodPost, "/api/settings", strings.NewReader(`{"notification_config":"","theme":"dark"}`)))
	if w.Code != 200 || strings.Contains(w.Body.String(), "notification_config") {
		t.Fatal(w.Code, w.Body.String())
	}
	loaded, err := h.Notifications.Load(context.Background())
	if err != nil || len(loaded.Channels) != 1 || loaded.Channels[0].Token != "123:secret-token" {
		t.Fatal("generic autosave overwrote configuration", err)
	}
	w = httptest.NewRecorder()
	HandleConfig(h, w, httptest.NewRequest(http.MethodGet, "/api/notifications/config", nil))
	if w.Header().Get("Cache-Control") != "no-store" || strings.Contains(w.Body.String(), "secret-token") {
		t.Fatal("secret response cacheable or unredacted")
	}
}

func TestRejectBadMethodsAndBodies(t *testing.T) {
	h := handler(t)
	for _, fn := range []func(*core.Handler, http.ResponseWriter, *http.Request){HandleTest, HandlePreview, HandleHistory} {
		w := httptest.NewRecorder()
		fn(h, w, httptest.NewRequest(http.MethodDelete, "/", nil))
		if w.Code != 405 {
			t.Fatal(w.Code)
		}
	}
	for _, body := range []string{`{}`, `{"enabled":true,"unexpected":1}`, `{} {}`, strings.Repeat("x", 129*1024)} {
		w := httptest.NewRecorder()
		HandleConfig(h, w, httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
	// Tests reject invalid destinations before making network requests.
	w := httptest.NewRecorder()
	HandleTest(h, w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"id":"bad","name":"bad","provider":"discord","webhook":"http://127.0.0.1/private"}`)))
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
