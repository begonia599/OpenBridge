# OpenBridge

OpenAI API 到 AssemblyAI 的智能网关,支持 API Key 轮询和流式请求自动转换

## 核心特性

- ✅ **OpenAI 完全兼容**: 标准 OpenAI API 格式
- 🔄 **API Key 轮询**: 支持多个后端 Key 自动轮询 (round_robin/random/least_used)
- 🔐 **客户端认证**: 支持多个客户端 API Key 管理
- 🌊 **流式智能处理**: 自动将流式请求转换为非流式(可配置)
- 📊 **详细日志**: 完整记录请求和响应,便于调试
- 🚀 **高性能**: 基于 Gin 框架
- 📝 **响应标准化**: 自动补全 OpenAI 标准字段

## 快速开始

### 1. 配置

复制示例配置并编辑:

```bash
cp config.example.yaml config.yaml
# 编辑 config.yaml,填入你的 API Keys
```

```yaml
# 客户端 API Keys
client_api_keys:
  - "sk-your-client-key-1"

# 后端 AssemblyAI Keys
assemblyai:
  api_keys:
    - "your-assemblyai-key-1"
  
  features:
    stream: false  # 是否支持流式
    unsupported_params:
      - "temperature"  # 不支持的参数列表
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

详见 [DEPLOYMENT.md](DEPLOYMENT.md)

### 3. 使用

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-your-client-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5-20250929",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## API 端点

- `POST /v1/chat/completions` - 对话补全
- `GET /v1/models` - 获取模型列表
- `GET /v1/models/:model` - 获取指定模型
- `GET /health` - 健康检查
- `GET /version` - 版本信息
- `GET /stats` - API Key 使用统计

## 配置说明

### 流式处理

当 `support_stream: false` 时,客户端的 `stream: true` 请求会自动转换为 `stream: false`,避免报错。

### 日志配置

```yaml
logging:
  level: debug  # 日志级别
  log_requests: true  # 记录请求
  log_responses: true  # 记录响应
```

## 测试脚本

- `test_client.py` - 完整功能测试
- `test_stream.py` - 流式处理测试
- `test_response_format.py` - 响应格式验证
- `test_rate_limit.py` - 速率限制测试
- `test_image.py` - 图片/多模态测试
- `test_model_limits.py` - 模型限制测试 (输入/输出 token)
- `test_compare_official.py` - 与官方 API 对比测试
- `test_assemblyai_direct.py` - 直接测试 AssemblyAI API

## 文档

- `DEPLOYMENT.md` - 部署指南 (Docker) 🐳
- `FEATURES.md` - 功能配置指南
- `PARAMETERS.md` - OpenAI API 参数详解
- `ERROR_HANDLING.md` - 错误处理说明
- `MODEL_VERIFICATION.md` - 模型验证指南
- `README.md` - 项目说明
- `ai.md` - AssemblyAI API 文档
