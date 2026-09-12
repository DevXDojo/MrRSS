package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

type profileInvalidatingTranslator struct{ calls int }

func (t *profileInvalidatingTranslator) Translate(text, targetLang string) (string, error) {
	return text, nil
}
func (t *profileInvalidatingTranslator) InvalidateCache() { t.calls++ }

func TestProfileChangesInvalidateTranslationClient(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete", "default", "invalid update"} {
		t.Run(operation, func(t *testing.T) {
			db, err := database.NewDB(":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if err := db.Init(); err != nil {
				t.Fatal(err)
			}
			id, err := db.CreateAIProfile(&models.AIProfile{Name: "Existing", Endpoint: "http://localhost:11434/v1", Model: "local", IsDefault: true})
			if err != nil {
				t.Fatal(err)
			}
			h := core.NewHandler(db, nil, nil, nil)
			translator := &profileInvalidatingTranslator{}
			h.Translator = translator
			path := fmt.Sprintf("/api/ai/profiles/%d", id)
			method := http.MethodPut
			handler := HandleUpdateAIProfile
			body := `{"name":"Changed","endpoint":"http://localhost:11434/v1","model":"local"}`
			wantStatus, wantCalls := http.StatusOK, 1
			switch operation {
			case "create":
				method, path, handler, wantStatus = http.MethodPost, "/api/ai/profiles", HandleCreateAIProfile, http.StatusCreated
			case "delete":
				method, handler, wantStatus = http.MethodDelete, HandleDeleteAIProfile, http.StatusNoContent
			case "default":
				method, path, handler = http.MethodPost, path+"/default", HandleSetDefaultAIProfile
			case "invalid update":
				body, wantStatus, wantCalls = `{}`, http.StatusBadRequest, 0
			}
			w := httptest.NewRecorder()
			handler(h, w, httptest.NewRequest(method, path, strings.NewReader(body)))
			if w.Code != wantStatus || translator.calls != wantCalls {
				t.Fatalf("status/calls = %d/%d, want %d/%d", w.Code, translator.calls, wantStatus, wantCalls)
			}
		})
	}
}
