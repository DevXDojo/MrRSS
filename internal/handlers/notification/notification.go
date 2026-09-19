package notification

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/notification"
)

func decode(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 128*1024)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return errors.New("invalid_request")
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		return errors.New("invalid_request")
	}
	return nil
}

func HandleConfig(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	switch r.Method {
	case http.MethodGet:
		handleGetConfig(ctx, h, w)
	case http.MethodPut:
		handleUpdateConfig(ctx, h, w, r)
	default:
		w.Header().Set("Allow", "GET, PUT")
		w.WriteHeader(405)
	}
}

// @Summary Read notification settings
// @Description Returns configuration and revision with secret placeholders. Automatic notifications are opt-in.
// @Tags notifications
// @Produce json
// @Success 200 {object} notification.Config
// @Router /notifications/config [get]
func handleGetConfig(ctx context.Context, h *core.Handler, w http.ResponseWriter) {
	c, err := h.Notifications.Load(ctx)
	if err != nil {
		response.Error(w, errors.New("config_unavailable"), http.StatusInternalServerError)
		return
	}
	response.JSON(w, c.Redacted())
}

// @Summary Save notification settings
// @Description Atomically saves validated configuration and new-article baselines. Preserve secret placeholders from GET. Edited or disabled rules cancel affected pending messages. A stale revision returns 409.
// @Tags notifications
// @Accept json
// @Produce json
// @Param config body notification.Config true "Complete configuration with current revision"
// @Success 200 {object} notification.Config
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /notifications/config [put]
func handleUpdateConfig(ctx context.Context, h *core.Handler, w http.ResponseWriter, r *http.Request) {
	var c notification.Config
	if err := decode(w, r, &c); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}
	saved, err := h.Notifications.Save(ctx, c)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, notification.ErrConflict) {
			status = http.StatusConflict
		}
		response.Error(w, err, status)
		return
	}
	response.JSON(w, saved)
}

// @Summary Test a notification destination
// @Description Sends a fixed test without article content. Uses draft credentials or matching saved placeholders, even when automatic notifications are disabled.
// @Tags notifications
// @Accept json
// @Produce json
// @Param channel body notification.Channel true "Destination to test"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.APIResponse
// @Router /notifications/test [post]
func HandleTest(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(405)
		return
	}
	var ch notification.Channel
	if err := decode(w, r, &ch); err != nil {
		response.Error(w, err, 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	if err := h.Notifications.Test(ctx, ch); err != nil {
		response.Error(w, err, 400)
		return
	}
	response.JSON(w, map[string]bool{"success": true})
}

// @Summary Preview a notification rule
// @Description Tests a draft rule against the latest 500 local articles, without sending messages or changing reading state. Production rules process articles received after activation.
// @Tags notifications
// @Accept json
// @Produce json
// @Param rule body notification.Rule true "Rule to preview"
// @Success 200 {object} notification.Preview
// @Failure 400 {object} response.APIResponse
// @Router /notifications/preview [post]
func HandlePreview(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(405)
		return
	}
	var rule notification.Rule
	if err := decode(w, r, &rule); err != nil {
		response.Error(w, err, 400)
		return
	}
	p, err := h.Notifications.Preview(r.Context(), rule)
	if err != nil {
		response.Error(w, err, 400)
		return
	}
	response.JSON(w, p)
}

type HistoryResponse []database.NotificationDelivery

// @Summary List notification deliveries
// @Description Latest 100 delivery records. Excludes message bodies and credentials. Completed records are retained for up to 30 days, capped at 1000.
// @Tags notifications
// @Produce json
// @Success 200 {object} HistoryResponse
// @Router /notifications/history [get]
func HandleHistory(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(405)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	history, err := h.Notifications.History(r.Context())
	if err != nil {
		response.Error(w, errors.New("history_unavailable"), 500)
		return
	}
	response.JSON(w, HistoryResponse(history))
}
