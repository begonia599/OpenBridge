package user

const userHTML = `<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenBridge 用户中心</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            width: 100%;
            max-width: 800px;
        }
        .card {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            margin-bottom: 20px;
        }
        h1 {
            color: #667eea;
            margin-bottom: 10px;
            font-size: 28px;
        }
        h2 {
            color: #333;
            font-size: 20px;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 2px solid #f0f0f0;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 14px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
            font-size: 14px;
        }
        input[type="text"],
        input[type="password"],
        input[type="email"] {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            transition: border 0.3s;
        }
        input:focus {
            outline: none;
            border-color: #667eea;
        }
        button {
            padding: 12px 24px;
            border: none;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.3s;
        }
        .btn-primary {
            background: #667eea;
            color: white;
            width: 100%;
        }
        .btn-primary:hover {
            background: #5568d3;
            transform: translateY(-1px);
        }
        .btn-success {
            background: #10b981;
            color: white;
        }
        .btn-success:hover {
            background: #059669;
        }
        .btn-danger {
            background: #ef4444;
            color: white;
            padding: 8px 16px;
            font-size: 13px;
        }
        .btn-danger:hover {
            background: #dc2626;
        }
        .btn-outline {
            background: white;
            color: #667eea;
            border: 2px solid #667eea;
            width: 100%;
            margin-top: 10px;
        }
        .btn-outline:hover {
            background: #f8f9ff;
        }
        .hidden {
            display: none;
        }
        .key-item {
            background: #f8f9fa;
            padding: 16px;
            border-radius: 8px;
            margin-bottom: 12px;
            border-left: 4px solid #667eea;
        }
        .key-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
        }
        .key-name {
            font-weight: 600;
            color: #333;
            font-size: 15px;
        }
        .key-value {
            font-family: 'Courier New', monospace;
            background: white;
            padding: 10px;
            border-radius: 6px;
            font-size: 13px;
            color: #555;
            word-break: break-all;
            margin: 8px 0;
            border: 1px solid #e0e0e0;
        }
        .key-meta {
            display: flex;
            gap: 16px;
            font-size: 12px;
            color: #666;
            margin-top: 8px;
        }
        .meta-item {
            display: flex;
            align-items: center;
            gap: 4px;
        }
        .copy-btn {
            background: #6366f1;
            color: white;
            padding: 6px 12px;
            font-size: 12px;
            border-radius: 6px;
            margin-left: 8px;
        }
        .copy-btn:hover {
            background: #4f46e5;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 16px;
            margin-bottom: 20px;
        }
        .stat-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
        }
        .stat-value {
            font-size: 32px;
            font-weight: bold;
            margin-bottom: 4px;
        }
        .stat-label {
            font-size: 13px;
            opacity: 0.9;
        }
        .toast {
            position: fixed;
            bottom: 30px;
            right: 30px;
            background: #333;
            color: white;
            padding: 16px 24px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
            opacity: 0;
            transform: translateY(20px);
            transition: all 0.3s;
            z-index: 1000;
        }
        .toast.show {
            opacity: 1;
            transform: translateY(0);
        }
        .toast.success {
            background: #10b981;
        }
        .toast.error {
            background: #ef4444;
        }
        .user-info {
            background: #f8f9ff;
            padding: 16px;
            border-radius: 8px;
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .user-name {
            font-size: 16px;
            font-weight: 600;
            color: #667eea;
        }
        .logout-btn {
            background: transparent;
            color: #666;
            border: 1px solid #ddd;
            padding: 6px 16px;
            font-size: 13px;
        }
        .logout-btn:hover {
            background: #f5f5f5;
        }
        .empty-state {
            text-align: center;
            padding: 40px 20px;
            color: #999;
        }
        .empty-state svg {
            width: 64px;
            height: 64px;
            margin-bottom: 16px;
            opacity: 0.5;
        }
        .link-btn {
            background: none;
            border: none;
            color: #667eea;
            cursor: pointer;
            font-size: 14px;
            text-decoration: underline;
            padding: 0;
            margin-top: 12px;
        }
        .link-btn:hover {
            color: #5568d3;
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- 登录页面 -->
        <div id="loginPage" class="card">
            <h1>🌉 OpenBridge</h1>
            <p class="subtitle">通用 LLM API 网关 - 用户中心</p>
            
            <div id="loginForm">
                <h2>登录</h2>
                <div class="form-group">
                    <label>用户名</label>
                    <input type="text" id="loginUsername" placeholder="输入用户名">
                </div>
                <div class="form-group">
                    <label>密码</label>
                    <input type="password" id="loginPassword" placeholder="输入密码">
                </div>
                <button class="btn-primary" onclick="login()">登录</button>
                <button class="btn-outline" onclick="showRegister()">没有账号？立即注册</button>
            </div>

            <div id="registerForm" class="hidden">
                <h2>注册新账号</h2>
                <div class="form-group">
                    <label>用户名</label>
                    <input type="text" id="regUsername" placeholder="3个字符以上">
                </div>
                <div class="form-group">
                    <label>密码</label>
                    <input type="password" id="regPassword" placeholder="6个字符以上">
                </div>
                <div class="form-group">
                    <label>邮箱 (可选)</label>
                    <input type="email" id="regEmail" placeholder="your@email.com">
                </div>
                <button class="btn-primary" onclick="register()">注册</button>
                <button class="btn-outline" onclick="showLogin()">已有账号？去登录</button>
            </div>
        </div>

        <!-- 用户主页 -->
        <div id="dashboardPage" class="card hidden">
            <div class="user-info">
                <div>
                    <h1>👋 你好，<span id="userName"></span></h1>
                    <p class="subtitle" style="margin:0">管理你的 API Keys 和使用情况</p>
                </div>
                <button class="logout-btn" onclick="logout()">登出</button>
            </div>

            <!-- 统计信息 -->
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-value" id="keyCount">0</div>
                    <div class="stat-label">API Keys</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="totalUsage">0</div>
                    <div class="stat-label">总调用次数</div>
                </div>
            </div>

            <!-- API Keys 管理 -->
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px;">
                <h2 style="margin:0;padding:0;border:none;">🔑 我的 API Keys</h2>
                <button class="btn-success" onclick="showGenerateKeyDialog()">+ 生成新 Key</button>
            </div>

            <div id="keysList"></div>

            <!-- 使用说明 -->
            <div style="background:#f8f9ff;padding:16px;border-radius:8px;margin-top:20px;">
                <h3 style="color:#667eea;margin-bottom:12px;font-size:16px;">📖 如何使用</h3>
                <p style="color:#666;font-size:14px;line-height:1.6;margin-bottom:8px;">
                    使用您的 API Key 调用 OpenBridge 服务：
                </p>
                <pre style="background:#fff;padding:12px;border-radius:6px;overflow-x:auto;font-size:13px;border:1px solid #e0e0e0;"><code>from openai import OpenAI

client = OpenAI(
    api_key="<strong style="color:#667eea;">你的API Key</strong>",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="claude-3-5-sonnet-20241022",
    messages=[{"role": "user", "content": "Hello!"}]
)</code></pre>
            </div>
        </div>

        <!-- 生成 Key 对话框 -->
        <div id="generateKeyDialog" class="card hidden">
            <h2>生成新的 API Key</h2>
            <div class="form-group">
                <label>Key 名称</label>
                <input type="text" id="keyName" placeholder="例如：生产环境、测试环境">
            </div>
            <div style="display:flex;gap:12px;">
                <button class="btn-primary" onclick="generateKey()">生成</button>
                <button class="btn-outline" onclick="hideGenerateKeyDialog()">取消</button>
            </div>
        </div>
    </div>

    <div class="toast" id="toast"></div>

    <script>
        let currentUser = null;

        // 页面加载时检查登录状态
        window.onload = function() {
            checkLoginStatus();
        };

        async function checkLoginStatus() {
            try {
                const res = await fetch('/user/api/profile');
                if (res.ok) {
                    const data = await res.json();
                    currentUser = data;
                    showDashboard();
                    loadUserData();
                } else {
                    showLoginPage();
                }
            } catch (e) {
                showLoginPage();
            }
        }

        async function login() {
            const username = document.getElementById('loginUsername').value.trim();
            const password = document.getElementById('loginPassword').value;

            if (!username || !password) {
                showToast('请填写用户名和密码', 'error');
                return;
            }

            try {
                const res = await fetch('/user/api/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, password })
                });

                const data = await res.json();

                if (res.ok) {
                    showToast('登录成功！', 'success');
                    currentUser = { username: data.username };
                    showDashboard();
                    loadUserData();
                } else {
                    showToast(data.error || '登录失败', 'error');
                }
            } catch (e) {
                showToast('网络错误', 'error');
            }
        }

        async function register() {
            const username = document.getElementById('regUsername').value.trim();
            const password = document.getElementById('regPassword').value;
            const email = document.getElementById('regEmail').value.trim();

            if (!username || !password) {
                showToast('请填写用户名和密码', 'error');
                return;
            }

            if (username.length < 3) {
                showToast('用户名至少3个字符', 'error');
                return;
            }

            if (password.length < 6) {
                showToast('密码至少6个字符', 'error');
                return;
            }

            try {
                const res = await fetch('/user/api/register', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, password, email })
                });

                const data = await res.json();

                if (res.ok) {
                    showToast('注册成功！请登录', 'success');
                    showLogin();
                    document.getElementById('loginUsername').value = username;
                } else {
                    showToast(data.error || '注册失败', 'error');
                }
            } catch (e) {
                showToast('网络错误', 'error');
            }
        }

        async function logout() {
            try {
                await fetch('/user/api/logout', { method: 'POST' });
                showToast('已登出', 'success');
                currentUser = null;
                showLoginPage();
            } catch (e) {
                showToast('网络错误', 'error');
            }
        }

        async function loadUserData() {
            try {
                // 加载用户信息
                const profileRes = await fetch('/user/api/profile');
                const profile = await profileRes.json();
                document.getElementById('userName').textContent = profile.username;

                // 加载 Keys
                const keysRes = await fetch('/user/api/keys');
                const keysData = await keysRes.json();
                renderKeys(keysData.keys || []);

                // 加载使用统计
                const usageRes = await fetch('/user/api/usage');
                const usageData = await usageRes.json();
                document.getElementById('keyCount').textContent = (keysData.keys || []).length;
                document.getElementById('totalUsage').textContent = usageData.total_usage || 0;
            } catch (e) {
                console.error('加载数据失败', e);
            }
        }

        function renderKeys(keys) {
            const container = document.getElementById('keysList');
            
            if (keys.length === 0) {
                container.innerHTML = '<div class="empty-state">' +
                    '<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">' +
                    '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"/>' +
                    '</svg>' +
                    '<p>还没有 API Key</p>' +
                    '<p style="font-size:13px;margin-top:8px;">点击上方按钮生成您的第一个 Key</p>' +
                    '</div>';
                return;
            }

            container.innerHTML = keys.map(function(key) {
                var lastUsedHTML = '';
                if (key.last_used) {
                    lastUsedHTML = '<div class="meta-item">' +
                        '<span>🕐 最后使用:</span>' +
                        '<span>' + new Date(key.last_used).toLocaleString() + '</span>' +
                        '</div>';
                }
                
                return '<div class="key-item">' +
                    '<div class="key-header">' +
                    '<span class="key-name">' + key.name + '</span>' +
                    '<button class="btn-danger" onclick="deleteKey(\'' + key.key + '\')">删除</button>' +
                    '</div>' +
                    '<div class="key-value">' +
                    key.key +
                    '<button class="copy-btn" onclick="copyKey(\'' + key.key + '\')">复制</button>' +
                    '</div>' +
                    '<div class="key-meta">' +
                    '<div class="meta-item">' +
                    '<span>📅 创建:</span>' +
                    '<span>' + new Date(key.created_at).toLocaleDateString() + '</span>' +
                    '</div>' +
                    '<div class="meta-item">' +
                    '<span>📊 使用:</span>' +
                    '<span>' + key.usage + ' 次</span>' +
                    '</div>' +
                    lastUsedHTML +
                    '</div>' +
                    '</div>';
            }).join('');
        }

        async function generateKey() {
            const name = document.getElementById('keyName').value.trim() || 'API Key';

            try {
                const res = await fetch('/user/api/keys/generate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name })
                });

                const data = await res.json();

                if (res.ok) {
                    showToast('Key 生成成功！', 'success');
                    hideGenerateKeyDialog();
                    loadUserData();
                } else {
                    showToast(data.error || '生成失败', 'error');
                }
            } catch (e) {
                showToast('网络错误', 'error');
            }
        }

        async function deleteKey(key) {
            if (!confirm('确定要删除此 Key 吗？删除后无法恢复！')) return;

            try {
                const res = await fetch('/user/api/keys/' + encodeURIComponent(key), {
                    method: 'DELETE'
                });

                const data = await res.json();

                if (res.ok) {
                    showToast('Key 已删除', 'success');
                    loadUserData();
                } else {
                    showToast(data.error || '删除失败', 'error');
                }
            } catch (e) {
                showToast('网络错误', 'error');
            }
        }

        function copyKey(key) {
            navigator.clipboard.writeText(key).then(() => {
                showToast('已复制到剪贴板', 'success');
            }).catch(() => {
                showToast('复制失败', 'error');
            });
        }

        function showLogin() {
            document.getElementById('loginForm').classList.remove('hidden');
            document.getElementById('registerForm').classList.add('hidden');
        }

        function showRegister() {
            document.getElementById('loginForm').classList.add('hidden');
            document.getElementById('registerForm').classList.remove('hidden');
        }

        function showLoginPage() {
            document.getElementById('loginPage').classList.remove('hidden');
            document.getElementById('dashboardPage').classList.add('hidden');
            document.getElementById('generateKeyDialog').classList.add('hidden');
        }

        function showDashboard() {
            document.getElementById('loginPage').classList.add('hidden');
            document.getElementById('dashboardPage').classList.remove('hidden');
            document.getElementById('generateKeyDialog').classList.add('hidden');
        }

        function showGenerateKeyDialog() {
            document.getElementById('dashboardPage').classList.add('hidden');
            document.getElementById('generateKeyDialog').classList.remove('hidden');
            document.getElementById('keyName').value = '';
        }

        function hideGenerateKeyDialog() {
            document.getElementById('generateKeyDialog').classList.add('hidden');
            document.getElementById('dashboardPage').classList.remove('hidden');
        }

        function showToast(msg, type = '') {
            const toast = document.getElementById('toast');
            toast.textContent = msg;
            toast.className = 'toast show ' + type;
            setTimeout(() => {
                toast.classList.remove('show');
            }, 3000);
        }

        // Enter 键快捷登录/注册
        document.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                if (!document.getElementById('loginPage').classList.contains('hidden')) {
                    if (!document.getElementById('loginForm').classList.contains('hidden')) {
                        login();
                    } else {
                        register();
                    }
                }
            }
        });
    </script>
</body>
</html>`

