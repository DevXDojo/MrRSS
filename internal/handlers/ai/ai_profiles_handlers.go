package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/models"
)

// ProfileRequest represents the request body for creating/updating an AI profile
type ProfileRequest struct {
	Name          string `json:"name"`
	APIKey        string `json:"api_key"`
	Endpoint      string `json:"endpoint"`
	Model         string `json:"model"`
	CustomHeaders string `json:"custom_headers"`
	IsDefault     bool   `json:"is_default"`
}

// ProfileTestRequest represents the request body for testing a configuration without saving
type ProfileTestRequest struct {
	APIKey        string `json:"api_key"`
	Endpoint      string `json:"endpoint"`
	Model         string `json:"model"`
	CustomHeaders string `json:"custom_headers"`
}

// ProfileTestResult represents the result of testing an AI profile
type ProfileTestResult struct {
	ProfileID         int64  `json:"profile_id"`
	ProfileName       string `json:"profile_name"`
	ConfigValid       bool   `json:"config_valid"`
	ConnectionSuccess bool   `json:"connection_success"`
	ModelAvailable    bool   `json:"model_available"`
	ResponseTimeMs    int64  `json:"response_time_ms"`
	ErrorMessage      string `json:"error_message,omitempty"`
}

// HandleListAIProfiles handles GET /api/ai/profiles
// @Summary      List AI profiles
// @Description  Get all AI configuration profiles (without API keys for security)
// @Tags         ai-profiles
// @Produce      json
// @Success      200  {array}   models.AIProfile  "List of AI profiles"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/profiles [get]
func HandleListAIProfiles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	profiles, err := h.DB.GetAllAIProfilesWithoutKeys()
	if err != nil {
		log.Printf("Error listing AI profiles: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	if profiles == nil {
		profiles = []models.AIProfile{}
	}

	response.JSON(w, profiles)
}

// HandleGetAIProfile handles GET /api/ai/profiles/:id
// @Summary      Get AI profile
// @Description  Get a specific AI configuration profile by ID (includes API key)
// @Tags         ai-profiles
// @Produce      json
// @Param        id   path      int  true  "Profile ID"
// @Success      200  {object}  models.AIProfile  "AI profile"
// @Failure      400  {object}  map[string]string  "Invalid profile ID"
// @Failure      404  {object}  map[string]string  "Profile not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/profiles/{id} [get]
func HandleGetAIProfile(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/ai/profiles/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid profile ID"), http.StatusBadRequest)
		return
	}

	profile, err := h.DB.GetAIProfile(id)
	if err != nil {
		log.Printf("Error getting AI profile: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	if profile == nil {
		response.Error(w, fmt.Errorf("profile not found"), http.StatusNotFound)
		return
	}

	// Mask API key for security (only show last 4 chars)
	if len(profile.APIKey) > 4 {
		profile.APIKey = "****" + profile.APIKey[len(profile.APIKey)-4:]
	} else if profile.APIKey != "" {
		profile.APIKey = "****"
	}

	response.JSON(w, profile)
}

// HandleCreateAIProfile handles POST /api/ai/profiles
// @Summary      Create AI profile
// @Description  Create a new AI configuration profile
// @Tags         ai-profiles
// @Accept       json
// @Produce      json
// @Param        request  body      ProfileRequest  true  "Profile data"
// @Success      201  {object}  models.AIProfile  "Created profile"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/profiles [post]
func HandleCreateAIProfile(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req ProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		response.Error(w, fmt.Errorf("name is required"), http.StatusBadRequest)
		return
	}
	if req.Endpoint == "" {
		response.Error(w, fmt.Errorf("endpoint is required"), http.StatusBadRequest)
		return
	}
	if req.Model == "" {
		response.Error(w, fmt.Errorf("model is required"), http.StatusBadRequest)
		return
	}

	profile := &models.AIProfile{
		Name:          req.Name,
		APIKey:        req.APIKey,
		Endpoint:      req.Endpoint,
		Model:         req.Model,
		CustomHeaders: req.CustomHeaders,
		IsDefault:     req.IsDefault,
	}

	id, err := h.DB.CreateAIProfile(profile)
	if err != nil {
		log.Printf("Error creating AI profile: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	profile.ID = id
	profile.APIKey = "" // Don't return API key in response

	w.WriteHeader(http.StatusCreated)
	response.JSON(w, profile)
}

// HandleUpdateAIProfile handles PUT /api/ai/profiles/:id
// @Summary      Update AI profile
// @Description  Update an existing AI configuration profile
// @Tags         ai-profiles
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "Profile ID"
// @Param        request  body      ProfileRequest  true  "Profile data"
// @Success      200  {object}  models.AIProfile  "Updated profile"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      404  {object}  map[string]string  "Profile not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/profiles/{id} [put]
func HandleUpdateAIProfile(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/ai/profiles/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid profile ID"), http.StatusBadRequest)
		return
	}

	// Check if profile exists
	existing, err := h.DB.GetAIProfile(id)
	if err != nil {
		log.Printf("Error getting AI profile: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	if existing == nil {
		response.Error(w, fmt.Errorf("profile not found"), http.StatusNotFound)
		return
	}

	var req ProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		response.Error(w, fmt.Errorf("name is required"), http.StatusBadRequest)
		return
	}
	if req.Endpoint == "" {
		response.Error(w, fmt.Errorf("endpoint is required"), http.StatusBadRequest)
		return
	}
	if req.Model == "" {
		response.Error(w, fmt.Errorf("model is required"), http.StatusBadRequest)
		return
	}

	// If API key is masked or empty, keep the existing key
	apiKey := req.APIKey
	if apiKey == "" || strings.HasPrefix(apiKey, "****") {
		apiKey = existing.APIKey
	}

	profile := &models.AIProfile{
		ID:            id,
		Name:          req.Name,
		APIKey:        apiKey,
		Endpoint:      req.Endpoint,
		Model:         req.Model,
		CustomHeaders: req.CustomHeaders,
		IsDefault:     req.IsDefault,
	}

	if err := h.DB.UpdateAIProfile(profile); err != nil {
		log.Printf("Error updating AI profile: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	profile.APIKey = "" // Don't return API key in response
	response.JSON(w, profile)
}

// HandleDeleteAIProfile handles DELETE /api/ai/profiles/:id
// @Summary      Delete AI profile
// @Description  Delete an AI configuration profile
// @Tags         ai-profiles
// @Param        id   path      int  true  "Profile ID"
// @Success      204  "No content"
// @Failure      400  {object}  map[string]string  "Invalid profile ID"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/profiles/{id} [delete]
func HandleDeleteAIProfile(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/ai/profiles/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid profile ID"), http.StatusBadRequest)
		return
	}

	if err := h.DB.DeleteAIProfile(id); err != nil {
		log.Printf("Error deleting AI profile: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleSetDefaultAIProfile handles POST /api/ai/profiles/:id/default
// @Summary      Set default AI profile
// @Description  Set an AI profile as the default
// @Tags         ai-profiles
// @Param        id   path      int  true  "Profile ID"
// @Success      200  {object}  map[string]string  "Success message"
// @Failure      400  {object}  map[string]string  "Invalid profile ID"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/profiles/{id}/default [post]
func HandleSetDefaultAIProfile(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path - handle both formats
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/api/ai/profiles/")
	path = strings.TrimSuffix(path, "/default")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid profile ID"), http.StatusBadRequest)
		return
	}

	if err := h.DB.SetDefaultAIProfile(id); err != nil {
		log.Printf("Error setting default AI profile: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]string{"message": "default profile set"})
}

// HandleTestAIProfile handles POST /api/ai/profiles/:id/test
// @Summary      Test AI profile
// @Description  Test an AI profile's connection and configuration
// @Tags         ai-profiles
// @Produce      json
// @Param        id   path      int  true  "Profile ID"
// @Success      200  {object}  ProfileTestResult  "Test result"
// @Failure      400  {object}  map[string]string  "Invalid profile ID"
// @Failure      404  {object}  map[string]string  "Profile not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/profiles/{id}/test [post]
func HandleTestAIProfile(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/api/ai/profiles/")
	path = strings.TrimSuffix(path, "/test")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid profile ID"), http.StatusBadRequest)
		return
	}

	profile, err := h.DB.GetAIProfile(id)
	if err != nil {
		log.Printf("Error getting AI profile: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	if profile == nil {
		response.Error(w, fmt.Errorf("profile not found"), http.StatusNotFound)
		return
	}

	result := testAIProfileConnection(h, profile)
	response.JSON(w, result)
}

// HandleTestAllAIProfiles handles POST /api/ai/profiles/test-all
// @Summary      Test all AI profiles
// @Description  Test all AI profiles' connections and configurations
// @Tags         ai-profiles
// @Produce      json
// @Success      200  {array}   ProfileTestResult  "Test results for all profiles"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/profiles/test-all [post]
func HandleTestAllAIProfiles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	profiles, err := h.DB.GetAllAIProfiles()
	if err != nil {
		log.Printf("Error listing AI profiles: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	if len(profiles) == 0 {
		response.JSON(w, []ProfileTestResult{})
		return
	}

	// Test all profiles concurrently
	results := make([]ProfileTestResult, len(profiles))
	var wg sync.WaitGroup

	for i, profile := range profiles {
		wg.Add(1)
		go func(idx int, p models.AIProfile) {
			defer wg.Done()
			results[idx] = testAIProfileConnection(h, &p)
		}(i, profile)
	}

	wg.Wait()
	response.JSON(w, results)
}

// HandleTestAIProfileConfig handles POST /api/ai/profiles/test-config
// @Summary      Test AI configuration without saving
// @Description  Test an AI configuration without creating/saving a profile
// @Tags         ai-profiles
// @Accept       json
// @Produce      json
// @Param        request  body      ProfileTestRequest  true  "Configuration to test"
// @Success      200  {object}  ProfileTestResult  "Test result"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/profiles/test-config [post]
func HandleTestAIProfileConfig(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req ProfileTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Create a temporary profile for testing
	tempProfile := &models.AIProfile{
		ID:            0,
		Name:          "Test",
		APIKey:        req.APIKey,
		Endpoint:      req.Endpoint,
		Model:         req.Model,
		CustomHeaders: req.CustomHeaders,
	}

	result := testAIProfileConnection(h, tempProfile)
	response.JSON(w, result)
}

// testAIProfileConnection tests a single AI profile
func testAIProfileConnection(h *core.Handler, profile *models.AIProfile) ProfileTestResult {
	result := ProfileTestResult{
		ProfileID:   profile.ID,
		ProfileName: profile.Name,
	}

	startTime := time.Now()

	// Validate configuration
	result.ConfigValid = true
	validationErrors := []string{}

	if profile.Endpoint == "" {
		validationErrors = append(validationErrors, "endpoint is required")
		result.ConfigValid = false
	}

	if profile.Model == "" {
		validationErrors = append(validationErrors, "model is required")
		result.ConfigValid = false
	}

	if !result.ConfigValid {
		result.ErrorMessage = "Configuration incomplete: " + strings.Join(validationErrors, ", ")
		result.ResponseTimeMs = time.Since(startTime).Milliseconds()
		return result
	}

	// Validate endpoint URL format
	parsedURL, err := url.Parse(profile.Endpoint)
	if err != nil {
		result.ConfigValid = false
		result.ErrorMessage = "Invalid endpoint URL: " + err.Error()
		result.ResponseTimeMs = time.Since(startTime).Milliseconds()
		return result
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.ConfigValid = false
		result.ErrorMessage = "API endpoint must use HTTP or HTTPS"
		result.ResponseTimeMs = time.Since(startTime).Milliseconds()
		return result
	}

	// Create HTTP client with proxy support if configured
	httpClient, err := createHTTPClientWithProxyForProfile(h)
	if err != nil {
		result.ConnectionSuccess = false
		result.ModelAvailable = false
		result.ErrorMessage = fmt.Sprintf("Failed to create HTTP client: %v", err)
		result.ResponseTimeMs = time.Since(startTime).Milliseconds()
		return result
	}
	httpClient.Timeout = 30 * time.Second

	// Create AI client for testing
	clientConfig := ai.ClientConfig{
		APIKey:        profile.APIKey,
		Endpoint:      profile.Endpoint,
		Model:         profile.Model,
		Timeout:       30 * time.Second,
		CustomHeaders: profile.CustomHeaders, // Keep as string, client will parse it
	}

	client := ai.NewClientWithHTTPClient(clientConfig, httpClient)

	// Try a simple test request
	_, err = client.Request("", "test")

	if err != nil {
		result.ConnectionSuccess = false
		result.ModelAvailable = false
		result.ErrorMessage = fmt.Sprintf("Connection failed: %v", err)
	} else {
		result.ConnectionSuccess = true
		result.ModelAvailable = true
	}

	result.ResponseTimeMs = time.Since(startTime).Milliseconds()
	return result
}

// createHTTPClientWithProxyForProfile creates an HTTP client with global proxy settings
func createHTTPClientWithProxyForProfile(h *core.Handler) (*http.Client, error) {
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	if proxyEnabled != "true" {
		return &http.Client{}, nil
	}

	proxyType, _ := h.DB.GetSetting("proxy_type")
	proxyHost, _ := h.DB.GetSetting("proxy_host")
	proxyPort, _ := h.DB.GetSetting("proxy_port")
	proxyUsername, _ := h.DB.GetEncryptedSetting("proxy_username")
	proxyPassword, _ := h.DB.GetEncryptedSetting("proxy_password")

	proxyURL := buildProxyURLForProfile(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
	if proxyURL == "" {
		return &http.Client{}, nil
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(u),
		},
	}, nil
}

// buildProxyURLForProfile builds a proxy URL from components
func buildProxyURLForProfile(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword string) string {
	if proxyHost == "" || proxyPort == "" {
		return ""
	}

	scheme := "http"
	switch proxyType {
	case "socks5":
		scheme = "socks5"
	case "https":
		scheme = "http" // HTTPS proxies use HTTP CONNECT
	}

	if proxyUsername != "" && proxyPassword != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%s", scheme, proxyUsername, proxyPassword, proxyHost, proxyPort)
	}
	return fmt.Sprintf("%s://%s:%s", scheme, proxyHost, proxyPort)
}
