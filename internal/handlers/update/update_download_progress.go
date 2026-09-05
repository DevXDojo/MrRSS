package update

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
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
