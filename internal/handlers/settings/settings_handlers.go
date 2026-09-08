package settings

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// safeGetEncryptedSetting safely retrieves an encrypted setting, returning empty string on error.
// This prevents JSON encoding errors when encrypted data is corrupted or cannot be decrypted.
func safeGetEncryptedSetting(h *core.Handler, key string) string {
	value, err := h.DB.GetEncryptedSetting(key)
	if err != nil {
		log.Printf("Warning: Failed to decrypt setting %s: %v. Returning empty string.", key, err)
		return ""
	}
	return sanitizeValue(value)
}

// safeGetSetting safely retrieves a setting, returning empty string on error.
func safeGetSetting(h *core.Handler, key string) string {
	value, err := h.DB.GetSetting(key)
	if err != nil {
		log.Printf("Warning: Failed to retrieve setting %s: %v. Returning empty string.", key, err)
		return ""
	}
	return sanitizeValue(value)
}

// sanitizeValue removes control characters that could break JSON encoding.
func sanitizeValue(value string) string {
	// Remove control characters that could break JSON
	return strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1 // Remove control characters except tab, newline, carriage return
		}
		return r
	}, value)
}

// HandleSettings handles GET and POST requests for application settings.
// Uses the definition-driven approach from settings_base.go for cleaner code.
func HandleSettings(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Get all settings using the definition-driven approach
		settings := GetAllSettings(h)
		response.JSON(w, settings)

	case http.MethodPost:
		// Parse request body as a generic map
		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.Error(w, err, http.StatusBadRequest)
			return
		}

		wasFreshRSSEnabled := false
		if _, ok := req["freshrss_enabled"]; ok {
			currentValue, err := h.DB.GetSetting("freshrss_enabled")
			if err == nil {
				wasFreshRSSEnabled = currentValue == "true"
			}
		}

		if rawStartupOnBoot, ok := req["startup_on_boot"]; ok {
			startupOnBoot, err := strconv.ParseBool(strings.TrimSpace(rawStartupOnBoot))
			if err != nil {
				response.Error(w, fmt.Errorf("invalid startup_on_boot value: %w", err), http.StatusBadRequest)
				return
			}
			req["startup_on_boot"] = strconv.FormatBool(startupOnBoot)

			currentValue, _ := h.DB.GetSetting("startup_on_boot")
			if h.SetStartupOnBoot != nil && currentValue != req["startup_on_boot"] {
				if err := h.SetStartupOnBoot(startupOnBoot); err != nil {
					log.Printf("Failed to update system startup integration: %v", err)
					response.Error(w, fmt.Errorf("update system startup integration: %w", err), http.StatusInternalServerError)
					return
				}
			}
		}

		// Save settings using the definition-driven approach
		if err := SaveSettings(h, req); err != nil {
			log.Printf("Failed to save settings: %v", err)
			response.Error(w, err, http.StatusInternalServerError)
			return
		}

		if shouldCleanupFreshRSSData(wasFreshRSSEnabled, req["freshrss_enabled"]) {
			if err := h.DB.CleanupFreshRSSData(); err != nil {
				log.Printf("Failed to cleanup FreshRSS data after disabling sync: %v", err)
				response.Error(w, err, http.StatusInternalServerError)
				return
			}
		}

		// Re-fetch all settings after save to return updated values
		settings := GetAllSettings(h)
		response.JSON(w, settings)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func shouldCleanupFreshRSSData(wasEnabled bool, newValue string) bool {
	return wasEnabled && strings.EqualFold(strings.TrimSpace(newValue), "false")
}
