const password = new URLSearchParams(window.location.search).get('password') || '';
const headers = { 'Content-Type': 'application/json', 'X-Admin-Password': password };

// 模型状态存储 (modelId -> status)
const modelStatus = new Map();

async function loadConfig() {
    try {
        const res = await fetch('/admin/api/config?password=' + password);
        const data = await res.json();
        renderKeys(data.client_api_keys || []);
        renderProviders(data.providers || {});
        loadModels();
    } catch (e) {
        console.error(e);
    }
}

async function loadModels() {
    try {
        const res = await fetch('/v1/models', {
            headers: { 'Authorization': 'Bearer ' + (document.querySelector('#keysTable tbody tr:first-child .key-display')?.textContent || 'sk-openbridge-key-1') }
        });
        const data = await res.json();
        window.currentModels = data.data || [];
        renderModels(window.currentModels);
    } catch (e) {
        console.error('Failed to load models:', e);
        document.getElementById('modelsContainer').innerHTML = '<div class="status">无法加载模型列表</div>';
    }
}

function renderModels(models) {
    const container = document.getElementById('modelsContainer');
    if (models.length === 0) {
        container.innerHTML = '<div class="status">暂无模型</div>';
        return;
    }
    container.innerHTML = models.map(m => {
        const status = modelStatus.get(m.id) || 'untested';
        const statusText = {
            'available': '✓ 可用',
            'unavailable': '✗ 不可用',
            'untested': '○ 未测试'
        }[status];
        
        return `
            <div class="model-card ${status}" onclick="openTestModal('${m.id}')">
                <div class="model-name">${m.id}</div>
                <div class="model-provider">by ${m.owned_by}</div>
                <div class="model-status">
                    <span class="status-dot ${status}"></span>
                    <span class="status-text ${status}">${statusText}</span>
                </div>
            </div>
        `;
    }).join('');
}

function openTestModal(modelId) {
    document.getElementById('testModelName').textContent = modelId;
    document.getElementById('testMessage').value = 'Hello!';
    document.getElementById('testMaxTokens').value = '100';
    document.getElementById('testOutput').style.display = 'none';
    document.getElementById('testModal').classList.add('show');
}

function closeTestModal() {
    document.getElementById('testModal').classList.remove('show');
}

async function runTest() {
    const modelId = document.getElementById('testModelName').textContent;
    const message = document.getElementById('testMessage').value;
    const maxTokens = parseInt(document.getElementById('testMaxTokens').value);
    const outputDiv = document.getElementById('testOutput');

    outputDiv.textContent = '发送中...';
    outputDiv.style.display = 'block';

    try {
        const clientKey = document.querySelector('#keysTable tbody tr:first-child .key-display')?.textContent || 'sk-openbridge-key-1';
        const res = await fetch('/v1/chat/completions', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + clientKey,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                model: modelId,
                messages: [{ role: 'user', content: message }],
                max_tokens: maxTokens
            })
        });

        const data = await res.json();
        if (res.ok) {
            // 标记为可用
            modelStatus.set(modelId, 'available');
            outputDiv.textContent = JSON.stringify(data, null, 2);
            showToast('✓ 模型可用');
        } else {
            // 标记为不可用
            modelStatus.set(modelId, 'unavailable');
            outputDiv.textContent = 'Error: ' + JSON.stringify(data, null, 2);
            showToast('✗ 模型不可用');
        }
        // 更新模型卡片
        renderModels(window.currentModels || []);
    } catch (e) {
        // 标记为不可用
        modelStatus.set(modelId, 'unavailable');
        outputDiv.textContent = 'Request failed: ' + e.message;
        showToast('✗ 请求失败');
        // 更新模型卡片
        renderModels(window.currentModels || []);
    }
}

function renderKeys(keys) {
    const tbody = document.querySelector('#keysTable tbody');
    if (keys.length === 0) {
        tbody.innerHTML = '<tr><td colspan="2" class="status">暂无 Key，点击上方按钮生成</td></tr>';
        return;
    }
    tbody.innerHTML = keys.map(key => `
        <tr>
            <td><span class="key-display">${key}</span><button class="btn-primary copy-btn" onclick="copyKey('${key}')">复制</button></td>
            <td><button class="btn-danger" onclick="deleteKey('${key}')">删除</button></td>
        </tr>
    `).join('');
}

function renderProviders(providers) {
    const tbody = document.querySelector('#providersTable tbody');
    const entries = Object.entries(providers);
    if (entries.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="status">暂无 Provider</td></tr>';
        return;
    }
    tbody.innerHTML = entries.map(([name, p]) => `
        <tr>
            <td><strong>${name}</strong></td>
            <td><span class="tag tag-${p.type}">${p.type}</span></td>
            <td style="font-size:13px">${p.base_url}</td>
            <td>${p.api_keys.map(k => '<span class="tag">' + k + '</span>').join('')}</td>
            <td><button class="btn-danger" onclick="deleteProvider('${name}')">删除</button></td>
        </tr>
    `).join('');
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
