package chat

import (
	"encoding/json"
	"net/http"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// CancelChatRequest identifies one logical generation in a saved session.
type CancelChatRequest struct {
	SessionID int64  `json:"session_id"`
	RequestID string `json:"request_id"`
}

func validChatRequestID(id string) bool {
	return len(id) <= 128 && strings.TrimSpace(id) == id && id != ""
}

// HandleCancelAIChat cancels generation, including a request not yet registered.
// @Summary Stop AI chat generation
// @Description Cancel one request by session_id and request_id. Repeated, finished, and early cancellation all return success.
// @Tags chat
// @Accept json
// @Produce json
// @Param request body chat.CancelChatRequest true "Request to cancel"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /ai-chat/cancel [post]
func HandleCancelAIChat(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}
	var req CancelChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}
	if req.SessionID <= 0 || !validChatRequestID(req.RequestID) {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}
	h.ChatRequests.Cancel(req.SessionID, req.RequestID)
	response.JSON(w, map[string]bool{"success": true})
}
