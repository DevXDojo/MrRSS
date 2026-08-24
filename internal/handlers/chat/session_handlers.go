package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/utils/textutil"
)

// CreateSessionRequest represents the request to create a new chat session
type CreateSessionRequest struct {
	ArticleID int64  `json:"article_id"`
	Title     string `json:"title"`
}

// UpdateSessionRequest represents the request to update a chat session
type UpdateSessionRequest struct {
	Title string `json:"title"`
}

// HandleListSessions handles GET requests to list all chat sessions for an article
// @Summary      List chat sessions
// @Description  Get all chat sessions for a specific article
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        article_id  query     int64   true  "Article ID"
// @Success      200  {array}   database.ChatSession  "List of chat sessions"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid article_id)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat/sessions [get]
func HandleListSessions(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get article_id from query parameter
	articleIDStr := r.URL.Query().Get("article_id")
	if articleIDStr == "" {
		response.Error(w, fmt.Errorf("missing article_id parameter"), http.StatusBadRequest)
		return
	}

	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid article_id"), http.StatusBadRequest)
		return
	}

	sessions, err := h.DB.GetChatSessionsByArticle(articleID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, sessions)
}

// HandleCreateSession handles POST requests to create a new chat session
// @Summary      Create chat session
// @Description  Create a new chat session for an article
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        request  body      chat.CreateSessionRequest  true  "Session creation request (article_id, title)"
// @Success      200  {object}  database.ChatSession  "Created chat session"
// @Failure      400  {object}  map[string]string  "Bad request (missing article_id)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat/sessions [post]
func HandleCreateSession(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if req.ArticleID == 0 {
		response.Error(w, fmt.Errorf("missing article_id"), http.StatusBadRequest)
		return
	}

	// Generate default title if not provided
	title := req.Title
	if title == "" {
		title = "New Chat"
	}

	sessionID, err := h.DB.CreateChatSession(req.ArticleID, title)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Get the created session
	session, err := h.DB.GetChatSession(sessionID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, session)
}

// HandleGetSession handles GET requests to retrieve a specific chat session
// @Summary      Get chat session
// @Description  Get details of a specific chat session
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        session_id  query     int64   true  "Session ID"
// @Success      200  {object}  database.ChatSession  "Chat session details"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid session_id)"
// @Failure      404  {object}  map[string]string  "Session not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat/session [get]
func HandleGetSession(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get session_id from query parameter
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		response.Error(w, fmt.Errorf("missing session_id parameter"), http.StatusBadRequest)
		return
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid session_id"), http.StatusBadRequest)
		return
	}

	session, err := h.DB.GetChatSession(sessionID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	if session == nil {
		response.Error(w, fmt.Errorf("session not found"), http.StatusNotFound)
		return
	}

	response.JSON(w, session)
}

// HandleUpdateSession handles PUT requests to update a chat session
// @Summary      Update chat session
// @Description  Update the title of a chat session
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        session_id  query     int64                            true  "Session ID"
// @Param        request    body      chat.UpdateSessionRequest  true  "Update request (title)"
// @Success      200  {object}  database.ChatSession  "Updated chat session"
// @Failure      400  {object}  map[string]string  "Bad request (missing session_id or title)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat/session [put]
func HandleUpdateSession(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get session_id from query parameter
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		response.Error(w, fmt.Errorf("missing session_id parameter"), http.StatusBadRequest)
		return
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid session_id"), http.StatusBadRequest)
		return
	}

	var req UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		response.Error(w, fmt.Errorf("missing title"), http.StatusBadRequest)
		return
	}

	err = h.DB.UpdateChatSessionTitle(sessionID, req.Title)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Get the updated session
	session, err := h.DB.GetChatSession(sessionID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, session)
}

// HandleDeleteSession handles DELETE requests to delete a chat session
// @Summary      Delete chat session
// @Description  Delete a specific chat session and all its messages
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        session_id  query     int64   true  "Session ID"
// @Success      200  {string}  string  "Session deleted successfully"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid session_id)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat/session [delete]
func HandleDeleteSession(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get session_id from query parameter
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		response.Error(w, fmt.Errorf("missing session_id parameter"), http.StatusBadRequest)
		return
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid session_id"), http.StatusBadRequest)
		return
	}

	err = h.DB.DeleteChatSession(sessionID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]string{"status": "deleted"})
}

// HandleListMessages handles GET requests to list all messages in a session
// HandleListMessages lists all messages in a chat session
// @Summary      List chat messages
// @Description  Get all messages in a specific chat session
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        session_id  query     int64   true  "Session ID"
// @Success      200  {array}   object  "List of chat messages (with HTML for assistant messages)"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid session_id)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat/messages [get]
func HandleListMessages(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get session_id from query parameter
	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		response.Error(w, fmt.Errorf("missing session_id parameter"), http.StatusBadRequest)
		return
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid session_id"), http.StatusBadRequest)
		return
	}

	messages, err := h.DB.GetChatMessages(sessionID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Convert markdown to HTML for assistant messages
	type MessageWithHTML struct {
		ID        int64  `json:"id"`
		SessionID int64  `json:"session_id"`
		Role      string `json:"role"`
		Content   string `json:"content"`
		HTML      string `json:"html,omitempty"` // Pre-rendered HTML for assistant messages
		Thinking  string `json:"thinking,omitempty"`
		CreatedAt string `json:"created_at"`
	}

	result := make([]MessageWithHTML, len(messages))
	for i, msg := range messages {
		result[i] = MessageWithHTML{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Role:      msg.Role,
			Content:   msg.Content,
			Thinking:  msg.Thinking,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		// Generate HTML for assistant messages
		if msg.Role == "assistant" {
			result[i].HTML = textutil.ConvertMarkdownToHTML(msg.Content)
		}
	}

	response.JSON(w, result)
}

// HandleDeleteMessage handles DELETE requests to delete a specific message
// @Summary      Delete chat message
// @Description  Delete a specific chat message
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        message_id  query     int64   true  "Message ID"
// @Success      200  {object}  map[string]string  "Deletion status (status: 'deleted')"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid message_id)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat/message [delete]
func HandleDeleteMessage(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get message_id from query parameter
	messageIDStr := r.URL.Query().Get("message_id")
	if messageIDStr == "" {
		response.Error(w, fmt.Errorf("missing message_id parameter"), http.StatusBadRequest)
		return
	}

	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid message_id"), http.StatusBadRequest)
		return
	}

	err = h.DB.DeleteChatMessage(messageID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]string{"status": "deleted"})
}

// HandleDeleteAllSessions handles DELETE requests to delete all chat sessions
// @Summary      Delete all chat sessions
// @Description  Delete all chat sessions and their messages
// @Tags         chat
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Deletion result (status, count)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat/sessions/all [delete]
func HandleDeleteAllSessions(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	count, err := h.DB.DeleteAllChatSessions()
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]interface{}{
		"status": "deleted",
		"count":  count,
	})
}
