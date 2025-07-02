package models

import (
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// APIResponse 统一的API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Subscription 订阅信息
type Subscription struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	URL        string      `json:"url"`
	Nodes      []*NodeInfo `json:"nodes"` // 改为包含状态信息的节点
	NodeCount  int         `json:"node_count"`
	LastUpdate string      `json:"last_update"`
	Status     string      `json:"status"`
	CreateTime string      `json:"create_time"`
}

// NodeInfo 包含状态信息的节点
type NodeInfo struct {
	*types.Node                  // 嵌入原始节点信息
	Index       int              `json:"index"`        // 节点索引
	Status      string           `json:"status"`       // 节点状态: idle, connecting, connected, testing, error
	IsRunning   bool             `json:"is_running"`   // 是否正在运行
	HTTPPort    int              `json:"http_port"`    // HTTP代理端口
	SOCKSPort   int              `json:"socks_port"`   // SOCKS代理端口
	TestResult  *NodeTestResult  `json:"test_result"`  // 最新测试结果
	SpeedResult *SpeedTestResult `json:"speed_result"` // 最新速度测试结果
	LastTest    time.Time        `json:"last_test"`    // 最后测试时间
	ConnectTime time.Time        `json:"connect_time"` // 连接时间
}

// NodeTestResult 节点测试结果
type NodeTestResult struct {
	NodeName string    `json:"node_name"`
	Success  bool      `json:"success"`
	Latency  string    `json:"latency,omitempty"`
	Error    string    `json:"error,omitempty"`
	TestTime time.Time `json:"test_time"`
	TestType string    `json:"test_type"` // tcp, http, full
}

// SpeedTestResult 速度测试结果
type SpeedTestResult struct {
	NodeName      string    `json:"node_name"`
	DownloadSpeed string    `json:"download_speed"`
	UploadSpeed   string    `json:"upload_speed"`
	Latency       string    `json:"latency"`
	TestTime      time.Time `json:"test_time"`
	TestDuration  string    `json:"test_duration"`
}

// ProxyStatus 代理状态
type ProxyStatus struct {
	V2RayRunning     bool        `json:"v2ray_running"`
	Hysteria2Running bool        `json:"hysteria2_running"`
	HTTPPort         int         `json:"http_port"`
	SOCKSPort        int         `json:"socks_port"`
	CurrentNode      string      `json:"current_node"`
	ActiveNodes      []*NodeInfo `json:"active_nodes"`      // 活跃节点列表
	TotalConnections int         `json:"total_connections"` // 总连接数
}

// SystemStatus 系统状态
type SystemStatus struct {
	ProxyPorts   map[string]int `json:"proxy_ports"`
	ServerStatus string         `json:"server_status"`
	Timestamp    int64          `json:"timestamp"`
	Version      string         `json:"version"`
	ActiveNodes  int            `json:"active_nodes"` // 活跃节点数
	TotalNodes   int            `json:"total_nodes"`  // 总节点数
	TestResults  int            `json:"test_results"` // 测试结果数
}

// AddSubscriptionRequest 添加订阅请求
type AddSubscriptionRequest struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// ParseSubscriptionRequest 解析订阅请求
type ParseSubscriptionRequest struct {
	ID string `json:"id"`
}

// DeleteSubscriptionRequest 删除订阅请求
type DeleteSubscriptionRequest struct {
	ID string `json:"id"`
}

// TestSubscriptionRequest 测试订阅请求
type TestSubscriptionRequest struct {
	ID string `json:"id"`
}

// NodeOperationRequest 节点操作请求
type NodeOperationRequest struct {
	SubscriptionID string `json:"subscription_id"`
	NodeIndex      int    `json:"node_index"`
	Operation      string `json:"operation"`
}

// BatchNodeOperationRequest 批量节点操作请求
type BatchNodeOperationRequest struct {
	SubscriptionID string `json:"subscription_id"`
	NodeIndexes    []int  `json:"node_indexes"`
	Operation      string `json:"operation"`
}

// ConnectNodeResponse 连接节点响应
type ConnectNodeResponse struct {
	Port      int    `json:"port,omitempty"`
	HTTPPort  int    `json:"http_port,omitempty"`
	SOCKSPort int    `json:"socks_port,omitempty"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// BatchTestResponse 批量测试响应
type BatchTestResponse struct {
	Results      []*NodeTestResult `json:"results"`
	SuccessCount int               `json:"success_count"`
	FailureCount int               `json:"failure_count"`
	TotalCount   int               `json:"total_count"`
}

// Settings 系统设置
type Settings struct {
	HTTPPort  int    `json:"http_port"`
	SOCKSPort int    `json:"socks_port"`
	TestURL   string `json:"test_url"`
}

// BatchTestProgress 批量测试进度
type BatchTestProgress struct {
	Type          string          `json:"type"`           // 事件类型: start, progress, complete, error
	Message       string          `json:"message"`        // 状态消息
	NodeIndex     int             `json:"node_index"`     // 当前测试的节点索引
	NodeName      string          `json:"node_name"`      // 节点名称
	Progress      int             `json:"progress"`       // 进度百分比
	Total         int             `json:"total"`          // 总节点数
	Completed     int             `json:"completed"`      // 已完成数
	SuccessCount  int             `json:"success_count"`  // 成功数
	FailureCount  int             `json:"failure_count"`  // 失败数
	CurrentResult *NodeTestResult `json:"current_result"` // 当前节点测试结果
	Timestamp     string          `json:"timestamp"`      // 时间戳
}

// NewAPIResponse 创建API响应
func NewAPIResponse() *APIResponse {
	return &APIResponse{
		Success: true,
		Message: "",
	}
}

// SetSuccess 设置成功响应
func (r *APIResponse) SetSuccess(data interface{}, message string) *APIResponse {
	r.Success = true
	r.Data = data
	r.Message = message
	return r
}

// SetError 设置错误响应
func (r *APIResponse) SetError(err error, message string) *APIResponse {
	r.Success = false
	r.Error = err.Error()
	r.Message = message
	return r
}

// NewSubscription 创建新订阅
func NewSubscription(id, name, url string) *Subscription {
	return &Subscription{
		ID:         id,
		Name:       name,
		URL:        url,
		Nodes:      []*NodeInfo{},
		NodeCount:  0,
		LastUpdate: "从未更新",
		Status:     "inactive",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
}

// NewNodeInfo 创建节点信息
func NewNodeInfo(node *types.Node, index int) *NodeInfo {
	return &NodeInfo{
		Node:      node,
		Index:     index,
		Status:    "idle",
		IsRunning: false,
		HTTPPort:  0,
		SOCKSPort: 0,
		LastTest:  time.Time{},
	}
}

// UpdateStatus 更新节点状态
func (n *NodeInfo) UpdateStatus(status string) {
	n.Status = status
}

// SetPorts 设置代理端口
func (n *NodeInfo) SetPorts(httpPort, socksPort int) {
	n.HTTPPort = httpPort
	n.SOCKSPort = socksPort
	if httpPort > 0 || socksPort > 0 {
		n.IsRunning = true
		n.ConnectTime = time.Now()
		n.Status = "connected"
	} else {
		n.IsRunning = false
		n.Status = "idle"
	}
}

// SetTestResult 设置测试结果
func (n *NodeInfo) SetTestResult(result *NodeTestResult) {
	n.TestResult = result
	n.LastTest = time.Now()
}

// SetSpeedResult 设置速度测试结果
func (n *NodeInfo) SetSpeedResult(result *SpeedTestResult) {
	n.SpeedResult = result
	n.LastTest = time.Now()
}
