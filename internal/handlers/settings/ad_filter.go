package settings

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"MrRSS/internal/adfilter"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/utils/textutil"
)

func decodeAdFilter(w http.ResponseWriter, r *http.Request, value interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(value) != nil {
		return errors.New("invalid_request")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return errors.New("invalid_request")
	}
	return nil
}

// HandleAdFilterConfig manages validated local advertising filter settings.
// @Summary Read or update advertising filters
// @Tags settings
// @Produce json
// @Success 200 {object} adfilter.Config
// @Router /ad-filter/config [get]
func HandleAdFilterConfig(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if r.Method == http.MethodGet {
		config, err := h.AdFilterConfig(ctx)
		if err != nil {
			response.Error(w, errors.New("config_unavailable"), 500)
			return
		}
		response.JSON(w, config)
		return
	}
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", "GET, PUT")
		w.WriteHeader(405)
		return
	}
	saveAdFilterConfig(ctx, h, w, r)
}

// saveAdFilterConfig validates and atomically updates advertising filters.
// @Summary Save advertising filters with revision conflict protection
// @Tags settings
// @Accept json
// @Produce json
// @Param config body adfilter.Config true "Filtering options and current revision"
// @Success 200 {object} adfilter.Config
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /ad-filter/config [put]
func saveAdFilterConfig(ctx context.Context, h *core.Handler, w http.ResponseWriter, r *http.Request) {
	var config adfilter.Config
	if err := decodeAdFilter(w, r, &config); err != nil {
		response.Error(w, err, 400)
		return
	}
	if err := adfilter.Validate(config.Options); err != nil {
		response.Error(w, err, 400)
		return
	}
	current, err := h.DB.GetSettingContext(ctx, "ad_filter_config")
	if err != nil {
		response.Error(w, errors.New("config_unavailable"), 500)
		return
	}
	old := adfilter.DefaultConfig()
	if current != "" && json.Unmarshal([]byte(current), &old) != nil {
		response.Error(w, errors.New("config_unavailable"), 500)
		return
	}
	if config.Revision != old.Revision {
		response.Error(w, errors.New("conflict"), 409)
		return
	}
	config.Revision++
	encoded, _ := json.Marshal(config)
	saved, err := h.DB.CompareAndSwapSetting(ctx, "ad_filter_config", current, string(encoded))
	if err != nil {
		response.Error(w, errors.New("save_failed"), 500)
		return
	}
	if !saved {
		response.Error(w, errors.New("conflict"), 409)
		return
	}
	response.JSON(w, config)
}

type AdFilterPreviewRequest struct {
	Options adfilter.Options `json:"options"`
	URL     string           `json:"url"`
	HTML    string           `json:"html"`
}

// HandleAdFilterPreview evaluates a pasted HTML sample without network access.
// @Summary Preview advertising filters locally
// @Tags settings
// @Accept json
// @Produce json
// @Param preview body AdFilterPreviewRequest true "HTML sample and candidate options"
// @Success 200 {object} adfilter.Content
// @Router /ad-filter/preview [post]
func HandleAdFilterPreview(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(405)
		return
	}
	var request AdFilterPreviewRequest
	if err := decodeAdFilter(w, r, &request); err != nil {
		response.Error(w, err, 400)
		return
	}
	if adfilter.Host(request.URL) == "" {
		response.Error(w, errors.New("invalid_url"), 400)
		return
	}
	if err := adfilter.Validate(request.Options); err != nil {
		response.Error(w, err, 400)
		return
	}
	response.JSON(w, textutil.PrepareFilteredArticle(request.HTML, request.URL, request.Options))
}

// HandleAdFilterAnalyze performs bounded, opt-in AI review using existing profiles.
// @Summary Analyze advertising with the configured AI profile
// @Tags settings
// @Accept json
// @Produce json
// @Param preview body AdFilterPreviewRequest true "Article HTML and filter options"
// @Success 200 {object} adfilter.Analysis
// @Router /ad-filter/analyze [post]
func HandleAdFilterAnalyze(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(405)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	var request AdFilterPreviewRequest
	if err := decodeAdFilter(w, r, &request); err != nil {
		response.Error(w, err, 400)
		return
	}
	if adfilter.Host(request.URL) == "" {
		response.Error(w, errors.New("invalid_url"), 400)
		return
	}
	if err := adfilter.Validate(request.Options); err != nil {
		response.Error(w, err, 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 75*time.Second)
	defer cancel()
	result, err := h.AnalyzeArticleAds(ctx, request.HTML, request.URL, request.Options)
	if err != nil {
		code := err.Error()
		if ctx.Err() != nil {
			code = "ai_timeout"
		}
		response.Error(w, errors.New(code), http.StatusUnprocessableEntity)
		return
	}
	response.JSON(w, result)
}
