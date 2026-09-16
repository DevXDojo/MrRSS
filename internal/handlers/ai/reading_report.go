package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"MrRSS/internal/ai"
	"MrRSS/internal/config"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/readingreport"
	"MrRSS/internal/utils/httputil"
)

type ReadingReportRequest struct {
	ArticleIDs []int64 `json:"article_ids"`
	ProfileID  int64   `json:"profile_id,omitempty"`
	Focus      string  `json:"focus,omitempty"`
}

type ReadingReportPreview struct {
	Sources []readingreport.Source `json:"sources"`
}
type ReadingReportResponse struct {
	Report  readingreport.Report   `json:"report"`
	Sources []readingreport.Source `json:"sources"`
	Model   string                 `json:"model"`
}

// HandlePreviewReadingReport previews local evidence without contacting AI.
// @Summary Preview reading report sources
// @Description Inspect local content coverage for 1-20 distinct article IDs; no external requests or reading-state changes.
// @Tags ai
// @Accept json
// @Produce json
// @Param request body handlers.ReadingReportRequest true "Selected article IDs"
// @Success 200 {object} handlers.ReadingReportPreview
// @Failure 400 {object} map[string]interface{}
// @Router /ai/reading-report/preview [post]
func HandlePreviewReadingReport(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	HandleReadingReport(h, w, r, true)
}

// HandleGenerateReadingReport generates a report with validated source IDs.
// @Summary Generate AI reading report
// @Description Generate a transient report from 1-20 selected local articles, with optional AI profile and focus (500 characters maximum). Honors AI usage limits, proxy settings and request cancellation. Does not mark articles read or save reports.
// @Tags ai
// @Accept json
// @Produce json
// @Param request body handlers.ReadingReportRequest true "Selected articles, optional profile and focus"
// @Success 200 {object} handlers.ReadingReportResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 429 {object} map[string]interface{}
// @Failure 502 {object} map[string]interface{}
// @Router /ai/reading-report [post]
func HandleGenerateReadingReport(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	HandleReadingReport(h, w, r, false)
}

// HandleReadingReport previews local source coverage or generates an on-demand
// report. It never marks articles read or saves generated reports automatically.
func HandleReadingReport(h *core.Handler, w http.ResponseWriter, r *http.Request, preview bool) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}
	var req ReadingReportRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&req); err != nil || req.ProfileID < 0 || utf8.RuneCountInString(req.Focus) > 500 {
		response.Error(w, errors.New("invalid report selection"), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()
	sources, err := readingreport.LoadSources(ctx, h.DB, req.ArticleIDs)
	if err != nil {
		if errors.Is(err, readingreport.ErrSelection) {
			response.Error(w, err, http.StatusBadRequest)
		} else {
			reportError(w, ai.ErrorCodeRequestFailed)
		}
		return
	}
	if preview {
		response.JSON(w, ReadingReportPreview{Sources: sources})
		return
	}
	language, _ := h.DB.GetSetting("language")
	request, err := readingreport.BuildRequest(sources, language, strings.TrimSpace(req.Focus))
	if err != nil {
		reportError(w, "report_no_content")
		return
	}
	if h.AITracker.IsLimitReached() {
		reportError(w, ai.ErrorCodeUsageLimitReached)
		return
	}
	cfg, err := reportConfig(h, req.ProfileID)
	if err != nil {
		reportError(w, ai.ErrorCodeConfigurationInvalid)
		return
	}
	httpClient, err := httputil.CreateHTTPClientWithProxySettings(h.DB, 120*time.Second)
	if err != nil {
		reportError(w, ai.ErrorCodeConfigurationInvalid)
		return
	}
	defer httpClient.CloseIdleConnections()
	if err := h.AITracker.WaitForRateLimitContext(ctx); err != nil {
		reportError(w, ai.ClassifyUserFacingError(err).Code)
		return
	}
	request.Model = cfg.Model
	result, err := ai.NewClientWithHTTPClient(*cfg, httpClient).RequestWithConfigContext(ctx, request)
	if err != nil {
		reportError(w, ai.ClassifyUserFacingError(err).Code)
		return
	}
	if ctx.Err() != nil {
		return
	}
	// Account for the provider work even when its generated structure is invalid.
	h.AITracker.TrackSummary(request.UserPrompt, result.Content)
	report, err := readingreport.ParseReport(result.Content, sources)
	if err != nil {
		reportError(w, ai.ErrorCodeInvalidResponse)
		return
	}
	response.JSON(w, ReadingReportResponse{Report: report, Sources: sources, Model: cfg.Model})
}

func reportConfig(h *core.Handler, profileID int64) (*ai.ClientConfig, error) {
	if h.AIProfileProvider != nil {
		if profileID > 0 {
			cfg, err := h.AIProfileProvider.GetConfigForProfile(profileID)
			if err != nil || cfg == nil || cfg.Endpoint == "" || cfg.Model == "" {
				return nil, errors.New("missing report profile")
			}
			return cfg, nil
		}
		profile, err := h.AIProfileProvider.GetProfileForFeature(ai.FeatureSummary)
		if err != nil {
			return nil, err
		}
		if profile != nil {
			return reportConfig(h, profile.ID)
		}
	} else if profileID > 0 {
		return nil, errors.New("missing report profile")
	}
	endpoint, _ := h.DB.GetSetting("ai_endpoint")
	model, _ := h.DB.GetSetting("ai_model")
	key, _ := h.DB.GetEncryptedSetting("ai_api_key")
	headers, _ := h.DB.GetSetting("ai_custom_headers")
	if endpoint == "" || model == "" || (endpoint == config.Get().AIEndpoint && key == "" && headers == "") {
		return nil, errors.New("configure AI before creating a report")
	}
	return &ai.ClientConfig{Endpoint: endpoint, Model: model, APIKey: key, CustomHeaders: headers}, nil
}

func reportError(w http.ResponseWriter, code string) {
	err := ai.UserFacingErrorForCode(code)
	if code == "report_no_content" {
		err = ai.UserFacingError{Code: code, Message: "Selected articles have no locally available content", HTTPStatus: http.StatusBadRequest}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	response.JSON(w, map[string]string{"error": err.Message, "error_code": err.Code})
}
