package router

import (
	"net/http"
	"openbridge/internal/config"
	"openbridge/internal/handler"
	"openbridge/internal/middleware"
	"openbridge/internal/service"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, apiKeyManager *service.APIKeyManager) *gin.Engine {
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
			"version":     "1.0.1",
			"build_date":  "2025-11-24",
			"description": "OpenAI-compatible API Gateway for AssemblyAI",
		})
	})

	// Stats endpoint (no auth required)
	r.GET("/stats", func(c *gin.Context) {
		stats := apiKeyManager.GetStats()
		c.JSON(http.StatusOK, gin.H{
			"api_key_usage": stats,
			"strategy":      cfg.AssemblyAI.RotationStrategy,
		})
	})

	// Initialize handlers
	chatHandler := handler.NewChatHandler(cfg, apiKeyManager)
	modelsHandler := handler.NewModelsHandler(cfg, apiKeyManager)

	// OpenAI compatible endpoints (with auth)
	v1 := r.Group("/v1")
	v1.Use(middleware.AuthMiddleware(cfg.ClientAPIKeys))
	{
		// Chat completions
		v1.POST("/chat/completions", chatHandler.CreateChatCompletion)

		// Models
		v1.GET("/models", modelsHandler.ListModels)
		v1.GET("/models/:model", modelsHandler.RetrieveModel)
	}

	return r
}
