package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/proxy"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/workflow"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// NodeServiceImpl 节点服务实现
type NodeServiceImpl struct {
	subscriptionService SubscriptionService
	proxyService        ProxyService

	// 真实的代理管理器
	v2rayManager     *proxy.ProxyManager
	hysteria2Manager *proxy.Hysteria2ProxyManager

	// 节点状态管理
	nodeStates map[string]*models.NodeInfo // key: subscriptionID_nodeIndex
	stateMutex sync.RWMutex

	// MVP测试器
	mvpTester *workflow.MVPTester

	// 端口分配计数器（用于批量测试时避免端口冲突）
	portCounter int64
}

// NewNodeService 创建节点服务
func NewNodeService(subscriptionService SubscriptionService, proxyService ProxyService) NodeService {
	service := &NodeServiceImpl{
		subscriptionService: subscriptionService,
		proxyService:        proxyService,
		v2rayManager:        proxy.NewProxyManager(),
		hysteria2Manager:    proxy.NewHysteria2ProxyManager(),
		nodeStates:          make(map[string]*models.NodeInfo),
		portCounter:         9000, // 测试端口从9000开始
	}

	// 初始化MVP测试器
	service.mvpTester = workflow.NewMVPTester("")

	return service
}

// ConnectNode 连接节点
func (n *NodeServiceImpl) ConnectNode(subscriptionID string, nodeIndex int, operation string) (*models.ConnectNodeResponse, error) {
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

	switch operation {
	case "http_random":
		// 随机HTTP端口连接 - 自动分配可用端口
		actualHTTPPort, _, err := n.startProxyForNode(nodeInfo.Node, 0, 0) // 传入0表示随机分配
		if err != nil {
			n.updateNodeStatus(subscriptionID, nodeIndex, "error")
			return nil, fmt.Errorf("启动HTTP代理失败: %v", err)
		}
		// 只返回HTTP端口
		response.HTTPPort = actualHTTPPort
		response.Port = actualHTTPPort
		n.setNodePorts(subscriptionID, nodeIndex, actualHTTPPort, 0)

	case "socks_random":
		// 随机SOCKS端口连接 - 自动分配可用端口
		_, actualSOCKSPort, err := n.startProxyForNode(nodeInfo.Node, 0, 0) // 传入0表示随机分配
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

		// 检查端口是否被占用，如果被占用则停止之前的代理
		if n.isPortOccupied(fixedHTTPPort) {
			err := n.stopAllProxies()
			if err != nil {
				n.updateNodeStatus(subscriptionID, nodeIndex, "error")
				return nil, fmt.Errorf("停止之前的代理失败: %v", err)
			}
		}

		actualHTTPPort, _, err := n.startProxyForNode(nodeInfo.Node, fixedHTTPPort, 0)
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

		// 检查端口是否被占用，如果被占用则停止之前的代理
		if n.isPortOccupied(fixedSOCKSPort) {
			err := n.stopAllProxies()
			if err != nil {
				n.updateNodeStatus(subscriptionID, nodeIndex, "error")
				return nil, fmt.Errorf("停止之前的代理失败: %v", err)
			}
		}

		_, actualSOCKSPort, err := n.startProxyForNode(nodeInfo.Node, 0, fixedSOCKSPort)
		if err != nil {
			n.updateNodeStatus(subscriptionID, nodeIndex, "error")
			return nil, fmt.Errorf("启动固定SOCKS代理失败: %v", err)
		}
		response.SOCKSPort = actualSOCKSPort
		response.Port = actualSOCKSPort
		n.setNodePorts(subscriptionID, nodeIndex, 0, actualSOCKSPort)

	case "disable":
		// 禁用节点（停止代理）
		err := n.stopProxyForNode(nodeInfo.Node)
		if err != nil {
			return nil, fmt.Errorf("停止代理失败: %v", err)
		}
		n.setNodePorts(subscriptionID, nodeIndex, 0, 0)

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
	// 停止现有代理
	n.v2rayManager.StopProxy()
	n.hysteria2Manager.StopHysteria2Proxy()

	if node.Protocol == "hysteria2" {
		// 为每次连接创建新的代理管理器实例，确保端口独立分配
		manager := proxy.NewHysteria2ProxyManager()

		// 设置端口 - 只在指定了固定端口时才设置
		if httpPort > 0 || socksPort > 0 {
			manager.SetFixedPorts(httpPort, socksPort)
		}
		// 如果传入0，让管理器自动分配可用端口

		// 启动Hysteria2代理
		err := manager.StartHysteria2Proxy(node)
		if err != nil {
			return 0, 0, err
		}

		// 更新实例引用
		n.hysteria2Manager = manager
		status := manager.GetHysteria2Status()
		return status.HTTPPort, status.SOCKSPort, nil
	} else {
		// 为每次连接创建新的代理管理器实例，确保端口独立分配
		manager := proxy.NewProxyManager()

		// 设置端口 - 只在指定了固定端口时才设置
		if httpPort > 0 || socksPort > 0 {
			manager.SetFixedPorts(httpPort, socksPort)
		}
		// 如果传入0，让管理器自动分配可用端口

		// 启动V2Ray代理
		err := manager.StartProxy(node)
		if err != nil {
			return 0, 0, err
		}

		// 更新实例引用
		n.v2rayManager = manager
		status := manager.GetStatus()
		return status.HTTPPort, status.SOCKSPort, nil
	}
}

// stopProxyForNode 停止节点代理
func (n *NodeServiceImpl) stopProxyForNode(node *types.Node) error {
	if node.Protocol == "hysteria2" {
		return n.hysteria2Manager.StopHysteria2Proxy()
	} else {
		return n.v2rayManager.StopProxy()
	}
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
	if n.v2rayManager.IsPortOccupied(port) || n.hysteria2Manager.IsPortOccupied(port) {
		return true
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

// stopAllProxies 停止所有代理
func (n *NodeServiceImpl) stopAllProxies() error {
	err := n.v2rayManager.StopProxy()
	if err != nil {
		return err
	}
	err = n.hysteria2Manager.StopHysteria2Proxy()
	if err != nil {
		return err
	}
	return nil
}
