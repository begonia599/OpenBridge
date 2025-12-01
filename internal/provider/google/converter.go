package google

import (
	"openbridge/internal/models"
	"strings"
)

// ConvertFromOpenAI 将 OpenAI 格式转换为 Gemini 格式
func ConvertFromOpenAI(req *models.ChatCompletionRequest) (*GenerateContentRequest, error) {
	geminiReq := &GenerateContentRequest{
		Contents: make([]Content, 0),
		GenerationConfig: &GenerationConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			MaxOutputTokens: req.MaxTokens,
		},
	}

	// 默认安全设置（较宽松）
	geminiReq.SafetySettings = []SafetySetting{
		{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_NONE"},
		{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_NONE"},
	}

	// 提取 system message
	var systemParts []Part
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			if text, ok := msg.Content.(string); ok {
				systemParts = append(systemParts, Part{Text: text})
			}
		}
	}
	if len(systemParts) > 0 {
		geminiReq.SystemInstruction = &Content{
			Parts: systemParts,
		}
	}

	// 转换 messages (跳过 system)
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			continue
		}

		// Gemini 使用 "user" 和 "model"
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		content := Content{
			Role:  role,
			Parts: make([]Part, 0),
		}

		// 处理 content
		switch v := msg.Content.(type) {
		case string:
			content.Parts = append(content.Parts, Part{Text: v})

		case []interface{}:
			// 多模态内容
			for _, part := range v {
				partMap, ok := part.(map[string]interface{})
				if !ok {
					continue
				}

				typeStr, _ := partMap["type"].(string)
				switch typeStr {
				case "text":
					if text, ok := partMap["text"].(string); ok {
						content.Parts = append(content.Parts, Part{Text: text})
					}

				case "image_url":
					// 转换图片 URL 为 base64
					imageURL, ok := partMap["image_url"].(map[string]interface{})
					if !ok {
						continue
					}
					url, _ := imageURL["url"].(string)

					// 如果是 data URI，提取 base64 数据
					if strings.HasPrefix(url, "data:") {
						parts := strings.SplitN(url, ",", 2)
						if len(parts) == 2 {
							mimeType := "image/png"
							if strings.Contains(parts[0], "image/jpeg") || strings.Contains(parts[0], "image/jpg") {
								mimeType = "image/jpeg"
							} else if strings.Contains(parts[0], "image/webp") {
								mimeType = "image/webp"
							} else if strings.Contains(parts[0], "image/png") {
								mimeType = "image/png"
							}

							content.Parts = append(content.Parts, Part{
								InlineData: &InlineData{
									MimeType: mimeType,
									Data:     parts[1],
								},
							})
						}
					}
				}
			}
		}

		if len(content.Parts) > 0 {
			geminiReq.Contents = append(geminiReq.Contents, content)
		}
	}

	return geminiReq, nil
}

// ConvertToOpenAI 将 Gemini 格式转换为 OpenAI 格式
func ConvertToOpenAI(resp *GenerateContentResponse, requestID string, requestModel string) *models.ChatCompletionResponse {
	openaiResp := &models.ChatCompletionResponse{
		ID:      requestID,
		Object:  "chat.completion",
		Created: 0,
		Model:   requestModel,
		Choices: make([]models.Choice, 0),
	}

	// 转换 candidates
	for _, candidate := range resp.Candidates {
		// 提取文本内容
		var content strings.Builder
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				content.WriteString(part.Text)
			}
		}

		choice := models.Choice{
			Index: candidate.Index,
			Message: models.ResponseMessage{
				Role:    "assistant",
				Content: content.String(),
			},
			FinishReason: convertFinishReason(candidate.FinishReason),
		}

		openaiResp.Choices = append(openaiResp.Choices, choice)
	}

	// 转换 usage
	if resp.UsageMetadata != nil {
		openaiResp.Usage = models.Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		}
	}

	return openaiResp
}

// ConvertStreamResponseToChunk 将 Gemini 流式响应转换为 OpenAI 流式块
func ConvertStreamResponseToChunk(resp *GenerateContentResponse, chunkID string, requestModel string, isFirst bool) *models.ChatCompletionChunk {
	chunk := &models.ChatCompletionChunk{
		ID:      chunkID,
		Object:  "chat.completion.chunk",
		Created: 0,
		Model:   requestModel,
		Choices: make([]models.ChunkChoice, 0),
	}

	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]

		// 提取文本内容
		var content strings.Builder
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				content.WriteString(part.Text)
			}
		}

		delta := models.ChunkDelta{
			Content: content.String(),
		}

		// 第一个 chunk 包含 role
		if isFirst {
			delta.Role = "assistant"
		}

		chunkChoice := models.ChunkChoice{
			Index: candidate.Index,
			Delta: delta,
		}

		// 如果有 finishReason
		if candidate.FinishReason != "" && candidate.FinishReason != "FINISH_REASON_UNSPECIFIED" {
			finishReason := convertFinishReason(candidate.FinishReason)
			chunkChoice.FinishReason = &finishReason
		}

		chunk.Choices = append(chunk.Choices, chunkChoice)
	}

	// 添加 usage 信息
	if resp.UsageMetadata != nil {
		chunk.Usage = &models.Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		}
	}

	return chunk
}

func convertFinishReason(geminiReason string) string {
	switch geminiReason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY":
		return "content_filter"
	case "RECITATION":
		return "content_filter"
	case "OTHER":
		return "stop"
	default:
		return "stop"
	}
}

