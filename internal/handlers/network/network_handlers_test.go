package network

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MrRSS/internal/database"
	corepkg "MrRSS/internal/handlers/core"
)

func setupHandler(t *testing.T) *corepkg.Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	return corepkg.NewHandler(db, nil, nil, nil)
}

func TestHandleDetectNetwork_MethodNotAllowed(t *testing.T) {
	h := setupHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/network/detect", nil)
	rr := httptest.NewRecorder()

	HandleDetectNetwork(h, rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleGetNetworkInfo_Defaults(t *testing.T) {
	h := setupHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/network/info", nil)
	rr := httptest.NewRecorder()

	HandleGetNetworkInfo(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if _, ok := resp["bandwidth_mbps"]; !ok {
		t.Fatalf("missing bandwidth_mbps in response")
	}
}
