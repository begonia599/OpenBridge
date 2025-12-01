package anthropic

import (
	"openbridge/internal/models"
	"strings"
)

// ConvertFromOpenAI 将 OpenAI 格式转换为 Claude 格式
func ConvertFromOpenAI(req *models.ChatCompletionRequest) (*ChatRequest, error) {
	claudeReq := &ChatRequest{
		Model:       req.Model,
		Messages:    make([]Message, 0),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
	}

	// 默认 max_tokens (Claude 要求必须设置)
	if claudeReq.MaxTokens == 0 {
		claudeReq.MaxTokens = 4096
	}

	// 提取 system message
	var systemMessages []string
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			if text, ok := msg.Content.(string); ok {
				systemMessages = append(systemMessages, text)
			}
		}
	}
	if len(systemMessages) > 0 {
		claudeReq.System = strings.Join(systemMessages, "\n\n")
	}

	// 转换 messages (跳过 system)
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			continue
		}

		claudeMsg := Message{
			Role: msg.Role,
		}

		// 处理 content
		switch content := msg.Content.(type) {
		case string:
			claudeMsg.Content = content
		case []interface{}:
			// 多模态内容
			contentBlocks := make([]ContentBlock, 0)
			for _, part := range content {
				partMap, ok := part.(map[string]interface{})
				if !ok {
					continue
				}

				typeStr, _ := partMap["type"].(string)
				switch typeStr {
				case "text":
					if text, ok := partMap["text"].(string); ok {
						contentBlocks = append(contentBlocks, ContentBlock{
							Type: "text",
							Text: text,
						})
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
							mediaType := "image/png"
							if strings.Contains(parts[0], "image/jpeg") {
								mediaType = "image/jpeg"
							} else if strings.Contains(parts[0], "image/webp") {
								mediaType = "image/webp"
							} else if strings.Contains(parts[0], "image/gif") {
								mediaType = "image/gif"
							}
							
							contentBlocks = append(contentBlocks, ContentBlock{
								Type: "image",
								Source: &ImageSource{
									Type:      "base64",
									MediaType: mediaType,
									Data:      parts[1],
								},
							})
						}
					}
				}
			}
			if len(contentBlocks) > 0 {
				claudeMsg.Content = contentBlocks
			}
		}

		claudeReq.Messages = append(claudeReq.Messages, claudeMsg)
	}

	return claudeReq, nil
}

// ConvertToOpenAI 将 Claude 格式转换为 OpenAI 格式
func ConvertToOpenAI(resp *ChatResponse, requestModel string) *models.ChatCompletionResponse {
	// 提取文本内容
	var content string
	if len(resp.Content) > 0 {
		for _, block := range resp.Content {
			if block.Type == "text" {
				content += block.Text
			}
		}
	}

	return &models.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: 0, // Claude 不返回时间戳，可以用当前时间
		Model:   requestModel,
		Choices: []models.Choice{
			{
				Index: 0,
				Message: models.ResponseMessage{
					Role:    resp.Role,
					Content: content,
				},
				FinishReason: convertFinishReason(resp.StopReason),
			},
		},
		Usage: models.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}

// ConvertStreamEventToChunk 将 Claude 流式事件转换为 OpenAI 流式块
func ConvertStreamEventToChunk(event *StreamEvent, chunkID string, requestModel string) *models.ChatCompletionChunk {
	chunk := &models.ChatCompletionChunk{
		ID:      chunkID,
		Object:  "chat.completion.chunk",
		Created: 0,
		Model:   requestModel,
		Choices: []models.ChunkChoice{
			{
				Index: 0,
				Delta: models.ChunkDelta{},
			},
		},
	}

	switch event.Type {
	case "message_start":
		// 消息开始，发送 role
		chunk.Choices[0].Delta.Role = "assistant"

	case "content_block_delta":
		// 内容增量
		if event.Delta != nil && event.Delta.Type == "text_delta" {
			chunk.Choices[0].Delta.Content = event.Delta.Text
		}

	case "message_delta":
		// 消息结束，设置 finish_reason
		if event.Delta != nil && event.Delta.StopReason != "" {
			finishReason := convertFinishReason(event.Delta.StopReason)
			chunk.Choices[0].FinishReason = &finishReason
		}

	case "message_stop":
		// 流结束
		finishReason := "stop"
		chunk.Choices[0].FinishReason = &finishReason
		
		// 添加 usage 信息
		if event.Usage != nil {
			chunk.Usage = &models.Usage{
				PromptTokens:     event.Usage.InputTokens,
				CompletionTokens: event.Usage.OutputTokens,
				TotalTokens:      event.Usage.InputTokens + event.Usage.OutputTokens,
			}
		}
	}

	return chunk
}

func convertFinishReason(claudeReason string) string {
	switch claudeReason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "stop_sequence":
		return "stop"
	default:
		return "stop"
	}
}

