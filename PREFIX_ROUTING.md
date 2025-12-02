# 前缀路由机制说明

## 概述

OpenBridge 现在支持**前缀路由机制**，类似于 Docker Hub 的 `namespace/image:tag` 格式。这解决了多个 Provider 拥有相同模型名称的问题。

## 工作原理

### 模型列表返回格式

当调用 `/v1/models` 时，系统会从所有配置的 Provider 获取模型列表，并为每个模型添加 Provider 前缀：

```json
{
  "object": "list",
  "data": [
    {
      "id": "openai/gpt-4",
      "object": "model",
      "created": 1234567890,
      "owned_by": "openai"
    },
    {
      "id": "openai/gpt-3.5-turbo",
      "object": "model",
      "created": 1234567890,
      "owned_by": "openai"
    },
    {
      "id": "anthropic/claude-3-opus-20240229",
      "object": "model",
      "created": 1234567890,
      "owned_by": "anthropic"
    },
    {
      "id": "openrouter/gpt-4",
      "object": "model",
      "created": 1234567890,
      "owned_by": "openrouter"
    },
    {
      "id": "openrouter/claude-3-opus-20240229",
      "object": "model",
      "created": 1234567890,
      "owned_by": "openrouter"
    }
  ]
}
```

### 请求格式

客户端在请求时使用带前缀的模型名称：

```json
{
  "model": "openai/gpt-4",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ]
}
```

### 路由过程

1. 客户端请求 `model: "openai/gpt-4"`
2. 系统解析前缀：`provider_name = "openai"`, `actual_model = "gpt-4"`
3. 查找对应的 Provider（openai）
4. 使用实际模型名称（gpt-4）向上游服务商发送请求
5. 返回响应时，恢复原始模型名称（openai/gpt-4）

## 配置示例

```yaml
server:
  host: "0.0.0.0"
  port: "8080"

client_api_keys:
  - "sk-openbridge-key-1"

providers:
  # OpenAI 官方
  openai:
    type: openai
    base_url: "https://api.openai.com/v1"
    api_keys:
      - "sk-xxx"
  
  # OpenRouter（支持多个模型商）
  openrouter:
    type: openai
    base_url: "https://openrouter.ai/api/v1"
    api_keys:
      - "sk-or-xxx"
  
  # Anthropic 官方
  anthropic:
    type: anthropic
    api_keys:
      - "sk-ant-xxx"
```

## 使用示例

### Python

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-openbridge-key-1",
    base_url="http://localhost:8080/v1"
)

# 获取所有可用模型
models = client.models.list()
for model in models.data:
    print(f"Model: {model.id} (owned by {model.owned_by})")

# 使用 OpenAI 官方的 GPT-4
response = client.chat.completions.create(
    model="openai/gpt-4",
    messages=[{"role": "user", "content": "Hello!"}]
)

# 使用 OpenRouter 的 Claude
response = client.chat.completions.create(
    model="openrouter/claude-3-opus-20240229",
    messages=[{"role": "user", "content": "Hello!"}]
)

# 使用 Anthropic 官方的 Claude
response = client.chat.completions.create(
    model="anthropic/claude-3-opus-20240229",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

### cURL

```bash
# 获取模型列表
curl -X GET http://localhost:8080/v1/models \
  -H "Authorization: Bearer sk-openbridge-key-1"

# 使用带前缀的模型名称
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-openbridge-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## 向后兼容性

如果只配置了一个 Provider，系统仍然支持不带前缀的模型名称：

```yaml
providers:
  openai:
    type: openai
    api_keys:
      - "sk-xxx"
```

此时可以使用：
```json
{
  "model": "gpt-4",
  "messages": [...]
}
```

系统会自动路由到唯一的 Provider。

## 优势

✅ **解决冲突** - 多个 Provider 拥有相同模型名称时不再有歧义  
✅ **自动发现** - 客户端通过 `/v1/models` 可以看到所有可用的组合  
✅ **无需配置规则** - 不需要复杂的 `routes` 配置  
✅ **用户友好** - 格式类似 Docker Hub，用户容易理解  
✅ **向后兼容** - 单 Provider 场景下仍支持不带前缀的模型名称

## 缓存机制

系统会缓存模型到 Provider 的映射关系，以提高性能：

- 首次调用 `/v1/models` 时，从所有 Provider 获取模型列表
- 后续请求时，直接从缓存查找 Provider
- 缓存使用 `provider_name/model_id` 作为 key

## 流式响应

流式响应也支持前缀路由，每个 chunk 中的 `model` 字段会包含完整的前缀：

```json
{
  "id": "chatcmpl-xxx",
  "object": "text_completion.chunk",
  "created": 1234567890,
  "model": "openai/gpt-4",
  "choices": [
    {
      "index": 0,
      "delta": {
        "content": "Hello"
      },
      "finish_reason": null
    }
  ]
}
```
