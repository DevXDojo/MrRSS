package feed

import (
	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"encoding/json"
	"net/http"
	"strconv"
)

// HandleContentOptions reads or saves per-feed full-text extraction settings.
// @Summary Feed content extraction settings
// @Description Get or save content and removal CSS selectors for one feed.
// @Tags feeds
// @Accept json
// @Produce json
// @Param id query int64 true "Feed ID"
// @Success 200 {object} database.FeedContentOptions
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /feeds/content-options [get]
// @Router /feeds/content-options [post]
func HandleContentOptions(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}
	if _, err := h.DB.GetFeedByID(id); err != nil {
		response.Error(w, err, http.StatusNotFound)
		return
	}
	if r.Method == http.MethodPost {
		var options database.FeedContentOptions
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32768)).Decode(&options); err != nil {
			response.Error(w, err, http.StatusBadRequest)
			return
		}
		if err := options.Validate(); err != nil {
			response.Error(w, err, http.StatusBadRequest)
			return
		}
		if err := h.DB.SetFeedContentOptions(r.Context(), id, options); err != nil {
			response.Error(w, err, http.StatusInternalServerError)
			return
		}
	}
	options, err := h.DB.GetFeedContentOptions(r.Context(), id)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	options.Cookie = nil // Never return saved credentials to the UI.
	response.JSON(w, options)
}
