# OpenBridge

<div align="center">

**通用 LLM API 网关 - 统一接口，多Provider支持**

[![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)](https://github.com/yourusername/openbridge)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

</div>

## ✨ 特性

- 🔄 **统一接口**: 下游始终使用 OpenAI 格式 API，无需修改客户端代码
- 🎯 **多 Provider 支持**: 原生支持 OpenAI、Claude (Anthropic)、Google Gemini 等
- 🔀 **智能路由**: 基于模型名称自动路由到对应的 Provider
- 🔑 **API Key 管理**: 支持多个 API Key 轮询、负载均衡
- 📊 **使用统计**: 实时查看各 Provider 的使用情况
- 🎨 **管理后台**: Web 界面管理配置、Key 和路由规则
- ⚡ **流式支持**: 完整支持 Server-Sent Events (SSE) 流式响应
- 🔄 **自动转换**: 自动进行 API 格式转换，对下游透明

## 🚀 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/openbridge.git
cd openbridge

# 编译
go build -o openbridge .
```

### 配置

复制配置示例并修改：

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: "8080"

# 客户端 API Keys
client_api_keys:
  - "sk-openbridge-key-1"

# Provider 配置
providers:
  # OpenAI 官方
  openai:
    type: openai
    base_url: "https://api.openai.com/v1"
    api_keys:
      - "sk-your-openai-key"
  
  # Claude 原生 API
  claude:
    type: anthropic
    base_url: ""  # 留空使用默认
    api_keys:
      - "sk-ant-your-claude-key"
  
  # Google Gemini
  gemini:
    type: google
    base_url: ""  # 留空使用默认
    api_keys:
      - "your-google-api-key"

# 模型路由规则
routes:
  "gpt-*": openai
  "o1-*": openai
  "claude-*": claude
  "gemini-*": gemini
```

### 运行

```bash
./openbridge
```

服务器将在 `http://localhost:8080` 启动。

## 📖 使用示例

### Python

```python
from openai import OpenAI

# 使用 OpenBridge 作为代理
client = OpenAI(
    api_key="sk-openbridge-key-1",  # OpenBridge 的客户端 API Key
    base_url="http://localhost:8080/v1"
)

# 调用 OpenAI
response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}]
)

# 调用 Claude (自动路由)
response = client.chat.completions.create(
    model="claude-3-5-sonnet-20241022",
    messages=[{"role": "user", "content": "Hello!"}]
)

# 调用 Gemini (自动路由)
response = client.chat.completions.create(
    model="gemini-1.5-pro",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

### cURL

```bash
# 调用 Claude
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-openbridge-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# 调用 Gemini
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-openbridge-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-1.5-pro",
    "messages": [{"role": "user", "content": "你好！"}]
  }'
```

## 🎯 支持的 Provider 类型

### OpenAI 格式 (`type: openai`)

适用于所有兼容 OpenAI API 格式的服务商：
- OpenAI 官方
- Azure OpenAI
- DeepSeek
- Moonshot (Kimi)
- 智谱 AI (GLM)
- 阿里云百炼
- 各种第三方代理

**配置示例**：

```yaml
providers:
  openai:
    type: openai
    base_url: "https://api.openai.com/v1"
    api_keys:
      - "sk-xxx"
```

### Anthropic Claude (`type: anthropic` 或 `claude`)

Claude 官方原生 API 支持，自动进行格式转换。

**特性**：
- ✅ 完整支持 Messages API
- ✅ System prompt 转换
- ✅ 流式响应
- ✅ 多模态 (图片)

**配置示例**：

```yaml
providers:
  claude:
    type: anthropic
    base_url: ""  # 可选，默认 https://api.anthropic.com
    api_keys:
      - "sk-ant-xxx"
```

### Google Gemini (`type: google` 或 `gemini`)

Google Gemini 原生 API 支持，自动进行格式转换。

**特性**：
- ✅ 完整支持 Gemini API
- ✅ System instruction 转换
- ✅ 流式响应
- ✅ 多模态 (图片)
- ✅ 安全设置自动配置

**配置示例**：

```yaml
providers:
  gemini:
    type: google
    base_url: ""  # 可选，默认官方 API
    api_keys:
      - "AIzaSyxxx"
```

## 🔄 格式转换说明

OpenBridge 自动在 OpenAI 格式和各 Provider 原生格式之间转换：

### Claude 转换

| OpenAI | Claude |
|--------|--------|
| `messages` (role: system) | `system` 字段 |
| `messages` (role: user/assistant) | `messages` 数组 |
| `max_tokens` | `max_tokens` (必需) |
| `temperature` | `temperature` |
| `top_p` | `top_p` |
| 图片 (data URI) | `image` content block |

### Gemini 转换

| OpenAI | Gemini |
|--------|--------|
| `messages` (role: system) | `systemInstruction` |
| `messages` (role: user) | `contents` (role: user) |
| `messages` (role: assistant) | `contents` (role: model) |
| `max_tokens` | `maxOutputTokens` |
| `temperature` | `temperature` |
| `top_p` | `topP` |
| 图片 (data URI) | `inlineData` |

## 🎨 管理后台

访问 `http://localhost:8080/admin` 打开 Web 管理界面。

**功能**：
- 📊 查看所有 Provider 和路由配置
- ➕ 动态添加/删除 Provider
- 🔑 生成/管理客户端 API Key
- 🔀 配置模型路由规则
- 💾 实时保存配置

## 📡 API 端点

### 核心端点

- `POST /v1/chat/completions` - 聊天补全 (流式/非流式)
- `GET /v1/models` - 列出所有可用模型
- `GET /v1/models/{model}` - 获取模型详情

### 管理端点

- `GET /health` - 健康检查
- `GET /version` - 版本信息
- `GET /stats` - 使用统计
- `GET /providers` - Provider 列表
- `GET /admin` - 管理界面

## ⚙️ 配置选项

### Server 配置

```yaml
server:
  host: "0.0.0.0"  # 监听地址
  port: "8080"     # 监听端口
```

### Admin 配置

```yaml
admin:
  enabled: true              # 启用管理后台
  password: "your-password"  # 留空则无需密码
```

### Logging 配置

```yaml
logging:
  level: "info"           # debug, info, warn, error
  format: "text"          # text, json
  log_requests: false     # 记录请求详情
  log_responses: false    # 记录响应详情
```

### Provider 配置

```yaml
providers:
  <name>:
    type: openai|anthropic|google    # Provider 类型
    base_url: "https://..."           # API 地址
    api_keys:                         # API Keys 列表
      - "key1"
      - "key2"
    rotation_strategy: round_robin    # round_robin (轮询)
```

### Routes 配置

```yaml
routes:
  "<pattern>": <provider_name>
```

支持通配符 `*`，例如：
- `gpt-*` 匹配所有以 `gpt-` 开头的模型
- `claude-3-*` 匹配所有 Claude 3 系列模型

## 🔐 安全建议

1. **生产环境**：
   - 设置管理后台密码
   - 使用 HTTPS (建议通过 Nginx 反向代理)
   - 限制管理后台访问 IP

2. **API Key 管理**：
   - 定期轮换 API Keys
   - 使用环境变量存储敏感信息
   - 不要将配置文件提交到版本控制

3. **网络配置**：
   - 在生产环境中设置 `GIN_MODE=release`
   - 配置适当的防火墙规则
   - 使用负载均衡器分发请求

## 🛠️ 开发

### 项目结构

```
openbridge/
├── internal/
│   ├── admin/          # 管理后台
│   ├── config/         # 配置管理
│   ├── handler/        # HTTP 处理器
│   ├── middleware/     # 中间件
│   ├── models/         # 数据模型
│   ├── provider/       # Provider 实现
│   │   ├── openai/     # OpenAI Provider
│   │   ├── anthropic/  # Claude Provider
│   │   └── google/     # Gemini Provider
│   ├── router/         # 路由配置
│   └── service/        # 业务逻辑
├── main.go             # 入口文件
├── version.go          # 版本信息
└── config.example.yaml # 配置示例
```

### 添加新 Provider

1. 在 `internal/provider/` 下创建新目录
2. 实现 `Provider` 接口
3. 实现格式转换函数
4. 在 `main.go` 中注册

示例接口：

```go
type Provider interface {
    Name() string
    Type() string
    ChatCompletion(req *models.ChatCompletionRequest, apiKey string) (*models.ChatCompletionResponse, error)
    ChatCompletionStream(req *models.ChatCompletionRequest, apiKey string) (<-chan *models.ChatCompletionChunk, <-chan error)
    ListModels(apiKey string) (*models.ModelList, error)
    SupportsStreaming() bool
}
```

## 📝 更新日志

### v2.0.0 (2025-12-01)

- ✨ **新增** Claude (Anthropic) 原生 API 支持
- ✨ **新增** Google Gemini 原生 API 支持
- ✨ **新增** 自动格式转换功能
- ✨ **新增** 流式响应转换支持
- 🔧 **改进** Provider 架构，支持多种 API 格式
- 📚 **文档** 完善配置和使用说明

### v1.0.0

- 🎉 初始版本
- ✅ OpenAI 格式 Provider 支持
- ✅ 基础路由和 Key 管理

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

---

<div align="center">

**[官网](https://github.com/yourusername/openbridge)** • **[文档](https://github.com/yourusername/openbridge/wiki)** • **[问题反馈](https://github.com/yourusername/openbridge/issues)**

Made with ❤️ by OpenBridge Team

</div>

