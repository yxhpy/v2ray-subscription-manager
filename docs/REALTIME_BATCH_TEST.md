# V2Ray 订阅管理器 - 实时批量测试功能

## 概述

本文档介绍了V2Ray订阅管理器的实时批量测试功能。该功能通过Server-Sent Events (SSE)技术实现实时进度更新，为用户提供更好的测试体验。

## 功能特性

### 🚀 核心特性

1. **实时进度更新**: 使用SSE技术实现实时进度推送
2. **可视化进度界面**: 美观的进度弹窗显示测试状态
3. **实时统计数据**: 动态显示成功/失败数量
4. **消息流展示**: 实时显示每个节点的测试消息
5. **可取消测试**: 支持中途取消正在进行的批量测试

### 🛠️ 技术特性

- **并发控制**: 最大2个并发测试，避免端口冲突
- **端口管理**: 智能端口分配和冲突检测
- **配置隔离**: 每个测试实例使用独立配置文件
- **资源清理**: 自动清理临时文件和进程

## 架构设计

### 前端架构

```
Web UI (JavaScript)
├── batchTestNodes() - 主入口函数
├── showBatchTestProgress() - 显示进度界面
├── startBatchTestSSE() - 建立SSE连接
├── updateBatchTestProgress() - 更新进度显示
├── handleBatchTestComplete() - 处理完成事件
└── cancelBatchTest() - 取消测试
```

### 后端架构

```
Go Backend
├── BatchTestNodesSSE() - SSE处理器
├── BatchTestNodesWithProgress() - 带进度回调的测试
├── ProgressCallback - 进度回调函数类型
└── BatchTestProgress - 进度数据模型
```

### 通信协议

使用Server-Sent Events (SSE)协议进行实时通信：

```
Event Types:
- progress: 测试进度更新
- final_result: 最终测试结果
- error: 错误信息
- close: 连接关闭
```

## 使用方法

### 1. 启动Web UI

```bash
# 使用测试脚本
./test_realtime_batch.sh

# 或手动启动
cd cmd/web-ui
go run main.go
```

### 2. 访问界面

打开浏览器访问: `http://localhost:8888`

### 3. 添加订阅

1. 点击"订阅管理"标签
2. 点击"添加订阅"
3. 输入订阅URL和名称
4. 点击"解析订阅"获取节点列表

### 4. 批量测试

1. 进入"节点管理"页面
2. 选择要测试的节点（复选框）
3. 点击"批量测试"按钮
4. 观察实时进度弹窗

### 5. 进度监控

进度弹窗显示以下信息：
- **统计数据**: 总数/完成数/成功数/失败数
- **进度条**: 可视化进度百分比
- **消息流**: 实时测试消息
- **操作按钮**: 取消/关闭按钮

## API接口

### SSE端点

```
GET /api/nodes/batch-test-sse?subscription_id={id}&node_indexes={indexes}
```

**参数**:
- `subscription_id`: 订阅ID
- `node_indexes`: JSON格式的节点索引数组

**响应格式**:
```
event: progress
data: {
  "type": "progress",
  "message": "正在测试节点 1: 香港节点",
  "node_index": 1,
  "node_name": "香港节点", 
  "progress": 50,
  "total": 4,
  "completed": 2,
  "success_count": 1,
  "failure_count": 1,
  "current_result": {
    "node_name": "香港节点",
    "success": true,
    "latency": "120ms"
  },
  "timestamp": "2024-06-30 21:15:30"
}
```

### 传统API端点

```
POST /api/nodes/batch-test
```

保持向后兼容，返回完整结果。

## 前端实现

### 1. 进度弹窗HTML

```html
<div class="modal active">
  <div class="modal-content">
    <div class="modal-header">
      <h3>批量测试进度</h3>
      <button class="close-btn">&times;</button>
    </div>
    <div class="modal-body">
      <div class="progress-info">
        <div class="progress-stats">
          <span>总数: <span id="progressTotal">0</span></span>
          <span>完成: <span id="progressCompleted">0</span></span>
          <span>成功: <span id="progressSuccess">0</span></span>
          <span>失败: <span id="progressFailure">0</span></span>
        </div>
        <div class="progress-bar-container">
          <div class="progress-bar">
            <div id="progressBar" class="progress-fill"></div>
          </div>
          <span id="progressPercent">0%</span>
        </div>
      </div>
      <div class="progress-messages">
        <div id="progressMessages" class="message-list"></div>
      </div>
    </div>
    <div class="modal-footer">
      <button id="cancelBatchTestBtn">取消测试</button>
    </div>
  </div>
</div>
```

### 2. SSE连接代码

```javascript
async startBatchTestSSE(nodeIndexes) {
    return new Promise((resolve, reject) => {
        const nodeIndexesStr = JSON.stringify(nodeIndexes);
        const sseUrl = `/api/nodes/batch-test-sse?subscription_id=${encodeURIComponent(this.activeSubscriptionId)}&node_indexes=${encodeURIComponent(nodeIndexesStr)}`;
        
        const eventSource = new EventSource(sseUrl);

        eventSource.addEventListener('progress', (event) => {
            const progress = JSON.parse(event.data);
            this.updateBatchTestProgress(progress);
        });

        eventSource.addEventListener('final_result', (event) => {
            const result = JSON.parse(event.data);
            this.handleBatchTestComplete(result);
            eventSource.close();
            resolve(result);
        });

        eventSource.onerror = (error) => {
            eventSource.close();
            reject(new Error('SSE连接失败'));
        };

        this.currentBatchTestSSE = eventSource;
    });
}
```

## 后端实现

### 1. 进度回调类型

```go
type ProgressCallback func(progress *models.BatchTestProgress)
```

### 2. 进度数据模型

```go
type BatchTestProgress struct {
    Type          string          `json:"type"`
    Message       string          `json:"message"`
    NodeIndex     int             `json:"node_index"`
    NodeName      string          `json:"node_name"`
    Progress      int             `json:"progress"`
    Total         int             `json:"total"`
    Completed     int             `json:"completed"`
    SuccessCount  int             `json:"success_count"`
    FailureCount  int             `json:"failure_count"`
    CurrentResult *NodeTestResult `json:"current_result"`
    Timestamp     string          `json:"timestamp"`
}
```

### 3. SSE处理器

```go
func (h *NodeHandler) BatchTestNodesSSE(w http.ResponseWriter, r *http.Request) {
    // 设置SSE头部
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // 解析参数
    subscriptionID := r.URL.Query().Get("subscription_id")
    nodeIndexesStr := r.URL.Query().Get("node_indexes")

    // 进度回调
    progressCallback := func(progress *models.BatchTestProgress) {
        h.sendSSEEvent(w, "progress", progress)
    }

    // 启动批量测试
    go func() {
        results, err := h.nodeService.BatchTestNodesWithProgress(
            subscriptionID, nodeIndexes, progressCallback)
        
        if err != nil {
            h.sendSSEError(w, err.Error())
            return
        }

        h.sendSSEEvent(w, "final_result", results)
        h.sendSSEEvent(w, "close", nil)
    }()

    <-r.Context().Done()
}
```

## 样式设计

### 1. 进度弹窗样式

```css
.modal {
    position: fixed;
    z-index: 1000;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
}

.modal-content {
    background-color: #fff;
    border-radius: 8px;
    max-width: 600px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}
```

### 2. 进度条样式

```css
.progress-bar {
    height: 20px;
    background-color: #e9ecef;
    border-radius: 10px;
    overflow: hidden;
}

.progress-fill {
    height: 100%;
    background: linear-gradient(45deg, #007bff, #28a745);
    transition: width 0.3s ease;
}
```

## 测试验证

### 1. 功能测试

```bash
# 启动测试
./test_realtime_batch.sh

# 访问Web UI
open http://localhost:8888
```

### 2. 测试步骤

1. 添加订阅源
2. 解析获取节点
3. 选择多个节点
4. 执行批量测试
5. 观察实时进度
6. 验证取消功能

### 3. 验证要点

- ✅ 进度条实时更新
- ✅ 统计数据准确
- ✅ 消息流顺序正确
- ✅ 节点状态同步
- ✅ 取消功能有效
- ✅ 最终结果正确

## 问题排查

### 1. SSE连接失败

**现象**: 进度弹窗不显示或无更新

**解决方案**:
- 检查浏览器控制台错误
- 确认SSE端点URL正确
- 验证服务器响应头设置

### 2. 进度更新异常

**现象**: 进度条不动或跳跃

**解决方案**:
- 检查进度计算逻辑
- 验证回调函数调用
- 确认并发控制正确

### 3. 内存泄漏

**现象**: 长时间运行后性能下降

**解决方案**:
- 确保EventSource正确关闭
- 检查DOM元素清理
- 验证goroutine退出

## 版本历史

- **v1.0**: 基础批量测试功能
- **v2.0**: 新增实时进度更新 (当前版本)
  - 新增SSE实时通信
  - 新增进度可视化界面
  - 新增取消测试功能
  - 优化并发控制
  - 优化端口管理

## 后续规划

- [ ] 支持测试历史记录
- [ ] 支持自定义测试参数
- [ ] 支持测试结果导出
- [ ] 支持测试报告生成
- [ ] 支持WebSocket双向通信 