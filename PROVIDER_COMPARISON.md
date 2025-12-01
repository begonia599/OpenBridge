# Provider API 对比与说明

本文档说明不同 LLM Provider 的 API 特性差异以及 OpenBridge 的处理方式。

## 📊 API 端点支持对比

| 功能 | OpenAI | Anthropic (Claude) | Google (Gemini) | 说明 |
|------|--------|-------------------|-----------------|------|
| **Chat Completions** | ✅ | ✅ | ✅ | 所有 provider 都支持 |
| **Streaming** | ✅ | ✅ | ✅ | 流式响应支持 |
| **List Models** | ✅ API | ❌ 无 API | ✅ API | Claude 无此端点 |
| **Retrieve Model** | ✅ | ❌ | ✅ | Claude 无此端点 |
| **Multi-modal** | ✅ | ✅ | ✅ | 图片等多模态支持 |

## 🔧 OpenBridge 处理方式

### 1. List Models 端点

**问题**: Anthropic (Claude) 官方 API 没有提供获取模型列表的端点。

**解决方案**: OpenBridge 为 Claude provider 返回**预定义的硬编码模型列表**。

```go
// internal/provider/anthropic/anthropic.go
func (p *Provider) ListModels(apiKey string) (*models.ModelList, error) {
    // Claude API 不提供模型列表端点，返回预定义的模型列表
    return &models.ModelList{
        Object: "list",
        Data: []models.Model{
            {ID: "claude-3-5-sonnet-20241022", ...},
            {ID: "claude-3-5-haiku-20241022", ...},
            // ... 更多模型
        },
    }, nil
}
```

**优点**:
- ✅ 对下游客户端透明，API 保持一致
- ✅ 无需额外的 API 调用
- ✅ 响应速度快

**注意事项**:
- ⚠️ 新模型发布时需要手动更新代码
- ⚠️ 无法获取实时的模型可用性信息

### 2. Google Gemini 的模型列表

Google 提供了模型列表 API，OpenBridge 会实时调用：

```go
// internal/provider/google/google.go
func (p *Provider) ListModels(apiKey string) (*models.ModelList, error) {
    url := fmt.Sprintf("%s/models?key=%s", p.baseURL, apiKey)
    // 实际调用 Google API 获取模型列表
    // ...
}
```

### 3. OpenAI 格式的 Provider

对于使用 OpenAI 兼容格式的 provider（如 DeepSeek、Moonshot 等），直接调用其 `/v1/models` 端点。

## 📝 API 格式差异

### 请求格式转换

#### OpenAI → Claude

| OpenAI 字段 | Claude 字段 | 说明 |
|------------|------------|------|
| `messages[].role=system` | `system` | System prompt 单独字段 |
| `messages[].role=user` | `messages[].role=user` | 保持一致 |
| `messages[].role=assistant` | `messages[].role=assistant` | 保持一致 |
| `max_tokens` | `max_tokens` | **Claude 必需此字段** |
| `temperature` | `temperature` | 保持一致 |
| `top_p` | `top_p` | 保持一致 |
| 不支持 | `top_k` | Claude 特有参数 |

**示例转换**:

```json
// OpenAI 格式输入
{
  "model": "claude-3-5-sonnet-20241022",
  "messages": [
    {"role": "system", "content": "You are helpful"},
    {"role": "user", "content": "Hello"}
  ]
}

// 转换为 Claude 格式
{
  "model": "claude-3-5-sonnet-20241022",
  "system": "You are helpful",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "max_tokens": 4096  // 自动添加默认值
}
```

#### OpenAI → Gemini

| OpenAI 字段 | Gemini 字段 | 说明 |
|------------|-------------|------|
| `messages[].role=system` | `systemInstruction` | System instruction |
| `messages[].role=user` | `contents[].role=user` | 保持 user |
| `messages[].role=assistant` | `contents[].role=model` | 改为 model |
| `max_tokens` | `generationConfig.maxOutputTokens` | 嵌套字段 |
| `temperature` | `generationConfig.temperature` | 嵌套字段 |
| `top_p` | `generationConfig.topP` | 嵌套字段 |

**示例转换**:

```json
// OpenAI 格式输入
{
  "model": "gemini-1.5-pro",
  "messages": [
    {"role": "system", "content": "You are helpful"},
    {"role": "user", "content": "Hello"}
  ],
  "temperature": 0.7
}

// 转换为 Gemini 格式
{
  "systemInstruction": {
    "parts": [{"text": "You are helpful"}]
  },
  "contents": [
    {
      "role": "user",
      "parts": [{"text": "Hello"}]
    }
  ],
  "generationConfig": {
    "temperature": 0.7
  },
  "safetySettings": [...]  // 自动添加
}
```

### 响应格式转换

#### Claude → OpenAI

```json
// Claude 原始响应
{
  "id": "msg_xxx",
  "type": "message",
  "role": "assistant",
  "content": [
    {"type": "text", "text": "Hello!"}
  ],
  "usage": {
    "input_tokens": 10,
    "output_tokens": 5
  }
}

// 转换为 OpenAI 格式
{
  "id": "msg_xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "claude-3-5-sonnet-20241022",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello!"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 5,
    "total_tokens": 15
  }
}
```

#### Gemini → OpenAI

```json
// Gemini 原始响应
{
  "candidates": [{
    "content": {
      "parts": [{"text": "Hello!"}],
      "role": "model"
    },
    "finishReason": "STOP"
  }],
  "usageMetadata": {
    "promptTokenCount": 10,
    "candidatesTokenCount": 5,
    "totalTokenCount": 15
  }
}

// 转换为 OpenAI 格式
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "gemini-1.5-pro",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello!"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 5,
    "total_tokens": 15
  }
}
```

## 🔄 流式响应处理

### Claude Streaming

Claude 使用 Server-Sent Events (SSE)，发送多个事件类型：

```
event: message_start
data: {"type": "message_start", ...}

event: content_block_start
data: {"type": "content_block_start", ...}

event: content_block_delta
data: {"type": "content_block_delta", "delta": {"text": "Hello"}}

event: content_block_stop
data: {"type": "content_block_stop", ...}

event: message_stop
data: {"type": "message_stop", ...}
```

OpenBridge 将这些事件转换为 OpenAI 格式的 chunks。

### Gemini Streaming

Gemini 也使用 SSE，格式相对简单：

```
data: {"candidates": [...], "usageMetadata": {...}}

data: {"candidates": [...]}
```

## 🎯 最佳实践

### 1. 模型选择建议

根据不同场景选择合适的 provider：

| 场景 | 推荐 Provider | 原因 |
|------|--------------|------|
| 长文本理解 | Claude | 200K token 上下文 |
| 代码生成 | GPT-4 / Claude | 代码能力强 |
| 快速响应 | Gemini Flash | 响应速度快，成本低 |
| 多模态 | GPT-4V / Gemini | 图片理解能力强 |
| 中文对话 | 所有 | 都有良好中文支持 |

### 2. 配置 Provider 的注意事项

**Claude (Anthropic)**:
```yaml
providers:
  claude:
    type: anthropic  # 或 claude
    base_url: ""     # 留空使用默认 API
    api_keys:
      - "sk-ant-xxx"  # Claude API key 格式
```

**Google Gemini**:
```yaml
providers:
  gemini:
    type: google     # 或 gemini
    base_url: ""     # 留空使用默认 API
    api_keys:
      - "AIzaSyxxx"  # Google API key 格式
```

### 3. 错误处理

不同 provider 的错误响应格式不同，OpenBridge 会统一转换：

- **Claude**: `{"type": "error", "error": {...}}`
- **Gemini**: `{"error": {"code": 400, "message": "..."}}`
- **OpenAI**: `{"error": {"message": "...", "type": "..."}}`

OpenBridge 统一返回 OpenAI 格式的错误响应。

## 📚 参考文档

- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)
- [Anthropic Claude API Reference](https://docs.anthropic.com/en/api)
- [Google Gemini API Reference](https://ai.google.dev/api/rest)

## 🔄 更新日志

- **2025-12-01**: 添加 Claude 和 Gemini 原生 API 支持
- **2025-12-01**: 实现完整的格式转换功能

