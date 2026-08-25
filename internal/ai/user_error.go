package ai

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	ErrorCodeConfigurationInvalid    = "configuration_invalid"
	ErrorCodeUsageLimitReached       = "usage_limit_reached"
	ErrorCodeRateLimited             = "rate_limited"
	ErrorCodeAuthenticationFailed    = "authentication_failed"
	ErrorCodePaymentRequired         = "payment_required"
	ErrorCodeModelOrEndpointNotFound = "model_or_endpoint_not_found"
	ErrorCodeRequestTooLarge         = "request_too_large"
	ErrorCodeTimeout                 = "timeout"
	ErrorCodeNetwork                 = "network_error"
	ErrorCodeProviderUnavailable     = "provider_unavailable"
	ErrorCodeInvalidResponse         = "invalid_response"
	ErrorCodeProviderRejectedRequest = "provider_rejected_request"
	ErrorCodeRequestFailed           = "request_failed"
)

// UserFacingError is safe to expose over the HTTP API. It deliberately keeps
// provider response bodies and request credentials out of user-visible text.
type UserFacingError struct {
	Code       string
	Message    string
	HTTPStatus int
}

// RedactEndpoint keeps only the non-sensitive origin and path for diagnostic
// logs. Userinfo, query parameters, and fragments may contain API credentials.
func RedactEndpoint(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "configured endpoint"
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

// ClassifyUserFacingError converts provider, transport, and response parsing
// failures into a stable public code and a short, non-sensitive message.
func ClassifyUserFacingError(err error) UserFacingError {
	if err == nil {
		return userFacingError(ErrorCodeRequestFailed)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return userFacingError(ErrorCodeTimeout)
	}
	if errors.Is(err, context.Canceled) {
		return userFacingError(ErrorCodeRequestFailed)
	}

	message := strings.ToLower(err.Error())
	if statusCode := providerStatusCode(message); statusCode > 0 {
		switch {
		case statusCode == http.StatusTooManyRequests:
			return userFacingError(ErrorCodeRateLimited)
		case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
			return userFacingError(ErrorCodeAuthenticationFailed)
		case statusCode == http.StatusPaymentRequired:
			return userFacingError(ErrorCodePaymentRequired)
		case statusCode == http.StatusNotFound:
			return userFacingError(ErrorCodeModelOrEndpointNotFound)
		case statusCode == http.StatusRequestEntityTooLarge:
			return userFacingError(ErrorCodeRequestTooLarge)
		case statusCode >= http.StatusInternalServerError:
			return userFacingError(ErrorCodeProviderUnavailable)
		case statusCode >= http.StatusBadRequest:
			return userFacingError(ErrorCodeProviderRejectedRequest)
		}
	}

	var networkErr net.Error
	if errors.As(err, &networkErr) {
		if networkErr.Timeout() {
			return userFacingError(ErrorCodeTimeout)
		}
		return userFacingError(ErrorCodeNetwork)
	}

	switch {
	case strings.Contains(message, "usage limit"), strings.Contains(message, "token limit"):
		return userFacingError(ErrorCodeUsageLimitReached)
	case strings.Contains(message, "timeout"), strings.Contains(message, "deadline exceeded"):
		return userFacingError(ErrorCodeTimeout)
	case strings.Contains(message, "connection refused"), strings.Contains(message, "no such host"), strings.Contains(message, "network is unreachable"):
		return userFacingError(ErrorCodeNetwork)
	case strings.Contains(message, "invalid json"), strings.Contains(message, "decode"), strings.Contains(message, "empty response"), strings.Contains(message, "empty content"), strings.Contains(message, "no choices"):
		return userFacingError(ErrorCodeInvalidResponse)
	default:
		return userFacingError(ErrorCodeRequestFailed)
	}
}

var providerStatusPattern = regexp.MustCompile(`(?i)(?:status|http)[^0-9]{0,12}([1-5][0-9]{2})`)

func providerStatusCode(message string) int {
	match := providerStatusPattern.FindStringSubmatch(message)
	if len(match) != 2 {
		return 0
	}
	statusCode, _ := strconv.Atoi(match[1])
	return statusCode
}

// UserFacingErrorForCode returns the canonical safe message and HTTP status
// for an explicitly detected application error.
func UserFacingErrorForCode(code string) UserFacingError {
	return userFacingError(code)
}

func userFacingError(code string) UserFacingError {
	switch code {
	case ErrorCodeConfigurationInvalid:
		return UserFacingError{Code: code, Message: "AI configuration is incomplete or invalid", HTTPStatus: http.StatusBadRequest}
	case ErrorCodeUsageLimitReached:
		return UserFacingError{Code: code, Message: "The configured AI usage limit has been reached", HTTPStatus: http.StatusTooManyRequests}
	case ErrorCodeRateLimited:
		return UserFacingError{Code: code, Message: "The AI service is receiving too many requests", HTTPStatus: http.StatusTooManyRequests}
	case ErrorCodeAuthenticationFailed:
		return UserFacingError{Code: code, Message: "AI service authentication failed", HTTPStatus: http.StatusUnauthorized}
	case ErrorCodePaymentRequired:
		return UserFacingError{Code: code, Message: "The AI service account has insufficient quota or balance", HTTPStatus: http.StatusPaymentRequired}
	case ErrorCodeModelOrEndpointNotFound:
		return UserFacingError{Code: code, Message: "The configured AI model or endpoint was not found", HTTPStatus: http.StatusNotFound}
	case ErrorCodeRequestTooLarge:
		return UserFacingError{Code: code, Message: "The AI request is too large for the configured service", HTTPStatus: http.StatusRequestEntityTooLarge}
	case ErrorCodeTimeout:
		return UserFacingError{Code: code, Message: "The AI service response timed out", HTTPStatus: http.StatusGatewayTimeout}
	case ErrorCodeNetwork:
		return UserFacingError{Code: code, Message: "MRSS could not connect to the AI service", HTTPStatus: http.StatusBadGateway}
	case ErrorCodeProviderUnavailable:
		return UserFacingError{Code: code, Message: "The AI service is temporarily unavailable", HTTPStatus: http.StatusBadGateway}
	case ErrorCodeInvalidResponse:
		return UserFacingError{Code: code, Message: "The AI service returned an invalid response", HTTPStatus: http.StatusBadGateway}
	case ErrorCodeProviderRejectedRequest:
		return UserFacingError{Code: code, Message: "The AI service rejected the request", HTTPStatus: http.StatusBadRequest}
	default:
		return UserFacingError{Code: ErrorCodeRequestFailed, Message: "The AI request failed", HTTPStatus: http.StatusBadGateway}
	}
}
