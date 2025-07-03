package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// SubscriptionDB 订阅数据库操作
type SubscriptionDB struct {
	db *Database
}

// NodeDB 节点数据库操作
type NodeDB struct {
	db *Database
}

// TestResultDB 测试结果数据库操作
type TestResultDB struct {
	db *Database
}

// ProxyStatusDB 代理状态数据库操作
type ProxyStatusDB struct {
	db *Database
}

// NewSubscriptionDB 创建订阅数据库操作实例
func NewSubscriptionDB(db *Database) *SubscriptionDB {
	return &SubscriptionDB{db: db}
}

// NewNodeDB 创建节点数据库操作实例
func NewNodeDB(db *Database) *NodeDB {
	return &NodeDB{db: db}
}

// NewTestResultDB 创建测试结果数据库操作实例
func NewTestResultDB(db *Database) *TestResultDB {
	return &TestResultDB{db: db}
}

// NewProxyStatusDB 创建代理状态数据库操作实例
func NewProxyStatusDB(db *Database) *ProxyStatusDB {
	return &ProxyStatusDB{db: db}
}

// SubscriptionDB 方法

// Create 创建订阅
func (s *SubscriptionDB) Create(subscription *models.Subscription) error {
	query := `
	INSERT INTO subscriptions (id, name, url, node_count, last_update, status, create_time)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.DB.Exec(query,
		subscription.ID,
		subscription.Name,
		subscription.URL,
		subscription.NodeCount,
		subscription.LastUpdate,
		subscription.Status,
		subscription.CreateTime,
	)
	return err
}

// GetAll 获取所有订阅
func (s *SubscriptionDB) GetAll() ([]*models.Subscription, error) {
	query := `SELECT id, name, url, node_count, last_update, status, create_time FROM subscriptions ORDER BY create_time DESC`
	
	rows, err := s.db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.Subscription
	for rows.Next() {
		sub := &models.Subscription{}
		err := rows.Scan(
			&sub.ID,
			&sub.Name,
			&sub.URL,
			&sub.NodeCount,
			&sub.LastUpdate,
			&sub.Status,
			&sub.CreateTime,
		)
		if err != nil {
			return nil, err
		}

		// 加载节点数据
		nodeDB := NewNodeDB(s.db)
		nodes, err := nodeDB.GetBySubscriptionID(sub.ID)
		if err != nil {
			return nil, fmt.Errorf("加载订阅 %s 的节点失败: %v", sub.ID, err)
		}
		sub.Nodes = nodes

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// GetByID 根据ID获取订阅
func (s *SubscriptionDB) GetByID(id string) (*models.Subscription, error) {
	query := `SELECT id, name, url, node_count, last_update, status, create_time FROM subscriptions WHERE id = ?`
	
	sub := &models.Subscription{}
	err := s.db.DB.QueryRow(query, id).Scan(
		&sub.ID,
		&sub.Name,
		&sub.URL,
		&sub.NodeCount,
		&sub.LastUpdate,
		&sub.Status,
		&sub.CreateTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("订阅不存在: %s", id)
		}
		return nil, err
	}

	// 加载节点数据
	nodeDB := NewNodeDB(s.db)
	nodes, err := nodeDB.GetBySubscriptionID(sub.ID)
	if err != nil {
		return nil, fmt.Errorf("加载订阅 %s 的节点失败: %v", sub.ID, err)
	}
	sub.Nodes = nodes

	return sub, nil
}

// Update 更新订阅
func (s *SubscriptionDB) Update(subscription *models.Subscription) error {
	query := `
	UPDATE subscriptions 
	SET name = ?, url = ?, node_count = ?, last_update = ?, status = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?`

	result, err := s.db.DB.Exec(query,
		subscription.Name,
		subscription.URL,
		subscription.NodeCount,
		subscription.LastUpdate,
		subscription.Status,
		subscription.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("订阅不存在: %s", subscription.ID)
	}

	return nil
}

// Delete 删除订阅
func (s *SubscriptionDB) Delete(id string) error {
	// 先删除相关节点（由于外键约束会自动删除）
	query := `DELETE FROM subscriptions WHERE id = ?`
	
	result, err := s.db.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("订阅不存在: %s", id)
	}

	return nil
}

// NodeDB 方法

// Create 创建节点
func (n *NodeDB) Create(node *models.NodeInfo, subscriptionID string) error {
	// 序列化参数
	parametersJSON, err := json.Marshal(node.Node.Parameters)
	if err != nil {
		return fmt.Errorf("序列化节点参数失败: %v", err)
	}

	query := `
	INSERT INTO nodes (subscription_id, node_index, name, protocol, server, port, method, password, parameters, status, is_running, http_port, socks_port, last_test, connect_time)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = n.db.DB.Exec(query,
		subscriptionID,
		node.Index,
		node.Name,
		node.Protocol,
		node.Server,
		node.Port,
		node.Method,
		node.Password,
		string(parametersJSON),
		node.Status,
		node.IsRunning,
		node.HTTPPort,
		node.SOCKSPort,
		node.LastTest.Format(time.RFC3339),
		node.ConnectTime.Format(time.RFC3339),
	)
	return err
}

// GetBySubscriptionID 根据订阅ID获取节点
func (n *NodeDB) GetBySubscriptionID(subscriptionID string) ([]*models.NodeInfo, error) {
	query := `
	SELECT id, node_index, name, protocol, server, port, method, password, parameters, status, is_running, http_port, socks_port, last_test, connect_time
	FROM nodes WHERE subscription_id = ? ORDER BY node_index`
	
	rows, err := n.db.DB.Query(query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*models.NodeInfo
	for rows.Next() {
		var dbID int
		var parametersJSON string
		var lastTestStr, connectTimeStr string
		
		nodeInfo := &models.NodeInfo{
			Node: &types.Node{},
		}

		err := rows.Scan(
			&dbID,
			&nodeInfo.Index,
			&nodeInfo.Name,
			&nodeInfo.Protocol,
			&nodeInfo.Server,
			&nodeInfo.Port,
			&nodeInfo.Method,
			&nodeInfo.Password,
			&parametersJSON,
			&nodeInfo.Status,
			&nodeInfo.IsRunning,
			&nodeInfo.HTTPPort,
			&nodeInfo.SOCKSPort,
			&lastTestStr,
			&connectTimeStr,
		)
		if err != nil {
			return nil, err
		}

		// 反序列化参数
		if err := json.Unmarshal([]byte(parametersJSON), &nodeInfo.Node.Parameters); err != nil {
			return nil, fmt.Errorf("反序列化节点参数失败: %v", err)
		}

		// 解析时间
		if lastTestStr != "" {
			if lastTest, err := time.Parse(time.RFC3339, lastTestStr); err == nil {
				nodeInfo.LastTest = lastTest
			}
		}
		if connectTimeStr != "" {
			if connectTime, err := time.Parse(time.RFC3339, connectTimeStr); err == nil {
				nodeInfo.ConnectTime = connectTime
			}
		}

		// 设置节点基本信息
		nodeInfo.Node.Name = nodeInfo.Name
		nodeInfo.Node.Protocol = nodeInfo.Protocol
		nodeInfo.Node.Server = nodeInfo.Server
		nodeInfo.Node.Port = nodeInfo.Port
		nodeInfo.Node.Method = nodeInfo.Method
		nodeInfo.Node.Password = nodeInfo.Password

		// 加载测试结果
		testResultDB := NewTestResultDB(n.db)
		if testResult, err := testResultDB.GetLatestByNodeID(dbID, "connection"); err == nil {
			nodeInfo.TestResult = testResult
		}
		if speedResult, err := testResultDB.GetLatestSpeedByNodeID(dbID); err == nil {
			nodeInfo.SpeedResult = speedResult
		}

		nodes = append(nodes, nodeInfo)
	}

	return nodes, nil
}

// Update 更新节点
func (n *NodeDB) Update(node *models.NodeInfo, subscriptionID string) error {
	// 序列化参数
	parametersJSON, err := json.Marshal(node.Node.Parameters)
	if err != nil {
		return fmt.Errorf("序列化节点参数失败: %v", err)
	}

	query := `
	UPDATE nodes 
	SET name = ?, protocol = ?, server = ?, port = ?, method = ?, password = ?, parameters = ?, status = ?, is_running = ?, http_port = ?, socks_port = ?, last_test = ?, connect_time = ?, updated_at = CURRENT_TIMESTAMP
	WHERE subscription_id = ? AND node_index = ?`

	_, err = n.db.DB.Exec(query,
		node.Name,
		node.Protocol,
		node.Server,
		node.Port,
		node.Method,
		node.Password,
		string(parametersJSON),
		node.Status,
		node.IsRunning,
		node.HTTPPort,
		node.SOCKSPort,
		node.LastTest.Format(time.RFC3339),
		node.ConnectTime.Format(time.RFC3339),
		subscriptionID,
		node.Index,
	)
	return err
}

// DeleteByIndexes 根据索引删除节点
func (n *NodeDB) DeleteByIndexes(subscriptionID string, nodeIndexes []int) error {
	if len(nodeIndexes) == 0 {
		return nil
	}

	// 构建IN子句
	placeholders := make([]string, len(nodeIndexes))
	args := []interface{}{subscriptionID}
	for i, index := range nodeIndexes {
		placeholders[i] = "?"
		args = append(args, index)
	}

	query := fmt.Sprintf(`DELETE FROM nodes WHERE subscription_id = ? AND node_index IN (%s)`, 
		strings.Join(placeholders, ","))

	_, err := n.db.DB.Exec(query, args...)
	return err
}

// ReindexNodes 重新索引节点（删除后调用）
func (n *NodeDB) ReindexNodes(subscriptionID string) error {
	// 获取所有节点按索引排序
	query := `SELECT id, node_index FROM nodes WHERE subscription_id = ? ORDER BY node_index`
	rows, err := n.db.DB.Query(query, subscriptionID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var updates []struct {
		id       int
		newIndex int
	}

	newIndex := 0
	for rows.Next() {
		var id, oldIndex int
		if err := rows.Scan(&id, &oldIndex); err != nil {
			return err
		}
		if oldIndex != newIndex {
			updates = append(updates, struct {
				id       int
				newIndex int
			}{id, newIndex})
		}
		newIndex++
	}

	// 执行更新
	for _, update := range updates {
		updateQuery := `UPDATE nodes SET node_index = ? WHERE id = ?`
		if _, err := n.db.DB.Exec(updateQuery, update.newIndex, update.id); err != nil {
			return err
		}
	}

	return nil
}

// TestResultDB 方法

// Create 创建测试结果
func (t *TestResultDB) Create(nodeID int, result *models.NodeTestResult) error {
	query := `
	INSERT INTO test_results (node_id, test_type, success, latency, error_message, test_time)
	VALUES (?, ?, ?, ?, ?, ?)`

	_, err := t.db.DB.Exec(query,
		nodeID,
		result.TestType,
		result.Success,
		result.Latency,
		result.Error,
		result.TestTime.Format(time.RFC3339),
	)
	return err
}

// CreateSpeedResult 创建速度测试结果
func (t *TestResultDB) CreateSpeedResult(nodeID int, result *models.SpeedTestResult) error {
	query := `
	INSERT INTO test_results (node_id, test_type, success, latency, download_speed, upload_speed, test_time, test_duration)
	VALUES (?, 'speed', TRUE, ?, ?, ?, ?, ?)`

	_, err := t.db.DB.Exec(query,
		nodeID,
		result.Latency,
		result.DownloadSpeed,
		result.UploadSpeed,
		result.TestTime.Format(time.RFC3339),
		result.TestDuration,
	)
	return err
}

// GetLatestByNodeID 获取节点最新的连接测试结果
func (t *TestResultDB) GetLatestByNodeID(nodeID int, testType string) (*models.NodeTestResult, error) {
	query := `
	SELECT test_type, success, latency, error_message, test_time
	FROM test_results 
	WHERE node_id = ? AND test_type = ?
	ORDER BY created_at DESC 
	LIMIT 1`

	var testTimeStr string
	result := &models.NodeTestResult{}
	
	err := t.db.DB.QueryRow(query, nodeID, testType).Scan(
		&result.TestType,
		&result.Success,
		&result.Latency,
		&result.Error,
		&testTimeStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到测试结果")
		}
		return nil, err
	}

	// 解析时间
	if testTime, err := time.Parse(time.RFC3339, testTimeStr); err == nil {
		result.TestTime = testTime
	}

	return result, nil
}

// GetLatestSpeedByNodeID 获取节点最新的速度测试结果
func (t *TestResultDB) GetLatestSpeedByNodeID(nodeID int) (*models.SpeedTestResult, error) {
	query := `
	SELECT latency, download_speed, upload_speed, test_time, test_duration
	FROM test_results 
	WHERE node_id = ? AND test_type = 'speed'
	ORDER BY created_at DESC 
	LIMIT 1`

	var testTimeStr string
	result := &models.SpeedTestResult{}
	
	err := t.db.DB.QueryRow(query, nodeID).Scan(
		&result.Latency,
		&result.DownloadSpeed,
		&result.UploadSpeed,
		&testTimeStr,
		&result.TestDuration,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到速度测试结果")
		}
		return nil, err
	}

	// 解析时间
	if testTime, err := time.Parse(time.RFC3339, testTimeStr); err == nil {
		result.TestTime = testTime
	}

	return result, nil
}

// GetNodeIDBySubscriptionAndIndex 根据订阅ID和节点索引获取节点数据库ID
func (n *NodeDB) GetNodeIDBySubscriptionAndIndex(subscriptionID string, nodeIndex int) (int, error) {
	query := `SELECT id FROM nodes WHERE subscription_id = ? AND node_index = ?`
	
	var nodeID int
	err := n.db.DB.QueryRow(query, subscriptionID, nodeIndex).Scan(&nodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("节点不存在")
		}
		return 0, err
	}
	
	return nodeID, nil
}