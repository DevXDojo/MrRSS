package freshrss

import (
	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProviderStatusAndSyncBoundaries(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	h := &core.Handler{DB: db}
	for _, provider := range []string{"freshrss", "miniflux"} {
		if err := db.SetSetting(provider+"_last_sync_time", map[string]string{"freshrss": "2026-09-12T01:00:00Z", "miniflux": "2026-09-12T02:00:00Z"}[provider]); err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		HandleSyncStatus(h, w, httptest.NewRequest(http.MethodGet, "/api/"+provider+"/status", nil))
		var body map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		want, _ := db.GetSetting(provider + "_last_sync_time")
		if body["last_sync_time"] != want {
			t.Fatal("status used other provider settings")
		}
	}
	if err := db.SetSetting("freshrss_enabled", "true"); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	HandleSync(h, w, httptest.NewRequest(http.MethodPost, "/api/miniflux/sync", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatal("disabled Miniflux was enabled by FreshRSS setting")
	}
}
