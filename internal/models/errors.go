package models

// OpenAI Standard Error Response

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// Common error types
const (
	ErrorTypeInvalidRequest     = "invalid_request_error"
	ErrorTypeAuthentication     = "authentication_error"
	ErrorTypePermission         = "permission_error"
	ErrorTypeNotFound           = "not_found_error"
	ErrorTypeRateLimit          = "rate_limit_error"
	ErrorTypeAPIError           = "api_error"
	ErrorTypeTimeout            = "timeout_error"
	ErrorTypeServerError        = "server_error"
	ErrorTypeServiceUnavailable = "service_unavailable_error"
)

// Common error codes
const (
	ErrorCodeInvalidAPIKey         = "invalid_api_key"
	ErrorCodeRateLimitExceeded     = "rate_limit_exceeded"
	ErrorCodeQuotaExceeded         = "quota_exceeded"
	ErrorCodeModelNotFound         = "model_not_found"
	ErrorCodeContextLengthExceeded = "context_length_exceeded"
	ErrorCodeInvalidRequest        = "invalid_request"
	ErrorCodeServerError           = "server_error"
)

// NewErrorResponse creates a standard OpenAI error response
func NewErrorResponse(message, errorType, code string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetail{
			Message: message,
			Type:    errorType,
			Code:    code,
		},
	}
}
