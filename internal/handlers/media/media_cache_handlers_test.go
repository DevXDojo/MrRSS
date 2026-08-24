package media

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/utils/fileutil"
)

func TestHandleMediaCacheInfoAndCleanup(t *testing.T) {
	tmp := t.TempDir()
	// Ensure data dir resolves to temp dir
	t.Setenv("APPDATA", tmp)
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_DATA_HOME", tmp)

	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init failed: %v", err)
	}

	h := core.NewHandler(db, nil, nil, nil)
	// Enable media cache and set small thresholds
	if err := h.DB.SetSetting("media_cache_enabled", "true"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}
	if err := h.DB.SetSetting("media_cache_max_age_days", "1"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}
	if err := h.DB.SetSetting("media_cache_max_size_mb", "1"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	// Get the cache directory
	cacheDir, err := fileutil.GetMediaCacheDir()
	if err != nil {
		t.Fatalf("GetMediaCacheDir failed: %v", err)
	}

	// Create some test cache files
	testData := []byte("test image data")
	file1 := filepath.Join(cacheDir, "abc123.jpg")
	file2 := filepath.Join(cacheDir, "def456.png")
	if err := os.WriteFile(file1, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Call info (GET) - should show cache size > 0
	req := httptest.NewRequest(http.MethodGet, "/media/info", nil)
	rr := httptest.NewRecorder()
	HandleMediaCacheInfo(h, rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for info, got %d", rr.Code)
	}
	var info map[string]float64
	if err := json.NewDecoder(rr.Body).Decode(&info); err != nil {
		t.Fatalf("decode info failed: %v", err)
	}

	cacheSizeMB, ok := info["cache_size_mb"]
	if !ok {
		t.Fatalf("info missing cache_size_mb")
	}
	if cacheSizeMB <= 0 {
		t.Errorf("Expected cache size > 0, got %f", cacheSizeMB)
	}

	// Test 1: Cleanup without ?all=true parameter (automatic cleanup)
	// Should respect max_age_days setting and not clean new files
	req2 := httptest.NewRequest(http.MethodPost, "/media/cleanup", nil)
	rr2 := httptest.NewRecorder()
	HandleMediaCacheCleanup(h, rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 for cleanup, got %d", rr2.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr2.Body).Decode(&resp); err != nil {
		t.Fatalf("decode cleanup failed: %v", err)
	}
	if success, ok := resp["success"].(bool); !ok || !success {
		t.Fatalf("expected cleanup success true, got %v", resp)
	}
	filesCleaned := int(resp["files_cleaned"].(float64))
	if filesCleaned != 0 {
		t.Errorf("Expected 0 files cleaned (too new), got %d", filesCleaned)
	}

	// Test 2: Cleanup with ?all=true parameter (manual cleanup)
	// Should clean all files regardless of age
	req3 := httptest.NewRequest(http.MethodPost, "/media/cleanup?all=true", nil)
	rr3 := httptest.NewRecorder()
	HandleMediaCacheCleanup(h, rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("expected 200 for cleanup with all=true, got %d", rr3.Code)
	}
	var resp2 map[string]interface{}
	if err := json.NewDecoder(rr3.Body).Decode(&resp2); err != nil {
		t.Fatalf("decode cleanup failed: %v", err)
	}
	filesCleaned2 := int(resp2["files_cleaned"].(float64))
	if filesCleaned2 == 0 {
		t.Error("Expected files to be cleaned with ?all=true")
	}
	if filesCleaned2 != 2 {
		t.Errorf("Expected 2 files cleaned with ?all=true, got %d", filesCleaned2)
	}

	// Test 3: Verify cache is now empty
	req4 := httptest.NewRequest(http.MethodGet, "/media/info", nil)
	rr4 := httptest.NewRecorder()
	HandleMediaCacheInfo(h, rr4, req4)
	if rr4.Code != http.StatusOK {
		t.Fatalf("expected 200 for info, got %d", rr4.Code)
	}
	var info2 map[string]float64
	if err := json.NewDecoder(rr4.Body).Decode(&info2); err != nil {
		t.Fatalf("decode info failed: %v", err)
	}
	if info2["cache_size_mb"] != 0 {
		t.Errorf("Expected cache size 0 after cleanup, got %f", info2["cache_size_mb"])
	}
}
