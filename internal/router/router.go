package router

import (
	"net/http"
	"openbridge/internal/config"
	"openbridge/internal/handler"
	"openbridge/internal/provider"
	"openbridge/internal/service"
	"openbridge/internal/user"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, registry *provider.Registry, keyManagers *service.ProviderKeyManagers) *gin.Engine {
	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Health check (no auth required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "openbridge",
		})
	})

	// Version endpoint (no auth required)
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":     "2.0.0",
			"description": "Universal LLM API Gateway",
		})
	})

	// Stats endpoint (no auth required)
	r.GET("/stats", func(c *gin.Context) {
		stats := keyManagers.GetStats()
		c.JSON(http.StatusOK, gin.H{
			"providers": stats,
		})
	})

	// Providers endpoint (no auth required)
	r.GET("/providers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"providers": registry.ListProviders(),
		})
	})

	// Initialize handlers
	chatHandler := handler.NewChatHandler(cfg, registry, keyManagers)
	modelsHandler := handler.NewModelsHandler(cfg, registry, keyManagers)

	// OpenAI compatible endpoints (with auth)
	v1 := r.Group("/v1")
	v1.Use(authMiddleware(cfg.ClientAPIKeys))
	{
		// Chat completions
		v1.POST("/chat/completions", chatHandler.CreateChatCompletion)

		// Models
		v1.GET("/models", modelsHandler.ListModels)
		v1.GET("/models/:model", modelsHandler.RetrieveModel)
	}

	return r
}

// authMiddleware API Key 认证中间件（支持配置文件 Key 和用户 Key）
func authMiddleware(validKeys []string) gin.HandlerFunc {
	keySet := make(map[string]bool)
	for _, key := range validKeys {
		keySet[key] = true
	}

	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Missing Authorization header",
					"type":    "authentication_error",
					"code":    "invalid_api_key",
				},
			})
			c.Abort()
			return
		}

		// 支持 "Bearer sk-xxx" 格式
		key := auth
		if len(auth) > 7 && auth[:7] == "Bearer " {
			key = auth[7:]
		}

		// 1. 检查是否是配置文件中的客户端 Key
		if keySet[key] {
			c.Set("auth_type", "admin")
			c.Next()
			return
		}

		// 2. 检查是否是用户系统中的 Key
		userStore := user.GetStore()
		if userStore != nil {
			if username, valid := userStore.ValidateAPIKey(key); valid {
				c.Set("auth_type", "user")
				c.Set("username", username)
				// 记录 API Key 使用
				go userStore.RecordKeyUsage(key)
				c.Next()
				return
			}
		}

		// 3. 所有验证都失败
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"message": "Invalid API key",
				"type":    "authentication_error",
				"code":    "invalid_api_key",
			},
		})
		c.Abort()
	}
}
