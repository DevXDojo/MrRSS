package update

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// HandleDownloadUpdate downloads the update file.
// @Summary      Download update
// @Description  Desktop-only update download; requires a native capability and verifies the official release asset SHA-256 digest. Disabled in server mode.
// @Tags         update
// @Param        Authorization header string true "Native desktop capability (Bearer token)"
// @Failure      403  {object}  map[string]string  "Desktop update authorization required"
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Download request (download_url, asset_name, optional request_id)"
// @Success      200  {object}  map[string]interface{}  "Download success (success, request_id, file_path, total_bytes, bytes_written)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid URL, asset name, or request ID)"
// @Failure      500  {object}  map[string]string  "Download failed"
// @Router       /download-update [post]
func HandleDownloadUpdate(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}
	if !authorizeUpdate(h, w, r) {
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

	if _, err := parseOfficialAsset(req.DownloadURL, req.AssetName); err != nil {
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
	metadataCtx, metadataCancel := context.WithTimeout(r.Context(), 30*time.Second)
	digest, err := officialAssetDigest(metadataCtx, client, req.DownloadURL, req.AssetName)
	metadataCancel()
	if err != nil {
		failDownload(req.RequestID, "download_verification_failed")
		writeDownloadError(w, req.RequestID, "download_verification_failed")
		return
	}
	restrictUpdateRedirects(client, req.DownloadURL)

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

	if err := verifyUpdateDigest(r.Context(), filePath, digest); err != nil {
		cleanupUpdateDownload(filePath)
		failDownload(req.RequestID, "download_verification_failed")
		writeDownloadError(w, req.RequestID, "download_verification_failed")
		return
	}
	verifiedUpdates.Store(verifiedUpdateKey{h, filePath}, digest)
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
// @Router       /download-update/progress [get]
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
