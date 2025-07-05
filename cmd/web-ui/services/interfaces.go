package services

import (
	"context"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// ProgressCallback 进度回调函数类型
type ProgressCallback func(progress *models.BatchTestProgress)

// SubscriptionService 订阅服务接口
type SubscriptionService interface {
	// 添加订阅
	AddSubscription(url, name string) (*models.Subscription, error)
	// 获取所有订阅
	GetAllSubscriptions() []*models.Subscription
	// 根据ID获取订阅
	GetSubscriptionByID(id string) (*models.Subscription, error)
	// 解析订阅
	ParseSubscription(id string) (*models.Subscription, error)
	// 删除订阅
	DeleteSubscription(id string) error
	// 更新订阅
	UpdateSubscription(subscription *models.Subscription) error
	// 测试订阅
	TestSubscription(id string) ([]*models.NodeTestResult, error)
	// 关闭服务，释放资源
	Close() error
}

// NodeService 节点服务接口
type NodeService interface {
	// 连接节点
	ConnectNode(subscriptionID string, nodeIndex int, operation string) (*models.ConnectNodeResponse, error)
	// 测试节点
	TestNode(subscriptionID string, nodeIndex int) (*models.NodeTestResult, error)
	// 速度测试节点
	SpeedTestNode(subscriptionID string, nodeIndex int) (*models.SpeedTestResult, error)
	// 批量测试节点
	BatchTestNodes(subscriptionID string, nodeIndexes []int) ([]*models.NodeTestResult, error)
	// 带进度回调的批量测试节点
	BatchTestNodesWithProgress(subscriptionID string, nodeIndexes []int, callback ProgressCallback) ([]*models.NodeTestResult, error)
	// 带进度回调和上下文的批量测试节点
	BatchTestNodesWithProgressAndContext(ctx context.Context, subscriptionID string, nodeIndexes []int, callback ProgressCallback) ([]*models.NodeTestResult, error)
	// 删除节点
	DeleteNodes(subscriptionID string, nodeIndexes []int) error
	// 停止所有节点连接
	StopAllNodeConnections() error
	// 检查端口冲突
	CheckPortConflict(port int) (*models.PortConflictInfo, error)
}

// ProxyService 代理服务接口
type ProxyService interface {
	// 获取代理状态
	GetProxyStatus() (*models.ProxyStatus, error)
	// 停止所有代理
	StopAllProxies() error
	// 启动V2Ray代理
	StartV2RayProxy(node *types.Node) error
	// 停止V2Ray代理
	StopV2RayProxy() error
	// 启动Hysteria2代理
	StartHysteria2Proxy(node *types.Node) error
	// 停止Hysteria2代理
	StopHysteria2Proxy() error
	// 设置固定端口
	SetFixedPorts(httpPort, socksPort int)
	// 停止所有连接
	StopAllConnections() error
}

// SystemService 系统服务接口
type SystemService interface {
	// 获取系统状态
	GetSystemStatus() (*models.SystemStatus, error)
	// 获取设置
	GetSettings() (*models.Settings, error)
	// 保存设置
	SaveSettings(settings *models.Settings) error
}

// TemplateService 模板服务接口
type TemplateService interface {
	// 渲染主页
	RenderIndex() (string, error)
}

// AutoProxyService 自动代理服务接口
type AutoProxyService interface {
	// 启动自动代理
	StartAutoProxy(req *models.StartAutoProxyRequest) error
	// 停止自动代理
	StopAutoProxy() error
	// 获取自动代理状态
	GetAutoProxyStatus() (*models.AutoProxyStatus, error)
	// 更新自动代理配置
	UpdateAutoProxyConfig(req *models.UpdateAutoProxyConfigRequest) error
	// 获取自动代理配置
	GetAutoProxyConfig() (*models.AutoProxyConfig, error)
	// 获取最佳节点
	GetBestNode() (*models.NodeInfo, error)
	// 获取节点性能历史
	GetNodePerformanceHistory(subscriptionID string, nodeIndex int) ([]*models.NodePerformanceRecord, error)
	// 切换到最佳节点
	SwitchToBestNode() error
	// 获取故障转移记录
	GetFailoverRecords() ([]*models.FailoverRecord, error)
}

// SmartConnectionService 智能连接服务接口
type SmartConnectionService interface {
	// 启动智能连接管理器
	Start() error
	// 停止智能连接管理器
	Stop() error
	// 获取智能连接状态
	GetStatus() (*models.SmartConnectionStatus, error)
	// 创建连接池
	CreateConnectionPool(req *models.CreateConnectionPoolRequest) (*models.ConnectionPool, error)
	// 获取所有连接池
	GetAllConnectionPools() ([]*models.ConnectionPool, error)
	// 根据ID获取连接池
	GetConnectionPoolByID(id string) (*models.ConnectionPool, error)
	// 更新连接池
	UpdateConnectionPool(req *models.UpdateConnectionPoolRequest) error
	// 删除连接池
	DeleteConnectionPool(id string) error
	// 启动连接池
	StartConnectionPool(id string) error
	// 停止连接池
	StopConnectionPool(id string) error
	// 创建路由规则
	CreateRoutingRule(req *models.CreateRoutingRuleRequest) (*models.RoutingRule, error)
	// 获取所有路由规则
	GetAllRoutingRules() ([]*models.RoutingRule, error)
	// 更新路由规则
	UpdateRoutingRule(req *models.UpdateRoutingRuleRequest) error
	// 删除路由规则
	DeleteRoutingRule(id string) error
}

// IntelligentProxyService 智能代理服务接口
type IntelligentProxyService interface {
	// 启动智能代理
	StartIntelligentProxy(req *models.IntelligentProxyRequest) error
	// 停止智能代理
	StopIntelligentProxy() error
	// 获取智能代理状态
	GetIntelligentProxyStatus() (*models.IntelligentProxyStatus, error)
	// 手动切换节点
	SwitchToNode(req *models.SwitchNodeRequest) error
	// 获取队列状态
	GetQueue() ([]*models.QueuedNode, error)
	// 强制重新测试所有节点
	ForceRetestAllNodes() error
	// 暂停/恢复自动切换
	ToggleAutoSwitch(enabled bool) error
	// 更新配置
	UpdateConfig(config *models.IntelligentProxyConfig) error
	// 获取测试进度
	GetTestingProgress() (*models.TestingProgress, error)
	// 订阅事件流
	SubscribeEvents() (<-chan *models.IntelligentProxyEvent, error)
	// 获取上次保存的配置
	GetLastSavedConfig() (*models.IntelligentProxyConfig, string, error)
}
