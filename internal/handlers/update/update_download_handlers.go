package update

import (
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HandleDownloadUpdate downloads the update file.
// @Summary      Download update
// @Description  Download the update file from GitHub releases to the temp directory
// @Tags         update
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Download request (download_url, asset_name, optional request_id)"
// @Success      200  {object}  map[string]interface{}  "Download success (success, request_id, file_path, total_bytes, bytes_written)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid URL, asset name, or request ID)"
// @Failure      500  {object}  map[string]string  "Download failed"
// @Router       /update/download [post]
func HandleDownloadUpdate(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DownloadURL string `json:"download_url"`
		AssetName   string `json:"asset_name"`
		RequestID   string `json:"request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	const allowedURLPrefix = "https://github.com/DevXDojo/MrRSS/releases/download/"
	if !strings.HasPrefix(req.DownloadURL, allowedURLPrefix) {
		log.Printf("Invalid update download URL")
		response.Error(w, fmt.Errorf("invalid download URL"), http.StatusBadRequest)
		return
	}

	if !filepath.IsLocal(req.AssetName) || strings.Contains(req.AssetName, "..") || strings.ContainsAny(req.AssetName, "/\\:") {
		log.Printf("Invalid update asset name")
		response.Error(w, fmt.Errorf("invalid asset name"), http.StatusBadRequest)
		return
	}

	if req.RequestID == "" {
		req.RequestID = newDownloadRequestID()
	}
	if !validDownloadRequestID(req.RequestID) {
		response.Error(w, fmt.Errorf("invalid request ID"), http.StatusBadRequest)
		return
	}

	setDownloadProgress(downloadProgress{RequestID: req.RequestID, State: "starting", Indeterminate: true})
	client, err := createUpdateHTTPClient(h, downloadRequestTimeout)
	if err != nil {
		failDownload(req.RequestID, "download_proxy_error")
		writeDownloadError(w, req.RequestID, "download_proxy_error")
		return
	}

	// Keep simultaneous downloads and pre-existing temp files isolated.
	filePath, err := createUpdateDownloadPath(req.AssetName)
	if err != nil {
		failDownload(req.RequestID, "download_failed")
		writeDownloadError(w, req.RequestID, "download_failed")
		return
	}
	// On failure the partial file is removed below, so this removes the empty
	// directory. A completed installer keeps its directory until installation.
	defer func() { _ = os.Remove(filepath.Dir(filePath)) }()
	log.Printf("Downloading update asset %s (request %s)", req.AssetName, req.RequestID)
	written, total, err := downloadUpdateFile(r.Context(), client, req.DownloadURL, filePath, req.RequestID, downloadMaxAttempts, time.Second)
	if err != nil {
		code := classifyDownloadError(err)
		failDownload(req.RequestID, code)
		log.Printf("Update download failed (request %s, code %s): %v", req.RequestID, code, err)
		writeDownloadError(w, req.RequestID, code)
		return
	}

	setDownloadProgress(downloadProgress{
		RequestID: req.RequestID, State: "completed", BytesWritten: written,
		TotalBytes: total, Percentage: 100,
	})
	log.Printf("Update downloaded successfully (request %s, %.2f MB)", req.RequestID, float64(written)/(1024*1024))

	response.JSON(w, map[string]interface{}{
		"success": true, "request_id": req.RequestID, "file_path": filePath,
		"total_bytes": total, "bytes_written": written,
	})
}

func createUpdateDownloadPath(assetName string) (string, error) {
	dir, err := os.MkdirTemp("", "mrrss-update-*")
	if err != nil {
		return "", fmt.Errorf("create update directory: %w", err)
	}
	return filepath.Join(dir, assetName), nil
}

// HandleDownloadUpdateProgress reports download progress for a caller-supplied request ID.
// @Summary      Get update download progress
// @Description  Get real-time progress for an update download request
// @Tags         update
// @Produce      json
// @Param        request_id  query     string  true  "Download request ID"
// @Success      200  {object}  map[string]interface{}  "Download progress"
// @Failure      400  {object}  map[string]string  "Invalid request ID"
// @Failure      404  {object}  map[string]string  "Download request not found"
// @Router       /update/download/progress [get]
func HandleDownloadUpdateProgress(_ *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}
	requestID := r.URL.Query().Get("request_id")
	if !validDownloadRequestID(requestID) {
		response.Error(w, fmt.Errorf("invalid request ID"), http.StatusBadRequest)
		return
	}
	updateDownloads.RLock()
	progress, ok := updateDownloads.items[requestID]
	updateDownloads.RUnlock()
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		response.JSON(w, map[string]interface{}{"success": false, "error_code": "download_not_found"})
		return
	}
	response.JSON(w, progress)
}
