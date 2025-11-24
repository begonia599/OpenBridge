# OpenBridge

**OpenAI-Compatible API Gateway** - A universal gateway that bridges OpenAI-compatible clients with various LLM providers.

[中文文档](README_CN.md) | [English](README.md)

## Features

- ✅ **OpenAI Compatible**: Standard OpenAI API format support
- 🔄 **API Key Rotation**: Multiple backend keys with automatic rotation (round_robin/random/least_used)
- 🔐 **Client Authentication**: Multi-client API key management
- 🌊 **Smart Stream Handling**: Auto-convert between streaming and non-streaming modes
- 🎯 **Parameter Filtering**: Configurable unsupported parameter stripping
- 📊 **Detailed Logging**: Complete request/response logging for debugging
- 🚀 **High Performance**: Built on Gin framework
- 📝 **Response Normalization**: Auto-complete OpenAI standard fields
- 🔧 **Flexible Configuration**: YAML-based configuration system

## Quick Start

### 1. Configuration

Copy the example config and edit:

```bash
cp config.example.yaml config.yaml
# Edit config.yaml and fill in your API Keys
```

```yaml
# Client API Keys (for downstream clients)
client_api_keys:
  - "sk-your-client-key-1"

# Backend Provider Configuration (example: AssemblyAI)
assemblyai:
  base_url: "https://llm-gateway.assemblyai.com/v1"
  api_keys:
    - "your-backend-api-key-1"
  
  features:
    stream: false  # Streaming support
    tools: false   # Tool calling support
    unsupported_params:
      - "temperature"  # Parameters not supported by backend
```

### 2. Run

#### Development
```bash
go run main.go
```

#### Production (Docker)
```bash
# One-click deployment
sudo chmod +x deploy.sh
sudo ./deploy.sh

# Or manual deployment
docker compose up -d
```

### 3. Usage

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-your-client-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "your-model-name",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## API Endpoints

- `POST /v1/chat/completions` - Chat completions
- `GET /v1/models` - List available models
- `GET /v1/models/:model` - Retrieve specific model
- `GET /health` - Health check
- `GET /version` - Version information
- `GET /stats` - API key usage statistics

## Configuration

### Stream Handling

When backend doesn't support streaming (`stream: false`), client streaming requests are automatically converted to non-streaming mode with fake SSE responses.

### Parameter Filtering

Configure unsupported parameters in `features.unsupported_params` to automatically strip them from requests:

```yaml
features:
  unsupported_params:
    - "temperature"  # Will be removed from requests
    - "top_p"        # Add any unsupported parameters
```

### Logging

```yaml
logging:
  level: debug  # Log level: debug, info, warn, error
  log_requests: true   # Log request bodies
  log_responses: true  # Log response bodies
```

## Supported Backends

Currently tested with:
- **AssemblyAI** - Claude models via LLM Gateway

Easily extensible to other providers by adjusting configuration.

## Version

Current version: **v1.0.1**

Check version:
```bash
curl http://localhost:8080/version
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
