package types

import "time"

// AutoProxyConfig 自动代理配置
type AutoProxyConfig struct {
	SubscriptionURL  string        `json:"subscription_url"`
	HTTPPort         int           `json:"http_port"`          // 固定HTTP代理端口
	SOCKSPort        int           `json:"socks_port"`         // 固定SOCKS代理端口
	UpdateInterval   time.Duration `json:"update_interval"`    // 更新间隔
	TestConcurrency  int           `json:"test_concurrency"`   // 测试并发数
	TestTimeout      time.Duration `json:"test_timeout"`       // 测试超时
	TestURL          string        `json:"test_url"`           // 测试URL
	MaxNodes         int           `json:"max_nodes"`          // 最大测试节点数
	MinPassingNodes  int           `json:"min_passing_nodes"`  // 最少通过节点数
	StateFile        string        `json:"state_file"`         // 状态文件路径
	ValidNodesFile   string        `json:"valid_nodes_file"`   // 有效节点中间文件
	EnableAutoSwitch bool          `json:"enable_auto_switch"` // 是否启用自动切换
}

// ValidNode 有效节点信息
type ValidNode struct {
	Node         *Node     `json:"node"`
	TestTime     time.Time `json:"test_time"`
	Latency      int64     `json:"latency_ms"`
	Speed        float64   `json:"speed_mbps"`
	SuccessCount int       `json:"success_count"`
	FailCount    int       `json:"fail_count"`
	Score        float64   `json:"score"` // 综合评分
}

// AutoProxyState 自动代理状态
type AutoProxyState struct {
	Running         bool            `json:"running"`
	StartTime       time.Time       `json:"start_time"`
	LastUpdate      time.Time       `json:"last_update"`
	CurrentNode     *Node           `json:"current_node"`
	ValidNodes      []ValidNode     `json:"valid_nodes"`
	TotalTests      int             `json:"total_tests"`
	SuccessfulTests int             `json:"successful_tests"`
	LastError       string          `json:"last_error,omitempty"`
	Config          AutoProxyConfig `json:"config"`
}

// TestTask 测试任务
type TestTask struct {
	Node      *Node
	Retry     int
	Timestamp time.Time
}

// TestBatch 测试批次
type TestBatch struct {
	ID        string      `json:"id"`
	Tasks     []TestTask  `json:"tasks"`
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
	Results   []ValidNode `json:"results"`
}
