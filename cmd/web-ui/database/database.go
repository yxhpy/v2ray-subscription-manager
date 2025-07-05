package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
)

// Database 数据库管理器
type Database struct {
	DB *sql.DB
}

var dbInstance *Database

// GetDB 获取数据库实例（单例模式）
func GetDB() *Database {
	if dbInstance == nil {
		var err error
		dbInstance, err = NewDatabase("data/v2ray_manager.db")
		if err != nil {
			log.Fatalf("初始化数据库失败: %v", err)
		}
	}
	return dbInstance
}

// NewDatabase 创建新的数据库连接
func NewDatabase(dbPath string) (*Database, error) {
	// 确保数据目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %v", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	database := &Database{DB: db}

	// 初始化表结构
	if err := database.initTables(); err != nil {
		return nil, fmt.Errorf("初始化表结构失败: %v", err)
	}

	log.Printf("✅ SQLite数据库初始化成功: %s", dbPath)
	return database, nil
}

// initTables 初始化数据库表结构
func (d *Database) initTables() error {
	// 订阅表
	subscriptionsTable := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		node_count INTEGER DEFAULT 0,
		last_update TEXT DEFAULT '',
		status TEXT DEFAULT 'inactive',
		create_time TEXT NOT NULL,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	);`

	// 节点表
	nodesTable := `
	CREATE TABLE IF NOT EXISTS nodes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		subscription_id TEXT NOT NULL,
		node_index INTEGER NOT NULL,
		name TEXT NOT NULL,
		protocol TEXT NOT NULL,
		server TEXT NOT NULL,
		port TEXT NOT NULL,
		method TEXT DEFAULT '',
		password TEXT DEFAULT '',
		parameters TEXT DEFAULT '{}',
		status TEXT DEFAULT 'idle',
		is_running BOOLEAN DEFAULT FALSE,
		http_port INTEGER DEFAULT 0,
		socks_port INTEGER DEFAULT 0,
		last_test TEXT DEFAULT '',
		connect_time TEXT DEFAULT '',
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE,
		UNIQUE(subscription_id, node_index)
	);`

	// 测试结果表
	testResultsTable := `
	CREATE TABLE IF NOT EXISTS test_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		node_id INTEGER NOT NULL,
		test_type TEXT NOT NULL, -- 'connection' or 'speed'
		success BOOLEAN NOT NULL,
		latency TEXT DEFAULT '',
		download_speed TEXT DEFAULT '',
		upload_speed TEXT DEFAULT '',
		error_message TEXT DEFAULT '',
		test_time TEXT NOT NULL,
		test_duration TEXT DEFAULT '',
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE
	);`

	// 代理状态表
	proxyStatusTable := `
	CREATE TABLE IF NOT EXISTS proxy_status (
		id INTEGER PRIMARY KEY CHECK (id = 1), -- 确保只有一行
		v2ray_running BOOLEAN DEFAULT FALSE,
		hysteria2_running BOOLEAN DEFAULT FALSE,
		http_port INTEGER DEFAULT 8888,
		socks_port INTEGER DEFAULT 1080,
		current_node_id INTEGER DEFAULT NULL,
		total_connections INTEGER DEFAULT 0,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (current_node_id) REFERENCES nodes(id)
	);`

	// 系统设置表
	settingsTable := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		description TEXT DEFAULT '',
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	);`

	// 智能代理配置表
	intelligentProxyConfigTable := `
	CREATE TABLE IF NOT EXISTS intelligent_proxy_config (
		id INTEGER PRIMARY KEY CHECK (id = 1), -- 确保只有一行
		subscription_id TEXT NOT NULL,
		test_concurrency INTEGER DEFAULT 10,
		test_interval INTEGER DEFAULT 30,
		health_check_interval INTEGER DEFAULT 60,
		test_timeout INTEGER DEFAULT 30,
		test_url TEXT DEFAULT 'https://www.google.com',
		switch_threshold INTEGER DEFAULT 100,
		max_queue_size INTEGER DEFAULT 50,
		http_port INTEGER DEFAULT 7890,
		socks_port INTEGER DEFAULT 7891,
		enable_auto_switch BOOLEAN DEFAULT TRUE,
		enable_retesting BOOLEAN DEFAULT TRUE,
		enable_health_check BOOLEAN DEFAULT TRUE,
		is_running BOOLEAN DEFAULT FALSE,
		start_time TEXT DEFAULT '',
		last_update TEXT DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	);`

	// 智能代理队列表
	intelligentProxyQueueTable := `
	CREATE TABLE IF NOT EXISTS intelligent_proxy_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		subscription_id TEXT NOT NULL,
		node_index INTEGER NOT NULL,
		node_name TEXT NOT NULL,
		protocol TEXT NOT NULL,
		server TEXT NOT NULL,
		port TEXT NOT NULL,
		latency INTEGER DEFAULT 0,
		speed REAL DEFAULT 0.0,
		score REAL DEFAULT 0.0,
		last_test_time TEXT DEFAULT '',
		test_count INTEGER DEFAULT 0,
		fail_count INTEGER DEFAULT 0,
		success_rate REAL DEFAULT 0.0,
		is_active BOOLEAN DEFAULT FALSE,
		status TEXT DEFAULT 'queued',
		priority INTEGER DEFAULT 0,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE,
		UNIQUE(subscription_id, node_index)
	);`

	// 智能代理测试历史表
	intelligentProxyTestHistoryTable := `
	CREATE TABLE IF NOT EXISTS intelligent_proxy_test_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		subscription_id TEXT NOT NULL,
		node_index INTEGER NOT NULL,
		node_name TEXT NOT NULL,
		success BOOLEAN NOT NULL,
		latency INTEGER DEFAULT 0,
		speed REAL DEFAULT 0.0,
		error_message TEXT DEFAULT '',
		test_time TEXT NOT NULL,
		test_duration INTEGER DEFAULT 0,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
	);`

	// 智能代理切换记录表
	intelligentProxySwitchLogTable := `
	CREATE TABLE IF NOT EXISTS intelligent_proxy_switch_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_node_index INTEGER DEFAULT -1,
		from_node_name TEXT DEFAULT '',
		to_node_index INTEGER NOT NULL,
		to_node_name TEXT NOT NULL,
		switch_reason TEXT NOT NULL,
		switch_time TEXT NOT NULL,
		latency_before INTEGER DEFAULT 0,
		latency_after INTEGER DEFAULT 0,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_nodes_subscription_id ON nodes(subscription_id);",
		"CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);",
		"CREATE INDEX IF NOT EXISTS idx_test_results_node_id ON test_results(node_id);",
		"CREATE INDEX IF NOT EXISTS idx_test_results_test_type ON test_results(test_type);",
		"CREATE INDEX IF NOT EXISTS idx_intelligent_proxy_queue_subscription_id ON intelligent_proxy_queue(subscription_id);",
		"CREATE INDEX IF NOT EXISTS idx_intelligent_proxy_queue_status ON intelligent_proxy_queue(status);",
		"CREATE INDEX IF NOT EXISTS idx_intelligent_proxy_queue_score ON intelligent_proxy_queue(score DESC);",
		"CREATE INDEX IF NOT EXISTS idx_intelligent_proxy_test_history_subscription_id ON intelligent_proxy_test_history(subscription_id);",
		"CREATE INDEX IF NOT EXISTS idx_intelligent_proxy_test_history_test_time ON intelligent_proxy_test_history(test_time);",
		"CREATE INDEX IF NOT EXISTS idx_intelligent_proxy_switch_log_switch_time ON intelligent_proxy_switch_log(switch_time);",
	}

	// 执行表创建
	tables := []string{
		subscriptionsTable,
		nodesTable,
		testResultsTable,
		proxyStatusTable,
		settingsTable,
		intelligentProxyConfigTable,
		intelligentProxyQueueTable,
		intelligentProxyTestHistoryTable,
		intelligentProxySwitchLogTable,
	}

	for _, table := range tables {
		if _, err := d.DB.Exec(table); err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}
	}

	// 创建索引
	for _, index := range indexes {
		if _, err := d.DB.Exec(index); err != nil {
			return fmt.Errorf("创建索引失败: %v", err)
		}
	}

	// 初始化默认数据
	if err := d.initDefaultData(); err != nil {
		return fmt.Errorf("初始化默认数据失败: %v", err)
	}

	return nil
}

// initDefaultData 初始化默认数据
func (d *Database) initDefaultData() error {
	// 初始化代理状态记录
	proxyStatusSQL := `
	INSERT OR IGNORE INTO proxy_status (id, v2ray_running, hysteria2_running, http_port, socks_port)
	VALUES (1, FALSE, FALSE, 8888, 1080);`

	if _, err := d.DB.Exec(proxyStatusSQL); err != nil {
		return fmt.Errorf("初始化代理状态失败: %v", err)
	}

	// 初始化默认设置
	defaultSettings := map[string]string{
		"app_version":    "v2.1.0",
		"max_concurrent": "2",
		"test_timeout":   "30",
		"test_url":       "https://www.google.com",
	}

	for key, value := range defaultSettings {
		settingSQL := `
		INSERT OR IGNORE INTO settings (key, value, description)
		VALUES (?, ?, ?);`
		
		description := ""
		switch key {
		case "app_version":
			description = "应用版本号"
		case "max_concurrent":
			description = "最大并发测试数"
		case "test_timeout":
			description = "测试超时时间（秒）"
		case "test_url":
			description = "测试URL"
		}

		if _, err := d.DB.Exec(settingSQL, key, value, description); err != nil {
			return fmt.Errorf("初始化设置 %s 失败: %v", key, err)
		}
	}

	return nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// GetStats 获取数据库统计信息
func (d *Database) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计订阅数量
	var subscriptionCount int
	err := d.DB.QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&subscriptionCount)
	if err != nil {
		return nil, err
	}
	stats["subscription_count"] = subscriptionCount

	// 统计节点数量
	var nodeCount int
	err = d.DB.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&nodeCount)
	if err != nil {
		return nil, err
	}
	stats["node_count"] = nodeCount

	// 统计活跃节点数量
	var activeNodeCount int
	err = d.DB.QueryRow("SELECT COUNT(*) FROM nodes WHERE is_running = TRUE").Scan(&activeNodeCount)
	if err != nil {
		return nil, err
	}
	stats["active_node_count"] = activeNodeCount

	// 统计测试结果数量
	var testResultCount int
	err = d.DB.QueryRow("SELECT COUNT(*) FROM test_results").Scan(&testResultCount)
	if err != nil {
		return nil, err
	}
	stats["test_result_count"] = testResultCount

	return stats, nil
}

// CloseGlobalDB 关闭全局数据库连接
func CloseGlobalDB() error {
	if dbInstance != nil {
		fmt.Printf("💾 正在关闭SQLite数据库连接...\n")
		err := dbInstance.Close()
		dbInstance = nil
		if err == nil {
			fmt.Printf("✅ SQLite数据库连接已关闭\n")
		}
		return err
	}
	return nil
}

// SmartConnectionDB 智能连接数据库管理器
type SmartConnectionDB struct {
	DB *Database
}

// NewSmartConnectionDB 创建智能连接数据库管理器
func NewSmartConnectionDB(db *Database) *SmartConnectionDB {
	return &SmartConnectionDB{
		DB: db,
	}
}

// GetAllConnectionPools 获取所有连接池
func (s *SmartConnectionDB) GetAllConnectionPools() ([]*models.ConnectionPool, error) {
	// 这里应该从数据库查询连接池
	// 暂时返回空列表
	return []*models.ConnectionPool{}, nil
}

// GetAllRoutingRules 获取所有路由规则
func (s *SmartConnectionDB) GetAllRoutingRules() ([]*models.RoutingRule, error) {
	// 这里应该从数据库查询路由规则
	// 暂时返回空列表
	return []*models.RoutingRule{}, nil
}

// CreateConnectionPool 创建连接池
func (s *SmartConnectionDB) CreateConnectionPool(pool *models.ConnectionPool) error {
	// 这里应该实现数据库插入逻辑
	// 暂时返回nil
	return nil
}

// CreatePoolConnection 创建连接池连接
func (s *SmartConnectionDB) CreatePoolConnection(poolID string, connection *models.PoolConnection) error {
	// 这里应该实现数据库插入逻辑
	// 暂时返回nil
	return nil
}

// UpdateConnectionPool 更新连接池
func (s *SmartConnectionDB) UpdateConnectionPool(pool *models.ConnectionPool) error {
	// 这里应该实现数据库更新逻辑
	// 暂时返回nil
	return nil
}

// DeleteConnectionPool 删除连接池
func (s *SmartConnectionDB) DeleteConnectionPool(id string) error {
	// 这里应该实现数据库删除逻辑
	// 暂时返回nil
	return nil
}

// CreateRoutingRule 创建路由规则
func (s *SmartConnectionDB) CreateRoutingRule(rule *models.RoutingRule) error {
	// 这里应该实现数据库插入逻辑
	// 暂时返回nil
	return nil
}

// UpdateRoutingRule 更新路由规则
func (s *SmartConnectionDB) UpdateRoutingRule(rule *models.RoutingRule) error {
	// 这里应该实现数据库更新逻辑
	// 暂时返回nil
	return nil
}

// DeleteRoutingRule 删除路由规则
func (s *SmartConnectionDB) DeleteRoutingRule(id string) error {
	// 这里应该实现数据库删除逻辑
	// 暂时返回nil
	return nil
}

// UpdatePoolConnection 更新连接池连接
func (s *SmartConnectionDB) UpdatePoolConnection(connection *models.PoolConnection) error {
	// 这里应该实现数据库更新逻辑
	// 暂时返回nil
	return nil
}