package service

import (
	"log"
	"openbridge/internal/config"
	"openbridge/internal/models"
)

type FeatureValidator struct {
	config *config.Config
}

func NewFeatureValidator(cfg *config.Config) *FeatureValidator {
	return &FeatureValidator{config: cfg}
}

// ValidateRequest 验证请求中的功能是否支持
func (v *FeatureValidator) ValidateRequest(req *models.ChatCompletionRequest) error {
	features := v.config.AssemblyAI.Features
	autoConvert := v.config.AssemblyAI.AutoConvert

	// 检查流式
	if req.Stream && !features.Stream {
		if autoConvert.WarnOnUnsupported {
			log.Printf("⚠️  Stream requested but not supported by backend")
		}
		if autoConvert.StreamToFake {
			log.Printf("🔄 Will convert to fake streaming")
		} else if autoConvert.RejectUnsupported {
			return &FeatureNotSupportedError{Feature: "stream"}
		}
	}

	// 检查多模态/图片
	hasVision := v.hasVisionContent(req.Messages)
	if hasVision && !features.Vision {
		if autoConvert.WarnOnUnsupported {
			log.Printf("⚠️  Vision/images detected but not supported by backend")
		}
		if autoConvert.RejectUnsupported {
			return &FeatureNotSupportedError{Feature: "vision"}
		}
		// 否则继续,让后端决定如何处理
	}

	// 检查工具调用
	if len(req.Tools) > 0 && !features.Tools {
		if autoConvert.WarnOnUnsupported {
			log.Printf("⚠️  Tools/function calling requested but not supported")
		}
		if autoConvert.StripUnsupported {
			log.Printf("🔧 Stripping tools from request")
			req.Tools = nil
			req.ToolChoice = nil
		} else if autoConvert.RejectUnsupported {
			return &FeatureNotSupportedError{Feature: "tools"}
		}
	}

	// 检查 JSON 模式
	if req.ResponseFormat != nil && req.ResponseFormat.Type == "json_object" && !features.JSONMode {
		if autoConvert.WarnOnUnsupported {
			log.Printf("⚠️  JSON mode requested but not supported")
		}
		if autoConvert.StripUnsupported {
			log.Printf("🔧 Stripping response_format from request")
			req.ResponseFormat = nil
		} else if autoConvert.RejectUnsupported {
			return &FeatureNotSupportedError{Feature: "json_mode"}
		}
	}

	// 检查 logprobs
	if req.Logprobs && !features.Logprobs {
		if autoConvert.WarnOnUnsupported {
			log.Printf("⚠️  Logprobs requested but not supported")
		}
		if autoConvert.StripUnsupported {
			log.Printf("🔧 Stripping logprobs from request")
			req.Logprobs = false
			req.TopLogprobs = 0
		} else if autoConvert.RejectUnsupported {
			return &FeatureNotSupportedError{Feature: "logprobs"}
		}
	}

	// 检查多个选择
	if req.N > 1 && !features.MultipleChoices {
		if autoConvert.WarnOnUnsupported {
			log.Printf("⚠️  Multiple choices (n=%d) requested but not supported", req.N)
		}
		if autoConvert.StripUnsupported {
			log.Printf("🔧 Resetting n to 1")
			req.N = 1
		} else if autoConvert.RejectUnsupported {
			return &FeatureNotSupportedError{Feature: "multiple_choices"}
		}
	}

	// 检查并移除 AssemblyAI Claude 不支持的参数
	// 只有 temperature 不支持
	// top_p, presence_penalty, frequency_penalty 都支持
	// 注意: temperature 在 chat.go 中手动构建请求时已被排除

	return nil
}

// hasVisionContent 检查消息中是否包含图片
func (v *FeatureValidator) hasVisionContent(messages []models.Message) bool {
	for _, msg := range messages {
		if contentArray, ok := msg.Content.([]interface{}); ok {
			for _, part := range contentArray {
				if partMap, ok := part.(map[string]interface{}); ok {
					if partMap["type"] == "image_url" {
						return true
					}
				}
			}
		}
	}
	return false
}

// ShouldConvertToFakeStream 判断是否需要转换为假流式
func (v *FeatureValidator) ShouldConvertToFakeStream(clientWantsStream bool) bool {
	return clientWantsStream &&
		!v.config.AssemblyAI.Features.Stream &&
		v.config.AssemblyAI.AutoConvert.StreamToFake
}

// FeatureNotSupportedError 功能不支持错误
type FeatureNotSupportedError struct {
	Feature string
}

func (e *FeatureNotSupportedError) Error() string {
	return "Feature not supported: " + e.Feature
}
