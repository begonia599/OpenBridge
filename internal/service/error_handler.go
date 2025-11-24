package service

import (
	"encoding/json"
	"log"
	"openbridge/internal/models"
	"strings"
)

// ParseAssemblyAIError attempts to parse and convert AssemblyAI errors to OpenAI format
func ParseAssemblyAIError(statusCode int, body string) *models.ErrorResponse {
	log.Printf("🔍 Parsing error response (status %d): %s", statusCode, body)

	var message string

	// Try to parse as JSON
	var rawError map[string]interface{}
	if err := json.Unmarshal([]byte(body), &rawError); err != nil {
		// Not JSON, use plain text as message
		message = body
	} else {
		// Extract error message from JSON
		message = extractErrorMessage(rawError)
	}

	// Determine error type and code based on status code and message
	errorType, errorCode := categorizeError(statusCode, message)

	return models.NewErrorResponse(message, errorType, errorCode)
}

// extractErrorMessage tries to find error message in various formats
func extractErrorMessage(rawError map[string]interface{}) string {
	// Try common error message fields
	if msg, ok := rawError["message"].(string); ok && msg != "" {
		return msg
	}
	if msg, ok := rawError["error"].(string); ok && msg != "" {
		return msg
	}
	if errorObj, ok := rawError["error"].(map[string]interface{}); ok {
		if msg, ok := errorObj["message"].(string); ok && msg != "" {
			return msg
		}
	}
	if msg, ok := rawError["detail"].(string); ok && msg != "" {
		return msg
	}

	// Fallback to JSON string
	data, _ := json.Marshal(rawError)
	return string(data)
}

// categorizeError determines OpenAI error type and code based on status and message
func categorizeError(statusCode int, message string) (errorType, errorCode string) {
	messageLower := strings.ToLower(message)

	switch statusCode {
	case 400:
		// Bad Request
		if strings.Contains(messageLower, "context") || strings.Contains(messageLower, "token") {
			return models.ErrorTypeInvalidRequest, models.ErrorCodeContextLengthExceeded
		}
		if strings.Contains(messageLower, "model") {
			return models.ErrorTypeInvalidRequest, models.ErrorCodeModelNotFound
		}
		return models.ErrorTypeInvalidRequest, models.ErrorCodeInvalidRequest

	case 401:
		// Unauthorized
		return models.ErrorTypeAuthentication, models.ErrorCodeInvalidAPIKey

	case 403:
		// Forbidden
		return models.ErrorTypePermission, "permission_denied"

	case 404:
		// Not Found
		return models.ErrorTypeNotFound, models.ErrorCodeModelNotFound

	case 429:
		// Rate Limit
		if strings.Contains(messageLower, "quota") {
			return models.ErrorTypeRateLimit, models.ErrorCodeQuotaExceeded
		}
		return models.ErrorTypeRateLimit, models.ErrorCodeRateLimitExceeded

	case 500, 502, 503:
		// Server Errors
		if statusCode == 503 {
			return models.ErrorTypeServiceUnavailable, "service_unavailable"
		}
		return models.ErrorTypeServerError, models.ErrorCodeServerError

	case 504:
		// Gateway Timeout
		return models.ErrorTypeTimeout, "timeout"

	default:
		// Unknown error
		return models.ErrorTypeAPIError, models.ErrorCodeServerError
	}
}

// IsRetryableError checks if an error should trigger a retry with another API key
func IsRetryableError(statusCode int) bool {
	switch statusCode {
	case 429: // Rate limit
		return true
	case 500, 502, 503, 504: // Server errors
		return true
	default:
		return false
	}
}
