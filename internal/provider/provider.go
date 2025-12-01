package provider

import (
	"openbridge/internal/models"
)

// Provider 定义了 LLM 提供商的接口
type Provider interface {
	// Name 返回提供商名称
	Name() string

	// Type 返回提供商类型 (openai, anthropic, google)
	Type() string

	// ChatCompletion 发送聊天请求
	ChatCompletion(req *models.ChatCompletionRequest, apiKey string) (*models.ChatCompletionResponse, error)

	// ChatCompletionStream 发送流式聊天请求
	ChatCompletionStream(req *models.ChatCompletionRequest, apiKey string) (<-chan *models.ChatCompletionChunk, <-chan error)

	// ListModels 获取模型列表
	ListModels(apiKey string) (*models.ModelList, error)

	// SupportsStreaming 是否支持流式
	SupportsStreaming() bool
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`     // openai, anthropic, google
	BaseURL  string   `yaml:"base_url"`
	APIKeys  []string `yaml:"api_keys"`
	Models   []string `yaml:"models"`   // 该提供商支持的模型列表
}
