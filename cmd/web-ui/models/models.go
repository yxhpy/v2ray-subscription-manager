package models

import (
	"fmt"
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

// PortConflictInfo 端口冲突信息
type PortConflictInfo struct {
	HasConflict        bool   `json:"has_conflict"`
	Port               int    `json:"port"`
	ProtocolType       string `json:"protocol_type"`      // HTTP 或 SOCKS
	ConflictNodeIndex  int    `json:"conflict_node_index,omitempty"`
	ConflictNodeName   string `json:"conflict_node_name,omitempty"`
	SubscriptionID     string `json:"subscription_id,omitempty"`
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
	// 代理设置
	HTTPPort  int  `json:"http_port"`
	SOCKSPort int  `json:"socks_port"`
	AllowLan  bool `json:"allow_lan"`
	
	// 测试设置
	TestURL       string `json:"test_url"`
	TestTimeout   int    `json:"test_timeout"`
	MaxConcurrent int    `json:"max_concurrent"`
	RetryCount    int    `json:"retry_count"`
	
	// 订阅设置
	UpdateInterval   int    `json:"update_interval"`
	UserAgent        string `json:"user_agent"`
	AutoTestNewNodes bool   `json:"auto_test_new_nodes"`
	
	// 安全设置
	EnableLogs    bool   `json:"enable_logs"`
	LogLevel      string `json:"log_level"`
	DataRetention int    `json:"data_retention"`
}

// ActiveConnection 活跃的代理连接
type ActiveConnection struct {
	SubscriptionID   string    `json:"subscription_id"`
	SubscriptionName string    `json:"subscription_name"`
	NodeIndex        int       `json:"node_index"`
	NodeName         string    `json:"node_name"`
	Protocol         string    `json:"protocol"`
	HTTPPort         int       `json:"http_port"`
	SOCKSPort        int       `json:"socks_port"`
	Server           string    `json:"server"`
	ConnectTime      time.Time `json:"connect_time"`
	IsActive         bool      `json:"is_active"`
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

// SmartConnectionStatus 智能连接状态
type SmartConnectionStatus struct {
	IsRunning   bool                    `json:"is_running"`
	TotalPools  int                     `json:"total_pools"`
	ActivePools int                     `json:"active_pools"`
	TotalRules  int                     `json:"total_rules"`
	ActiveRules int                     `json:"active_rules"`
	GlobalStats *GlobalConnectionStats `json:"global_stats"`
	Uptime      int64                   `json:"uptime"`
	LastUpdate  time.Time               `json:"last_update"`
}

// GlobalConnectionStats 全局连接统计
type GlobalConnectionStats struct {
	TotalConnections   int `json:"total_connections"`
	ActiveConnections  int `json:"active_connections"`
	HealthyConnections int `json:"healthy_connections"`
}

// ConnectionPool 连接池
type ConnectionPool struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Config      *ConnectionPoolConfig     `json:"config"`
	Connections []*PoolConnection        `json:"connections"`
	Status      *ConnectionPoolStatus     `json:"status"`
	CreateTime  time.Time                 `json:"create_time"`
	UpdateTime  time.Time                 `json:"update_time"`
}

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	MaxConnections      int    `json:"max_connections"`
	MinHealthyNodes     int    `json:"min_healthy_nodes"`
	LoadBalanceMode     string `json:"load_balance_mode"`
	HealthCheckInterval int    `json:"health_check_interval"`
	HealthCheckTimeout  int    `json:"health_check_timeout"`
	HealthCheckURL      string `json:"health_check_url"`
	FailoverThreshold   int    `json:"failover_threshold"`
	RecoveryThreshold   int    `json:"recovery_threshold"`
	AutoRebalance       bool   `json:"auto_rebalance"`
	RebalanceInterval   int    `json:"rebalance_interval"`
}

// ConnectionPoolStatus 连接池状态
type ConnectionPoolStatus struct {
	IsRunning         bool      `json:"is_running"`
	ActiveConnections int       `json:"active_connections"`
	HealthyNodes      int       `json:"healthy_nodes"`
	TotalTraffic      int64     `json:"total_traffic"`
	LastUpdate        time.Time `json:"last_update"`
}

// PoolConnection 连接池连接
type PoolConnection struct {
	ID             string             `json:"id"`
	SubscriptionID string             `json:"subscription_id"`
	NodeIndex      int                `json:"node_index"`
	NodeName       string             `json:"node_name"`
	Protocol       string             `json:"protocol"`
	Server         string             `json:"server"`
	Status         string             `json:"status"`
	Health         *ConnectionHealth  `json:"health"`
	Stats          *ConnectionStats   `json:"stats"`
	Weight         int                `json:"weight"`
	Priority       int                `json:"priority"`
	CreateTime     time.Time          `json:"create_time"`
	LastCheck      time.Time          `json:"last_check"`
}

// ConnectionHealth 连接健康状态
type ConnectionHealth struct {
	IsHealthy         bool      `json:"is_healthy"`
	Latency           int       `json:"latency"`
	LastHealthCheck   time.Time `json:"last_health_check"`
	FailureCount      int       `json:"failure_count"`
	ConsecutiveErrors int       `json:"consecutive_errors"`
}

// ConnectionStats 连接统计
type ConnectionStats struct {
	TotalRequests   int64     `json:"total_requests"`
	SuccessRequests int64     `json:"success_requests"`
	FailedRequests  int64     `json:"failed_requests"`
	TotalTraffic    int64     `json:"total_traffic"`
	UploadTraffic   int64     `json:"upload_traffic"`
	DownloadTraffic int64     `json:"download_traffic"`
	LastUsed        time.Time `json:"last_used"`
}

// RoutingRule 路由规则
type RoutingRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	RuleType    string    `json:"rule_type"`
	Pattern     string    `json:"pattern"`
	Action      string    `json:"action"`
	TargetPool  string    `json:"target_pool"`
	Priority    int       `json:"priority"`
	Enabled     bool      `json:"enabled"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// NodeSelection 节点选择
type NodeSelection struct {
	SubscriptionID string `json:"subscription_id"`
	NodeIndex      int    `json:"node_index"`
	Weight         int    `json:"weight"`
	Priority       int    `json:"priority"`
}

// CreateConnectionPoolRequest 创建连接池请求
type CreateConnectionPoolRequest struct {
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Config         *ConnectionPoolConfig `json:"config"`
	NodeSelections []*NodeSelection `json:"node_selections"`
}

// UpdateConnectionPoolRequest 更新连接池请求
type UpdateConnectionPoolRequest struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Config      *ConnectionPoolConfig `json:"config"`
}

// CreateRoutingRuleRequest 创建路由规则请求
type CreateRoutingRuleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	RuleType    string `json:"rule_type"`
	Pattern     string `json:"pattern"`
	Action      string `json:"action"`
	TargetPool  string `json:"target_pool"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

// UpdateRoutingRuleRequest 更新路由规则请求
type UpdateRoutingRuleRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	RuleType    string `json:"rule_type"`
	Pattern     string `json:"pattern"`
	Action      string `json:"action"`
	TargetPool  string `json:"target_pool"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

// NewConnectionPool 创建新连接池
func NewConnectionPool(name, description string, config *ConnectionPoolConfig) *ConnectionPool {
	return &ConnectionPool{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:        name,
		Description: description,
		Config:      config,
		Connections: []*PoolConnection{},
		Status: &ConnectionPoolStatus{
			IsRunning:         false,
			ActiveConnections: 0,
			HealthyNodes:      0,
			TotalTraffic:      0,
			LastUpdate:        time.Now(),
		},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
}

// NewRoutingRule 创建新路由规则
func NewRoutingRule(name, description, ruleType, pattern, action, targetPool string, priority int, enabled bool) *RoutingRule {
	return &RoutingRule{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:        name,
		Description: description,
		RuleType:    ruleType,
		Pattern:     pattern,
		Action:      action,
		TargetPool:  targetPool,
		Priority:    priority,
		Enabled:     enabled,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
}

// StartAutoProxyRequest 启动自动代理请求
type StartAutoProxyRequest struct {
	SubscriptionID string `json:"subscription_id"`
	Mode           string `json:"mode"` // failover, load_balance, smart
	Config         *AutoProxyConfig `json:"config"`
}

// UpdateAutoProxyConfigRequest 更新自动代理配置请求
type UpdateAutoProxyConfigRequest struct {
	Config *AutoProxyConfig `json:"config"`
}

// AutoProxyConfig 自动代理配置
type AutoProxyConfig struct {
	TestInterval     int    `json:"test_interval"`     // 测试间隔（秒）
	HealthThreshold  int    `json:"health_threshold"`  // 健康阈值
	FailoverTimeout  int    `json:"failover_timeout"`  // 故障转移超时
	MaxRetries       int    `json:"max_retries"`       // 最大重试次数
	TestURL          string `json:"test_url"`          // 测试URL
	TestTimeout      int    `json:"test_timeout"`      // 测试超时
	SmartSwitching   bool   `json:"smart_switching"`   // 智能切换
	LoadBalanceMode  string `json:"load_balance_mode"` // 负载均衡模式
}

// AutoProxyStatus 自动代理状态
type AutoProxyStatus struct {
	IsRunning       bool                `json:"is_running"`
	Mode            string              `json:"mode"`
	CurrentNode     *NodeInfo           `json:"current_node"`
	AvailableNodes  []*NodeInfo         `json:"available_nodes"`
	FailedNodes     []*NodeInfo         `json:"failed_nodes"`
	TotalSwitches   int                 `json:"total_switches"`
	LastSwitchTime  time.Time           `json:"last_switch_time"`
	HealthStats     *AutoProxyHealthStats `json:"health_stats"`
	StartTime       time.Time           `json:"start_time"`
	Uptime          int64               `json:"uptime"`
	LastUpdate      time.Time           `json:"last_update"`
}

// AutoProxyHealthStats 自动代理健康统计
type AutoProxyHealthStats struct {
	TotalTests     int     `json:"total_tests"`
	SuccessTests   int     `json:"success_tests"`
	FailedTests    int     `json:"failed_tests"`
	SuccessRate    float64 `json:"success_rate"`
	AverageLatency int     `json:"average_latency"`
	LastTestTime   time.Time `json:"last_test_time"`
}

// NodePerformanceRecord 节点性能记录
type NodePerformanceRecord struct {
	NodeName       string    `json:"node_name"`
	Latency        int       `json:"latency"`
	DownloadSpeed  int64     `json:"download_speed"`
	UploadSpeed    int64     `json:"upload_speed"`
	SuccessRate    float64   `json:"success_rate"`
	Timestamp      time.Time `json:"timestamp"`
	TestCount      int       `json:"test_count"`
	FailCount      int       `json:"fail_count"`
}

// FailoverRecord 故障转移记录
type FailoverRecord struct {
	ID             string    `json:"id"`
	FromNode       string    `json:"from_node"`
	ToNode         string    `json:"to_node"`
	FailureReason  string    `json:"failure_reason"`
	SwitchTime     time.Time `json:"switch_time"`
	RecoveryTime   time.Time `json:"recovery_time"`
	DowntimeDuration int64   `json:"downtime_duration"`
	TriggerType    string    `json:"trigger_type"`
}

// IntelligentProxyRequest 智能代理请求
type IntelligentProxyRequest struct {
	SubscriptionID string                  `json:"subscription_id"`
	Config         *IntelligentProxyConfig `json:"config"`
}

// IntelligentProxyConfig 智能代理配置
type IntelligentProxyConfig struct {
	TestConcurrency      int    `json:"test_concurrency"`      // 测试并发数
	TestInterval         int    `json:"test_interval"`         // 定时重测间隔（分钟）
	HealthCheckInterval  int    `json:"health_check_interval"` // 健康检查间隔（秒）
	TestTimeout          int    `json:"test_timeout"`          // 测试超时（秒）
	TestURL              string `json:"test_url"`              // 测试URL
	SwitchThreshold      int    `json:"switch_threshold"`      // 切换阈值（延迟差异ms）
	MaxQueueSize         int    `json:"max_queue_size"`        // 队列最大大小
	HTTPPort             int    `json:"http_port"`             // HTTP代理端口
	SOCKSPort            int    `json:"socks_port"`            // SOCKS代理端口
	EnableAutoSwitch     bool   `json:"enable_auto_switch"`    // 启用自动切换
	EnableRetesting      bool   `json:"enable_retesting"`      // 启用定时重测
	EnableHealthCheck    bool   `json:"enable_health_check"`   // 启用健康检查
}

// IntelligentProxyStatus 智能代理状态
type IntelligentProxyStatus struct {
	IsRunning           bool                     `json:"is_running"`
	SubscriptionID      string                   `json:"subscription_id"`
	SubscriptionName    string                   `json:"subscription_name"`
	ActiveNode          *QueuedNode              `json:"active_node"`
	Queue               []*QueuedNode            `json:"queue"`
	QueueSize           int                      `json:"queue_size"`
	TestedNodes         int                      `json:"tested_nodes"`
	FailedNodes         int                      `json:"failed_nodes"`
	TotalSwitches       int                      `json:"total_switches"`
	LastSwitchTime      time.Time                `json:"last_switch_time"`
	LastTestTime        time.Time                `json:"last_test_time"`
	TestingProgress     *TestingProgress         `json:"testing_progress"`
	Config              *IntelligentProxyConfig  `json:"config"`
	HTTPPort            int                      `json:"http_port"`
	SOCKSPort           int                      `json:"socks_port"`
	StartTime           time.Time                `json:"start_time"`
	Uptime              int64                    `json:"uptime"`
	LastUpdate          time.Time                `json:"last_update"`
}

// QueuedNode 队列中的节点
type QueuedNode struct {
	SubscriptionID string    `json:"subscription_id"`
	NodeIndex      int       `json:"node_index"`
	NodeName       string    `json:"node_name"`
	Protocol       string    `json:"protocol"`
	Server         string    `json:"server"`
	Port           string    `json:"port"`
	Latency        int64     `json:"latency"`        // 延迟（毫秒）
	Speed          float64   `json:"speed"`          // 速度（Mbps）
	Score          float64   `json:"score"`          // 综合评分
	LastTestTime   time.Time `json:"last_test_time"`
	TestCount      int       `json:"test_count"`
	FailCount      int       `json:"fail_count"`
	SuccessRate    float64   `json:"success_rate"`
	IsActive       bool      `json:"is_active"`      // 是否为当前激活节点
	Status         string    `json:"status"`         // 状态：testing, queued, active, failed
}

// TestingProgress 测试进度
type TestingProgress struct {
	IsRunning     bool      `json:"is_running"`
	TotalNodes    int       `json:"total_nodes"`
	TestedNodes   int       `json:"tested_nodes"`
	SuccessNodes  int       `json:"success_nodes"`
	FailedNodes   int       `json:"failed_nodes"`
	CurrentNode   string    `json:"current_node"`
	Progress      int       `json:"progress"`      // 百分比
	StartTime     time.Time `json:"start_time"`
	EstimatedTime int       `json:"estimated_time"` // 预计剩余时间（秒）
}

// NodeSpeedTestResult 节点速度测试结果
type NodeSpeedTestResult struct {
	SubscriptionID string    `json:"subscription_id"`
	NodeIndex      int       `json:"node_index"`
	NodeName       string    `json:"node_name"`
	Success        bool      `json:"success"`
	Latency        int64     `json:"latency"`        // 延迟（毫秒）
	Speed          float64   `json:"speed"`          // 速度（Mbps）
	Error          string    `json:"error"`          // 错误信息
	TestTime       time.Time `json:"test_time"`
	TestDuration   int64     `json:"test_duration"`  // 测试耗时（毫秒）
}

// IntelligentProxyEvent 智能代理事件
type IntelligentProxyEvent struct {
	Type      string      `json:"type"`      // 事件类型：testing_start, testing_progress, testing_complete, node_switch, queue_update
	Data      interface{} `json:"data"`      // 事件数据
	Timestamp time.Time   `json:"timestamp"`
}

// SwitchNodeRequest 切换节点请求
type SwitchNodeRequest struct {
	NodeIndex int `json:"node_index"` // 要切换到的节点索引（队列中的位置）
}
