package advertisement

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
)

func TestSnoozePersistsForOneWeek(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	h := core.NewHandler(db, nil, nil, nil)
	before := time.Now().Add(7 * 24 * time.Hour).Unix()
	recorder := httptest.NewRecorder()
	HandleSnooze(h, recorder, httptest.NewRequest(http.MethodPost, "/api/ads/snooze", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("save failed: %s", recorder.Body.String())
	}
	stored, err := db.GetSetting("ads_snoozed_until")
	until, _ := strconv.ParseInt(stored, 10, 64)
	if err != nil || until < before || until > time.Now().Add(7*24*time.Hour).Unix() {
		t.Fatalf("invalid persisted snooze: %s, %v", stored, err)
	}
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if again, _ := db.GetSetting("ads_snoozed_until"); again != stored {
		t.Fatal("initialization reset saved preference")
	}
	for _, handler := range []func(*core.Handler, http.ResponseWriter, *http.Request){HandleAds, HandleSnooze} {
		recorder := httptest.NewRecorder()
		handler(h, recorder, httptest.NewRequest(http.MethodDelete, "/api/ads", nil))
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Fatal("unsupported method accepted")
		}
	}
}
