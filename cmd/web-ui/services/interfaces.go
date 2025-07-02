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
	// 测试订阅
	TestSubscription(id string) ([]*models.NodeTestResult, error)
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
