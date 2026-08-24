package errors

import "fmt"

// ErrorCode represents a unique error code for API responses
type ErrorCode string

const (
	// General errors (1000-1999)
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"

	// Feed errors (2000-2999)
	ErrCodeFeedNotFound    ErrorCode = "FEED_NOT_FOUND"
	ErrCodeFeedInvalidURL  ErrorCode = "FEED_INVALID_URL"
	ErrCodeFeedFetchFailed ErrorCode = "FEED_FETCH_FAILED"
	ErrCodeFeedParseFailed ErrorCode = "FEED_PARSE_FAILED"
	ErrCodeFeedDuplicate   ErrorCode = "FEED_DUPLICATE"

	// Article errors (3000-3999)
	ErrCodeArticleNotFound  ErrorCode = "ARTICLE_NOT_FOUND"
	ErrCodeArticleInvalidID ErrorCode = "ARTICLE_INVALID_ID"

	// AI errors (4000-4999)
	ErrCodeAIConfigFailed   ErrorCode = "AI_CONFIG_FAILED"
	ErrCodeAIRequestFailed  ErrorCode = "AI_REQUEST_FAILED"
	ErrCodeAIQuotaExceeded  ErrorCode = "AI_QUOTA_EXCEEDED"
	ErrCodeAIInvalidRequest ErrorCode = "AI_INVALID_REQUEST"

	// Translation errors (5000-5999)
	ErrCodeTranslationFailed ErrorCode = "TRANSLATION_FAILED"

	// Database errors (6000-6999)
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"
)

// AppError represents an application error with code and message
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewFeedError creates a feed-related error
func NewFeedError(code ErrorCode, message string, err error) *AppError {
	return NewAppError(code, message, err)
}

// NewArticleError creates an article-related error
func NewArticleError(code ErrorCode, message string, err error) *AppError {
	return NewAppError(code, message, err)
}

// NewAIError creates an AI-related error
func NewAIError(code ErrorCode, message string, err error) *AppError {
	return NewAppError(code, message, err)
}

// NewTranslationError creates a translation-related error
func NewTranslationError(code ErrorCode, message string, err error) *AppError {
	return NewAppError(code, message, err)
}
