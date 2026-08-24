package routes

import (
	"errors"
	"net/http"
	"strings"

	aihandlers "MrRSS/internal/handlers/ai"
	chat "MrRSS/internal/handlers/chat"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// registerAIRoutes registers all AI-related routes
func registerAIRoutes(mux *http.ServeMux, h *core.Handler) {
	// AI Chat
	mux.HandleFunc("/api/ai-chat", func(w http.ResponseWriter, r *http.Request) { chat.HandleAIChat(h, w, r) })
	mux.HandleFunc("/api/ai/chat/sessions/delete-all", func(w http.ResponseWriter, r *http.Request) { chat.HandleDeleteAllSessions(h, w, r) })
	mux.HandleFunc("/api/ai/chat/sessions", func(w http.ResponseWriter, r *http.Request) { chat.HandleListSessions(h, w, r) })
	mux.HandleFunc("/api/ai/chat/session/create", func(w http.ResponseWriter, r *http.Request) { chat.HandleCreateSession(h, w, r) })
	mux.HandleFunc("/api/ai/chat/session", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			chat.HandleGetSession(h, w, r)
		case http.MethodPut, http.MethodPatch:
			chat.HandleUpdateSession(h, w, r)
		case http.MethodDelete:
			chat.HandleDeleteSession(h, w, r)
		default:
			response.Error(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/ai/chat/messages", func(w http.ResponseWriter, r *http.Request) { chat.HandleListMessages(h, w, r) })
	mux.HandleFunc("/api/ai/chat/message/delete", func(w http.ResponseWriter, r *http.Request) { chat.HandleDeleteMessage(h, w, r) })

	// AI testing and search
	mux.HandleFunc("/api/ai/test", func(w http.ResponseWriter, r *http.Request) { aihandlers.HandleTestAIConfig(h, w, r) })
	mux.HandleFunc("/api/ai/test/info", func(w http.ResponseWriter, r *http.Request) { aihandlers.HandleGetAITestInfo(h, w, r) })
	mux.HandleFunc("/api/ai/search", func(w http.ResponseWriter, r *http.Request) { aihandlers.HandleAISearch(h, w, r) })

	// AI Profiles
	mux.HandleFunc("/api/ai/profiles/test-all", func(w http.ResponseWriter, r *http.Request) { aihandlers.HandleTestAllAIProfiles(h, w, r) })
	mux.HandleFunc("/api/ai/profiles/test-config", func(w http.ResponseWriter, r *http.Request) { aihandlers.HandleTestAIProfileConfig(h, w, r) })
	mux.HandleFunc("/api/ai/profiles", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			aihandlers.HandleListAIProfiles(h, w, r)
		case http.MethodPost:
			aihandlers.HandleCreateAIProfile(h, w, r)
		default:
			response.Error(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/ai/profiles/", func(w http.ResponseWriter, r *http.Request) {
		// Handle routes like /api/ai/profiles/:id, /api/ai/profiles/:id/test, /api/ai/profiles/:id/default
		path := r.URL.Path
		if strings.HasSuffix(path, "/test") {
			aihandlers.HandleTestAIProfile(h, w, r)
		} else if strings.HasSuffix(path, "/default") {
			aihandlers.HandleSetDefaultAIProfile(h, w, r)
		} else {
			// Direct profile access by ID
			switch r.Method {
			case http.MethodGet:
				aihandlers.HandleGetAIProfile(h, w, r)
			case http.MethodPut:
				aihandlers.HandleUpdateAIProfile(h, w, r)
			case http.MethodDelete:
				aihandlers.HandleDeleteAIProfile(h, w, r)
			default:
				response.Error(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
			}
		}
	})
}
