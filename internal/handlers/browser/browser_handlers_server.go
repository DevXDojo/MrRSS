//go:build server

// Package browser provides HTTP handlers for browser-related operations (server mode).
package browser

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	handlers "MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// HandleOpenURL handles URL opening requests in server mode.
// In server mode, this returns a redirect response that instructs the client to open the URL.
// @Summary      Open URL in browser (server mode)
// @Description  Returns redirect response for client-side URL opening
// @Tags         browser
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Open URL request (url)"
// @Success      200  {object}  map[string]string  "Redirect URL (redirect)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid URL)"
// @Router       /browser/open [post]
func HandleOpenURL(h *handlers.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Validate URL
	if req.URL == "" {
		response.Error(w, fmt.Errorf("URL is required"), http.StatusBadRequest)
		return
	}

	// Parse and validate URL scheme
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		log.Printf("Invalid URL format: %v", err)
		response.Error(w, fmt.Errorf("invalid URL format: %w", err), http.StatusBadRequest)
		return
	}

	// Only allow http and https schemes for security
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		log.Printf("Invalid URL scheme: %s", parsedURL.Scheme)
		response.Error(w, fmt.Errorf("only HTTP and HTTPS URLs are allowed"), http.StatusBadRequest)
		return
	}

	// Server mode: return redirect response for client-side handling
	log.Printf("Server mode detected, instructing client to open URL: %s", req.URL)
	w.WriteHeader(http.StatusOK)
	response.JSON(w, map[string]string{"redirect": req.URL})
}
