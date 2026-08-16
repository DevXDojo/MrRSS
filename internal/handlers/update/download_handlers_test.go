package update

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"MrRSS/internal/handlers/core"
)

func TestHandleDownloadUpdate_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/update/download", nil)
	rr := httptest.NewRecorder()

	HandleDownloadUpdate(&core.Handler{}, rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleDownloadUpdate_InvalidURLPrefix(t *testing.T) {
	body := bytes.NewReader([]byte(`{"download_url":"https://malicious.com/file","asset_name":"app.zip"}`))
	req := httptest.NewRequest(http.MethodPost, "/update/download", body)
	rr := httptest.NewRecorder()

	HandleDownloadUpdate(&core.Handler{}, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid url prefix, got %d", rr.Code)
	}
}

func TestHandleDownloadUpdate_InvalidAssetName(t *testing.T) {
	body := bytes.NewReader([]byte(`{"download_url":"https://github.com/DevXDojo/MrRSS/releases/download/v1/app.zip","asset_name":"../secret"}`))
	req := httptest.NewRequest(http.MethodPost, "/update/download", body)
	rr := httptest.NewRecorder()

	HandleDownloadUpdate(&core.Handler{}, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid asset name, got %d", rr.Code)
	}
}
