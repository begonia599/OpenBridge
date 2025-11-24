package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates client API keys
func AuthMiddleware(validKeys []string) gin.HandlerFunc {
	// Create a map for O(1) lookup
	keyMap := make(map[string]bool)
	for _, key := range validKeys {
		keyMap[key] = true
	}

	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Missing Authorization header",
					"type":    "invalid_request_error",
					"code":    "missing_api_key",
				},
			})
			c.Abort()
			return
		}

		// Extract token (Bearer token)
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			// No "Bearer " prefix found
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Invalid Authorization header format. Expected: Bearer <token>",
					"type":    "invalid_request_error",
					"code":    "invalid_api_key_format",
				},
			})
			c.Abort()
			return
		}

		// Validate token
		if !keyMap[token] {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Invalid API key",
					"type":    "invalid_request_error",
					"code":    "invalid_api_key",
				},
			})
			c.Abort()
			return
		}

		// Token is valid, continue
		c.Next()
	}
}
