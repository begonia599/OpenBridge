package models

// AssemblyAI Response Models

type AssemblyAIResponse struct {
	RequestID      string             `json:"request_id"`
	Choices        []AssemblyAIChoice `json:"choices"`
	Request        AssemblyAIRequest  `json:"request,omitempty"`
	Usage          AssemblyAIUsage    `json:"usage"`
	HTTPStatusCode int                `json:"http_status_code,omitempty"`
	ResponseTime   int64              `json:"response_time,omitempty"`
	LLMStatusCode  int                `json:"llm_status_code,omitempty"`
}

type AssemblyAIChoice struct {
	Message      AssemblyAIMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type AssemblyAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AssemblyAIRequest struct {
	Model       string  `json:"model,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type AssemblyAIUsage struct {
	InputTokens      int `json:"input_tokens"`
	OutputTokens     int `json:"output_tokens"`
	TotalTokens      int `json:"total_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}
