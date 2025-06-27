package types

// ProxyStatus 代理状态
type ProxyStatus struct {
	Running   bool   `json:"running"`
	NodeName  string `json:"node_name"`
	Protocol  string `json:"protocol"`
	HTTPPort  int    `json:"http_port"`
	SOCKSPort int    `json:"socks_port"`
	PID       int    `json:"pid,omitempty"`
}

// Hysteria2Status Hysteria2代理状态
type Hysteria2Status struct {
	Running   bool   `json:"running"`
	NodeName  string `json:"node_name"`
	Protocol  string `json:"protocol"`
	HTTPPort  int    `json:"http_port"`
	SOCKSPort int    `json:"socks_port"`
	PID       int    `json:"pid,omitempty"`
}

// SpeedTestResult 测速结果
type SpeedTestResult struct {
	NodeName     string  `json:"node_name"`
	Protocol     string  `json:"protocol"`
	Server       string  `json:"server"`
	Port         string  `json:"port"`
	Success      bool    `json:"success"`
	ResponseTime float64 `json:"response_time_ms"`
	Error        string  `json:"error,omitempty"`
}

// SpeedTestConfig 测速配置
type SpeedTestConfig struct {
	Concurrency int    `json:"concurrency"`
	Timeout     int    `json:"timeout"`
	TestURL     string `json:"test_url"`
	OutputFile  string `json:"output_file"`
	MaxNodes    int    `json:"max_nodes"`
}
