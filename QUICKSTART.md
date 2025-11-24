# 快速开始

5 分钟部署 OpenBridge 到你的服务器!

---

## 🚀 最快部署 (Ubuntu/Debian)

### 1. 下载并运行

```bash
# 克隆项目
git clone https://github.com/your-repo/openbridge.git
cd openbridge

# 一键部署
sudo chmod +x deploy.sh
sudo ./deploy.sh
```

### 2. 输入 API Key

```
请输入第一个 AssemblyAI API Key: [粘贴你的 Key]
```

### 3. 完成!

```
╔═══════════════════════════════════════════╗
║          🎉 部署成功! 🎉                  ║
╚═══════════════════════════════════════════╝

服务信息:
  • 本地访问: http://localhost:8080
```

---

## 📝 获取 AssemblyAI API Key

1. 访问 https://www.assemblyai.com/
2. 注册账号
3. 进入 Dashboard
4. 复制 API Key

---

## 🧪 测试

```bash
# 健康检查
curl http://localhost:8080/health

# 发送请求
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-openbridge-test-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5-20250929",
    "messages": [{"role": "user", "content": "你好"}],
    "max_tokens": 100
  }'
```

---

## 🔧 常用命令

```bash
# 查看日志
docker compose logs -f

# 重启服务
docker compose restart

# 停止服务
docker compose stop

# 启动服务
docker compose start
```

---

## 📚 下一步

- 阅读 [DEPLOYMENT.md](DEPLOYMENT.md) 了解详细部署
- 阅读 [FEATURES.md](FEATURES.md) 了解功能配置
- 阅读 [PARAMETERS.md](PARAMETERS.md) 了解 API 参数

---

## ❓ 常见问题

### Q: 如何修改端口?

编辑 `docker-compose.yml`:
```yaml
ports:
  - "9000:8080"  # 改为 9000
```

### Q: 如何添加更多 API Keys?

编辑 `config.yaml`:
```yaml
assemblyai:
  api_keys:
    - "key-1"
    - "key-2"
    - "key-3"  # 新增
```

然后重启: `docker compose restart`

### Q: 如何查看日志?

```bash
docker compose logs -f
```

---

## 🎯 就这么简单!

现在你有了一个完整的 OpenAI 兼容 API 网关,可以使用 AssemblyAI 的 Claude 模型了!
