package article

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/siyuan"
	"MrRSS/internal/utils/httputil"
)

func HandleExportToSiYuan(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		ArticleID int64 `json:"article_id"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&input); err != nil || input.ArticleID <= 0 {
		response.Error(w, fmt.Errorf("invalid article ID"), http.StatusBadRequest)
		return
	}
	enabled, _ := h.DB.GetSetting("siyuan_enabled")
	endpointSetting, _ := h.DB.GetSetting("siyuan_endpoint")
	notebook, _ := h.DB.GetSetting("siyuan_notebook_id")
	folder, _ := h.DB.GetSetting("siyuan_folder")
	token, err := h.DB.GetEncryptedSetting("siyuan_api_token")
	if err != nil {
		response.Error(w, fmt.Errorf("could not load SiYuan token"), http.StatusInternalServerError)
		return
	}
	endpoint, err := siyuan.ParseEndpoint(endpointSetting)
	if err != nil || enabled != "true" || !siyuan.ValidBlockID(notebook) {
		response.Error(w, fmt.Errorf("configure SiYuan integration, endpoint, and notebook first"), http.StatusBadRequest)
		return
	}
	article, err := h.DB.GetArticleByID(input.ArticleID)
	if err != nil || article == nil {
		response.Error(w, fmt.Errorf("article not found"), http.StatusNotFound)
		return
	}
	path, err := siyuan.DocumentPath(folder, article.Title, article.ID)
	if err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	content, _, err := h.GetArticleContentContext(ctx, article.ID)
	if err != nil || strings.TrimSpace(content) == "" {
		response.Error(w, fmt.Errorf("load article content before exporting"), http.StatusUnprocessableEntity)
		return
	}
	markdown, err := siyuan.ArticleMarkdown(article, content)
	if err != nil {
		response.Error(w, fmt.Errorf("could not convert article to Markdown"), http.StatusInternalServerError)
		return
	}
	var proxySettings httputil.ProxySettingsProvider = h.DB
	if siyuan.IsLoopback(endpoint) {
		proxySettings = nil
	}
	client, err := httputil.CreateHTTPClientWithProxySettings(proxySettings, 30*time.Second)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid proxy configuration"), http.StatusBadRequest)
		return
	}
	defer client.CloseIdleConnections()
	id, err := siyuan.CreateDocument(ctx, client, endpoint, token, notebook, path, markdown)
	if err != nil {
		response.Error(w, err, http.StatusBadGateway)
		return
	}
	response.JSON(w, map[string]string{"document_id": id})
}
