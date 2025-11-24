package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"openbridge/internal/config"
	"openbridge/internal/models"
	"openbridge/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	config           *config.Config
	apiKeyManager    *service.APIKeyManager
	featureValidator *service.FeatureValidator
}

func NewChatHandler(cfg *config.Config, apiKeyManager *service.APIKeyManager) *ChatHandler {
	return &ChatHandler{
		config:           cfg,
		apiKeyManager:    apiKeyManager,
		featureValidator: service.NewFeatureValidator(cfg),
	}
}

func (h *ChatHandler) CreateChatCompletion(c *gin.Context) {
	var req models.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log client request
	if h.config.Logging.LogRequests {
		reqJSON, _ := json.MarshalIndent(req, "", "  ")
		log.Printf("📥 Client Request:\n%s", string(reqJSON))
	}

	// 验证功能支持
	if err := h.featureValidator.ValidateRequest(&req); err != nil {
		if featureErr, ok := err.(*service.FeatureNotSupportedError); ok {
			log.Printf("❌ Feature not supported: %s", featureErr.Feature)
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				"Feature not supported: "+featureErr.Feature,
				models.ErrorTypeInvalidRequest,
				models.ErrorCodeInvalidRequest,
			))
			return
		}
		log.Printf("❌ Validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 记录客户端是否请求流式
	clientWantsStream := req.Stream

	// 如果需要,转换流式请求
	if h.featureValidator.ShouldConvertToFakeStream(clientWantsStream) {
		log.Printf("🔄 Converting stream request: client wants stream, backend doesn't support")
		req.Stream = false
	}

	// Get next API key using rotation strategy
	apiKey := h.apiKeyManager.GetNextKey()
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No API keys configured",
		})
		return
	}

	// Mask API key for logging
	maskedKey := apiKey
	if len(apiKey) > 8 {
		maskedKey = apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
	}
	log.Printf("🔑 Using API Key: %s", maskedKey)

	// Forward request to AssemblyAI
	aaiResp, err := h.forwardToAssemblyAI(&req, apiKey)
	if err != nil {
		log.Printf("❌ Error forwarding to AssemblyAI: %v", err)

		// If it's an API error, convert to OpenAI format
		if apiErr, ok := err.(*APIError); ok {
			errorResp := service.ParseAssemblyAIError(apiErr.StatusCode, apiErr.Message)
			c.JSON(apiErr.StatusCode, errorResp)
			return
		}

		// Generic error
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			err.Error(),
			models.ErrorTypeAPIError,
			models.ErrorCodeServerError,
		))
		return
	}

	// Log backend response
	if h.config.Logging.LogResponses {
		respJSON, _ := json.MarshalIndent(aaiResp, "", "  ")
		log.Printf("📤 Backend Response:\n%s", string(respJSON))
	}

	// Convert to OpenAI format
	openAIResp := service.ConvertToOpenAIResponse(aaiResp, req.Model)

	// If client wants stream but backend doesn't support it, convert to SSE format
	if h.featureValidator.ShouldConvertToFakeStream(clientWantsStream) {
		log.Printf("🔄 Converting non-stream response to stream format for client")
		h.sendAsStream(c, openAIResp)
		return
	}

	// Log final response
	if h.config.Logging.LogResponses {
		finalJSON, _ := json.MarshalIndent(openAIResp, "", "  ")
		log.Printf("✅ Final Response to Client:\n%s", string(finalJSON))
	}

	c.JSON(http.StatusOK, openAIResp)
}

func (h *ChatHandler) forwardToAssemblyAI(req *models.ChatCompletionRequest, apiKey string) (*models.AssemblyAIResponse, error) {
	// Prepare request body - manually build to exclude unsupported fields
	reqMap := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
	}

	// Only add non-zero/non-nil values
	if req.MaxTokens > 0 {
		reqMap["max_tokens"] = req.MaxTokens
	}

	// Add parameters only if not in unsupported list
	unsupportedParams := make(map[string]bool)
	for _, param := range h.config.AssemblyAI.Features.UnsupportedParams {
		unsupportedParams[param] = true
	}

	if req.Temperature != 0 && !unsupportedParams["temperature"] {
		reqMap["temperature"] = req.Temperature
	}
	if req.TopP != 0 && !unsupportedParams["top_p"] {
		reqMap["top_p"] = req.TopP
	}
	if req.PresencePenalty != 0 && !unsupportedParams["presence_penalty"] {
		reqMap["presence_penalty"] = req.PresencePenalty
	}
	if req.FrequencyPenalty != 0 && !unsupportedParams["frequency_penalty"] {
		reqMap["frequency_penalty"] = req.FrequencyPenalty
	}
	if req.Stream {
		reqMap["stream"] = req.Stream
	}
	if req.StreamOptions != nil {
		reqMap["stream_options"] = req.StreamOptions
	}
	if len(req.Tools) > 0 {
		reqMap["tools"] = req.Tools
	} else if !h.config.AssemblyAI.Features.Tools {
		// Explicitly disable tools if not supported
		reqMap["tools"] = []interface{}{}
	}
	if req.ToolChoice != nil {
		reqMap["tool_choice"] = req.ToolChoice
	}
	if req.ResponseFormat != nil {
		reqMap["response_format"] = req.ResponseFormat
	}
	if req.N > 0 {
		reqMap["n"] = req.N
	}
	if req.Logprobs {
		reqMap["logprobs"] = req.Logprobs
	}
	if req.TopLogprobs > 0 {
		reqMap["top_logprobs"] = req.TopLogprobs
	}

	reqBody, err := json.Marshal(reqMap)
	if err != nil {
		return nil, err
	}

	// Log request to backend
	log.Printf("🚀 Forwarding to AssemblyAI: %s", h.config.AssemblyAI.BaseURL+"/chat/completions")
	if h.config.Logging.LogRequests {
		log.Printf("📦 Request Body to Backend:\n%s", string(reqBody))
	}

	// Create HTTP request
	url := h.config.AssemblyAI.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("📊 Backend Status Code: %d", resp.StatusCode)

	// Log rate limit headers if present
	if limit := resp.Header.Get("X-Ratelimit-Limit"); limit != "" {
		remaining := resp.Header.Get("X-Ratelimit-Remaining")
		reset := resp.Header.Get("X-Ratelimit-Reset")
		log.Printf("⏱️  Rate Limit: %s requests, %s remaining, reset in %ss", limit, remaining, reset)
	}
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		log.Printf("⏳ Retry-After: %s seconds", retryAfter)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ AssemblyAI error response (Status %d):\n%s", resp.StatusCode, string(body))

		// Check if error is retryable (rate limit, server error)
		if service.IsRetryableError(resp.StatusCode) {
			log.Printf("⚠️  Error is retryable, consider implementing retry logic")
		}

		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if h.config.Logging.LogResponses {
		log.Printf("📥 Raw Backend Response:\n%s", string(body))
	}

	// Parse response
	var aaiResp models.AssemblyAIResponse
	if err := json.Unmarshal(body, &aaiResp); err != nil {
		log.Printf("❌ Failed to parse backend response: %v", err)
		return nil, err
	}

	return &aaiResp, nil
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

// sendAsStream converts a non-stream response to SSE (Server-Sent Events) format
func (h *ChatHandler) sendAsStream(c *gin.Context, resp *models.ChatCompletionResponse) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		log.Printf("❌ Streaming not supported")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Send each chunk in SSE format
	for i, choice := range resp.Choices {
		// Create stream chunk
		streamChunk := map[string]interface{}{
			"id":      resp.ID,
			"object":  "chat.completion.chunk",
			"created": resp.Created,
			"model":   resp.Model,
			"choices": []map[string]interface{}{
				{
					"index": i,
					"delta": map[string]interface{}{
						"role":    choice.Message.Role,
						"content": choice.Message.Content,
					},
					"finish_reason": nil,
				},
			},
		}

		data, err := json.Marshal(streamChunk)
		if err != nil {
			log.Printf("❌ Error marshaling stream chunk: %v", err)
			return
		}

		// Write SSE format: "data: {json}\n\n"
		c.Writer.Write([]byte("data: "))
		c.Writer.Write(data)
		c.Writer.Write([]byte("\n\n"))
		flusher.Flush()

		// Send finish chunk
		finishChunk := map[string]interface{}{
			"id":      resp.ID,
			"object":  "chat.completion.chunk",
			"created": resp.Created,
			"model":   resp.Model,
			"choices": []map[string]interface{}{
				{
					"index":         i,
					"delta":         map[string]interface{}{},
					"finish_reason": choice.FinishReason,
				},
			},
		}

		data, err = json.Marshal(finishChunk)
		if err != nil {
			log.Printf("❌ Error marshaling finish chunk: %v", err)
			return
		}

		c.Writer.Write([]byte("data: "))
		c.Writer.Write(data)
		c.Writer.Write([]byte("\n\n"))
		flusher.Flush()
	}

	// Send usage chunk (optional but recommended)
	usageChunk := map[string]interface{}{
		"id":      resp.ID,
		"object":  "chat.completion.chunk",
		"created": resp.Created,
		"model":   resp.Model,
		"choices": []map[string]interface{}{},
		"usage":   resp.Usage,
	}

	data, err := json.Marshal(usageChunk)
	if err == nil {
		c.Writer.Write([]byte("data: "))
		c.Writer.Write(data)
		c.Writer.Write([]byte("\n\n"))
		flusher.Flush()
	}

	// Send [DONE] signal
	c.Writer.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()

	log.Printf("✅ Stream response sent successfully")
}
