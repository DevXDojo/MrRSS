package update

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

const (
	downloadMaxAttempts    = 3
	downloadRequestTimeout = 30 * time.Minute
	downloadProgressTTL    = 30 * time.Minute
)

type downloadProgress struct {
	RequestID     string  `json:"request_id"`
	State         string  `json:"state"`
	BytesWritten  int64   `json:"bytes_written"`
	TotalBytes    int64   `json:"total_bytes"`
	Percentage    float64 `json:"percentage"`
	Indeterminate bool    `json:"indeterminate"`
	ErrorCode     string  `json:"error_code,omitempty"`
	updatedAt     time.Time
}

var updateDownloads = struct {
	sync.RWMutex
	items map[string]downloadProgress
}{items: make(map[string]downloadProgress)}

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

	if req.AssetName == "" || strings.Contains(req.AssetName, "..") || strings.Contains(req.AssetName, "/") || strings.Contains(req.AssetName, "\\") {
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

	filePath := filepath.Join(os.TempDir(), req.AssetName)
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

func downloadUpdateFile(ctx context.Context, client *http.Client, downloadURL, filePath, requestID string, maxAttempts int, baseDelay time.Duration) (int64, int64, error) {
	partPath := filePath + ".part"
	_ = os.Remove(partPath)
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		written, total, retry, err := downloadUpdateAttempt(ctx, client, downloadURL, partPath, requestID)
		if err == nil {
			if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
				_ = os.Remove(partPath)
				return 0, 0, fmt.Errorf("replace existing update: %w", err)
			}
			if err := os.Rename(partPath, filePath); err != nil {
				_ = os.Remove(partPath)
				return 0, 0, fmt.Errorf("finalize update: %w", err)
			}
			return written, total, nil
		}
		lastErr = err
		if !retry || attempt == maxAttempts || ctx.Err() != nil {
			break
		}
		delay := baseDelay * time.Duration(1<<(attempt-1))
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				lastErr = ctx.Err()
				attempt = maxAttempts
			case <-timer.C:
			}
		}
	}

	_ = os.Remove(partPath)
	return 0, 0, lastErr
}

func downloadUpdateAttempt(ctx context.Context, client *http.Client, downloadURL, partPath, requestID string) (int64, int64, bool, error) {
	offset := int64(0)
	if info, err := os.Stat(partPath); err == nil {
		offset = info.Size()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return 0, 0, false, err
	}
	req.Header.Set("User-Agent", "MrRSS-Updater")
	req.Header.Set("Accept", "application/octet-stream")
	if offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, isRetryableDownloadError(err), err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable && offset > 0 {
		_ = os.Remove(partPath)
		return 0, 0, true, fmt.Errorf("range no longer valid")
	}
	if resp.StatusCode >= 500 {
		return 0, 0, true, fmt.Errorf("download server returned %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return 0, 0, false, fmt.Errorf("download server returned %d", resp.StatusCode)
	}

	resume := offset > 0 && resp.StatusCode == http.StatusPartialContent
	if resume {
		start, _, ok := parseContentRange(resp.Header.Get("Content-Range"))
		if !ok || start != offset {
			_ = os.Remove(partPath)
			return 0, 0, true, fmt.Errorf("invalid range response")
		}
	} else {
		offset = 0
	}

	flags := os.O_CREATE | os.O_WRONLY
	if resume {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	out, err := os.OpenFile(partPath, flags, 0o600)
	if err != nil {
		return 0, 0, false, err
	}

	total := responseTotalSize(resp, offset)
	setDownloadProgress(downloadProgress{
		RequestID: requestID, State: "downloading", BytesWritten: offset,
		TotalBytes: total, Percentage: downloadPercentage(offset, total), Indeterminate: total <= 0,
	})

	buffer := make([]byte, 64*1024)
	written := offset
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			wn, writeErr := out.Write(buffer[:n])
			written += int64(wn)
			setDownloadProgress(downloadProgress{
				RequestID: requestID, State: "downloading", BytesWritten: written,
				TotalBytes: total, Percentage: downloadPercentage(written, total), Indeterminate: total <= 0,
			})
			if writeErr != nil {
				_ = out.Close()
				return written, total, false, writeErr
			}
			if wn != n {
				_ = out.Close()
				return written, total, false, io.ErrShortWrite
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			_ = out.Close()
			return written, total, isRetryableDownloadError(readErr), readErr
		}
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		return written, total, false, err
	}
	if err := out.Close(); err != nil {
		return written, total, false, err
	}
	if total > 0 && written != total {
		return written, total, true, fmt.Errorf("download incomplete: expected %d bytes, got %d", total, written)
	}
	return written, total, false, nil
}

func responseTotalSize(resp *http.Response, offset int64) int64 {
	if resp.StatusCode == http.StatusPartialContent {
		_, total, ok := parseContentRange(resp.Header.Get("Content-Range"))
		if ok {
			return total
		}
	}
	if resp.ContentLength > 0 {
		return offset + resp.ContentLength
	}
	return 0
}

func parseContentRange(value string) (int64, int64, bool) {
	if !strings.HasPrefix(value, "bytes ") {
		return 0, 0, false
	}
	parts := strings.Split(strings.TrimPrefix(value, "bytes "), "/")
	if len(parts) != 2 {
		return 0, 0, false
	}
	rangeParts := strings.Split(parts[0], "-")
	if len(rangeParts) != 2 {
		return 0, 0, false
	}
	start, err1 := strconv.ParseInt(rangeParts[0], 10, 64)
	total, err2 := strconv.ParseInt(parts[1], 10, 64)
	return start, total, err1 == nil && err2 == nil && start >= 0 && total > 0
}

func isRetryableDownloadError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	return errors.As(err, &netErr) || errors.Is(err, io.ErrUnexpectedEOF)
}

func classifyDownloadError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return "download_timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return "download_timeout"
		}
		return "download_network_error"
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "server returned 5") {
		return "download_server_error"
	}
	if strings.Contains(message, "incomplete") {
		return "download_incomplete"
	}
	return "download_failed"
}

func downloadPercentage(written, total int64) float64 {
	if total <= 0 {
		return 0
	}
	percentage := float64(written) * 100 / float64(total)
	if percentage > 100 {
		return 100
	}
	return percentage
}

func setDownloadProgress(progress downloadProgress) {
	progress.updatedAt = time.Now()
	updateDownloads.Lock()
	updateDownloads.items[progress.RequestID] = progress
	for id, item := range updateDownloads.items {
		if progress.updatedAt.Sub(item.updatedAt) > downloadProgressTTL {
			delete(updateDownloads.items, id)
		}
	}
	updateDownloads.Unlock()
}

func failDownload(requestID, errorCode string) {
	updateDownloads.RLock()
	progress := updateDownloads.items[requestID]
	updateDownloads.RUnlock()
	progress.RequestID = requestID
	progress.State = "failed"
	progress.ErrorCode = errorCode
	setDownloadProgress(progress)
}

func writeDownloadError(w http.ResponseWriter, requestID, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false, "request_id": requestID, "error_code": code,
	})
}

func validDownloadRequestID(requestID string) bool {
	if requestID == "" || len(requestID) > 128 {
		return false
	}
	for _, r := range requestID {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func newDownloadRequestID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err == nil {
		return hex.EncodeToString(buffer)
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
