package script

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"MrRSS/internal/database"
	corepkg "MrRSS/internal/handlers/core"
	"MrRSS/internal/utils/fileutil"
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

func TestHandleGetScriptsDir_GET(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/scripts/dir", nil)
	rr := httptest.NewRecorder()

	HandleGetScriptsDir(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if _, ok := resp["scripts_dir"]; !ok {
		t.Fatalf("missing scripts_dir in response")
	}
}

func TestHandleListScripts_IncludesCreatedScript(t *testing.T) {
	h := setupHandler(t)

	// Ensure scripts dir exists
	scriptsDir, err := fileutil.GetScriptsDir()
	if err != nil {
		t.Fatalf("GetScriptsDir failed: %v", err)
	}

	// Create a fake script file
	testScript := filepath.Join(scriptsDir, "test_script.py")
	if err := os.WriteFile(testScript, []byte("print(1)"), fs.FileMode(0644)); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	defer os.Remove(testScript)

	req := httptest.NewRequest(http.MethodGet, "/scripts/list", nil)
	rr := httptest.NewRecorder()

	HandleListScripts(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	scripts, ok := resp["scripts"].([]interface{})
	if !ok {
		t.Fatalf("scripts field missing or wrong type")
	}

	found := false
	for _, s := range scripts {
		m, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		if m["name"] == "test_script.py" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("test_script.py not listed in scripts")
	}
}
