package service

import (
	"fmt"
	"openbridge/internal/models"
	"time"

	"github.com/google/uuid"
)

// ConvertToOpenAIResponse converts AssemblyAI response to OpenAI format
func ConvertToOpenAIResponse(aaiResp *models.AssemblyAIResponse, requestModel string) *models.ChatCompletionResponse {
	// Generate OpenAI-style completion ID
	completionID := fmt.Sprintf("chatcmpl-%s", uuid.New().String()[:24])

	// Convert choices
	choices := make([]models.Choice, len(aaiResp.Choices))
	for i, aaiChoice := range aaiResp.Choices {
		choices[i] = models.Choice{
			Index: i,
			Message: models.ResponseMessage{
				Role:    aaiChoice.Message.Role,
				Content: aaiChoice.Message.Content,
			},
			FinishReason: aaiChoice.FinishReason,
		}
	}

	// Convert usage - AssemblyAI uses input_tokens/output_tokens
	usage := models.Usage{
		PromptTokens:     aaiResp.Usage.InputTokens,
		CompletionTokens: aaiResp.Usage.OutputTokens,
		TotalTokens:      aaiResp.Usage.TotalTokens,
	}

	// Ensure total is calculated if not provided
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return &models.ChatCompletionResponse{
		ID:      completionID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   requestModel,
		Choices: choices,
		Usage:   usage,
	}
}
