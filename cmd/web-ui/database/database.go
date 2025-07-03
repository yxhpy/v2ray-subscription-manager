package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
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

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_nodes_subscription_id ON nodes(subscription_id);",
		"CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);",
		"CREATE INDEX IF NOT EXISTS idx_test_results_node_id ON test_results(node_id);",
		"CREATE INDEX IF NOT EXISTS idx_test_results_test_type ON test_results(test_type);",
	}

	// 执行表创建
	tables := []string{
		subscriptionsTable,
		nodesTable,
		testResultsTable,
		proxyStatusTable,
		settingsTable,
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