package google

// Google Gemini API 原生格式定义

// GenerateContentRequest Gemini API 请求
type GenerateContentRequest struct {
	Contents         []Content               `json:"contents"`
	SystemInstruction *Content               `json:"systemInstruction,omitempty"`
	GenerationConfig *GenerationConfig       `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting         `json:"safetySettings,omitempty"`
}

type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 编码
}

type GenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GenerateContentResponse Gemini API 响应
type GenerateContentResponse struct {
	Candidates     []Candidate    `json:"candidates"`
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *UsageMetadata `json:"usageMetadata,omitempty"`
}

type Candidate struct {
	Content       Content        `json:"content"`
	FinishReason  string         `json:"finishReason"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type PromptFeedback struct {
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
	BlockReason   string         `json:"blockReason,omitempty"`
}

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// StreamResponse Gemini 流式响应（与非流式相同）
type StreamResponse = GenerateContentResponse

// ErrorResponse Gemini API 错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ModelsListResponse 模型列表响应
type ModelsListResponse struct {
	Models        []ModelInfo `json:"models"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}

type ModelInfo struct {
	Name                       string   `json:"name"`
	BaseModelID               string   `json:"baseModelId,omitempty"`
	Version                    string   `json:"version"`
	DisplayName                string   `json:"displayName"`
	Description                string   `json:"description"`
	InputTokenLimit           int      `json:"inputTokenLimit"`
	OutputTokenLimit          int      `json:"outputTokenLimit"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	Temperature                float64  `json:"temperature,omitempty"`
	TopP                       float64  `json:"topP,omitempty"`
	TopK                       int      `json:"topK,omitempty"`
}

