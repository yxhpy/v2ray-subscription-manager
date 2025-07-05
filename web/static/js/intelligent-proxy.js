// 智能代理管理JavaScript

class IntelligentProxyManager {
    constructor() {
        this.isRunning = false;
        this.eventSource = null;
        this.statusTimer = null;
        this.autoSwitchEnabled = true;
        
        this.initializeEventListeners();
        this.loadInitialStatus();
    }

    // 初始化事件监听器
    initializeEventListeners() {
        // 启动表单
        const startForm = document.getElementById('startForm');
        if (startForm) {
            startForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.startIntelligentProxy();
            });
        }

        // 停止按钮
        const stopBtn = document.getElementById('stopBtn');
        if (stopBtn) {
            stopBtn.addEventListener('click', () => {
                this.stopIntelligentProxy();
            });
        }

        // 重新测试按钮
        const retestBtn = document.getElementById('retestBtn');
        if (retestBtn) {
            retestBtn.addEventListener('click', () => {
                this.forceRetestAllNodes();
            });
        }

        // 切换自动切换按钮
        const toggleAutoSwitchBtn = document.getElementById('toggleAutoSwitchBtn');
        if (toggleAutoSwitchBtn) {
            toggleAutoSwitchBtn.addEventListener('click', () => {
                this.toggleAutoSwitch();
            });
        }

        // 清空日志按钮
        const clearLogBtn = document.getElementById('clearLogBtn');
        if (clearLogBtn) {
            clearLogBtn.addEventListener('click', () => {
                this.clearEventLog();
            });
        }
    }

    // 加载初始状态
    async loadInitialStatus() {
        try {
            await this.updateStatus();
        } catch (error) {
            console.error('加载初始状态失败:', error);
        }
    }

    // 启动智能代理
    async startIntelligentProxy() {
        const subscriptionId = document.getElementById('subscriptionSelect').value;
        if (!subscriptionId) {
            this.showAlert('请选择订阅', 'warning');
            return;
        }

        const config = this.getConfigFromForm();
        
        try {
            const response = await fetch('/api/intelligent-proxy/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    subscription_id: subscriptionId,
                    config: config
                })
            });

            const result = await response.json();
            
            if (result.success) {
                this.showAlert('智能代理启动成功', 'success');
                this.isRunning = true;
                this.updateUI();
                this.startEventStream();
                this.startStatusUpdater();
            } else {
                this.showAlert(`启动失败: ${result.error}`, 'danger');
            }
        } catch (error) {
            console.error('启动智能代理失败:', error);
            this.showAlert('启动失败，请检查网络连接', 'danger');
        }
    }

    // 停止智能代理
    async stopIntelligentProxy() {
        try {
            const response = await fetch('/api/intelligent-proxy/stop', {
                method: 'POST'
            });

            const result = await response.json();
            
            if (result.success) {
                this.showAlert('智能代理停止成功', 'success');
                this.isRunning = false;
                this.updateUI();
                this.stopEventStream();
                this.stopStatusUpdater();
                this.clearStatus();
            } else {
                this.showAlert(`停止失败: ${result.error}`, 'danger');
            }
        } catch (error) {
            console.error('停止智能代理失败:', error);
            this.showAlert('停止失败，请检查网络连接', 'danger');
        }
    }

    // 从表单获取配置
    getConfigFromForm() {
        // 获取队列大小配置，确保不为0或负数
        let maxQueueSize = parseInt(document.getElementById('maxQueueSize')?.value) || 50;
        if (maxQueueSize <= 0) {
            maxQueueSize = 50; // 确保使用合理的默认值
        }
        
        return {
            test_concurrency: parseInt(document.getElementById('testConcurrency').value) || 10,
            test_interval: parseInt(document.getElementById('testInterval').value) || 30,
            health_check_interval: parseInt(document.getElementById('healthCheckInterval').value) || 60,
            test_timeout: parseInt(document.getElementById('testTimeout').value) || 30,
            test_url: document.getElementById('testURL').value || 'https://www.google.com',
            switch_threshold: parseInt(document.getElementById('switchThreshold').value) || 100,
            max_queue_size: maxQueueSize,
            http_port: parseInt(document.getElementById('httpPort').value) || 7890,
            socks_port: parseInt(document.getElementById('socksPort').value) || 7891,
            enable_auto_switch: document.getElementById('enableAutoSwitch').checked,
            enable_retesting: document.getElementById('enableRetesting').checked,
            enable_health_check: document.getElementById('enableHealthCheck').checked
        };
    }

    // 更新状态
    async updateStatus() {
        try {
            const response = await fetch('/api/intelligent-proxy/status');
            const result = await response.json();
            
            if (result.success && result.data) {
                this.updateStatusDisplay(result.data);
                this.updateQueueDisplay(result.data.queue || []);
                this.isRunning = result.data.is_running || false;
                this.autoSwitchEnabled = result.data.config?.enable_auto_switch || false;
                this.updateUI();
            }
        } catch (error) {
            console.error('更新状态失败:', error);
        }
    }

    // 更新状态显示
    updateStatusDisplay(status) {
        // 运行状态
        const runningStatusEl = document.getElementById('runningStatus');
        const statusPanel = document.getElementById('statusPanel');
        if (runningStatusEl) {
            runningStatusEl.textContent = status.is_running ? '运行中' : '已停止';
            
            // 更新状态卡片样式
            const statusCards = statusPanel.querySelectorAll('.status-card');
            statusCards.forEach(card => {
                card.classList.remove('running', 'stopped');
                card.classList.add(status.is_running ? 'running' : 'stopped');
            });
        }

        // 当前节点
        const currentNodeEl = document.getElementById('currentNode');
        if (currentNodeEl) {
            if (status.active_node) {
                currentNodeEl.textContent = `${status.active_node.node_name} (${status.active_node.latency}ms)`;
            } else {
                currentNodeEl.textContent = '无';
            }
        }

        // 队列大小
        const queueSizeEl = document.getElementById('queueSize');
        if (queueSizeEl) {
            queueSizeEl.textContent = status.queue_size || 0;
        }

        // 切换次数
        const switchCountEl = document.getElementById('switchCount');
        if (switchCountEl) {
            switchCountEl.textContent = status.total_switches || 0;
        }

        // 测试节点数
        const testedNodesEl = document.getElementById('testedNodes');
        if (testedNodesEl) {
            testedNodesEl.textContent = `${status.tested_nodes || 0}/${status.tested_nodes + status.failed_nodes || 0}`;
        }

        // 运行时间
        const uptimeEl = document.getElementById('uptime');
        if (uptimeEl) {
            uptimeEl.textContent = this.formatDuration(status.uptime || 0);
        }

        // 测试进度
        if (status.testing_progress) {
            this.updateTestingProgress(status.testing_progress);
        }
    }

    // 更新队列显示
    updateQueueDisplay(queue) {
        const queueList = document.getElementById('queueList');
        if (!queueList) return;

        if (queue.length === 0) {
            queueList.innerHTML = '<div style="text-align: center; color: #666; padding: 40px;">暂无数据</div>';
            return;
        }

        let html = '';
        queue.forEach((node, index) => {
            const isActive = node.is_active;
            const itemClass = isActive ? 'queue-item active' : 'queue-item';
            
            html += `
                <div class="${itemClass}">
                    <div class="node-info">
                        <div style="font-weight: 500;">
                            ${isActive ? '🟢 ' : ''}${node.node_name}
                        </div>
                        <div style="font-size: 0.9em; color: #666;">
                            ${node.protocol} | ${node.server}:${node.port}
                        </div>
                        <div class="node-stats">
                            <span>延迟: ${node.latency}ms</span>
                            <span>速度: ${node.speed.toFixed(2)}Mbps</span>
                            <span>评分: ${node.score.toFixed(2)}</span>
                            <span>成功率: ${node.success_rate.toFixed(1)}%</span>
                        </div>
                    </div>
                    <div>
                        ${!isActive ? `<button class="btn btn-primary" onclick="intelligentProxy.switchToNode(${index})">切换</button>` : '<span style="color: #28a745; font-weight: 500;">当前激活</span>'}
                    </div>
                </div>
            `;
        });

        queueList.innerHTML = html;
    }

    // 更新测试进度
    updateTestingProgress(progress) {
        const progressSection = document.getElementById('progressSection');
        const currentTestNodeEl = document.getElementById('currentTestNode');
        const testProgressEl = document.getElementById('testProgress');
        const testCompletedEl = document.getElementById('testCompleted');
        const testTotalEl = document.getElementById('testTotal');
        const testSuccessEl = document.getElementById('testSuccess');
        const testFailedEl = document.getElementById('testFailed');
        const progressFillEl = document.getElementById('progressFill');

        if (progress.is_running) {
            progressSection.style.display = 'block';
            
            if (currentTestNodeEl) currentTestNodeEl.textContent = progress.current_node || '-';
            if (testProgressEl) testProgressEl.textContent = `${progress.progress || 0}%`;
            if (testCompletedEl) testCompletedEl.textContent = progress.tested_nodes || 0;
            if (testTotalEl) testTotalEl.textContent = progress.total_nodes || 0;
            if (testSuccessEl) testSuccessEl.textContent = progress.success_nodes || 0;
            if (testFailedEl) testFailedEl.textContent = progress.failed_nodes || 0;
            if (progressFillEl) progressFillEl.style.width = `${progress.progress || 0}%`;
        } else {
            progressSection.style.display = 'none';
        }
    }

    // 切换到指定节点
    async switchToNode(nodeIndex) {
        try {
            const response = await fetch('/api/intelligent-proxy/switch', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    node_index: nodeIndex
                })
            });

            const result = await response.json();
            
            if (result.success) {
                this.showAlert('节点切换成功', 'success');
                await this.updateStatus();
            } else {
                this.showAlert(`切换失败: ${result.error}`, 'danger');
            }
        } catch (error) {
            console.error('切换节点失败:', error);
            this.showAlert('切换失败，请检查网络连接', 'danger');
        }
    }

    // 强制重新测试所有节点
    async forceRetestAllNodes() {
        try {
            const response = await fetch('/api/intelligent-proxy/retest', {
                method: 'POST'
            });

            const result = await response.json();
            
            if (result.success) {
                this.showAlert('重新测试已启动', 'success');
            } else {
                this.showAlert(`启动重测失败: ${result.error}`, 'danger');
            }
        } catch (error) {
            console.error('启动重新测试失败:', error);
            this.showAlert('启动重测失败，请检查网络连接', 'danger');
        }
    }

    // 切换自动切换模式
    async toggleAutoSwitch() {
        const newState = !this.autoSwitchEnabled;
        
        try {
            const response = await fetch('/api/intelligent-proxy/toggle-auto-switch', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    enabled: newState
                })
            });

            const result = await response.json();
            
            if (result.success) {
                this.autoSwitchEnabled = newState;
                this.updateUI();
                this.showAlert(newState ? '自动切换已启用' : '自动切换已暂停', 'success');
            } else {
                this.showAlert(`切换模式失败: ${result.error}`, 'danger');
            }
        } catch (error) {
            console.error('切换自动模式失败:', error);
            this.showAlert('切换模式失败，请检查网络连接', 'danger');
        }
    }

    // 启动事件流
    startEventStream() {
        if (this.eventSource) {
            this.eventSource.close();
        }

        this.eventSource = new EventSource('/api/intelligent-proxy/events');
        
        this.eventSource.onopen = () => {
            this.addEventLog('事件流连接成功', 'info');
        };

        this.eventSource.onerror = (error) => {
            console.error('事件流错误:', error);
            this.addEventLog('事件流连接中断', 'error');
        };

        // 监听各种事件
        this.eventSource.addEventListener('connected', (e) => {
            const data = JSON.parse(e.data);
            this.addEventLog(`事件流已连接: ${data.message}`, 'info');
        });

        this.eventSource.addEventListener('service_started', (e) => {
            const data = JSON.parse(e.data);
            this.addEventLog('智能代理服务已启动', 'success');
        });

        this.eventSource.addEventListener('service_stopped', (e) => {
            this.addEventLog('智能代理服务已停止', 'warning');
        });

        this.eventSource.addEventListener('testing_start', (e) => {
            const data = JSON.parse(e.data);
            this.addEventLog(`开始测试节点，共 ${data.total_nodes} 个节点`, 'info');
        });

        this.eventSource.addEventListener('testing_progress', (e) => {
            const data = JSON.parse(e.data);
            this.updateTestingProgress(data);
        });

        this.eventSource.addEventListener('testing_complete', (e) => {
            const data = JSON.parse(e.data);
            this.addEventLog(`节点测试完成，成功: ${data.success_nodes}，失败: ${data.failed_nodes}`, 'success');
            this.updateStatus(); // 更新状态和队列
        });

        this.eventSource.addEventListener('node_switch', (e) => {
            const data = JSON.parse(e.data);
            const fromNode = data.from_node ? data.from_node.node_name : '无';
            const toNode = data.to_node.node_name;
            const reason = this.getSwitchReasonText(data.switch_reason);
            this.addEventLog(`节点切换: ${fromNode} → ${toNode} (${reason})`, 'warning');
            this.updateStatus(); // 更新状态显示
        });

        this.eventSource.addEventListener('queue_update', (e) => {
            this.updateStatus(); // 更新队列显示
        });

        this.eventSource.addEventListener('auto_switch_toggled', (e) => {
            const data = JSON.parse(e.data);
            this.addEventLog(`自动切换${data.enabled ? '已启用' : '已暂停'}`, 'info');
        });

        this.eventSource.addEventListener('config_updated', (e) => {
            this.addEventLog('配置已更新', 'info');
        });

        this.eventSource.addEventListener('heartbeat', (e) => {
            // 心跳事件，不需要显示
        });
    }

    // 停止事件流
    stopEventStream() {
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }
    }

    // 启动状态更新器
    startStatusUpdater() {
        if (this.statusTimer) {
            clearInterval(this.statusTimer);
        }
        
        this.statusTimer = setInterval(() => {
            this.updateStatus();
        }, 5000); // 每5秒更新一次状态
    }

    // 停止状态更新器
    stopStatusUpdater() {
        if (this.statusTimer) {
            clearInterval(this.statusTimer);
            this.statusTimer = null;
        }
    }

    // 清空状态
    clearStatus() {
        const runningStatusEl = document.getElementById('runningStatus');
        const currentNodeEl = document.getElementById('currentNode');
        const queueSizeEl = document.getElementById('queueSize');
        const switchCountEl = document.getElementById('switchCount');
        const testedNodesEl = document.getElementById('testedNodes');
        const uptimeEl = document.getElementById('uptime');
        const queueList = document.getElementById('queueList');
        const progressSection = document.getElementById('progressSection');

        if (runningStatusEl) runningStatusEl.textContent = '未运行';
        if (currentNodeEl) currentNodeEl.textContent = '无';
        if (queueSizeEl) queueSizeEl.textContent = '0';
        if (switchCountEl) switchCountEl.textContent = '0';
        if (testedNodesEl) testedNodesEl.textContent = '0';
        if (uptimeEl) uptimeEl.textContent = '0秒';
        if (queueList) queueList.innerHTML = '<div style="text-align: center; color: #666; padding: 40px;">暂无数据</div>';
        if (progressSection) progressSection.style.display = 'none';
    }

    // 更新UI状态
    updateUI() {
        const startFormElements = document.querySelectorAll('#startForm input, #startForm select, #startForm button[type="submit"]');
        const stopBtn = document.getElementById('stopBtn');
        const retestBtn = document.getElementById('retestBtn');
        const toggleAutoSwitchBtn = document.getElementById('toggleAutoSwitchBtn');

        // 启动表单
        startFormElements.forEach(el => {
            el.disabled = this.isRunning;
        });

        // 停止按钮
        if (stopBtn) {
            stopBtn.disabled = !this.isRunning;
        }

        // 重新测试按钮
        if (retestBtn) {
            retestBtn.disabled = !this.isRunning;
        }

        // 自动切换按钮
        if (toggleAutoSwitchBtn) {
            toggleAutoSwitchBtn.disabled = !this.isRunning;
            toggleAutoSwitchBtn.textContent = this.autoSwitchEnabled ? '暂停自动切换' : '启用自动切换';
            toggleAutoSwitchBtn.className = this.autoSwitchEnabled ? 'btn btn-warning' : 'btn btn-success';
        }
    }

    // 添加事件日志
    addEventLog(message, type = 'info') {
        const eventLog = document.getElementById('eventLog');
        if (!eventLog) return;

        const timestamp = new Date().toLocaleTimeString();
        const colorMap = {
            'info': '#007bff',
            'success': '#28a745',
            'warning': '#ffc107',
            'error': '#dc3545'
        };

        const eventItem = document.createElement('div');
        eventItem.className = 'event-item';
        eventItem.style.borderLeftColor = colorMap[type] || '#007bff';
        eventItem.innerHTML = `
            <div style="font-size: 11px; color: #666;">${timestamp}</div>
            <div>${message}</div>
        `;

        // 插入到顶部
        eventLog.insertBefore(eventItem, eventLog.firstChild);

        // 限制日志条数
        const items = eventLog.querySelectorAll('.event-item');
        if (items.length > 100) {
            items[items.length - 1].remove();
        }

        // 自动滚动到顶部
        eventLog.scrollTop = 0;
    }

    // 清空事件日志
    clearEventLog() {
        const eventLog = document.getElementById('eventLog');
        if (eventLog) {
            eventLog.innerHTML = '<div style="text-align: center; color: #666;">等待事件...</div>';
        }
    }

    // 显示警告消息
    showAlert(message, type = 'info') {
        // 创建警告框
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert alert-${type}`;
        alertDiv.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 1000;
            padding: 12px 20px;
            border-radius: 4px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            max-width: 400px;
            word-wrap: break-word;
        `;

        const colorMap = {
            'info': { bg: '#d1ecf1', color: '#0c5460', border: '#bee5eb' },
            'success': { bg: '#d4edda', color: '#155724', border: '#c3e6cb' },
            'warning': { bg: '#fff3cd', color: '#856404', border: '#ffeaa7' },
            'danger': { bg: '#f8d7da', color: '#721c24', border: '#f5c6cb' }
        };

        const colors = colorMap[type] || colorMap['info'];
        alertDiv.style.backgroundColor = colors.bg;
        alertDiv.style.color = colors.color;
        alertDiv.style.border = `1px solid ${colors.border}`;

        alertDiv.innerHTML = `
            <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>${message}</span>
                <button style="background: none; border: none; font-size: 18px; cursor: pointer; margin-left: 10px;" onclick="this.parentElement.parentElement.remove()">×</button>
            </div>
        `;

        document.body.appendChild(alertDiv);

        // 3秒后自动移除
        setTimeout(() => {
            if (alertDiv.parentNode) {
                alertDiv.remove();
            }
        }, 3000);
    }

    // 格式化持续时间
    formatDuration(seconds) {
        if (seconds < 60) {
            return `${seconds}秒`;
        } else if (seconds < 3600) {
            return `${Math.floor(seconds / 60)}分${seconds % 60}秒`;
        } else {
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            return `${hours}小时${minutes}分`;
        }
    }

    // 获取切换原因文本
    getSwitchReasonText(reason) {
        const reasonMap = {
            'manual_switch': '手动切换',
            'initial_activation': '初始激活',
            'better_node_available': '发现更快节点',
            'health_check_failed': '健康检查失败',
            'auto_failover': '自动故障转移'
        };
        return reasonMap[reason] || reason;
    }
}

// 全局实例
let intelligentProxy;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    intelligentProxy = new IntelligentProxyManager();
});

// 页面卸载时清理资源
window.addEventListener('beforeunload', () => {
    if (intelligentProxy) {
        intelligentProxy.stopEventStream();
        intelligentProxy.stopStatusUpdater();
    }
});