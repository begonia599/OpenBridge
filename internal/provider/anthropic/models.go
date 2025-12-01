package anthropic

// Claude API 原生格式定义

// ChatRequest Claude API 聊天请求
type ChatRequest struct {
	Model         string    `json:"model"`
	Messages      []Message `json:"messages"`
	MaxTokens     int       `json:"max_tokens"`
	Temperature   float64   `json:"temperature,omitempty"`
	TopP          float64   `json:"top_p,omitempty"`
	TopK          int       `json:"top_k,omitempty"`
	Stream        bool      `json:"stream,omitempty"`
	StopSequences []string  `json:"stop_sequences,omitempty"`
	System        string    `json:"system,omitempty"`
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // 可以是 string 或 ContentBlock 数组
}

type ContentBlock struct {
	Type   string      `json:"type"`
	Text   string      `json:"text,omitempty"`
	Source *ImageSource `json:"source,omitempty"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// ChatResponse Claude API 聊天响应
type ChatResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// StreamEvent Claude API 流式事件
type StreamEvent struct {
	Type         string         `json:"type"`
	Message      *ChatResponse  `json:"message,omitempty"`
	Index        int            `json:"index,omitempty"`
	ContentBlock *ContentBlock  `json:"content_block,omitempty"`
	Delta        *StreamDelta   `json:"delta,omitempty"`
	Usage        *Usage         `json:"usage,omitempty"`
}

type StreamDelta struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

// ErrorResponse Claude API 错误响应
type ErrorResponse struct {
	Type  string `json:"type"`
	Error Error  `json:"error"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

