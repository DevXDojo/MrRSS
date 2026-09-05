package feed

import (
	"encoding/json"
	"errors"
	"net/http"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// HandleCategory changes all local subscriptions in a category and its descendants.
// @Summary Dissolve a category or unsubscribe its feeds
// @Description Atomically dissolve a category into uncategorized feeds or unsubscribe the category tree. FreshRSS categories are protected.
// @Tags feeds
// @Accept json
// @Produce json
// @Param request body object true "category and action (dissolve or unsubscribe)"
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /feeds/category [post]
func HandleCategory(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Category *string `json:"category"`
		Action   string  `json:"action"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&req); err != nil || req.Category == nil || len(*req.Category) > 4096 || (req.Action != "dissolve" && req.Action != "unsubscribe") || (req.Action == "dissolve" && *req.Category == "") {
		response.Error(w, errors.New("invalid category action"), http.StatusBadRequest)
		return
	}
	count, err := h.DB.ChangeFeedCategory(r.Context(), *req.Category, req.Action)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, database.ErrSyncedCategory) {
			status = http.StatusForbidden
		}
		response.Error(w, err, status)
		return
	}
	response.JSON(w, map[string]int{"affected": count})
}
