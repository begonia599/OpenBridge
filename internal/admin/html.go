package admin

const adminHTML = `<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenBridge Admin</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; color: #333; }
        .container { max-width: 1000px; margin: 0 auto; padding: 20px; }
        h1 { margin-bottom: 20px; color: #1a1a1a; }
        h2 { font-size: 18px; margin-bottom: 15px; color: #444; border-bottom: 2px solid #007bff; padding-bottom: 8px; }
        .card { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .form-row { display: flex; gap: 10px; margin-bottom: 10px; flex-wrap: wrap; }
        input, select { padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; }
        input:focus, select:focus { outline: none; border-color: #007bff; }
        input[type="text"] { flex: 1; min-width: 150px; }
        button { padding: 8px 16px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; transition: background 0.2s; }
        .btn-primary { background: #007bff; color: white; }
        .btn-primary:hover { background: #0056b3; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-danger:hover { background: #c82333; }
        .btn-success { background: #28a745; color: white; }
        .btn-success:hover { background: #218838; }
        table { width: 100%; border-collapse: collapse; margin-top: 10px; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #eee; }
        th { background: #f8f9fa; font-weight: 600; }
        .tag { display: inline-block; padding: 2px 8px; background: #e9ecef; border-radius: 4px; font-size: 12px; margin: 2px; }
        .tag-openai { background: #d4edda; color: #155724; }
        .tag-anthropic { background: #fff3cd; color: #856404; }
        .tag-google { background: #cce5ff; color: #004085; }
        .key-display { font-family: monospace; background: #f8f9fa; padding: 4px 8px; border-radius: 4px; }
        .copy-btn { padding: 4px 8px; font-size: 12px; margin-left: 8px; }
        .status { padding: 20px; text-align: center; color: #666; }
        .toast { position: fixed; bottom: 20px; right: 20px; padding: 12px 20px; background: #333; color: white; border-radius: 4px; opacity: 0; transition: opacity 0.3s; }
        .toast.show { opacity: 1; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌉 OpenBridge Admin</h1>

        <!-- Client Keys -->
        <div class="card">
            <h2>🔑 客户端 API Keys</h2>
            <p style="color:#666;margin-bottom:15px;font-size:14px">下游客户端使用这些 Key 访问本服务</p>
            <button class="btn-success" onclick="generateKey()">+ 生成新 Key</button>
            <table id="keysTable">
                <thead><tr><th>API Key</th><th>操作</th></tr></thead>
                <tbody></tbody>
            </table>
        </div>

        <!-- Providers -->
        <div class="card">
            <h2>🔌 上游 Providers</h2>
            <p style="color:#666;margin-bottom:15px;font-size:14px">配置上游 LLM 服务商</p>
            <div class="form-row">
                <input type="text" id="providerName" placeholder="名称 (如: openai)">
                <select id="providerType">
                    <option value="openai">OpenAI 格式</option>
                    <option value="anthropic">Anthropic 格式</option>
                    <option value="google">Google 格式</option>
                </select>
                <input type="text" id="providerUrl" placeholder="Base URL (如: https://api.openai.com/v1)">
                <input type="text" id="providerKeys" placeholder="API Keys (逗号分隔)">
                <button class="btn-primary" onclick="addProvider()">添加</button>
            </div>
            <table id="providersTable">
                <thead><tr><th>名称</th><th>类型</th><th>Base URL</th><th>API Keys</th><th>操作</th></tr></thead>
                <tbody></tbody>
            </table>
        </div>

        <!-- Routes -->
        <div class="card">
            <h2>🔀 模型路由</h2>
            <p style="color:#666;margin-bottom:15px;font-size:14px">根据模型名称路由到对应 Provider（支持通配符 *）</p>
            <div class="form-row">
                <input type="text" id="routePattern" placeholder="模型匹配 (如: gpt-* 或 claude-3-opus)">
                <select id="routeProvider"></select>
                <button class="btn-primary" onclick="addRoute()">添加</button>
            </div>
            <table id="routesTable">
                <thead><tr><th>模型匹配</th><th>Provider</th><th>操作</th></tr></thead>
                <tbody></tbody>
            </table>
        </div>
    </div>

    <div class="toast" id="toast"></div>

    <script>
        const password = new URLSearchParams(window.location.search).get('password') || '';
        const headers = { 'Content-Type': 'application/json', 'X-Admin-Password': password };

        async function loadConfig() {
            try {
                const res = await fetch('/admin/api/config?password=' + password);
                const data = await res.json();
                renderKeys(data.client_api_keys || []);
                renderProviders(data.providers || {});
                renderRoutes(data.routes || {});
                updateProviderSelect(data.providers || {});
            } catch (e) {
                console.error(e);
            }
        }

        function renderKeys(keys) {
            const tbody = document.querySelector('#keysTable tbody');
            if (keys.length === 0) {
                tbody.innerHTML = '<tr><td colspan="2" class="status">暂无 Key，点击上方按钮生成</td></tr>';
                return;
            }
            tbody.innerHTML = keys.map(key => ` + "`" + `
                <tr>
                    <td><span class="key-display">${key}</span><button class="btn-primary copy-btn" onclick="copyKey('${key}')">复制</button></td>
                    <td><button class="btn-danger" onclick="deleteKey('${key}')">删除</button></td>
                </tr>
            ` + "`" + `).join('');
        }

        function renderProviders(providers) {
            const tbody = document.querySelector('#providersTable tbody');
            const entries = Object.entries(providers);
            if (entries.length === 0) {
                tbody.innerHTML = '<tr><td colspan="5" class="status">暂无 Provider</td></tr>';
                return;
            }
            tbody.innerHTML = entries.map(([name, p]) => ` + "`" + `
                <tr>
                    <td><strong>${name}</strong></td>
                    <td><span class="tag tag-${p.type}">${p.type}</span></td>
                    <td style="font-size:13px">${p.base_url}</td>
                    <td>${p.api_keys.map(k => '<span class="tag">' + k + '</span>').join('')}</td>
                    <td><button class="btn-danger" onclick="deleteProvider('${name}')">删除</button></td>
                </tr>
            ` + "`" + `).join('');
        }

        function renderRoutes(routes) {
            const tbody = document.querySelector('#routesTable tbody');
            const entries = Object.entries(routes);
            if (entries.length === 0) {
                tbody.innerHTML = '<tr><td colspan="3" class="status">暂无路由规则</td></tr>';
                return;
            }
            tbody.innerHTML = entries.map(([pattern, provider]) => ` + "`" + `
                <tr>
                    <td><code>${pattern}</code></td>
                    <td>${provider}</td>
                    <td><button class="btn-danger" onclick="deleteRoute('${encodeURIComponent(pattern)}')">删除</button></td>
                </tr>
            ` + "`" + `).join('');
        }

        function updateProviderSelect(providers) {
            const select = document.getElementById('routeProvider');
            select.innerHTML = Object.keys(providers).map(name => 
                ` + "`" + `<option value="${name}">${name}</option>` + "`" + `
            ).join('');
        }

        async function generateKey() {
            const res = await fetch('/admin/api/keys/generate', { method: 'POST', headers });
            const data = await res.json();
            showToast('Key 已生成: ' + data.key);
            loadConfig();
        }

        async function deleteKey(key) {
            if (!confirm('确定删除此 Key？')) return;
            await fetch('/admin/api/keys/' + encodeURIComponent(key), { method: 'DELETE', headers });
            showToast('Key 已删除');
            loadConfig();
        }

        async function addProvider() {
            const name = document.getElementById('providerName').value.trim();
            const type = document.getElementById('providerType').value;
            const url = document.getElementById('providerUrl').value.trim();
            const keys = document.getElementById('providerKeys').value.split(',').map(k => k.trim()).filter(k => k);
            
            if (!name || !url || keys.length === 0) {
                showToast('请填写完整信息');
                return;
            }

            await fetch('/admin/api/providers', {
                method: 'POST',
                headers,
                body: JSON.stringify({ name, type, base_url: url, api_keys: keys, rotation_strategy: 'round_robin' })
            });
            showToast('Provider 已添加');
            document.getElementById('providerName').value = '';
            document.getElementById('providerUrl').value = '';
            document.getElementById('providerKeys').value = '';
            loadConfig();
        }

        async function deleteProvider(name) {
            if (!confirm('确定删除 Provider: ' + name + '？')) return;
            await fetch('/admin/api/providers/' + name, { method: 'DELETE', headers });
            showToast('Provider 已删除');
            loadConfig();
        }

        async function addRoute() {
            const pattern = document.getElementById('routePattern').value.trim();
            const provider = document.getElementById('routeProvider').value;
            
            if (!pattern || !provider) {
                showToast('请填写完整信息');
                return;
            }

            await fetch('/admin/api/routes', {
                method: 'POST',
                headers,
                body: JSON.stringify({ pattern, provider })
            });
            showToast('路由已添加');
            document.getElementById('routePattern').value = '';
            loadConfig();
        }

        async function deleteRoute(pattern) {
            if (!confirm('确定删除此路由？')) return;
            await fetch('/admin/api/routes/' + pattern, { method: 'DELETE', headers });
            showToast('路由已删除');
            loadConfig();
        }

        function copyKey(key) {
            navigator.clipboard.writeText(key);
            showToast('已复制到剪贴板');
        }

        function showToast(msg) {
            const toast = document.getElementById('toast');
            toast.textContent = msg;
            toast.classList.add('show');
            setTimeout(() => toast.classList.remove('show'), 2000);
        }

        loadConfig();
    </script>
</body>
</html>`
