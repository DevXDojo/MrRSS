package response

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrors "MrRSS/internal/errors"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error information in API responses
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON writes a JSON response with success status
// Note: This serializes data directly without wrapping to maintain backward compatibility
// with the existing frontend API contract
func JSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

// Error writes an error response with appropriate status code
func Error(w http.ResponseWriter, err error, defaultStatus int) {
	status := defaultStatus
	var appErr *apperrors.AppError

	if errors.As(err, &appErr) {
		switch appErr.Code {
		case apperrors.ErrCodeInvalidInput, apperrors.ErrCodeFeedInvalidURL, apperrors.ErrCodeArticleInvalidID:
			status = http.StatusBadRequest
		case apperrors.ErrCodeNotFound, apperrors.ErrCodeFeedNotFound, apperrors.ErrCodeArticleNotFound:
			status = http.StatusNotFound
		case apperrors.ErrCodeUnauthorized:
			status = http.StatusUnauthorized
		case apperrors.ErrCodeAIQuotaExceeded:
			status = http.StatusTooManyRequests
		default:
			status = http.StatusInternalServerError
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorInfo := &ErrorInfo{
		Code:    string(apperrors.ErrCodeInternal),
		Message: "An internal error occurred",
	}

	if appErr != nil {
		errorInfo.Code = string(appErr.Code)
		errorInfo.Message = appErr.Message
	} else if err != nil {
		errorInfo.Message = err.Error()
	}

	resp := APIResponse{
		Success: false,
		Error:   errorInfo,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
