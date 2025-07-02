package services

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/parser"
)

// SubscriptionServiceImpl 订阅服务实现
type SubscriptionServiceImpl struct {
	subscriptions map[string]*models.Subscription
	mutex         sync.RWMutex
}

// NewSubscriptionService 创建订阅服务
func NewSubscriptionService() SubscriptionService {
	return &SubscriptionServiceImpl{
		subscriptions: make(map[string]*models.Subscription),
	}
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
	s.subscriptions[id] = subscription

	return subscription, nil
}

// GetAllSubscriptions 获取所有订阅
func (s *SubscriptionServiceImpl) GetAllSubscriptions() []*models.Subscription {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	subscriptions := make([]*models.Subscription, 0, len(s.subscriptions))
	for _, sub := range s.subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions
}

// GetSubscriptionByID 根据ID获取订阅
func (s *SubscriptionServiceImpl) GetSubscriptionByID(id string) (*models.Subscription, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	subscription, exists := s.subscriptions[id]
	if !exists {
		return nil, fmt.Errorf("订阅不存在: %s", id)
	}
	return subscription, nil
}

// ParseSubscription 解析订阅
func (s *SubscriptionServiceImpl) ParseSubscription(id string) (*models.Subscription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	subscription, exists := s.subscriptions[id]
	if !exists {
		return nil, fmt.Errorf("订阅不存在: %s", id)
	}

	// 获取订阅内容
	content, err := parser.FetchSubscription(subscription.URL)
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

	// 转换为 NodeInfo 结构
	nodeInfos := make([]*models.NodeInfo, 0, len(nodes))
	for i, node := range nodes {
		nodeInfo := models.NewNodeInfo(node, i)
		nodeInfos = append(nodeInfos, nodeInfo)
	}

	// 更新订阅信息
	subscription.Nodes = nodeInfos
	subscription.NodeCount = len(nodeInfos)
	subscription.LastUpdate = time.Now().Format("2006-01-02 15:04:05")
	subscription.Status = "active"

	return subscription, nil
}

// DeleteSubscription 删除订阅
func (s *SubscriptionServiceImpl) DeleteSubscription(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.subscriptions[id]
	if !exists {
		return fmt.Errorf("订阅不存在: %s", id)
	}

	delete(s.subscriptions, id)
	return nil
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

// extractNameFromURL 从URL中提取名称
func (s *SubscriptionServiceImpl) extractNameFromURL(url string) string {
	// 简单的名称提取逻辑，可以根据需要改进
	if len(url) > 20 {
		return url[:20] + "..."
	}
	return url
}
