package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/database"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/proxy"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/workflow"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// NodeServiceImpl 节点服务实现
type NodeServiceImpl struct {
	subscriptionService SubscriptionService
	proxyService        ProxyService

	// 数据库操作
	nodeDB         *database.NodeDB
	testResultDB   *database.TestResultDB

	// 节点连接管理 - 每个连接独立的代理管理器
	nodeConnections map[string]*NodeConnection // key: subscriptionID_nodeIndex
	connectionMutex sync.RWMutex

	// 节点状态管理
	nodeStates map[string]*models.NodeInfo // key: subscriptionID_nodeIndex
	stateMutex sync.RWMutex

	// MVP测试器
	mvpTester *workflow.MVPTester

	// 端口分配计数器（用于批量测试时避免端口冲突）
	portCounter int64
}

// NodeConnection 节点连接信息
type NodeConnection struct {
	V2RayManager     *proxy.ProxyManager
	Hysteria2Manager *proxy.Hysteria2ProxyManager
	HTTPPort         int
	SOCKSPort        int
	Protocol         string
	IsActive         bool
}

// NewNodeService 创建节点服务
func NewNodeService(subscriptionService SubscriptionService, proxyService ProxyService) NodeService {
	db := database.GetDB()
	service := &NodeServiceImpl{
		subscriptionService: subscriptionService,
		proxyService:        proxyService,
		nodeDB:              database.NewNodeDB(db),
		testResultDB:        database.NewTestResultDB(db),
		nodeConnections:     make(map[string]*NodeConnection),
		nodeStates:          make(map[string]*models.NodeInfo),
		portCounter:         9000, // 测试端口从9000开始
	}

	// 初始化MVP测试器
	service.mvpTester = workflow.NewMVPTester("")

	return service
}

// ConnectNode 连接节点（并发安全）
func (n *NodeServiceImpl) ConnectNode(subscriptionID string, nodeIndex int, operation string) (*models.ConnectNodeResponse, error) {
	// 防止快速点击导致的并发问题
	operationKey := fmt.Sprintf("connect_%s_%d", subscriptionID, nodeIndex)

	fmt.Printf("DEBUG: 开始处理节点连接操作 %s (订阅:%s, 节点:%d, 操作:%s)\n", operationKey, subscriptionID, nodeIndex, operation)

	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err != nil {
		return nil, err
	}

	if nodeIndex < 0 || nodeIndex >= len(subscription.Nodes) {
		return nil, fmt.Errorf("节点索引无效: %d", nodeIndex)
	}

	nodeInfo := subscription.Nodes[nodeIndex]
	response := &models.ConnectNodeResponse{}

	// 确保节点状态存在
	n.ensureNodeState(subscriptionID, nodeIndex, nodeInfo)

	// 更新节点状态为连接中
	n.updateNodeStatus(subscriptionID, nodeIndex, "connecting")

	fmt.Printf("DEBUG: 准备执行连接操作 %s\n", operation)

	switch operation {
	case "http_random":
		// 随机HTTP端口连接 - 自动分配可用端口
		fmt.Printf("DEBUG: 开始启动HTTP随机端口代理\n")
		actualHTTPPort, _, err := n.startProxyForNodeWithConnection(subscriptionID, nodeIndex, nodeInfo.Node, 0, 0) // 传入0表示随机分配
		if err != nil {
			fmt.Printf("DEBUG: 启动HTTP代理失败: %v\n", err)
			n.updateNodeStatus(subscriptionID, nodeIndex, "error")
			return nil, fmt.Errorf("启动HTTP代理失败: %v", err)
		}
		// 只返回HTTP端口
		fmt.Printf("DEBUG: HTTP代理启动成功，端口: %d\n", actualHTTPPort)
		response.HTTPPort = actualHTTPPort
		response.Port = actualHTTPPort
		n.setNodePorts(subscriptionID, nodeIndex, actualHTTPPort, 0)

	case "socks_random":
		// 随机SOCKS端口连接 - 自动分配可用端口
		_, actualSOCKSPort, err := n.startProxyForNodeWithConnection(subscriptionID, nodeIndex, nodeInfo.Node, 0, 0) // 传入0表示随机分配
		if err != nil {
			n.updateNodeStatus(subscriptionID, nodeIndex, "error")
			return nil, fmt.Errorf("启动SOCKS代理失败: %v", err)
		}
		// 只返回SOCKS端口
		response.SOCKSPort = actualSOCKSPort
		response.Port = actualSOCKSPort
		n.setNodePorts(subscriptionID, nodeIndex, 0, actualSOCKSPort)

	case "http_fixed":
		// 固定HTTP端口连接 - 使用系统配置的固定端口
		fixedHTTPPort := 8090 // 系统配置的固定HTTP端口

		// 检查端口是否被占用，如果被占用则停止占用该端口的连接
		if n.isPortOccupied(fixedHTTPPort) {
			n.stopConnectionsByPort(fixedHTTPPort)
		}

		actualHTTPPort, _, err := n.startProxyForNodeWithConnection(subscriptionID, nodeIndex, nodeInfo.Node, fixedHTTPPort, 0)
		if err != nil {
			n.updateNodeStatus(subscriptionID, nodeIndex, "error")
			return nil, fmt.Errorf("启动固定HTTP代理失败: %v", err)
		}
		response.HTTPPort = actualHTTPPort
		response.Port = actualHTTPPort
		n.setNodePorts(subscriptionID, nodeIndex, actualHTTPPort, 0)

	case "socks_fixed":
		// 固定SOCKS端口连接 - 使用系统配置的固定端口
		fixedSOCKSPort := 1088 // 系统配置的固定SOCKS端口

		// 检查端口是否被占用，如果被占用则停止占用该端口的连接
		if n.isPortOccupied(fixedSOCKSPort) {
			n.stopConnectionsByPort(fixedSOCKSPort)
		}

		_, actualSOCKSPort, err := n.startProxyForNodeWithConnection(subscriptionID, nodeIndex, nodeInfo.Node, 0, fixedSOCKSPort)
		if err != nil {
			n.updateNodeStatus(subscriptionID, nodeIndex, "error")
			return nil, fmt.Errorf("启动固定SOCKS代理失败: %v", err)
		}
		response.SOCKSPort = actualSOCKSPort
		response.Port = actualSOCKSPort
		n.setNodePorts(subscriptionID, nodeIndex, 0, actualSOCKSPort)

	case "disable":
		// 禁用节点（停止代理）
		n.removeNodeConnection(subscriptionID, nodeIndex)
		n.setNodePorts(subscriptionID, nodeIndex, 0, 0)
		n.updateNodeStatus(subscriptionID, nodeIndex, "idle")

	default:
		n.updateNodeStatus(subscriptionID, nodeIndex, "error")
		return nil, fmt.Errorf("不支持的操作: %s", operation)
	}

	response.Status = "success"
	response.Message = "操作完成"
	return response, nil
}

// TestNode 测试节点
func (n *NodeServiceImpl) TestNode(subscriptionID string, nodeIndex int) (*models.NodeTestResult, error) {
	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err != nil {
		return nil, err
	}

	if nodeIndex < 0 || nodeIndex >= len(subscription.Nodes) {
		return nil, fmt.Errorf("节点索引无效: %d", nodeIndex)
	}

	nodeInfo := subscription.Nodes[nodeIndex]

	// 确保节点状态存在
	n.ensureNodeState(subscriptionID, nodeIndex, nodeInfo)

	// 更新节点状态为测试中
	n.updateNodeStatus(subscriptionID, nodeIndex, "testing")

	// 创建测试结果
	result := &models.NodeTestResult{
		NodeName: nodeInfo.Name,
		TestTime: time.Now(),
		TestType: "connection",
	}

	// 执行真实的TCP连接测试
	startTime := time.Now()

	// 根据协议选择合适的测试方法
	var testErr error
	if nodeInfo.Protocol == "hysteria2" {
		// 测试Hysteria2节点
		testErr = n.testHysteria2Node(nodeInfo.Node)
	} else {
		// 测试V2Ray节点
		testErr = n.testV2RayNode(nodeInfo.Node)
	}

	latency := time.Since(startTime)

	if testErr != nil {
		result.Success = false
		result.Error = testErr.Error()
		n.updateNodeStatus(subscriptionID, nodeIndex, "error")
	} else {
		result.Success = true
		result.Latency = fmt.Sprintf("%dms", latency.Milliseconds())
		n.updateNodeStatus(subscriptionID, nodeIndex, "idle")
	}

	// 保存测试结果到节点状态和订阅数据
	n.setNodeTestResult(subscriptionID, nodeIndex, result)

	// 保存到数据库
	if nodeID, err := n.nodeDB.GetNodeIDBySubscriptionAndIndex(subscriptionID, nodeIndex); err == nil {
		if err := n.testResultDB.Create(nodeID, result); err != nil {
			fmt.Printf("WARNING: 保存测试结果到数据库失败: %v\n", err)
		}
	}

	return result, nil
}

// SpeedTestNode 速度测试节点
func (n *NodeServiceImpl) SpeedTestNode(subscriptionID string, nodeIndex int) (*models.SpeedTestResult, error) {
	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err != nil {
		return nil, err
	}

	if nodeIndex < 0 || nodeIndex >= len(subscription.Nodes) {
		return nil, fmt.Errorf("节点索引无效: %d", nodeIndex)
	}

	nodeInfo := subscription.Nodes[nodeIndex]

	// 确保节点状态存在
	n.ensureNodeState(subscriptionID, nodeIndex, nodeInfo)

	// 更新节点状态为测试中
	n.updateNodeStatus(subscriptionID, nodeIndex, "testing")

	// 创建速度测试结果
	result := &models.SpeedTestResult{
		NodeName: nodeInfo.Name,
		TestTime: time.Now(),
	}

	startTime := time.Now()

	// 执行真实的速度测试
	var testErr error
	var downloadSpeed, uploadSpeed, latency float64

	if nodeInfo.Protocol == "hysteria2" {
		downloadSpeed, uploadSpeed, latency, testErr = n.speedTestHysteria2Node(nodeInfo.Node)
	} else {
		downloadSpeed, uploadSpeed, latency, testErr = n.speedTestV2RayNode(nodeInfo.Node)
	}

	testDuration := time.Since(startTime)

	if testErr != nil {
		result.DownloadSpeed = "0 Mbps"
		result.UploadSpeed = "0 Mbps"
		result.Latency = "超时"
		n.updateNodeStatus(subscriptionID, nodeIndex, "error")
	} else {
		result.DownloadSpeed = fmt.Sprintf("%.1f Mbps", downloadSpeed)
		result.UploadSpeed = fmt.Sprintf("%.1f Mbps", uploadSpeed)
		result.Latency = fmt.Sprintf("%.0fms", latency)
		n.updateNodeStatus(subscriptionID, nodeIndex, "idle")
	}

	result.TestDuration = fmt.Sprintf("%.1fs", testDuration.Seconds())

	// 保存速度测试结果到节点状态和订阅数据
	n.setNodeSpeedResult(subscriptionID, nodeIndex, result)

	// 保存到数据库
	if nodeID, err := n.nodeDB.GetNodeIDBySubscriptionAndIndex(subscriptionID, nodeIndex); err == nil {
		if err := n.testResultDB.CreateSpeedResult(nodeID, result); err != nil {
			fmt.Printf("WARNING: 保存速度测试结果到数据库失败: %v\n", err)
		}
	}

	return result, nil
}

// BatchTestNodesWithProgress 带进度回调的批量测试节点
func (n *NodeServiceImpl) BatchTestNodesWithProgress(subscriptionID string, nodeIndexes []int, callback ProgressCallback) ([]*models.NodeTestResult, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR: BatchTestNodesWithProgress panic: %v\n", r)
		}
	}()

	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("获取订阅失败: %v", err)
	}

	if subscription.Nodes == nil || len(subscription.Nodes) == 0 {
		return nil, fmt.Errorf("订阅中没有可用节点")
	}

	total := len(nodeIndexes)
	if total == 0 {
		return []*models.NodeTestResult{}, nil
	}

	results := make([]*models.NodeTestResult, 0, total)
	successCount := 0
	failureCount := 0

	// 发送开始事件
	if callback != nil {
		callback(&models.BatchTestProgress{
			Type:      "start",
			Message:   fmt.Sprintf("开始批量测试 %d 个节点", total),
			Total:     total,
			Completed: 0,
			Progress:  0,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	// 使用信号量控制并发数（最大2个并发）
	semaphore := make(chan struct{}, 2)
	var wg sync.WaitGroup
	var mu sync.Mutex

	completed := 0

	for i, nodeIndex := range nodeIndexes {
		wg.Add(1)
		go func(idx, nodeIndex int) {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					fmt.Printf("ERROR: 节点测试 goroutine panic: %v\n", r)
					mu.Lock()
					result := &models.NodeTestResult{
						NodeName: fmt.Sprintf("节点 %d", nodeIndex),
						Success:  false,
						Error:    fmt.Sprintf("测试过程中发生内部错误: %v", r),
						TestTime: time.Now(),
						TestType: "batch",
					}
					results = append(results, result)
					completed++
					failureCount++
					mu.Unlock()
				}
			}()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 任务间延迟
			if idx > 0 {
				time.Sleep(500 * time.Millisecond)
			}

			// 验证节点索引
			if nodeIndex < 0 || nodeIndex >= len(subscription.Nodes) {
				mu.Lock()
				result := &models.NodeTestResult{
					NodeName: fmt.Sprintf("节点 %d", nodeIndex),
					Success:  false,
					Error:    "节点索引无效",
					TestTime: time.Now(),
					TestType: "batch",
				}
				results = append(results, result)
				completed++
				failureCount++

				// 发送进度更新
				if callback != nil {
					func() {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("ERROR: 发送进度回调时 panic: %v\n", r)
							}
						}()
						callback(&models.BatchTestProgress{
							Type:          "progress",
							Message:       fmt.Sprintf("节点 %d 测试失败: 节点索引无效", nodeIndex),
							NodeIndex:     nodeIndex,
							NodeName:      fmt.Sprintf("节点 %d", nodeIndex),
							Progress:      (completed * 100) / total,
							Total:         total,
							Completed:     completed,
							SuccessCount:  successCount,
							FailureCount:  failureCount,
							CurrentResult: result,
							Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
						})
					}()
				}
				mu.Unlock()
				return
			}

			node := subscription.Nodes[nodeIndex]

			// 发送当前节点开始测试的消息
			if callback != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("ERROR: 发送开始测试回调时 panic: %v\n", r)
						}
					}()
					callback(&models.BatchTestProgress{
						Type:      "progress",
						Message:   fmt.Sprintf("正在测试节点 %d: %s", nodeIndex, node.Name),
						NodeIndex: nodeIndex,
						NodeName:  node.Name,
						Progress:  (completed * 100) / total,
						Total:     total,
						Completed: completed,
						Timestamp: time.Now().Format("2006-01-02 15:04:05"),
					})
				}()
			}

			// 执行单个节点测试，增加超时保护
			var result *models.NodeTestResult
			var err error

			testDone := make(chan struct{})
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("ERROR: TestNode panic: %v\n", r)
						err = fmt.Errorf("测试过程中发生内部错误: %v", r)
					}
					close(testDone)
				}()
				result, err = n.TestNode(subscriptionID, nodeIndex)
			}()

			// 等待测试完成或超时（30秒）
			select {
			case <-testDone:
				// 测试完成
			case <-time.After(30 * time.Second):
				err = fmt.Errorf("测试超时")
				fmt.Printf("WARNING: 节点 %d 测试超时\n", nodeIndex)
			}

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result = &models.NodeTestResult{
					NodeName: node.Name,
					Success:  false,
					Error:    err.Error(),
					TestTime: time.Now(),
					TestType: "batch",
				}
				failureCount++
			} else {
				if result.Success {
					successCount++
				} else {
					failureCount++
				}
			}

			results = append(results, result)
			completed++

			// 发送进度更新
			if callback != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("ERROR: 发送完成回调时 panic: %v\n", r)
						}
					}()
					statusMsg := "测试完成"
					if result.Success {
						statusMsg = fmt.Sprintf("测试成功 (延迟: %s)", result.Latency)
					} else {
						statusMsg = fmt.Sprintf("测试失败: %s", result.Error)
					}

					callback(&models.BatchTestProgress{
						Type:          "progress",
						Message:       fmt.Sprintf("节点 %d (%s): %s", nodeIndex, node.Name, statusMsg),
						NodeIndex:     nodeIndex,
						NodeName:      node.Name,
						Progress:      (completed * 100) / total,
						Total:         total,
						Completed:     completed,
						SuccessCount:  successCount,
						FailureCount:  failureCount,
						CurrentResult: result,
						Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
					})
				}()
			}
		}(i, nodeIndex)
	}

	wg.Wait()

	// 发送完成事件
	if callback != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("ERROR: 发送完成事件时 panic: %v\n", r)
				}
			}()
			callback(&models.BatchTestProgress{
				Type:         "complete",
				Message:      fmt.Sprintf("批量测试完成: 成功 %d，失败 %d", successCount, failureCount),
				Progress:     100,
				Total:        total,
				Completed:    completed,
				SuccessCount: successCount,
				FailureCount: failureCount,
				Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
			})
		}()
	}

	return results, nil
}

// BatchTestNodes 批量测试节点（保持原有接口）
func (n *NodeServiceImpl) BatchTestNodes(subscriptionID string, nodeIndexes []int) ([]*models.NodeTestResult, error) {
	return n.BatchTestNodesWithProgress(subscriptionID, nodeIndexes, nil)
}

// BatchTestNodesWithProgressAndContext 带进度回调和上下文的批量测试节点
func (n *NodeServiceImpl) BatchTestNodesWithProgressAndContext(ctx context.Context, subscriptionID string, nodeIndexes []int, callback ProgressCallback) ([]*models.NodeTestResult, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR: BatchTestNodesWithProgressAndContext panic: %v\n", r)
		}
	}()

	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("获取订阅失败: %v", err)
	}

	if subscription.Nodes == nil || len(subscription.Nodes) == 0 {
		return nil, fmt.Errorf("订阅中没有可用节点")
	}

	total := len(nodeIndexes)
	if total == 0 {
		return []*models.NodeTestResult{}, nil
	}

	results := make([]*models.NodeTestResult, 0, total)
	successCount := 0
	failureCount := 0

	// 发送开始事件
	if callback != nil {
		callback(&models.BatchTestProgress{
			Type:      "start",
			Message:   fmt.Sprintf("开始批量测试 %d 个节点", total),
			Total:     total,
			Completed: 0,
			Progress:  0,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	// 检查是否已被取消
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("批量测试已被取消: %v", ctx.Err())
	default:
	}

	// 使用信号量控制并发数（最大2个并发）
	semaphore := make(chan struct{}, 2)
	var wg sync.WaitGroup
	var mu sync.Mutex

	completed := 0

	// 使用标志来控制循环退出
	cancelled := false

	for i, nodeIndex := range nodeIndexes {
		// 检查是否已被取消
		select {
		case <-ctx.Done():
			fmt.Printf("DEBUG: 批量测试被取消，停止启动新的测试任务\n")
			cancelled = true
		default:
		}

		// 如果已取消，不再启动新的测试任务
		if cancelled {
			break
		}

		wg.Add(1)
		go func(idx, nodeIndex int) {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					fmt.Printf("ERROR: 节点测试 goroutine panic: %v\n", r)
					mu.Lock()
					result := &models.NodeTestResult{
						NodeName: fmt.Sprintf("节点 %d", nodeIndex),
						Success:  false,
						Error:    fmt.Sprintf("测试过程中发生内部错误: %v", r),
						TestTime: time.Now(),
						TestType: "batch",
					}
					results = append(results, result)
					completed++
					failureCount++
					mu.Unlock()
				}
			}()

			// 获取信号量
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				fmt.Printf("DEBUG: 节点 %d 测试被取消（获取信号量时）\n", nodeIndex)
				return
			}
			defer func() { <-semaphore }()

			// 任务间延迟
			if idx > 0 {
				select {
				case <-time.After(500 * time.Millisecond):
				case <-ctx.Done():
					fmt.Printf("DEBUG: 节点 %d 测试被取消（延迟期间）\n", nodeIndex)
					return
				}
			}

			// 验证节点索引
			if nodeIndex < 0 || nodeIndex >= len(subscription.Nodes) {
				mu.Lock()
				result := &models.NodeTestResult{
					NodeName: fmt.Sprintf("节点 %d", nodeIndex),
					Success:  false,
					Error:    "节点索引无效",
					TestTime: time.Now(),
					TestType: "batch",
				}
				results = append(results, result)
				completed++
				failureCount++

				// 发送进度更新
				if callback != nil {
					func() {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("ERROR: 发送进度回调时 panic: %v\n", r)
							}
						}()
						callback(&models.BatchTestProgress{
							Type:          "progress",
							Message:       fmt.Sprintf("节点 %d 测试失败: 节点索引无效", nodeIndex),
							NodeIndex:     nodeIndex,
							NodeName:      fmt.Sprintf("节点 %d", nodeIndex),
							Progress:      (completed * 100) / total,
							Total:         total,
							Completed:     completed,
							SuccessCount:  successCount,
							FailureCount:  failureCount,
							CurrentResult: result,
							Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
						})
					}()
				}
				mu.Unlock()
				return
			}

			node := subscription.Nodes[nodeIndex]

			// 检查是否被取消
			select {
			case <-ctx.Done():
				fmt.Printf("DEBUG: 节点 %d 测试被取消（开始测试前）\n", nodeIndex)
				return
			default:
			}

			// 发送当前节点开始测试的消息
			if callback != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("ERROR: 发送开始测试回调时 panic: %v\n", r)
						}
					}()
					callback(&models.BatchTestProgress{
						Type:      "progress",
						Message:   fmt.Sprintf("正在测试节点 %d: %s", nodeIndex, node.Name),
						NodeIndex: nodeIndex,
						NodeName:  node.Name,
						Progress:  (completed * 100) / total,
						Total:     total,
						Completed: completed,
						Timestamp: time.Now().Format("2006-01-02 15:04:05"),
					})
				}()
			}

			// 执行单个节点测试，增加取消检查
			var result *models.NodeTestResult
			var err error

			testDone := make(chan struct{})
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("ERROR: TestNode panic: %v\n", r)
						err = fmt.Errorf("测试过程中发生内部错误: %v", r)
					}
					close(testDone)
				}()
				result, err = n.TestNode(subscriptionID, nodeIndex)
			}()

			// 等待测试完成、超时或取消
			select {
			case <-testDone:
				// 测试完成
			case <-time.After(30 * time.Second):
				err = fmt.Errorf("测试超时")
				fmt.Printf("WARNING: 节点 %d 测试超时\n", nodeIndex)
			case <-ctx.Done():
				err = fmt.Errorf("测试被取消")
				fmt.Printf("DEBUG: 节点 %d 测试被取消\n", nodeIndex)
			}

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result = &models.NodeTestResult{
					NodeName: node.Name,
					Success:  false,
					Error:    err.Error(),
					TestTime: time.Now(),
					TestType: "batch",
				}
				failureCount++
			} else {
				if result.Success {
					successCount++
				} else {
					failureCount++
				}
			}

			results = append(results, result)
			completed++

			// 发送进度更新
			if callback != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("ERROR: 发送完成回调时 panic: %v\n", r)
						}
					}()
					statusMsg := "测试完成"
					if result.Success {
						statusMsg = fmt.Sprintf("测试成功 (延迟: %s)", result.Latency)
					} else {
						statusMsg = fmt.Sprintf("测试失败: %s", result.Error)
					}

					callback(&models.BatchTestProgress{
						Type:          "progress",
						Message:       fmt.Sprintf("节点 %d (%s): %s", nodeIndex, node.Name, statusMsg),
						NodeIndex:     nodeIndex,
						NodeName:      node.Name,
						Progress:      (completed * 100) / total,
						Total:         total,
						Completed:     completed,
						SuccessCount:  successCount,
						FailureCount:  failureCount,
						CurrentResult: result,
						Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
					})
				}()
			}
		}(i, nodeIndex)
	}

	wg.Wait()

	// 检查最终状态
	select {
	case <-ctx.Done():
		fmt.Printf("DEBUG: 批量测试被取消，但已完成的测试结果仍会返回\n")
		// 发送取消事件
		if callback != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("ERROR: 发送取消事件时 panic: %v\n", r)
					}
				}()
				callback(&models.BatchTestProgress{
					Type:         "cancelled",
					Message:      fmt.Sprintf("批量测试已取消: 完成 %d，成功 %d，失败 %d", completed, successCount, failureCount),
					Progress:     (completed * 100) / total,
					Total:        total,
					Completed:    completed,
					SuccessCount: successCount,
					FailureCount: failureCount,
					Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
				})
			}()
		}
		return results, fmt.Errorf("批量测试被取消")
	default:
		// 发送完成事件
		if callback != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("ERROR: 发送完成事件时 panic: %v\n", r)
					}
				}()
				callback(&models.BatchTestProgress{
					Type:         "complete",
					Message:      fmt.Sprintf("批量测试完成: 成功 %d，失败 %d", successCount, failureCount),
					Progress:     100,
					Total:        total,
					Completed:    completed,
					SuccessCount: successCount,
					FailureCount: failureCount,
					Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
				})
			}()
		}
	}

	return results, nil
}

// startProxyForNode 为节点启动代理
func (n *NodeServiceImpl) startProxyForNode(node *types.Node, httpPort, socksPort int) (int, int, error) {
	// 为每个连接创建新的代理管理器实例，确保端口独立分配
	v2rayManager := proxy.NewProxyManager()
	hysteria2Manager := proxy.NewHysteria2ProxyManager()

	// 设置端口 - 只在指定了固定端口时才设置
	if httpPort > 0 || socksPort > 0 {
		v2rayManager.SetFixedPorts(httpPort, socksPort)
		hysteria2Manager.SetFixedPorts(httpPort, socksPort)
	}
	// 如果传入0，让管理器自动分配可用端口

	// 启动代理
	var err error
	var actualHTTPPort, actualSOCKSPort int

	if node.Protocol == "hysteria2" {
		// 启动Hysteria2代理
		err = hysteria2Manager.StartHysteria2Proxy(node)
		if err != nil {
			return 0, 0, err
		}
		status := hysteria2Manager.GetHysteria2Status()
		actualHTTPPort = status.HTTPPort
		actualSOCKSPort = status.SOCKSPort
	} else {
		// 启动V2Ray代理
		err = v2rayManager.StartProxy(node)
		if err != nil {
			return 0, 0, err
		}
		status := v2rayManager.GetStatus()
		actualHTTPPort = status.HTTPPort
		actualSOCKSPort = status.SOCKSPort
	}

	return actualHTTPPort, actualSOCKSPort, nil
}

// startProxyForNodeWithConnection 为节点启动代理并管理连接
func (n *NodeServiceImpl) startProxyForNodeWithConnection(subscriptionID string, nodeIndex int, node *types.Node, httpPort, socksPort int) (int, int, error) {
	fmt.Printf("DEBUG: 开始为节点启动代理 (协议:%s, HTTP端口:%d, SOCKS端口:%d)\n", node.Protocol, httpPort, socksPort)

	// 停止已存在的连接
	n.removeNodeConnection(subscriptionID, nodeIndex)
	fmt.Printf("DEBUG: 已清理旧连接\n")

	// 为每个连接创建新的代理管理器实例
	fmt.Printf("DEBUG: 创建代理管理器实例\n")
	v2rayManager := proxy.NewProxyManager()
	hysteria2Manager := proxy.NewHysteria2ProxyManager()

	// 设置端口 - 只在指定了固定端口时才设置
	if httpPort > 0 || socksPort > 0 {
		fmt.Printf("DEBUG: 设置固定端口 HTTP:%d, SOCKS:%d\n", httpPort, socksPort)
		v2rayManager.SetFixedPorts(httpPort, socksPort)
		hysteria2Manager.SetFixedPorts(httpPort, socksPort)
	} else {
		fmt.Printf("DEBUG: 使用自动端口分配\n")
	}

	// 启动代理
	var err error
	var actualHTTPPort, actualSOCKSPort int

	fmt.Printf("DEBUG: 启动%s代理\n", node.Protocol)
	if node.Protocol == "hysteria2" {
		// 启动Hysteria2代理
		err = hysteria2Manager.StartHysteria2Proxy(node)
		if err != nil {
			return 0, 0, err
		}
		status := hysteria2Manager.GetHysteria2Status()
		actualHTTPPort = status.HTTPPort
		actualSOCKSPort = status.SOCKSPort
	} else {
		// 启动V2Ray代理
		err = v2rayManager.StartProxy(node)
		if err != nil {
			return 0, 0, err
		}
		status := v2rayManager.GetStatus()
		actualHTTPPort = status.HTTPPort
		actualSOCKSPort = status.SOCKSPort
	}

	// 创建连接记录
	connection := &NodeConnection{
		V2RayManager:     v2rayManager,
		Hysteria2Manager: hysteria2Manager,
		HTTPPort:         actualHTTPPort,
		SOCKSPort:        actualSOCKSPort,
		Protocol:         node.Protocol,
		IsActive:         true,
	}

	// 添加到连接管理
	n.addNodeConnection(subscriptionID, nodeIndex, connection)

	// 更新节点状态为已连接，并保存端口信息到数据库
	n.setNodePorts(subscriptionID, nodeIndex, actualHTTPPort, actualSOCKSPort)
	n.updateNodeStatus(subscriptionID, nodeIndex, "connected")

	return actualHTTPPort, actualSOCKSPort, nil
}

// stopConnectionsByPort 停止占用指定端口的连接
func (n *NodeServiceImpl) stopConnectionsByPort(port int) {
	n.connectionMutex.Lock()
	defer n.connectionMutex.Unlock()

	var keysToRemove []string
	for key, connection := range n.nodeConnections {
		if connection.IsActive && (connection.HTTPPort == port || connection.SOCKSPort == port) {
			n.stopNodeConnection(connection)
			keysToRemove = append(keysToRemove, key)
		}
	}

	// 移除已停止的连接
	for _, key := range keysToRemove {
		delete(n.nodeConnections, key)
	}
}

// getNodeConnectionKey 获取节点连接键
func (n *NodeServiceImpl) getNodeConnectionKey(subscriptionID string, nodeIndex int) string {
	return fmt.Sprintf("%s_%d", subscriptionID, nodeIndex)
}

// addNodeConnection 添加节点连接
func (n *NodeServiceImpl) addNodeConnection(subscriptionID string, nodeIndex int, connection *NodeConnection) {
	key := n.getNodeConnectionKey(subscriptionID, nodeIndex)

	n.connectionMutex.Lock()
	defer n.connectionMutex.Unlock()

	// 停止已存在的连接
	if existingConn, exists := n.nodeConnections[key]; exists {
		n.stopNodeConnection(existingConn)
	}

	n.nodeConnections[key] = connection
}

// getNodeConnection 获取节点连接
func (n *NodeServiceImpl) getNodeConnection(subscriptionID string, nodeIndex int) *NodeConnection {
	key := n.getNodeConnectionKey(subscriptionID, nodeIndex)

	n.connectionMutex.RLock()
	defer n.connectionMutex.RUnlock()

	return n.nodeConnections[key]
}

// removeNodeConnection 移除节点连接
func (n *NodeServiceImpl) removeNodeConnection(subscriptionID string, nodeIndex int) {
	key := n.getNodeConnectionKey(subscriptionID, nodeIndex)

	n.connectionMutex.Lock()
	defer n.connectionMutex.Unlock()

	if connection, exists := n.nodeConnections[key]; exists {
		fmt.Printf("DEBUG: 移除节点连接 %s (端口: HTTP:%d, SOCKS:%d)\n", key, connection.HTTPPort, connection.SOCKSPort)
		n.stopNodeConnection(connection)
		delete(n.nodeConnections, key)
		fmt.Printf("DEBUG: 节点连接 %s 已成功移除\n", key)
	} else {
		fmt.Printf("DEBUG: 未找到要移除的节点连接 %s\n", key)
	}
}

// stopNodeConnection 停止节点连接
func (n *NodeServiceImpl) stopNodeConnection(connection *NodeConnection) {
	if connection == nil {
		fmt.Printf("DEBUG: 尝试停止空连接\n")
		return
	}

	fmt.Printf("DEBUG: 停止节点连接 (协议:%s, HTTP:%d, SOCKS:%d)\n", connection.Protocol, connection.HTTPPort, connection.SOCKSPort)
	connection.IsActive = false

	var err error
	if connection.V2RayManager != nil {
		err = connection.V2RayManager.StopProxy()
		if err != nil {
			fmt.Printf("DEBUG: 停止V2Ray代理失败: %v\n", err)
		} else {
			fmt.Printf("DEBUG: V2Ray代理已停止\n")
		}
	}
	if connection.Hysteria2Manager != nil {
		err = connection.Hysteria2Manager.StopHysteria2Proxy()
		if err != nil {
			fmt.Printf("DEBUG: 停止Hysteria2代理失败: %v\n", err)
		} else {
			fmt.Printf("DEBUG: Hysteria2代理已停止\n")
		}
	}
}

// stopProxyForNode 停止节点代理
func (n *NodeServiceImpl) stopProxyForNode(node *types.Node) error {
	// 这个方法现在需要知道具体是哪个节点，暂时保留向后兼容
	// 在实际使用中应该使用 removeNodeConnection
	return nil
}

// testV2RayNode 测试V2Ray节点
func (n *NodeServiceImpl) testV2RayNode(node *types.Node) error {
	// 获取唯一端口号，增加更大的间隔避免冲突
	portBase := int(atomic.AddInt64(&n.portCounter, 20))
	httpPort := portBase
	socksPort := portBase + 1

	// 确保端口可用
	for i := 0; i < 10; i++ {
		if n.isPortAvailable(httpPort) && n.isPortAvailable(socksPort) {
			break
		}
		portBase = int(atomic.AddInt64(&n.portCounter, 20))
		httpPort = portBase
		socksPort = portBase + 1
	}

	// 创建临时测试专用代理管理器，确保配置文件独立
	tempManager := proxy.NewTestProxyManager()
	tempManager.SetFixedPorts(httpPort, socksPort)
	defer func() {
		// 确保清理代理
		tempManager.StopProxy()
		// 给清理一些时间
		time.Sleep(1 * time.Second)
	}()

	// 启动代理
	err := tempManager.StartProxy(node)
	if err != nil {
		return fmt.Errorf("启动代理失败: %v", err)
	}

	// 等待代理启动，增加等待时间确保稳定
	time.Sleep(5 * time.Second)

	// 验证代理是否真正运行
	if !tempManager.IsRunning() {
		return fmt.Errorf("代理启动后未能正常运行")
	}

	// 测试代理连接
	return tempManager.TestProxy()
}

// testHysteria2Node 测试Hysteria2节点
func (n *NodeServiceImpl) testHysteria2Node(node *types.Node) error {
	// 获取唯一端口号，增加更大的间隔避免冲突
	portBase := int(atomic.AddInt64(&n.portCounter, 20))
	httpPort := portBase
	socksPort := portBase + 1

	// 确保端口可用
	for i := 0; i < 10; i++ {
		if n.isPortAvailable(httpPort) && n.isPortAvailable(socksPort) {
			break
		}
		portBase = int(atomic.AddInt64(&n.portCounter, 20))
		httpPort = portBase
		socksPort = portBase + 1
	}

	// 创建临时测试专用代理管理器，确保配置文件独立
	tempManager := proxy.NewTestHysteria2ProxyManager()
	tempManager.SetFixedPorts(httpPort, socksPort)
	defer func() {
		// 确保清理代理
		tempManager.StopHysteria2Proxy()
		// 给清理一些时间
		time.Sleep(1 * time.Second)
	}()

	// 启动代理
	err := tempManager.StartHysteria2Proxy(node)
	if err != nil {
		return fmt.Errorf("启动代理失败: %v", err)
	}

	// 等待代理启动，增加等待时间确保稳定
	time.Sleep(5 * time.Second)

	// 验证代理是否真正运行
	if !tempManager.IsHysteria2Running() {
		return fmt.Errorf("代理启动后未能正常运行")
	}

	// 测试代理连接
	return tempManager.TestHysteria2Proxy()
}

// speedTestV2RayNode V2Ray节点速度测试
func (n *NodeServiceImpl) speedTestV2RayNode(node *types.Node) (float64, float64, float64, error) {
	// 获取唯一端口号，增加更大的间隔避免冲突
	portBase := int(atomic.AddInt64(&n.portCounter, 20))
	httpPort := portBase
	socksPort := portBase + 1

	// 确保端口可用
	for i := 0; i < 10; i++ {
		if n.isPortAvailable(httpPort) && n.isPortAvailable(socksPort) {
			break
		}
		portBase = int(atomic.AddInt64(&n.portCounter, 20))
		httpPort = portBase
		socksPort = portBase + 1
	}

	// 创建临时测试专用代理管理器，确保配置文件独立
	tempManager := proxy.NewTestProxyManager()
	tempManager.SetFixedPorts(httpPort, socksPort)
	defer func() {
		tempManager.StopProxy()
		time.Sleep(1 * time.Second)
	}()

	// 启动代理
	err := tempManager.StartProxy(node)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("启动代理失败: %v", err)
	}

	// 等待代理启动，增加等待时间确保稳定
	time.Sleep(5 * time.Second)

	// 验证代理是否真正运行
	if !tempManager.IsRunning() {
		return 0, 0, 0, fmt.Errorf("代理启动后未能正常运行")
	}

	// 测试代理连接
	err = tempManager.TestProxy()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("代理测试失败: %v", err)
	}

	// 执行真实的速度测试
	return n.performRealSpeedTest(tempManager.GetStatus().HTTPPort, tempManager.GetStatus().SOCKSPort)
}

// speedTestHysteria2Node Hysteria2节点速度测试
func (n *NodeServiceImpl) speedTestHysteria2Node(node *types.Node) (float64, float64, float64, error) {
	// 获取唯一端口号，增加更大的间隔避免冲突
	portBase := int(atomic.AddInt64(&n.portCounter, 20))
	httpPort := portBase
	socksPort := portBase + 1

	// 确保端口可用
	for i := 0; i < 10; i++ {
		if n.isPortAvailable(httpPort) && n.isPortAvailable(socksPort) {
			break
		}
		portBase = int(atomic.AddInt64(&n.portCounter, 20))
		httpPort = portBase
		socksPort = portBase + 1
	}

	// 创建临时测试专用代理管理器，确保配置文件独立
	tempManager := proxy.NewTestHysteria2ProxyManager()
	tempManager.SetFixedPorts(httpPort, socksPort)
	defer func() {
		tempManager.StopHysteria2Proxy()
		time.Sleep(1 * time.Second)
	}()

	// 启动代理
	err := tempManager.StartHysteria2Proxy(node)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("启动代理失败: %v", err)
	}

	// 等待代理启动，增加等待时间确保稳定
	time.Sleep(5 * time.Second)

	// 验证代理是否真正运行
	if !tempManager.IsHysteria2Running() {
		return 0, 0, 0, fmt.Errorf("代理启动后未能正常运行")
	}

	// 测试代理连接
	err = tempManager.TestHysteria2Proxy()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("代理测试失败: %v", err)
	}

	// 执行真实的速度测试
	status := tempManager.GetHysteria2Status()
	return n.performRealSpeedTest(status.HTTPPort, status.SOCKSPort)
}

// performRealSpeedTest 执行真实的速度测试
func (n *NodeServiceImpl) performRealSpeedTest(httpPort, socksPort int) (float64, float64, float64, error) {
	// 使用HTTP代理进行速度测试
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", httpPort)

	// 测试延迟
	latencyStart := time.Now()
	err := n.testProxyLatency(proxyURL)
	latency := float64(time.Since(latencyStart).Milliseconds())

	if err != nil {
		return 0, 0, latency, fmt.Errorf("延迟测试失败: %v", err)
	}

	// 测试下载速度 - 通过代理下载测试文件
	downloadSpeed, err := n.testDownloadSpeed(proxyURL)
	if err != nil {
		return 0, 0, latency, fmt.Errorf("下载速度测试失败: %v", err)
	}

	// 测试上传速度 - 通过代理上传测试数据
	uploadSpeed, err := n.testUploadSpeed(proxyURL)
	if err != nil {
		return downloadSpeed, 0, latency, fmt.Errorf("上传速度测试失败: %v", err)
	}

	return downloadSpeed, uploadSpeed, latency, nil
}

// testProxyLatency 测试代理延迟
func (n *NodeServiceImpl) testProxyLatency(proxyURL string) error {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(proxyURL)
			},
		},
		Timeout: 10 * time.Second,
	}

	// 测试访问Google
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		// 如果Google不可达，尝试其他网站
		resp, err = client.Get("https://httpbin.org/ip")
		if err != nil {
			return fmt.Errorf("无法通过代理访问测试网站: %v", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	return nil
}

// testDownloadSpeed 测试下载速度
func (n *NodeServiceImpl) testDownloadSpeed(proxyURL string) (float64, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(proxyURL)
			},
		},
		Timeout: 30 * time.Second,
	}

	// 使用较小的测试文件进行快速测试
	testURL := "https://httpbin.org/bytes/1048576" // 1MB测试文件

	start := time.Now()
	resp, err := client.Get(testURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// 读取数据并计算速度
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	duration := time.Since(start).Seconds()
	if duration == 0 {
		duration = 0.001 // 避免除零
	}

	// 计算速度 (Mbps)
	bytes := float64(len(data))
	mbps := (bytes * 8) / (duration * 1024 * 1024)

	return mbps, nil
}

// testUploadSpeed 测试上传速度
func (n *NodeServiceImpl) testUploadSpeed(proxyURL string) (float64, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(proxyURL)
			},
		},
		Timeout: 30 * time.Second,
	}

	// 创建1MB的测试数据
	testData := make([]byte, 1048576) // 1MB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	start := time.Now()
	resp, err := client.Post("https://httpbin.org/post", "application/octet-stream", bytes.NewReader(testData))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	duration := time.Since(start).Seconds()
	if duration == 0 {
		duration = 0.001 // 避免除零
	}

	// 计算上传速度 (Mbps)
	bytes := float64(len(testData))
	mbps := (bytes * 8) / (duration * 1024 * 1024)

	return mbps, nil
}

// 状态管理方法

// updateNodeStatus 更新节点状态
func (n *NodeServiceImpl) updateNodeStatus(subscriptionID string, nodeIndex int, status string) {
	key := fmt.Sprintf("%s_%d", subscriptionID, nodeIndex)

	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// 更新内部状态
	if nodeState, exists := n.nodeStates[key]; exists {
		nodeState.UpdateStatus(status)
	}

	// 同步到订阅数据
	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err == nil && nodeIndex < len(subscription.Nodes) {
		subscription.Nodes[nodeIndex].UpdateStatus(status)
		
		// 更新数据库中的节点状态
		if err := n.nodeDB.Update(subscription.Nodes[nodeIndex], subscriptionID); err != nil {
			fmt.Printf("WARNING: 更新节点状态到数据库失败: %v\n", err)
		}
	}
}

// setNodePorts 设置节点端口
func (n *NodeServiceImpl) setNodePorts(subscriptionID string, nodeIndex int, httpPort, socksPort int) {
	key := fmt.Sprintf("%s_%d", subscriptionID, nodeIndex)

	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// 更新内部状态
	if nodeState, exists := n.nodeStates[key]; exists {
		nodeState.SetPorts(httpPort, socksPort)
	}

	// 同步到订阅数据
	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err == nil && nodeIndex < len(subscription.Nodes) {
		subscription.Nodes[nodeIndex].SetPorts(httpPort, socksPort)
		
		// 更新数据库中的节点状态
		if err := n.nodeDB.Update(subscription.Nodes[nodeIndex], subscriptionID); err != nil {
			fmt.Printf("WARNING: 更新节点端口到数据库失败: %v\n", err)
		}
	}
}

// setNodeTestResult 设置节点测试结果
func (n *NodeServiceImpl) setNodeTestResult(subscriptionID string, nodeIndex int, result *models.NodeTestResult) {
	key := fmt.Sprintf("%s_%d", subscriptionID, nodeIndex)

	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// 更新内部状态
	if nodeState, exists := n.nodeStates[key]; exists {
		nodeState.SetTestResult(result)
	}

	// 同步到订阅数据
	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err == nil && nodeIndex < len(subscription.Nodes) {
		subscription.Nodes[nodeIndex].SetTestResult(result)
	}
}

// setNodeSpeedResult 设置节点速度测试结果
func (n *NodeServiceImpl) setNodeSpeedResult(subscriptionID string, nodeIndex int, result *models.SpeedTestResult) {
	key := fmt.Sprintf("%s_%d", subscriptionID, nodeIndex)

	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// 更新内部状态
	if nodeState, exists := n.nodeStates[key]; exists {
		nodeState.SetSpeedResult(result)
	}

	// 同步到订阅数据
	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err == nil && nodeIndex < len(subscription.Nodes) {
		subscription.Nodes[nodeIndex].SetSpeedResult(result)
	}
}

// ensureNodeState 确保节点状态存在
func (n *NodeServiceImpl) ensureNodeState(subscriptionID string, nodeIndex int, nodeInfo *models.NodeInfo) {
	key := fmt.Sprintf("%s_%d", subscriptionID, nodeIndex)

	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	if _, exists := n.nodeStates[key]; !exists {
		n.nodeStates[key] = nodeInfo
	}
}

// isPortOccupied 检查端口是否被占用
func (n *NodeServiceImpl) isPortOccupied(port int) bool {
	n.connectionMutex.RLock()
	defer n.connectionMutex.RUnlock()

	// 检查所有活跃的连接是否占用了该端口
	for _, connection := range n.nodeConnections {
		if connection.IsActive && (connection.HTTPPort == port || connection.SOCKSPort == port) {
			return true
		}
	}
	return false
}

// isPortAvailable 检查端口是否可用
func (n *NodeServiceImpl) isPortAvailable(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// DeleteNodes 删除节点
func (n *NodeServiceImpl) DeleteNodes(subscriptionID string, nodeIndexes []int) error {
	if subscriptionID == "" {
		return fmt.Errorf("订阅ID不能为空")
	}

	if len(nodeIndexes) == 0 {
		return fmt.Errorf("节点索引列表不能为空")
	}

	fmt.Printf("DEBUG: 删除节点请求 - subscription_id: %s, node_indexes: %v\n", subscriptionID, nodeIndexes)

	// 获取订阅
	subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err != nil {
		return fmt.Errorf("获取订阅失败: %v", err)
	}

	if subscription.Nodes == nil || len(subscription.Nodes) == 0 {
		return fmt.Errorf("订阅中没有节点")
	}

	// 验证所有节点索引的有效性
	for _, nodeIndex := range nodeIndexes {
		if nodeIndex < 0 || nodeIndex >= len(subscription.Nodes) {
			return fmt.Errorf("节点索引 %d 无效", nodeIndex)
		}
	}

	// 停止要删除节点的连接
	for _, nodeIndex := range nodeIndexes {
		// 移除节点连接
		n.removeNodeConnection(subscriptionID, nodeIndex)
		
		// 清理节点状态
		key := fmt.Sprintf("%s_%d", subscriptionID, nodeIndex)
		n.stateMutex.Lock()
		delete(n.nodeStates, key)
		n.stateMutex.Unlock()
		
		fmt.Printf("DEBUG: 已清理节点 %d 的连接和状态\n", nodeIndex)
	}

	// 从数据库删除节点
	err = n.nodeDB.DeleteByIndexes(subscriptionID, nodeIndexes)
	if err != nil {
		return fmt.Errorf("从数据库删除节点失败: %v", err)
	}

	// 重新索引剩余节点
	err = n.nodeDB.ReindexNodes(subscriptionID)
	if err != nil {
		return fmt.Errorf("重新索引节点失败: %v", err)
	}

	// 从订阅对象中删除节点（从大到小排序，避免索引变化影响）
	sortedIndexes := make([]int, len(nodeIndexes))
	copy(sortedIndexes, nodeIndexes)
	
	// 简单的冒泡排序（从大到小）
	for i := 0; i < len(sortedIndexes)-1; i++ {
		for j := 0; j < len(sortedIndexes)-1-i; j++ {
			if sortedIndexes[j] < sortedIndexes[j+1] {
				sortedIndexes[j], sortedIndexes[j+1] = sortedIndexes[j+1], sortedIndexes[j]
			}
		}
	}

	// 从大到小删除节点，避免索引偏移
	for _, nodeIndex := range sortedIndexes {
		if nodeIndex < len(subscription.Nodes) {
			subscription.Nodes = append(subscription.Nodes[:nodeIndex], subscription.Nodes[nodeIndex+1:]...)
			fmt.Printf("DEBUG: 已从订阅中删除节点索引 %d\n", nodeIndex)
		}
	}

	// 更新节点计数
	subscription.NodeCount = len(subscription.Nodes)
	
	// 更新订阅（通过subscription service保存）
	err = n.subscriptionService.UpdateSubscription(subscription)
	if err != nil {
		return fmt.Errorf("更新订阅失败: %v", err)
	}

	fmt.Printf("DEBUG: 成功删除 %d 个节点\n", len(nodeIndexes))
	return nil
}

// stopAllProxies 停止所有代理
func (n *NodeServiceImpl) stopAllProxies() error {
	n.connectionMutex.Lock()
	defer n.connectionMutex.Unlock()

	for key, connection := range n.nodeConnections {
		n.stopNodeConnection(connection)
		delete(n.nodeConnections, key)
	}

	return nil
}

// GetActiveConnections 获取所有活跃的代理连接
func (n *NodeServiceImpl) GetActiveConnections() []*models.ActiveConnection {
	n.connectionMutex.RLock()
	defer n.connectionMutex.RUnlock()

	var connections []*models.ActiveConnection
	for key, connection := range n.nodeConnections {
		if connection.IsActive {
			// 解析订阅ID和节点索引
			parts := strings.Split(key, "_")
			if len(parts) >= 2 {
				subscriptionID := strings.Join(parts[:len(parts)-1], "_")
				nodeIndex := parts[len(parts)-1]
				
				// 获取节点信息
				if subscription, err := n.subscriptionService.GetSubscriptionByID(subscriptionID); err == nil {
					if idx, err := strconv.Atoi(nodeIndex); err == nil && idx >= 0 && idx < len(subscription.Nodes) {
						nodeInfo := subscription.Nodes[idx]
						
						activeConn := &models.ActiveConnection{
							SubscriptionID:   subscriptionID,
							SubscriptionName: subscription.Name,
							NodeIndex:        idx,
							NodeName:         nodeInfo.Name,
							Protocol:         connection.Protocol,
							HTTPPort:         connection.HTTPPort,
							SOCKSPort:        connection.SOCKSPort,
							Server:           nodeInfo.Server,
							ConnectTime:      nodeInfo.ConnectTime,
							IsActive:         connection.IsActive,
						}
						connections = append(connections, activeConn)
					}
				}
			}
		}
	}

	return connections
}

// StopAllActiveConnections 停止所有活跃连接
func (n *NodeServiceImpl) StopAllActiveConnections() error {
	n.connectionMutex.Lock()
	defer n.connectionMutex.Unlock()

	var keys []string
	for key, connection := range n.nodeConnections {
		if connection.IsActive {
			n.stopNodeConnection(connection)
			keys = append(keys, key)
		}
	}

	// 清理连接记录
	for _, key := range keys {
		delete(n.nodeConnections, key)
		
		// 更新节点状态
		parts := strings.Split(key, "_")
		if len(parts) >= 2 {
			subscriptionID := strings.Join(parts[:len(parts)-1], "_")
			if nodeIndex, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				n.setNodePorts(subscriptionID, nodeIndex, 0, 0)
				n.updateNodeStatus(subscriptionID, nodeIndex, "idle")
			}
		}
	}

	return nil
}

// StopAllNodeConnections 停止所有节点连接
func (n *NodeServiceImpl) StopAllNodeConnections() error {
	n.connectionMutex.Lock()
	defer n.connectionMutex.Unlock()
	
	fmt.Printf("🛑 正在停止所有节点连接...\n")
	var errors []error
	stoppedCount := 0
	
	for key, connection := range n.nodeConnections {
		fmt.Printf("🔌 停止连接: %s (协议:%s, HTTP:%d, SOCKS:%d)\n", 
			key, connection.Protocol, connection.HTTPPort, connection.SOCKSPort)
		
		n.stopNodeConnection(connection)
		stoppedCount++
	}
	
	// 清空所有连接
	n.nodeConnections = make(map[string]*NodeConnection)
	
	fmt.Printf("✅ 已停止 %d 个节点连接\n", stoppedCount)
	
	if len(errors) > 0 {
		return fmt.Errorf("停止部分连接时发生错误: %v", errors)
	}
	
	return nil
}
