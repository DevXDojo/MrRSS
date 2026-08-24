package update

import (
	"net/http"
	"strconv"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/version"
)

// HandleVersion returns the current application version.
// @Summary      Get application version
// @Description  Get the current application version string
// @Tags         update
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "Application version (version)"
// @Router       /version [get]
func HandleVersion(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	response.JSON(w, map[string]string{
		"version": version.Version,
	})
}

// compareVersions compares two semantic versions (e.g., "1.1.0" vs "1.0.0")
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			p2, _ = strconv.Atoi(parts2[i])
		}

		if p1 > p2 {
			return 1
		} else if p1 < p2 {
			return -1
		}
	}

	return 0
}
