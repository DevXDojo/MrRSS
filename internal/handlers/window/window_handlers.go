package window

import (
	"encoding/json"
	"fmt"
	"net/http"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// WindowState represents the saved window state
type WindowState struct {
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximized bool `json:"maximized"`
}

// HandleGetWindowState returns the saved window state from database
// @Summary      Get window state
// @Description  Get the saved window state (position, size, maximized status)
// @Tags         window
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "Window state (x, y, width, height, maximized)"
// @Router       /window/state [get]
func HandleGetWindowState(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	x, err := h.DB.GetSetting("window_x")
	if err != nil {
		// Return empty state if not found
		response.JSON(w, map[string]string{
			"x":         "",
			"y":         "",
			"width":     "",
			"height":    "",
			"maximized": "",
		})
		return
	}
	y, _ := h.DB.GetSetting("window_y")
	width, _ := h.DB.GetSetting("window_width")
	height, _ := h.DB.GetSetting("window_height")
	maximized, _ := h.DB.GetSetting("window_maximized")

	response.JSON(w, map[string]string{
		"x":         x,
		"y":         y,
		"width":     width,
		"height":    height,
		"maximized": maximized,
	})
}

// HandleSaveWindowState saves the current window state to database
// @Summary      Save window state
// @Description  Save the current window state (position, size, maximized status)
// @Tags         window
// @Accept       json
// @Produce      json
// @Param        state  body      WindowState  true  "Window state to save"
// @Success      200  {string}  string  "Window state saved successfully"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /window/state [post]
func HandleSaveWindowState(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var state WindowState
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Convert to strings for database storage and check for errors
	if err := h.DB.SetSetting("window_x", fmt.Sprintf("%d", state.X)); err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	if err := h.DB.SetSetting("window_y", fmt.Sprintf("%d", state.Y)); err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	if err := h.DB.SetSetting("window_width", fmt.Sprintf("%d", state.Width)); err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	if err := h.DB.SetSetting("window_height", fmt.Sprintf("%d", state.Height)); err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	if err := h.DB.SetSetting("window_maximized", fmt.Sprintf("%t", state.Maximized)); err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]string{"status": "ok"})
}
