package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"openbridge/internal/config"
	"openbridge/internal/models"
	"openbridge/internal/provider"
	"openbridge/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	config      *config.Config
	registry    *provider.Registry
	keyManagers *service.ProviderKeyManagers
}

func NewChatHandler(cfg *config.Config, registry *provider.Registry, keyManagers *service.ProviderKeyManagers) *ChatHandler {
	return &ChatHandler{
		config:      cfg,
		registry:    registry,
		keyManagers: keyManagers,
	}
}

func (h *ChatHandler) CreateChatCompletion(c *gin.Context) {
	var req models.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			err.Error(),
			models.ErrorTypeInvalidRequest,
			models.ErrorCodeInvalidRequest,
		))
		return
	}

	// Log client request
	if h.config.Logging.LogRequests {
		reqJSON, _ := json.MarshalIndent(req, "", "  ")
		log.Printf("📥 Client Request:\n%s", string(reqJSON))
	}

	// 根据 model 路由到对应的 Provider
	p, err := h.registry.RouteModel(req.Model)
	if err != nil {
		log.Printf("❌ No provider for model %s: %v", req.Model, err)
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			"Model not found: "+req.Model,
			models.ErrorTypeNotFound,
			models.ErrorCodeModelNotFound,
		))
		return
	}

	log.Printf("🔀 Routing model %s to provider: %s (%s)", req.Model, p.Name(), p.Type())

	// 获取 API Key
	apiKey := h.keyManagers.GetKey(p.Name())
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"No API keys configured for provider: "+p.Name(),
			models.ErrorTypeServerError,
			models.ErrorCodeServerError,
		))
		return
	}

	// Mask API key for logging
	maskedKey := apiKey
	if len(apiKey) > 8 {
		maskedKey = apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
	}
	log.Printf("🔑 Using API Key: %s", maskedKey)

	// 处理流式请求
	if req.Stream {
		h.handleStreamRequest(c, p, &req, apiKey)
		return
	}

	// 非流式请求
	resp, err := p.ChatCompletion(&req, apiKey)
	if err != nil {
		log.Printf("❌ Provider error: %v", err)
		h.handleProviderError(c, err)
		return
	}

	// Log response
	if h.config.Logging.LogResponses {
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		log.Printf("📤 Response:\n%s", string(respJSON))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) handleStreamRequest(c *gin.Context, p provider.Provider, req *models.ChatCompletionRequest, apiKey string) {
	// 设置 SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		log.Printf("❌ Streaming not supported")
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			"Streaming not supported",
			models.ErrorTypeServerError,
			models.ErrorCodeServerError,
		))
		return
	}

	chunkChan, errChan := p.ChatCompletionStream(req, apiKey)

	for {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				// Channel closed, send [DONE]
				c.Writer.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
				log.Printf("✅ Stream completed")
				return
			}

			data, err := json.Marshal(chunk)
			if err != nil {
				log.Printf("❌ Error marshaling chunk: %v", err)
				continue
			}

			c.Writer.Write([]byte("data: "))
			c.Writer.Write(data)
			c.Writer.Write([]byte("\n\n"))
			flusher.Flush()

		case err := <-errChan:
			if err != nil {
				log.Printf("❌ Stream error: %v", err)
				// 尝试发送错误信息
				errResp := models.NewErrorResponse(
					err.Error(),
					models.ErrorTypeServerError,
					models.ErrorCodeServerError,
				)
				data, _ := json.Marshal(errResp)
				c.Writer.Write([]byte("data: "))
				c.Writer.Write(data)
				c.Writer.Write([]byte("\n\n"))
				flusher.Flush()
			}
			return
		}
	}
}

func (h *ChatHandler) handleProviderError(c *gin.Context, err error) {
	// 尝试解析 API 错误
	statusCode := http.StatusInternalServerError
	message := err.Error()

	// 检查是否是 API 错误类型
	if apiErr, ok := err.(interface{ StatusCode() int }); ok {
		statusCode = apiErr.StatusCode()
	}

	c.JSON(statusCode, models.NewErrorResponse(
		message,
		models.ErrorTypeAPIError,
		models.ErrorCodeServerError,
	))
}
