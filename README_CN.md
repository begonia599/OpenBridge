# OpenBridge

**OpenAI 兼容 API 网关** - 连接 OpenAI 兼容客户端与各种 LLM 提供商的通用网关。

[中文文档](README_CN.md) | [English](README.md)

## 特性

- ✅ **OpenAI 兼容**: 支持标准 OpenAI API 格式
- 🔄 **API Key 轮询**: 多个后端密钥自动轮询 (round_robin/random/least_used)
- 🔐 **客户端认证**: 多客户端 API Key 管理
- 🌊 **智能流式处理**: 自动在流式和非流式模式间转换
- 🎯 **参数过滤**: 可配置的不支持参数自动剔除
- 📊 **详细日志**: 完整的请求/响应日志用于调试
- 🚀 **高性能**: 基于 Gin 框架构建
- 📝 **响应标准化**: 自动补全 OpenAI 标准字段
- 🔧 **灵活配置**: 基于 YAML 的配置系统

## 快速开始

### 1. 配置

复制示例配置并编辑:

```bash
cp config.example.yaml config.yaml
# 编辑 config.yaml 并填入你的 API Keys
```

```yaml
# 客户端 API Keys (供下游客户端使用)
client_api_keys:
  - "sk-your-client-key-1"

# 后端提供商配置 (示例: AssemblyAI)
assemblyai:
  base_url: "https://llm-gateway.assemblyai.com/v1"
  api_keys:
    - "your-backend-api-key-1"
  
  features:
    stream: false  # 流式支持
    tools: false   # 工具调用支持
    unsupported_params:
      - "temperature"  # 后端不支持的参数
```

### 2. 运行

#### 开发环境
```bash
go run main.go
```

#### 生产环境 (Docker)
```bash
# 一键部署
sudo chmod +x deploy.sh
sudo ./deploy.sh

# 或手动部署
docker compose up -d
```

### 3. 使用

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-your-client-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "your-model-name",
    "messages": [{"role": "user", "content": "你好!"}]
  }'
```

## API 端点

- `POST /v1/chat/completions` - 对话补全
- `GET /v1/models` - 列出可用模型
- `GET /v1/models/:model` - 获取指定模型信息
- `GET /health` - 健康检查
- `GET /version` - 版本信息
- `GET /stats` - API Key 使用统计

## 配置说明

### 流式处理

当后端不支持流式 (`stream: false`) 时,客户端的流式请求会自动转换为非流式模式,并返回伪 SSE 响应。

### 参数过滤

在 `features.unsupported_params` 中配置不支持的参数,将自动从请求中剔除:

```yaml
features:
  unsupported_params:
    - "temperature"  # 将从请求中移除
    - "top_p"        # 添加任何不支持的参数
```

### 日志配置

```yaml
logging:
  level: debug  # 日志级别: debug, info, warn, error
  log_requests: true   # 记录请求体
  log_responses: true  # 记录响应体
```

## 支持的后端

目前已测试:
- **AssemblyAI** - 通过 LLM Gateway 访问 Claude 模型

通过调整配置可轻松扩展到其他提供商。

## 版本

当前版本: **v1.0.1**

查看版本:
```bash
curl http://localhost:8080/version
```

## 许可证

MIT License

## 贡献

欢迎贡献! 请随时提交 Pull Request。
