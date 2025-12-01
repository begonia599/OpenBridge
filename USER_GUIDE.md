# OpenBridge 用户系统指南

## 🎯 概述

OpenBridge 现在提供了完整的用户系统，用户可以：
- 🔐 注册和登录账户
- 🔑 生成和管理自己的 API Keys
- 📊 查看 API 使用统计
- 🎨 通过美观的 Web 界面管理

## 🚀 快速开始

### 1. 访问用户中心

打开浏览器访问：**http://localhost:8080/user**

### 2. 注册账户

首次使用需要注册账户：

1. 点击 **"没有账号？立即注册"**
2. 填写信息：
   - **用户名**：至少 3 个字符
   - **密码**：至少 6 个字符
   - **邮箱**：可选
3. 点击 **"注册"**
4. 注册成功后会自动为你生成一个默认 API Key

### 3. 登录

使用注册的用户名和密码登录：

1. 输入用户名和密码
2. 点击 **"登录"**
3. 登录后进入用户控制面板

### 4. 管理 API Keys

#### 查看 Keys

登录后可以看到所有 API Keys：
- Key 值
- Key 名称
- 创建时间
- 使用次数
- 最后使用时间

#### 生成新 Key

1. 点击 **"+ 生成新 Key"**
2. 输入 Key 名称（例如：生产环境、测试环境）
3. 点击 **"生成"**
4. 复制新生成的 Key 并妥善保存

#### 复制 Key

点击 Key 右侧的 **"复制"** 按钮，Key 将被复制到剪贴板。

#### 删除 Key

1. 点击 Key 右侧的 **"删除"** 按钮
2. 确认删除
3. 删除后该 Key 立即失效

## 📊 使用统计

用户控制面板顶部显示：
- **API Keys 数量**：拥有的 Key 总数
- **总调用次数**：所有 Key 的累计调用次数

每个 Key 单独显示：
- 该 Key 的调用次数
- 最后使用时间

## 💻 使用 API Key

获取 API Key 后，可以在程序中使用：

### Python 示例

```python
from openai import OpenAI

# 使用你的用户 API Key
client = OpenAI(
    api_key="sk-user-xxxxxxxxxxxxxx",  # 从用户中心复制的 Key
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="claude-3-5-sonnet-20241022",
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
```

### cURL 示例

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-user-xxxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### JavaScript 示例

```javascript
const OpenAI = require('openai');

const client = new OpenAI({
  apiKey: 'sk-user-xxxxxxxxxxxxxx',
  baseURL: 'http://localhost:8080/v1'
});

const response = await client.chat.completions.create({
  model: 'claude-3-5-sonnet-20241022',
  messages: [{ role: 'user', content: 'Hello!' }]
});

console.log(response.choices[0].message.content);
```

## 🔒 安全建议

### API Key 安全

1. **妥善保管**：将 API Key 当作密码一样保管
2. **不要分享**：不要在公开场合分享你的 API Key
3. **不要提交到代码库**：不要将 Key 提交到 Git 等版本控制系统
4. **使用环境变量**：建议使用环境变量存储 Key

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.getenv("OPENBRIDGE_API_KEY"),
    base_url="http://localhost:8080/v1"
)
```

### 账户安全

1. **使用强密码**：密码至少 6 个字符，建议包含大小写字母、数字和符号
2. **定期更换密码**：建议定期更换密码
3. **及时删除泄露的 Key**：如果 Key 泄露，立即删除并生成新的

## 📝 默认用户

系统初始化时会创建一个默认测试用户：

- **用户名**：`demo`
- **密码**：`demo123`

⚠️ **生产环境请立即修改或删除此默认用户！**

## 🆚 用户 Key vs 管理员 Key

OpenBridge 支持两种类型的 API Key：

### 用户 Key (sk-user-xxx)

- ✅ 通过用户系统生成
- ✅ 每个用户独立管理
- ✅ 可查看使用统计
- ✅ 可随时生成/删除
- ✅ 有使用记录

### 管理员 Key (sk-openbridge-xxx 或 sk-ob-xxx)

- ✅ 在配置文件中定义
- ✅ 通过管理后台生成
- ✅ 全局共享
- ⚠️ 无单独使用统计

## 🌐 多用户场景

OpenBridge 用户系统适用于：

### 团队协作

```
公司团队
├── 用户A（前端开发）
│   ├── Key 1: 开发环境
│   └── Key 2: 测试环境
├── 用户B（后端开发）
│   ├── Key 1: 本地开发
│   └── Key 2: CI/CD
└── 用户C（产品经理）
    └── Key 1: 产品测试
```

### API 售卖

如果你想提供 API 服务：
1. 客户注册账户
2. 获取自己的 API Key
3. 使用 Key 调用服务
4. 查看自己的使用量

### 多环境管理

一个用户可以为不同环境生成不同的 Key：
- 开发环境 Key
- 测试环境 Key
- 生产环境 Key
- 个人项目 Key

## 📁 数据存储

用户数据存储在 `users.json` 文件中：

```json
{
  "demo": {
    "username": "demo",
    "password": "hash...",
    "email": "demo@openbridge.local",
    "api_keys": [
      {
        "key": "sk-user-xxx",
        "name": "默认 Key",
        "created_at": "2025-12-01T13:25:28Z",
        "usage": 42
      }
    ],
    "created_at": "2025-12-01T13:25:28Z"
  }
}
```

⚠️ **备份建议**：定期备份 `users.json` 文件

## 🔧 管理员功能

### 查看用户文件

```bash
cat users.json
```

### 备份用户数据

```bash
cp users.json users.backup.json
```

### 重置用户系统

```bash
rm users.json
# 重启服务，将自动创建默认用户
```

## 🎯 常见问题

### Q: 忘记密码怎么办？

A: 目前需要管理员手动编辑 `users.json` 文件。未来版本将支持密码重置功能。

### Q: 可以修改用户名吗？

A: 目前不支持修改用户名。如需更改，请注册新账户。

### Q: API Key 有数量限制吗？

A: 没有硬性限制，但建议每个用户保持在 10 个以内。

### Q: Key 使用有频率限制吗？

A: 系统本身没有限制，取决于上游 Provider 的限制。

### Q: 如何批量导入用户？

A: 可以直接编辑 `users.json` 文件。注意密码需要使用 SHA256 哈希。

### Q: 支持 OAuth 登录吗？

A: 当前版本不支持，未来版本可能会添加。

## 🔮 未来功能

计划中的功能：

- [ ] 密码重置
- [ ] 邮箱验证
- [ ] 使用配额限制
- [ ] OAuth 登录支持
- [ ] 用户角色和权限
- [ ] API 调用日志详情
- [ ] 使用量图表
- [ ] Key 过期时间设置
- [ ] 双因素认证 (2FA)

## 📞 技术支持

如有问题，请：
1. 查看日志：服务器终端输出
2. 查看用户数据：`users.json`
3. 提交 Issue：[GitHub Issues](https://github.com/yourusername/openbridge/issues)

---

<div align="center">

**[返回主文档](README.md)** • **[Provider 对比](PROVIDER_COMPARISON.md)** • **[管理员指南](README.md#管理后台)**

Made with ❤️ by OpenBridge Team

</div>

