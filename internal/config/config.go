package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        ServerConfig     `yaml:"server"`
	ClientAPIKeys []string         `yaml:"client_api_keys"`
	AssemblyAI    AssemblyAIConfig `yaml:"assemblyai"`
	Logging       LoggingConfig    `yaml:"logging"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type AssemblyAIConfig struct {
	BaseURL          string            `yaml:"base_url"`
	APIKeys          []string          `yaml:"api_keys"`
	RotationStrategy string            `yaml:"rotation_strategy"`
	Features         FeaturesConfig    `yaml:"features"`
	AutoConvert      AutoConvertConfig `yaml:"auto_convert"`

	// 保留向后兼容 (deprecated)
	SupportStream bool `yaml:"support_stream"`
}

type FeaturesConfig struct {
	Stream            bool     `yaml:"stream"`
	Vision            bool     `yaml:"vision"`
	Tools             bool     `yaml:"tools"`
	JSONMode          bool     `yaml:"json_mode"`
	Logprobs          bool     `yaml:"logprobs"`
	MultipleChoices   bool     `yaml:"multiple_choices"`
	SystemFingerprint bool     `yaml:"system_fingerprint"`
	UnsupportedParams []string `yaml:"unsupported_params"` // 不支持的参数列表
}

type AutoConvertConfig struct {
	StreamToFake      bool `yaml:"stream_to_fake"`
	StripUnsupported  bool `yaml:"strip_unsupported"`
	WarnOnUnsupported bool `yaml:"warn_on_unsupported"`
	RejectUnsupported bool `yaml:"reject_unsupported"`
}

type LoggingConfig struct {
	Level        string `yaml:"level"`
	Format       string `yaml:"format"`
	LogRequests  bool   `yaml:"log_requests"`
	LogResponses bool   `yaml:"log_responses"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.AssemblyAI.BaseURL == "" {
		cfg.AssemblyAI.BaseURL = "https://llm-gateway.assemblyai.com/v1"
	}
	if cfg.AssemblyAI.RotationStrategy == "" {
		cfg.AssemblyAI.RotationStrategy = "round_robin"
	}

	// 向后兼容: 如果使用旧的 support_stream 配置
	if cfg.AssemblyAI.SupportStream {
		cfg.AssemblyAI.Features.Stream = true
	}

	// AutoConvert 默认值
	if !cfg.AssemblyAI.AutoConvert.StreamToFake && !cfg.AssemblyAI.Features.Stream {
		// 如果后端不支持流式,默认启用假流式转换
		cfg.AssemblyAI.AutoConvert.StreamToFake = true
	}

	return &cfg, nil
}
