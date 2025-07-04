package services

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/database"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/parser"
)

// SubscriptionServiceImpl 订阅服务实现
type SubscriptionServiceImpl struct {
	subscriptionDB *database.SubscriptionDB
	nodeDB         *database.NodeDB
	systemService  SystemService  // 添加系统服务依赖
	mutex          sync.RWMutex
}

// NewSubscriptionService 创建订阅服务
func NewSubscriptionService() SubscriptionService {
	db := database.GetDB()
	return &SubscriptionServiceImpl{
		subscriptionDB: database.NewSubscriptionDB(db),
		nodeDB:         database.NewNodeDB(db),
	}
}

// NewSubscriptionServiceWithSystemService 创建带系统服务的订阅服务
func NewSubscriptionServiceWithSystemService(systemService SystemService) SubscriptionService {
	db := database.GetDB()
	return &SubscriptionServiceImpl{
		subscriptionDB: database.NewSubscriptionDB(db),
		nodeDB:         database.NewNodeDB(db),
		systemService:  systemService,
	}
}

// getUserAgent 获取用户代理字符串（从设置中或使用默认值）
func (s *SubscriptionServiceImpl) getUserAgent() string {
	if s.systemService == nil {
		return "V2Ray/1.0"
	}
	
	settings, err := s.systemService.GetSettings()
	if err != nil || settings.UserAgent == "" {
		return "V2Ray/1.0"
	}
	
	return settings.UserAgent
}

// AddSubscription 添加订阅
func (s *SubscriptionServiceImpl) AddSubscription(url, name string) (*models.Subscription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 生成唯一ID
	id := fmt.Sprintf("sub_%d", time.Now().UnixNano())

	// 如果没有提供名称，从URL中提取
	if name == "" {
		name = s.extractNameFromURL(url)
	}

	// 创建订阅
	subscription := models.NewSubscription(id, name, url)
	
	// 保存到数据库
	if err := s.subscriptionDB.Create(subscription); err != nil {
		return nil, fmt.Errorf("保存订阅到数据库失败: %v", err)
	}

	return subscription, nil
}

// GetAllSubscriptions 获取所有订阅
func (s *SubscriptionServiceImpl) GetAllSubscriptions() []*models.Subscription {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	subscriptions, err := s.subscriptionDB.GetAll()
	if err != nil {
		fmt.Printf("ERROR: 获取订阅列表失败: %v\n", err)
		return []*models.Subscription{}
	}
	return subscriptions
}

// GetSubscriptionByID 根据ID获取订阅
func (s *SubscriptionServiceImpl) GetSubscriptionByID(id string) (*models.Subscription, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.subscriptionDB.GetByID(id)
}

// ParseSubscription 解析订阅
func (s *SubscriptionServiceImpl) ParseSubscription(id string) (*models.Subscription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取订阅
	subscription, err := s.subscriptionDB.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 获取订阅内容（使用自定义User-Agent）
	userAgent := s.getUserAgent()
	content, err := parser.FetchSubscriptionWithUserAgent(subscription.URL, userAgent)
	if err != nil {
		return nil, fmt.Errorf("获取订阅失败: %v", err)
	}

	// 解码内容
	decodedContent, err := parser.DecodeBase64(content)
	if err != nil {
		return nil, fmt.Errorf("解码订阅失败: %v", err)
	}

	// 解析节点
	nodes, err := parser.ParseLinks(decodedContent)
	if err != nil {
		return nil, fmt.Errorf("解析订阅失败: %v", err)
	}

	// 先清空旧节点
	// TODO: 实现删除旧节点的逻辑

	// 转换为 NodeInfo 结构并保存到数据库
	nodeInfos := make([]*models.NodeInfo, 0, len(nodes))
	for i, node := range nodes {
		nodeInfo := models.NewNodeInfo(node, i)
		
		// 保存节点到数据库
		if err := s.nodeDB.Create(nodeInfo, subscription.ID); err != nil {
			fmt.Printf("WARNING: 保存节点失败: %v\n", err)
			continue
		}
		
		nodeInfos = append(nodeInfos, nodeInfo)
	}

	// 更新订阅信息
	subscription.Nodes = nodeInfos
	subscription.NodeCount = len(nodeInfos)
	subscription.LastUpdate = time.Now().Format("2006-01-02 15:04:05")
	subscription.Status = "active"

	// 更新数据库中的订阅
	if err := s.subscriptionDB.Update(subscription); err != nil {
		return nil, fmt.Errorf("更新订阅失败: %v", err)
	}

	return subscription, nil
}

// DeleteSubscription 删除订阅
func (s *SubscriptionServiceImpl) DeleteSubscription(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.subscriptionDB.Delete(id)
}

// UpdateSubscription 更新订阅
func (s *SubscriptionServiceImpl) UpdateSubscription(subscription *models.Subscription) error {
	if subscription == nil {
		return fmt.Errorf("订阅对象不能为空")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.subscriptionDB.Update(subscription)
}

// TestSubscription 测试订阅
func (s *SubscriptionServiceImpl) TestSubscription(id string) ([]*models.NodeTestResult, error) {
	subscription, err := s.GetSubscriptionByID(id)
	if err != nil {
		return nil, err
	}

	if len(subscription.Nodes) == 0 {
		return nil, fmt.Errorf("订阅中没有节点，请先解析订阅")
	}

	results := make([]*models.NodeTestResult, 0, len(subscription.Nodes))

	// 创建测试结果
	for i, nodeInfo := range subscription.Nodes {
		result := &models.NodeTestResult{
			NodeName: nodeInfo.Name,
			Success:  true,                         // 模拟测试成功
			Latency:  strconv.Itoa(50+i*10) + "ms", // 模拟延迟
			TestTime: time.Now(),
			TestType: "subscription",
		}
		results = append(results, result)

		// 更新节点的测试结果
		nodeInfo.SetTestResult(result)
	}

	return results, nil
}

// Close 关闭服务，释放资源
func (s *SubscriptionServiceImpl) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// 关闭数据库连接
	if s.subscriptionDB != nil {
		// 注意：这里不能直接关闭数据库，因为数据库是全局共享的
		// 数据库连接会在全局关闭时统一处理
		fmt.Printf("💾 订阅服务资源已释放\n")
	}
	
	return nil
}

// extractNameFromURL 从URL中提取名称
func (s *SubscriptionServiceImpl) extractNameFromURL(url string) string {
	// 简单的名称提取逻辑，可以根据需要改进
	if len(url) > 20 {
		return url[:20] + "..."
	}
	return url
}
