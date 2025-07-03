package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/services"
)

// BatchTestManager 批量测试管理器
type BatchTestManager struct {
	mutex          sync.RWMutex
	activeSessions map[string]context.CancelFunc // sessionID -> cancelFunc
}

// 全局批量测试管理器
var batchTestManager = &BatchTestManager{
	activeSessions: make(map[string]context.CancelFunc),
}

// 添加活跃会话
func (btm *BatchTestManager) AddSession(sessionID string, cancel context.CancelFunc) {
	btm.mutex.Lock()
	defer btm.mutex.Unlock()
	btm.activeSessions[sessionID] = cancel
}

// 移除活跃会话
func (btm *BatchTestManager) RemoveSession(sessionID string) {
	btm.mutex.Lock()
	defer btm.mutex.Unlock()
	delete(btm.activeSessions, sessionID)
}

// 取消会话
func (btm *BatchTestManager) CancelSession(sessionID string) bool {
	btm.mutex.Lock()
	defer btm.mutex.Unlock()
	if cancel, exists := btm.activeSessions[sessionID]; exists {
		cancel()
		delete(btm.activeSessions, sessionID)
		return true
	}
	return false
}

// NodeHandler 节点处理器
type NodeHandler struct {
	nodeService services.NodeService
}

// NewNodeHandler 创建节点处理器
func NewNodeHandler(nodeService services.NodeService) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
	}
}

// ConnectNode 连接节点
func (h *NodeHandler) ConnectNode(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var req models.NodeOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	result, err := h.nodeService.ConnectNode(req.SubscriptionID, req.NodeIndex, req.Operation)
	if err != nil {
		response.SetError(err, "节点连接失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(result, "节点连接成功")
	h.writeJSONResponse(w, response)
}

// TestNode 测试节点
func (h *NodeHandler) TestNode(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var req models.NodeOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	result, err := h.nodeService.TestNode(req.SubscriptionID, req.NodeIndex)
	if err != nil {
		response.SetError(err, "节点测试失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(result, "节点测试完成")
	h.writeJSONResponse(w, response)
}

// SpeedTestNode 速度测试节点
func (h *NodeHandler) SpeedTestNode(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var req models.NodeOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	result, err := h.nodeService.SpeedTestNode(req.SubscriptionID, req.NodeIndex)
	if err != nil {
		response.SetError(err, "速度测试失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(result, "速度测试完成")
	h.writeJSONResponse(w, response)
}

// BatchTestNodes 批量测试节点
func (h *NodeHandler) BatchTestNodes(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var req models.BatchNodeOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	results, err := h.nodeService.BatchTestNodes(req.SubscriptionID, req.NodeIndexes)
	if err != nil {
		response.SetError(err, "批量测试失败")
		h.writeJSONResponse(w, response)
		return
	}

	// 统计结果
	batchResponse := &models.BatchTestResponse{
		Results:      results,
		TotalCount:   len(results),
		SuccessCount: 0,
		FailureCount: 0,
	}

	for _, result := range results {
		if result.Success {
			batchResponse.SuccessCount++
		} else {
			batchResponse.FailureCount++
		}
	}

	response.SetSuccess(batchResponse, "批量测试完成")
	h.writeJSONResponse(w, response)
}

// BatchTestNodesSSE 批量测试节点 - SSE版本
func (h *NodeHandler) BatchTestNodesSSE(w http.ResponseWriter, r *http.Request) {
	// 设置完整的CORS和SSE头部
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("X-Accel-Buffering", "no") // 禁用nginx缓冲

	// 立即发送连接测试消息
	fmt.Fprintf(w, "event: ping\ndata: {\"message\": \"connection test\"}\n\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// 从URL参数获取请求数据
	subscriptionID := r.URL.Query().Get("subscription_id")
	nodeIndexesStr := r.URL.Query().Get("node_indexes")

	fmt.Printf("DEBUG: SSE批量测试请求 - subscription_id: %s, node_indexes: %s\n", subscriptionID, nodeIndexesStr)

	if subscriptionID == "" || nodeIndexesStr == "" {
		h.sendSSEError(w, "缺少必要参数: subscription_id 或 node_indexes")
		return
	}

	// 解析节点索引
	var nodeIndexes []int
	if err := json.Unmarshal([]byte(nodeIndexesStr), &nodeIndexes); err != nil {
		h.sendSSEError(w, "解析节点索引失败: "+err.Error())
		return
	}

	if len(nodeIndexes) == 0 {
		h.sendSSEError(w, "节点索引列表为空")
		return
	}

	fmt.Printf("DEBUG: 解析到节点索引: %v\n", nodeIndexes)

	// 创建可取消的上下文和会话ID
	testCtx, testCancel := context.WithCancel(context.Background())
	sessionID := fmt.Sprintf("%s_%d", subscriptionID, time.Now().UnixNano())

	// 注册会话到管理器
	batchTestManager.AddSession(sessionID, testCancel)
	defer batchTestManager.RemoveSession(sessionID)

	// 发送连接成功消息（包含会话ID）
	h.sendSSEEvent(w, "connected", map[string]interface{}{
		"message":   "SSE连接已建立，开始批量测试",
		"total":     len(nodeIndexes),
		"sessionId": sessionID,
	})

	// 使用通道来传递测试结果和进度
	progressChan := make(chan *models.BatchTestProgress, 100)
	resultChan := make(chan *models.BatchTestResponse, 1)
	errorChan := make(chan error, 1)

	// 客户端连接状态
	clientCtx := r.Context()
	clientConnected := true

	// 进度回调函数
	progressCallback := func(progress *models.BatchTestProgress) {
		select {
		case progressChan <- progress:
			// 成功发送进度
		case <-time.After(2 * time.Second):
			fmt.Printf("WARNING: 进度通道阻塞，跳过进度更新\n")
		}
	}

	// 启动批量测试 goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("ERROR: 批量测试 panic: %v\n", r)
				select {
				case errorChan <- fmt.Errorf("批量测试发生内部错误: %v", r):
				default:
				}
			}
			close(progressChan)
			close(resultChan)
			close(errorChan)
		}()

		fmt.Printf("DEBUG: 开始批量测试 %d 个节点 (会话ID: %s)\n", len(nodeIndexes), sessionID)

		// 检查是否已被取消
		select {
		case <-testCtx.Done():
			fmt.Printf("DEBUG: 批量测试在开始前被取消\n")
			select {
			case errorChan <- fmt.Errorf("批量测试已被取消"):
			default:
			}
			return
		default:
		}

		results, err := h.nodeService.BatchTestNodesWithProgressAndContext(testCtx, subscriptionID, nodeIndexes, progressCallback)

		if err != nil {
			fmt.Printf("ERROR: 批量测试失败: %v\n", err)
			// 检查是否是取消导致的错误
			if testCtx.Err() != nil {
				select {
				case errorChan <- fmt.Errorf("批量测试被取消"):
				default:
				}
			} else {
				select {
				case errorChan <- err:
				default:
				}
			}
			return
		}

		// 发送最终结果
		batchResponse := &models.BatchTestResponse{
			Results:      results,
			TotalCount:   len(results),
			SuccessCount: 0,
			FailureCount: 0,
		}

		for _, result := range results {
			if result.Success {
				batchResponse.SuccessCount++
			} else {
				batchResponse.FailureCount++
			}
		}

		select {
		case resultChan <- batchResponse:
			fmt.Printf("DEBUG: 批量测试完成: 成功 %d，失败 %d\n", batchResponse.SuccessCount, batchResponse.FailureCount)
		default:
		}
	}()

	// 心跳检测
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// 主循环：处理进度、结果和客户端连接状态
	for {
		select {
		case <-clientCtx.Done():
			// 客户端断开连接
			if clientConnected {
				fmt.Printf("DEBUG: 客户端连接断开 - %v，但测试继续进行\n", clientCtx.Err())
				clientConnected = false
			}
			// 不立即返回，继续等待测试完成或超时

		case <-heartbeatTicker.C:
			// 发送心跳
			if clientConnected {
				h.sendSSEEvent(w, "heartbeat", map[string]interface{}{
					"timestamp": time.Now().Unix(),
				})
			}

		case progress, ok := <-progressChan:
			if !ok {
				// 进度通道关闭，准备退出
				fmt.Printf("DEBUG: 进度通道已关闭\n")
				return
			}

			if clientConnected {
				h.sendSSEEvent(w, "progress", progress)
			}

		case result, ok := <-resultChan:
			if !ok {
				fmt.Printf("DEBUG: 结果通道异常关闭\n")
				return
			}

			if clientConnected {
				h.sendSSEEvent(w, "final_result", result)
				h.sendSSEEvent(w, "close", nil)
			}
			fmt.Printf("DEBUG: 批量测试正常完成\n")
			return

		case err, ok := <-errorChan:
			if !ok {
				fmt.Printf("DEBUG: 错误通道异常关闭\n")
				return
			}

			if clientConnected {
				// 检查是否是取消错误
				if testCtx.Err() != nil {
					fmt.Printf("DEBUG: 发送取消事件: %v\n", err)
					h.sendSSEEvent(w, "cancelled", map[string]interface{}{
						"message": "批量测试已被取消",
						"reason":  err.Error(),
					})
				} else {
					fmt.Printf("DEBUG: 发送错误事件: %v\n", err)
					h.sendSSEError(w, "批量测试失败: "+err.Error())
				}
			}
			fmt.Printf("DEBUG: 批量测试出错完成: %v\n", err)
			return

		case <-time.After(15 * time.Minute):
			// 15分钟超时保护
			fmt.Printf("WARNING: 批量测试超时，强制退出\n")
			if clientConnected {
				h.sendSSEError(w, "批量测试超时（15分钟）")
			}
			return
		}
	}
}

// sendSSEEvent 发送SSE事件
func (h *NodeHandler) sendSSEEvent(w http.ResponseWriter, eventType string, data interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR: 发送SSE事件时发生 panic: %v\n", r)
		}
	}()

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("ERROR: JSON序列化失败: %v\n", err)
			fmt.Fprintf(w, "event: error\ndata: {\"error\": \"JSON序列化失败: %s\"}\n\n", err.Error())
		} else {
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(jsonData))
		}
	} else {
		fmt.Fprintf(w, "event: %s\ndata: {}\n\n", eventType)
	}

	// 强制刷新缓冲区
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	} else {
		fmt.Printf("WARNING: ResponseWriter 不支持 Flusher 接口\n")
	}
}

// sendSSEError 发送SSE错误事件
func (h *NodeHandler) sendSSEError(w http.ResponseWriter, message string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR: 发送SSE错误事件时发生 panic: %v\n", r)
		}
	}()

	fmt.Printf("ERROR: 发送SSE错误: %s\n", message)
	fmt.Fprintf(w, "event: error\ndata: {\"error\": \"%s\"}\n\n", message)

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	} else {
		fmt.Printf("WARNING: ResponseWriter 不支持 Flusher 接口\n")
	}
}

// writeJSONResponse 写入JSON响应
func (h *NodeHandler) writeJSONResponse(w http.ResponseWriter, response *models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteNodes 删除节点
func (h *NodeHandler) DeleteNodes(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var req models.BatchNodeOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	if req.SubscriptionID == "" {
		response.SetError(fmt.Errorf("缺少订阅ID"), "订阅ID不能为空")
		h.writeJSONResponse(w, response)
		return
	}

	if len(req.NodeIndexes) == 0 {
		response.SetError(fmt.Errorf("缺少节点索引"), "节点索引列表不能为空")
		h.writeJSONResponse(w, response)
		return
	}

	fmt.Printf("DEBUG: 删除节点请求 - subscription_id: %s, node_indexes: %v\n", req.SubscriptionID, req.NodeIndexes)

	// 调用节点服务删除节点
	err := h.nodeService.DeleteNodes(req.SubscriptionID, req.NodeIndexes)
	if err != nil {
		fmt.Printf("ERROR: 删除节点失败: %v\n", err)
		response.SetError(err, "删除节点失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(map[string]interface{}{
		"deleted_count": len(req.NodeIndexes),
		"node_indexes":  req.NodeIndexes,
	}, fmt.Sprintf("成功删除 %d 个节点", len(req.NodeIndexes)))
	h.writeJSONResponse(w, response)
}

// CancelBatchTest 取消批量测试
func (h *NodeHandler) CancelBatchTest(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	if r.Method != "POST" {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	if req.SessionID == "" {
		response.SetError(fmt.Errorf("缺少会话ID"), "会话ID不能为空")
		h.writeJSONResponse(w, response)
		return
	}

	// 尝试取消会话
	if batchTestManager.CancelSession(req.SessionID) {
		fmt.Printf("DEBUG: 成功取消批量测试会话: %s\n", req.SessionID)
		response.SetSuccess(map[string]interface{}{
			"cancelled": true,
			"sessionId": req.SessionID,
		}, "批量测试已取消")
	} else {
		fmt.Printf("DEBUG: 未找到要取消的批量测试会话: %s\n", req.SessionID)
		response.SetError(fmt.Errorf("会话不存在"), "未找到指定的批量测试会话，可能已经完成或不存在")
	}

	h.writeJSONResponse(w, response)
}
