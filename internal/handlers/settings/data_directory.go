package settings

import (
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/utils/fileutil"
	"encoding/json"
	"net/http"
)

func HandleDataDirectory(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if fileutil.IsServerMode() {
		http.Error(w, "desktop only", http.StatusNotImplemented)
		return
	}
	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		var input struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if _, err := fileutil.ScheduleDataDirectory(input.Path); err != nil {
			response.Error(w, err, http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		if err := fileutil.CancelDataDirectoryChange(); err != nil {
			response.Error(w, err, http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cfg, err := fileutil.DesktopStorageStatus()
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	response.JSON(w, map[string]interface{}{"data_directory": cfg.DataDirectory, "pending_directory": cfg.PendingDirectory, "last_error": cfg.LastError, "managed": fileutil.DesktopStorageManaged()})
}
