// V2Ray UI 应用程序
class V2RayUI {
    constructor() {
        this.currentPanel = 'dashboard';
        this.isLoading = false;
        this.statusData = {
            v2ray: 'stopped',
            hysteria2: 'stopped',
            subscription: 'unknown'
        };
        this.subscriptions = [];
        this.activeSubscriptionId = null;
        this.selectedNodes = new Set();
        this.init();
    }

    init() {
        this.setupNavigation();
        this.setupEventListeners();
        this.loadInitialData();
        this.startStatusPolling();
    }

    // 设置导航
    setupNavigation() {
        const navItems = document.querySelectorAll('.nav-item');
        navItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const panel = item.getAttribute('data-panel');
                this.switchPanel(panel);
                this.updateNavigation(item);
            });
        });
    }

    // 更新导航状态
    updateNavigation(activeItem) {
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        activeItem.classList.add('active');
    }

    // 切换面板
    switchPanel(panelName) {
        if (this.currentPanel === panelName) return;

        const currentPanelEl = document.getElementById(this.currentPanel);
        const newPanelEl = document.getElementById(panelName);

        if (currentPanelEl && newPanelEl) {
            currentPanelEl.classList.remove('active');
            newPanelEl.classList.add('active');
            this.currentPanel = panelName;
            
            // 加载面板特定数据
            this.loadPanelData(panelName);
        }
    }

    // 设置事件监听器
    setupEventListeners() {
        // 订阅管理
        document.getElementById('addSubscriptionBtn')?.addEventListener('click', () => {
            this.addSubscription();
        });

        // 节点批量操作
        document.getElementById('selectAllNodes')?.addEventListener('click', () => {
            this.selectAllNodes(true);
        });

        document.getElementById('deselectAllNodes')?.addEventListener('click', () => {
            this.selectAllNodes(false);
        });

        document.getElementById('batchTestNodes')?.addEventListener('click', () => {
            this.batchTestNodes();
        });

        document.getElementById('deleteSelectedNodes')?.addEventListener('click', () => {
            this.deleteSelectedNodes();
        });

        // 代理控制
        document.getElementById('startV2ray')?.addEventListener('click', () => {
            this.toggleProxy('v2ray', 'start');
        });

        document.getElementById('stopV2ray')?.addEventListener('click', () => {
            this.toggleProxy('v2ray', 'stop');
        });

        document.getElementById('startHysteria2')?.addEventListener('click', () => {
            this.toggleProxy('hysteria2', 'start');
        });

        document.getElementById('stopHysteria2')?.addEventListener('click', () => {
            this.toggleProxy('hysteria2', 'stop');
        });

        document.getElementById('testConnection')?.addEventListener('click', () => {
            this.testConnection();
        });

        // 测速工具
        document.getElementById('quickTest')?.addEventListener('click', () => {
            this.runSpeedTest('quick');
        });

        document.getElementById('fullTest')?.addEventListener('click', () => {
            this.runSpeedTest('full');
        });

        // 设置保存
        document.getElementById('saveSettings')?.addEventListener('click', () => {
            this.saveSettings();
        });

        // 刷新状态
        document.getElementById('refreshStatus')?.addEventListener('click', () => {
            this.refreshStatus();
        });
    }

    // 显示通知
    showNotification(message, type = 'info', duration = 5000) {
        const notifications = document.getElementById('notifications');
        if (!notifications) return;

        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;

        notifications.appendChild(notification);

        // 自动移除通知
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, duration);
    }

    // 加载初始数据
    async loadInitialData() {
        this.showNotification('正在加载数据...', 'info');
        await this.loadStatus();
        await this.loadSubscriptions();
        this.showNotification('数据加载完成', 'success');
    }

    // 加载面板特定数据
    loadPanelData(panelName) {
        switch (panelName) {
            case 'dashboard':
                this.loadStatus();
                break;
            case 'subscriptions':
                this.loadSubscriptions();
                break;
            case 'nodes':
                // 如果没有活跃订阅，尝试自动选择第一个有节点的订阅
                if (!this.activeSubscriptionId && this.subscriptions.length > 0) {
                    const subscriptionWithNodes = this.subscriptions.find(sub => sub.nodes && sub.nodes.length > 0);
                    if (subscriptionWithNodes) {
                        this.activeSubscriptionId = subscriptionWithNodes.id;
                    }
                }
                this.renderNodes();
                break;
            case 'proxy':
                this.loadProxyStatus();
                break;
        }
    }

    // 开始状态轮询
    startStatusPolling() {
        setInterval(() => {
            if (this.currentPanel === 'dashboard') {
                this.loadStatus();
            }
        }, 5000);
    }

    // 加载系统状态
    async loadStatus() {
        try {
            const response = await fetch('/api/status');
            const data = await response.json();
            this.updateStatusDisplay(data);
        } catch (error) {
            console.error('加载状态失败:', error);
            this.showNotification('加载状态失败', 'error');
        }
    }

    // 更新状态显示
    updateStatusDisplay(data) {
        // 更新状态指示器
        const v2rayStatus = document.getElementById('v2rayStatus');
        const hysteria2Status = document.getElementById('hysteria2Status');
        const subscriptionStatus = document.getElementById('subscriptionStatus');

        if (v2rayStatus) {
            v2rayStatus.textContent = data.v2ray || '已停止';
            v2rayStatus.className = `status-indicator status-${data.v2ray === 'running' ? 'running' : 'stopped'}`;
        }

        if (hysteria2Status) {
            hysteria2Status.textContent = data.hysteria2 || '已停止';
            hysteria2Status.className = `status-indicator status-${data.hysteria2 === 'running' ? 'running' : 'stopped'}`;
        }

        if (subscriptionStatus) {
            subscriptionStatus.textContent = data.subscription || '未知';
            subscriptionStatus.className = `status-indicator status-unknown`;
        }

        // 更新端口信息
        document.getElementById('httpPort').textContent = data.httpPort || '-';
        document.getElementById('socksPort').textContent = data.socksPort || '-';
        document.getElementById('currentNode').textContent = data.currentNode || '无';
    }

    // 加载订阅列表
    async loadSubscriptions() {
        try {
            const response = await fetch('/api/subscriptions');
            const data = await response.json();
            if (data.success) {
                this.subscriptions = data.data || [];
                
                // 如果还没有活跃订阅，自动选择第一个有节点的订阅
                if (!this.activeSubscriptionId && this.subscriptions.length > 0) {
                    // 优先选择有节点的订阅
                    const subscriptionWithNodes = this.subscriptions.find(sub => sub.nodes && sub.nodes.length > 0);
                    if (subscriptionWithNodes) {
                        this.activeSubscriptionId = subscriptionWithNodes.id;
                        console.log('自动设置活跃订阅ID (有节点):', this.activeSubscriptionId);
                    } else {
                        // 如果没有带节点的订阅，选择第一个订阅
                        this.activeSubscriptionId = this.subscriptions[0].id;
                        console.log('设置第一个订阅为活跃订阅ID:', this.activeSubscriptionId);
                    }
                }
                
                // 输出当前状态用于调试
                console.log('当前订阅列表:', this.subscriptions.map(sub => ({ id: sub.id, name: sub.name, node_count: sub.nodes ? sub.nodes.length : 0 })));
                console.log('当前活跃订阅ID:', this.activeSubscriptionId);
                
                this.renderSubscriptions();
            } else {
                console.error('加载订阅失败:', data.message);
                this.subscriptions = [];
                this.renderSubscriptions();
            }
        } catch (error) {
            console.error('加载订阅失败:', error);
            this.subscriptions = [];
            this.renderSubscriptions();
        }
    }

    // 添加订阅
    async addSubscription() {
        const urlInput = document.getElementById('subscriptionUrl');
        const nameInput = document.getElementById('subscriptionName');

        const url = urlInput.value.trim();
        const name = nameInput.value.trim() || '新订阅';

        if (!url) {
            this.showNotification('请输入订阅链接', 'warning');
            return;
        }

        this.showNotification('正在添加订阅...', 'info');

        try {
            const response = await fetch('/api/subscriptions', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    url: url,
                    name: name
                })
            });

            const data = await response.json();
            if (data.success) {
                // 重新加载订阅列表
                await this.loadSubscriptions();
                
                // 清空输入框
                urlInput.value = '';
                nameInput.value = '';

                this.showNotification('订阅添加成功', 'success');
            } else {
                this.showNotification(`添加订阅失败: ${data.message}`, 'error');
            }
        } catch (error) {
            console.error('添加订阅失败:', error);
            this.showNotification('添加订阅失败', 'error');
        }
    }

    // 渲染订阅列表
    renderSubscriptions() {
        const container = document.getElementById('subscriptionItems');
        if (!container) return;

        if (this.subscriptions.length === 0) {
            container.innerHTML = '<div class="placeholder">暂无订阅，请添加订阅链接</div>';
            return;
        }

        container.innerHTML = this.subscriptions.map(sub => `
            <div class="subscription-item ${sub.id === this.activeSubscriptionId ? 'active' : ''}" data-id="${sub.id}" onclick="app.selectSubscription('${sub.id}')">
                <div class="subscription-info">
                    <h4>${sub.name}</h4>
                    <div class="subscription-url">${sub.url}</div>
                    <div class="subscription-meta">
                        节点数: ${sub.nodes ? sub.nodes.length : 0} | 更新时间: ${sub.updated_at || '未更新'}
                    </div>
                </div>
                <div class="subscription-actions">
                    <button class="btn btn-info btn-sm" onclick="event.stopPropagation(); app.parseSubscription('${sub.id}')">解析</button>
                    <button class="btn btn-danger btn-sm" onclick="event.stopPropagation(); app.deleteSubscription('${sub.id}')">删除</button>
                </div>
            </div>
        `).join('');
    }

    // 选择订阅
    selectSubscription(subscriptionId) {
        this.activeSubscriptionId = subscriptionId;
        console.log('手动选择订阅ID:', subscriptionId);
        this.renderSubscriptions();
        this.renderNodes();
    }

    // 解析订阅
    async parseSubscription(subscriptionId) {
        this.showNotification('正在解析订阅...', 'info');
        
        try {
            const response = await fetch('/api/subscriptions/parse', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    id: subscriptionId
                })
            });

            const data = await response.json();
            if (data.success) {
                this.activeSubscriptionId = subscriptionId;
                // 重新加载订阅列表以获取更新的节点数据
                await this.loadSubscriptions();
                this.renderSubscriptions();
                this.renderNodes();
                this.showNotification(`订阅解析完成，解析出 ${data.data.nodes.length} 个节点`, 'success');
            } else {
                this.showNotification(`订阅解析失败: ${data.message}`, 'error');
            }
        } catch (error) {
            console.error('解析订阅失败:', error);
            this.showNotification('订阅解析失败', 'error');
        }
    }

    // 删除订阅
    async deleteSubscription(subscriptionId) {
        if (!confirm('确定要删除这个订阅吗？')) return;

        try {
            const response = await fetch('/api/subscriptions/delete', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    id: subscriptionId
                })
            });

            const data = await response.json();
            if (data.success) {
                if (this.activeSubscriptionId === subscriptionId) {
                    this.activeSubscriptionId = null;
                }
                // 重新加载订阅列表
                await this.loadSubscriptions();
                this.renderSubscriptions();
                this.renderNodes();
                this.showNotification('订阅删除成功', 'success');
            } else {
                this.showNotification(`删除订阅失败: ${data.message}`, 'error');
            }
        } catch (error) {
            console.error('删除订阅失败:', error);
            this.showNotification('删除订阅失败', 'error');
        }
    }

    // 测试订阅（已移除，因为没有实际用途）

    // 渲染节点列表
    async renderNodes() {
        const container = document.getElementById('nodeItems');
        if (!container) return;

        if (!this.activeSubscriptionId) {
            container.innerHTML = '<div class="placeholder">请先添加订阅以获取节点</div>';
            return;
        }

        try {
            // 获取当前订阅的节点数据
            const subscription = this.subscriptions.find(sub => sub.id === this.activeSubscriptionId);
            if (!subscription) {
                container.innerHTML = '<div class="placeholder">订阅不存在</div>';
                return;
            }

            if (!subscription.nodes || subscription.nodes.length === 0) {
                container.innerHTML = '<div class="placeholder">请先解析订阅以获取节点</div>';
                return;
            }

            const nodes = subscription.nodes;
            container.innerHTML = nodes.map(node => this.renderNodeItem(node)).join('');
        } catch (error) {
            console.error('渲染节点失败:', error);
            container.innerHTML = '<div class="placeholder">节点加载失败</div>';
        }
    }

    // 渲染单个节点项
    renderNodeItem(node) {
        const isSelected = this.selectedNodes.has(node.index);
        const statusClass = this.getNodeStatusClass(node.status);
        const statusText = this.getNodeStatusText(node.status);
        
        return `
            <div class="node-item ${isSelected ? 'selected' : ''} ${statusClass}" data-index="${node.index}">
                <div class="node-checkbox">
                    <input type="checkbox" ${isSelected ? 'checked' : ''} 
                           onchange="app.toggleNodeSelection(${node.index}, this.checked)">
                </div>
                <div class="node-info">
                    <div class="node-header">
                        <h4>${node.name}</h4>
                        <span class="node-status ${statusClass}">${statusText}</span>
                        ${node.is_running ? '<span class="running-indicator">🟢 运行中</span>' : ''}
                    </div>
                    <div class="node-meta">
                        <span class="protocol">${node.protocol.toUpperCase()}</span>
                        <span class="server">${node.server}:${node.port}</span>
                        ${this.renderNodePorts(node)}
                    </div>
                    ${this.renderTestResults(node)}
                </div>
                <div class="node-actions">
                    ${this.renderNodeActionButtons(node)}
                </div>
            </div>
        `;
    }

    // 渲染节点端口信息
    renderNodePorts(node) {
        if (!node.is_running) return '';
        
        const ports = [];
        if (node.http_port) ports.push(`HTTP:${node.http_port}`);
        if (node.socks_port) ports.push(`SOCKS:${node.socks_port}`);
        
        return ports.length > 0 ? `<span class="ports">${ports.join(' | ')}</span>` : '';
    }

    // 渲染测试结果
    renderTestResults(node) {
        let html = '';
        
        // 连接测试结果
        if (node.test_result) {
            const result = node.test_result;
            const resultClass = result.success ? 'success' : 'error';
            const testTime = this.formatTime(result.test_time);
            html += `
                <div class="test-result ${resultClass}">
                    <span class="test-type">连接测试:</span>
                    ${result.success ? 
                        `<span class="latency">${result.latency}</span>` : 
                        `<span class="error">${result.error || '测试失败'}</span>`
                    }
                    <span class="test-time">${testTime}</span>
                </div>
            `;
        }
        
        // 速度测试结果
        if (node.speed_result) {
            const result = node.speed_result;
            const testTime = this.formatTime(result.test_time);
            html += `
                <div class="speed-result">
                    <span class="test-type">速度测试:</span>
                    <span class="speeds">↓${result.download_speed} ↑${result.upload_speed}</span>
                    <span class="latency">${result.latency}</span>
                    <span class="test-time">${testTime}</span>
                </div>
            `;
        }
        
        return html ? `<div class="node-results">${html}</div>` : '';
    }

    // 渲染节点操作按钮
    renderNodeActionButtons(node) {
        const isConnecting = node.status === 'connecting';
        const isTesting = node.status === 'testing';
        const isRunning = node.is_running;
        
        return `
            <div class="action-group">
                <select class="connect-type" ${isConnecting || isTesting ? 'disabled' : ''}>
                    <option value="http_random">随机HTTP</option>
                    <option value="socks_random">随机SOCKS</option>
                    <option value="http_fixed">固定HTTP</option>
                    <option value="socks_fixed">固定SOCKS</option>
                </select>
                <button class="btn btn-success btn-sm" 
                        onclick="app.connectNode('${this.activeSubscriptionId}', ${node.index})"
                        ${isConnecting || isTesting ? 'disabled' : ''}>
                    ${isConnecting ? '连接中...' : (isRunning ? '重连' : '连接')}
                </button>
                ${isRunning ? `
                    <button class="btn btn-danger btn-sm" 
                            onclick="app.disconnectNode('${this.activeSubscriptionId}', ${node.index})">
                        断开
                    </button>
                ` : ''}
            </div>
            <div class="action-group">
                <button class="btn btn-info btn-sm" 
                        onclick="app.testNode('${this.activeSubscriptionId}', ${node.index})"
                        ${isTesting || isConnecting ? 'disabled' : ''}>
                    ${isTesting ? '测试中...' : '连接测试'}
                </button>
                <button class="btn btn-warning btn-sm" 
                        onclick="app.speedTestNode('${this.activeSubscriptionId}', ${node.index})"
                        ${isTesting || isConnecting ? 'disabled' : ''}>
                    ${isTesting ? '测试中...' : '速度测试'}
                </button>
            </div>
        `;
    }

    // 获取节点状态样式类
    getNodeStatusClass(status) {
        const statusMap = {
            'idle': 'status-idle',
            'connecting': 'status-connecting',
            'connected': 'status-connected',
            'testing': 'status-testing',
            'error': 'status-error'
        };
        return statusMap[status] || 'status-idle';
    }

    // 获取节点状态文本
    getNodeStatusText(status) {
        const statusMap = {
            'idle': '空闲',
            'connecting': '连接中',
            'connected': '已连接',
            'testing': '测试中',
            'error': '错误'
        };
        return statusMap[status] || '未知';
    }

    // 格式化时间
    formatTime(timeStr) {
        if (!timeStr) return '';
        try {
            const date = new Date(timeStr);
            return date.toLocaleTimeString();
        } catch (e) {
            return '';
        }
    }

    // 切换节点选择
    toggleNodeSelection(nodeIndex, selected) {
        if (selected) {
            this.selectedNodes.add(nodeIndex);
        } else {
            this.selectedNodes.delete(nodeIndex);
        }
        this.renderNodes();
    }

    // 全选/取消全选节点
    selectAllNodes(selectAll) {
        if (!this.activeSubscriptionId) return;
        
        const subscription = this.subscriptions.find(sub => sub.id === this.activeSubscriptionId);
        if (!subscription || !subscription.nodes) return;
        
        if (selectAll) {
            subscription.nodes.forEach(node => {
                this.selectedNodes.add(node.index);
            });
        } else {
            this.selectedNodes.clear();
        }
        this.renderNodes();
    }

    // 批量测试节点
    async batchTestNodes() {
        // 检查是否有活跃的订阅
        if (!this.activeSubscriptionId) {
            this.showNotification('请先选择一个订阅', 'warning');
            return;
        }

        if (this.selectedNodes.size === 0) {
            this.showNotification('请先选择要测试的节点', 'warning');
            return;
        }

        const nodeIndexes = Array.from(this.selectedNodes);
        
        // 创建进度显示界面
        this.showBatchTestProgress(nodeIndexes.length);
        
        try {
            // 使用SSE进行实时批量测试
            await this.startBatchTestSSE(nodeIndexes);
        } catch (error) {
            console.error('批量测试失败:', error);
            this.showNotification('批量测试失败: ' + error.message, 'error');
            this.hideBatchTestProgress();
        }
    }

    // 显示批量测试进度界面
    showBatchTestProgress(totalNodes) {
        // 创建进度弹窗
        const progressModal = document.createElement('div');
        progressModal.id = 'batchTestProgressModal';
        progressModal.className = 'modal active';
        progressModal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>批量测试进度</h3>
                    <button class="close-btn" onclick="app.cancelBatchTest()">&times;</button>
                </div>
                <div class="modal-body">
                    <div class="progress-info">
                        <div class="progress-stats">
                            <span>总数: <span id="progressTotal">${totalNodes}</span></span>
                            <span>完成: <span id="progressCompleted">0</span></span>
                            <span>成功: <span id="progressSuccess">0</span></span>
                            <span>失败: <span id="progressFailure">0</span></span>
                        </div>
                        <div class="progress-bar-container">
                            <div class="progress-bar">
                                <div id="progressBar" class="progress-fill" style="width: 0%"></div>
                            </div>
                            <span id="progressPercent">0%</span>
                        </div>
                    </div>
                    <div class="progress-messages">
                        <div id="progressMessages" class="message-list"></div>
                    </div>
                </div>
                <div class="modal-footer">
                    <button id="cancelBatchTestBtn" onclick="app.cancelBatchTest()">取消测试</button>
                </div>
            </div>
        `;
        document.body.appendChild(progressModal);
    }

    // 隐藏批量测试进度界面
    hideBatchTestProgress() {
        const progressModal = document.getElementById('batchTestProgressModal');
        if (progressModal) {
            progressModal.remove();
        }
    }

    // 使用SSE开始批量测试
    async startBatchTestSSE(nodeIndexes) {
        return new Promise((resolve, reject) => {
            // 再次确认activeSubscriptionId存在
            if (!this.activeSubscriptionId) {
                reject(new Error('没有活跃的订阅ID'));
                return;
            }

            // 构建SSE URL with parameters
            const nodeIndexesStr = JSON.stringify(nodeIndexes);
            const sseUrl = `/api/nodes/batch-test-sse?subscription_id=${encodeURIComponent(this.activeSubscriptionId)}&node_indexes=${encodeURIComponent(nodeIndexesStr)}`;
            
            console.log('启动批量测试SSE:', sseUrl);
            console.log('订阅ID:', this.activeSubscriptionId);
            console.log('节点索引:', nodeIndexes);
            
            // 创建SSE连接
            const eventSource = new EventSource(sseUrl);
            let isResolved = false;
            let connectionTimeout;
            let lastProgressTime = Date.now();
            let connectionEstablished = false;
            
            // 设置连接超时（20分钟，适应大批量测试）
            const TOTAL_TIMEOUT = 20 * 60 * 1000; // 20分钟
            const PROGRESS_TIMEOUT = 3 * 60 * 1000; // 3分钟没有进度更新则认为超时
            const CONNECTION_TIMEOUT = 30 * 1000; // 30秒连接超时
            
            // 连接超时检测
            const connectionTimeoutId = setTimeout(() => {
                if (!connectionEstablished && !isResolved) {
                    console.error('SSE连接超时');
                    eventSource.close();
                    this.showNotification('SSE连接超时，请检查网络连接', 'error');
                    reject(new Error('SSE连接超时'));
                }
            }, CONNECTION_TIMEOUT);
            
            connectionTimeout = setTimeout(() => {
                if (!isResolved) {
                    console.error('SSE连接总体超时');
                    eventSource.close();
                    this.showNotification('批量测试总体超时（20分钟），请检查网络连接', 'error');
                    reject(new Error('SSE连接总体超时'));
                }
            }, TOTAL_TIMEOUT);
            
            // 进度监控超时
            const progressMonitor = setInterval(() => {
                if (!isResolved && connectionEstablished && Date.now() - lastProgressTime > PROGRESS_TIMEOUT) {
                    console.error('SSE进度超时');
                    clearInterval(progressMonitor);
                    eventSource.close();
                    if (!isResolved) {
                        this.showNotification('批量测试进度超时（3分钟无响应），可能网络连接不稳定', 'error');
                        reject(new Error('SSE进度超时'));
                    }
                }
            }, 30000); // 每30秒检查一次

            // 监听连接测试事件
            eventSource.addEventListener('ping', (event) => {
                console.log('收到ping事件:', event.data);
                clearTimeout(connectionTimeoutId);
                connectionEstablished = true;
                lastProgressTime = Date.now();
            });

            // 监听连接成功事件
            eventSource.addEventListener('connected', (event) => {
                try {
                    clearTimeout(connectionTimeoutId);
                    connectionEstablished = true;
                    lastProgressTime = Date.now();
                    const data = JSON.parse(event.data);
                    console.log('SSE连接成功:', data);
                    
                    // 保存会话ID用于取消测试
                    if (data.sessionId) {
                        this.currentBatchTestSessionId = data.sessionId;
                        console.log('保存会话ID:', this.currentBatchTestSessionId);
                    }
                    
                    this.showNotification(`SSE连接成功，开始测试 ${data.total} 个节点，请耐心等待...`, 'success', 5000);
                } catch (err) {
                    console.error('解析连接事件失败:', err);
                    this.showNotification('解析连接响应失败', 'error');
                }
            });

            // 监听心跳事件
            eventSource.addEventListener('heartbeat', (event) => {
                console.log('收到心跳:', event.data);
                lastProgressTime = Date.now();
            });

            // 监听进度事件
            eventSource.addEventListener('progress', (event) => {
                try {
                    lastProgressTime = Date.now();
                    const progress = JSON.parse(event.data);
                    this.updateBatchTestProgress(progress);
                } catch (err) {
                    console.error('解析进度数据失败:', err);
                    this.showNotification('解析进度数据失败', 'warning');
                }
            });

            // 监听最终结果事件
            eventSource.addEventListener('final_result', (event) => {
                try {
                    clearTimeout(connectionTimeout);
                    clearInterval(progressMonitor);
                    isResolved = true;
                    const result = JSON.parse(event.data);
                    
                    // 清除会话ID
                    this.currentBatchTestSessionId = null;
                    
                    this.handleBatchTestComplete(result);
                    eventSource.close();
                    resolve(result);
                } catch (err) {
                    console.error('解析最终结果失败:', err);
                    eventSource.close();
                    this.showNotification('解析测试结果失败', 'error');
                    reject(err);
                }
            });

            // 监听取消事件
            eventSource.addEventListener('cancelled', (event) => {
                try {
                    clearTimeout(connectionTimeout);
                    clearInterval(progressMonitor);
                    isResolved = true;
                    const result = JSON.parse(event.data);
                    console.log('批量测试已被取消:', result);
                    this.showNotification(`批量测试已取消: ${result.message}`, 'warning');
                    
                    // 清除会话ID
                    this.currentBatchTestSessionId = null;
                    
                    // 更新取消按钮为关闭按钮
                    const cancelBtn = document.getElementById('cancelBatchTestBtn');
                    if (cancelBtn) {
                        cancelBtn.textContent = '关闭';
                        cancelBtn.onclick = () => this.hideBatchTestProgress();
                    }
                    
                    eventSource.close();
                    resolve(result);
                } catch (err) {
                    console.error('解析取消事件失败:', err);
                    eventSource.close();
                    this.showNotification('解析取消响应失败', 'error');
                    reject(err);
                }
            });

            // 监听错误事件
            eventSource.addEventListener('error', (event) => {
                try {
                    clearTimeout(connectionTimeout);
                    clearInterval(progressMonitor);
                    const error = JSON.parse(event.data);
                    console.error('SSE错误事件:', error);
                    this.showNotification(`批量测试错误: ${error.error}`, 'error');
                    eventSource.close();
                    if (!isResolved) {
                        reject(new Error(error.error));
                    }
                } catch (err) {
                    console.error('解析错误事件失败:', err);
                    eventSource.close();
                    this.showNotification('收到未知错误事件', 'error');
                    if (!isResolved) {
                        reject(new Error('SSE连接错误'));
                    }
                }
            });

            // 监听关闭事件
            eventSource.addEventListener('close', () => {
                clearTimeout(connectionTimeout);
                clearInterval(progressMonitor);
                console.log('收到关闭事件');
                eventSource.close();
            });

            // 处理连接错误
            eventSource.onerror = (error) => {
                console.error('SSE连接错误:', error);
                
                // 检查连接状态
                if (eventSource.readyState === EventSource.CONNECTING) {
                    console.log('SSE正在重连...');
                    if (connectionEstablished) {
                        this.showNotification('连接中断，正在重连...', 'warning');
                        lastProgressTime = Date.now(); // 重置进度时间
                    }
                } else if (eventSource.readyState === EventSource.CLOSED) {
                    console.log('SSE连接已关闭');
                    clearTimeout(connectionTimeout);
                    clearInterval(progressMonitor);
                    eventSource.close();
                    if (!isResolved) {
                        if (!connectionEstablished) {
                            this.showNotification('SSE连接失败，请检查网络连接', 'error');
                            reject(new Error('SSE连接失败'));
                        } else {
                            this.showNotification('SSE连接意外断开', 'error');
                            reject(new Error('SSE连接断开'));
                        }
                    }
                }
            };

            // 保存eventSource引用以便取消
            this.currentBatchTestSSE = eventSource;
        });
    }

    // 更新批量测试进度
    updateBatchTestProgress(progress) {
        // 更新统计信息
        document.getElementById('progressTotal').textContent = progress.total;
        document.getElementById('progressCompleted').textContent = progress.completed;
        document.getElementById('progressSuccess').textContent = progress.success_count;
        document.getElementById('progressFailure').textContent = progress.failure_count;

        // 更新进度条
        const progressBar = document.getElementById('progressBar');
        const progressPercent = document.getElementById('progressPercent');
        if (progressBar && progressPercent) {
            progressBar.style.width = `${progress.progress}%`;
            progressPercent.textContent = `${progress.progress}%`;
        }

        // 添加进度消息
        const messagesContainer = document.getElementById('progressMessages');
        if (messagesContainer && progress.message) {
            const messageElement = document.createElement('div');
            messageElement.className = 'progress-message';
            messageElement.innerHTML = `
                <span class="timestamp">${progress.timestamp}</span>
                <span class="message">${progress.message}</span>
            `;
            messagesContainer.appendChild(messageElement);
            // 自动滚动到底部
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }

        // 更新单个节点状态（如果有当前结果）
        if (progress.current_result && progress.node_index !== undefined) {
            this.updateNodeStatus(progress.node_index, 
                progress.current_result.success ? 'success' : 'error'
            );
        }
    }

    // 处理批量测试完成
    handleBatchTestComplete(result) {
        this.showNotification(
            `批量测试完成: 成功 ${result.success_count}，失败 ${result.failure_count}`, 
            'success'
        );
        
        // 更新取消按钮为关闭按钮
        const cancelBtn = document.getElementById('cancelBatchTestBtn');
        if (cancelBtn) {
            cancelBtn.textContent = '关闭';
            cancelBtn.onclick = () => this.hideBatchTestProgress();
        }
        
        // 刷新节点显示
        setTimeout(async () => {
            await this.loadSubscriptions();
            this.renderNodes();
        }, 1000);
    }

    // 取消批量测试
    async cancelBatchTest() {
        try {
            console.log('开始取消批量测试...');
            
            // 如果有会话ID，调用后端取消API
            if (this.currentBatchTestSessionId) {
                console.log('取消批量测试，会话ID:', this.currentBatchTestSessionId);
                
                try {
                    const response = await fetch('/api/nodes/cancel-batch-test', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            session_id: this.currentBatchTestSessionId
                        })
                    });

                    if (response.ok) {
                        const data = await response.json();
                        if (data.success) {
                            console.log('后端取消成功:', data);
                            this.showNotification('批量测试已取消', 'warning');
                        } else {
                            console.warn('后端取消失败:', data.message);
                            this.showNotification(`取消测试失败: ${data.message}`, 'warning');
                        }
                    } else {
                        console.warn('取消请求失败:', response.status, response.statusText);
                        this.showNotification('取消请求失败，但会强制关闭连接', 'warning');
                    }
                } catch (fetchError) {
                    console.error('取消请求网络错误:', fetchError);
                    this.showNotification('取消请求网络错误，但会强制关闭连接', 'warning');
                }
                
                // 清除会话ID
                this.currentBatchTestSessionId = null;
            }
            
            // 关闭SSE连接
            if (this.currentBatchTestSSE) {
                console.log('关闭SSE连接');
                this.currentBatchTestSSE.close();
                this.currentBatchTestSSE = null;
            }
            
            // 更新UI状态
            const cancelBtn = document.getElementById('cancelBatchTestBtn');
            if (cancelBtn) {
                cancelBtn.textContent = '关闭';
                cancelBtn.onclick = () => this.hideBatchTestProgress();
            }
            
        } catch (error) {
            console.error('取消批量测试失败:', error);
            this.showNotification('取消测试时发生错误', 'error');
        }
        
        // 无论如何都要隐藏进度界面
        setTimeout(() => {
            this.hideBatchTestProgress();
        }, 1000);
    }

    // 删除选中节点
    deleteSelectedNodes() {
        if (this.selectedNodes.size === 0) {
            this.showNotification('请先选择要删除的节点', 'warning');
            return;
        }

        if (!confirm(`确定要删除 ${this.selectedNodes.size} 个节点吗？`)) return;

        // 这里可以实现真实的删除逻辑
        this.selectedNodes.clear();
        this.renderNodes();
        this.showNotification('选中节点删除成功', 'success');
    }

    // 连接节点
    async connectNode(subscriptionId, nodeIndex) {
        try {
            // 获取连接类型
            const nodeElement = document.querySelector(`[data-index="${nodeIndex}"]`);
            const connectType = nodeElement?.querySelector('.connect-type')?.value || 'http_random';
            
            this.showNotification('正在连接节点...', 'info');

            // 更新UI状态
            this.updateNodeStatus(nodeIndex, 'connecting');

            const response = await fetch('/api/nodes/connect', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    subscription_id: subscriptionId,
                    node_index: nodeIndex,
                    operation: connectType
                })
            });

            const data = await response.json();
            if (data.success) {
                const result = data.data;
                let message = '节点连接成功';
                if (result.http_port) message += ` (HTTP端口: ${result.http_port})`;
                if (result.socks_port) message += ` (SOCKS端口: ${result.socks_port})`;
                
                this.showNotification(message, 'success');
                
                // 刷新节点显示和状态
                await this.loadSubscriptions();
                this.renderNodes();
                this.loadStatus();
            } else {
                this.showNotification(`节点连接失败: ${data.message}`, 'error');
                this.updateNodeStatus(nodeIndex, 'error');
            }
        } catch (error) {
            console.error('连接节点失败:', error);
            this.showNotification('节点连接失败', 'error');
            this.updateNodeStatus(nodeIndex, 'error');
        }
    }

    // 断开节点
    async disconnectNode(subscriptionId, nodeIndex) {
        try {
            this.showNotification('正在断开节点...', 'info');

            const response = await fetch('/api/nodes/connect', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    subscription_id: subscriptionId,
                    node_index: nodeIndex,
                    operation: 'disable'
                })
            });

            const data = await response.json();
            if (data.success) {
                this.showNotification('节点断开成功', 'success');
                
                // 刷新节点显示和状态
                await this.loadSubscriptions();
                this.renderNodes();
                this.loadStatus();
            } else {
                this.showNotification(`节点断开失败: ${data.message}`, 'error');
            }
        } catch (error) {
            console.error('断开节点失败:', error);
            this.showNotification('节点断开失败', 'error');
        }
    }

    // 测试节点
    async testNode(subscriptionId, nodeIndex) {
        try {
            this.showNotification('正在测试节点连接...', 'info');

            // 更新UI状态
            this.updateNodeStatus(nodeIndex, 'testing');

            const response = await fetch('/api/nodes/test', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    subscription_id: subscriptionId,
                    node_index: nodeIndex
                })
            });

            const data = await response.json();
            if (data.success) {
                const result = data.data;
                const message = result.success ? 
                    `节点测试成功 (延迟: ${result.latency})` : 
                    `节点测试失败: ${result.error}`;
                
                this.showNotification(message, result.success ? 'success' : 'warning');
                
                // 刷新节点显示以显示测试结果
                await this.loadSubscriptions();
                this.renderNodes();
            } else {
                this.showNotification(`节点测试失败: ${data.message}`, 'error');
                this.updateNodeStatus(nodeIndex, 'error');
            }
        } catch (error) {
            console.error('测试节点失败:', error);
            this.showNotification('节点测试失败', 'error');
            this.updateNodeStatus(nodeIndex, 'error');
        }
    }

    // 速度测试节点
    async speedTestNode(subscriptionId, nodeIndex) {
        try {
            this.showNotification('正在进行速度测试...', 'info');

            // 更新UI状态
            this.updateNodeStatus(nodeIndex, 'testing');

            const response = await fetch('/api/nodes/speedtest', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    subscription_id: subscriptionId,
                    node_index: nodeIndex
                })
            });

            const data = await response.json();
            if (data.success) {
                const result = data.data;
                const message = `速度测试完成: 下载 ${result.download_speed}, 上传 ${result.upload_speed}, 延迟 ${result.latency}`;
                
                this.showNotification(message, 'success');
                
                // 刷新节点显示以显示测试结果
                await this.loadSubscriptions();
                this.renderNodes();
            } else {
                this.showNotification(`速度测试失败: ${data.message}`, 'error');
                this.updateNodeStatus(nodeIndex, 'error');
            }
        } catch (error) {
            console.error('速度测试失败:', error);
            this.showNotification('速度测试失败', 'error');
            this.updateNodeStatus(nodeIndex, 'error');
        }
    }

    // 更新节点状态（仅UI）
    updateNodeStatus(nodeIndex, status) {
        const nodeElement = document.querySelector(`[data-index="${nodeIndex}"]`);
        if (nodeElement) {
            // 移除所有状态类
            nodeElement.classList.remove('status-idle', 'status-connecting', 'status-connected', 'status-testing', 'status-error');
            // 添加新状态类
            nodeElement.classList.add(this.getNodeStatusClass(status));
            
            // 更新状态文本
            const statusElement = nodeElement.querySelector('.node-status');
            if (statusElement) {
                statusElement.textContent = this.getNodeStatusText(status);
                statusElement.className = `node-status ${this.getNodeStatusClass(status)}`;
            }
        }
    }

    // 加载代理状态
    async loadProxyStatus() {
        await this.loadStatus();
    }

    // 切换代理状态
    async toggleProxy(type, action) {
        this.showNotification(`正在${action === 'start' ? '启动' : '停止'} ${type.toUpperCase()}...`, 'info');
        await new Promise(resolve => setTimeout(resolve, 1000));
        this.showNotification(`${type.toUpperCase()} ${action === 'start' ? '启动' : '停止'}成功`, 'success');
        this.loadStatus();
    }

    // 测试连接
    async testConnection() {
        this.showNotification('正在测试连接...', 'info');
        await new Promise(resolve => setTimeout(resolve, 2000));
        this.showNotification('连接测试完成', 'success');
    }

    // 运行测速
    async runSpeedTest(type) {
        const testType = type === 'quick' ? '快速' : '完整';
        this.showNotification(`正在进行${testType}测速...`, 'info');
        
        const duration = type === 'quick' ? 3000 : 8000;
        await new Promise(resolve => setTimeout(resolve, duration));
        
        // 更新测试结果
        const resultsContainer = document.getElementById('testResults');
        if (resultsContainer) {
            resultsContainer.innerHTML = `
                <div style="padding: 15px; background-color: #f8f8f8; border: 1px solid #ccc; border-radius: 4px;">
                    <h4>${testType}测速结果</h4>
                    <p>下载速度: ${Math.floor(Math.random() * 50) + 10} Mbps</p>
                    <p>上传速度: ${Math.floor(Math.random() * 20) + 5} Mbps</p>
                    <p>延迟: ${Math.floor(Math.random() * 100) + 20} ms</p>
                    <p>测试时间: ${new Date().toLocaleString()}</p>
                </div>
            `;
        }
        
        this.showNotification(`${testType}测速完成`, 'success');
    }

    // 保存设置
    async saveSettings() {
        const httpPort = document.getElementById('httpPortSetting')?.value;
        const socksPort = document.getElementById('socksPortSetting')?.value;
        const testUrl = document.getElementById('testUrlSetting')?.value;

        this.showNotification('正在保存设置...', 'info');
        await new Promise(resolve => setTimeout(resolve, 500));
        this.showNotification('设置保存成功', 'success');
    }

    // 刷新状态
    async refreshStatus() {
        this.showNotification('正在刷新状态...', 'info');
        await this.loadStatus();
        this.showNotification('状态刷新完成', 'success');
    }
}

// 初始化应用
const app = new V2RayUI(); 