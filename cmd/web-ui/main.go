package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/handlers"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/services"
)

// WebUIServer Web UI服务器
type WebUIServer struct {
	// 服务层
	subscriptionService services.SubscriptionService
	nodeService         services.NodeService
	proxyService        services.ProxyService
	systemService       services.SystemService
	templateService     services.TemplateService

	// 处理器层
	subscriptionHandler *handlers.SubscriptionHandler
	nodeHandler         *handlers.NodeHandler
	proxyHandler        *handlers.ProxyHandler
	statusHandler       *handlers.StatusHandler

	// 服务器配置
	port string
}

// NewWebUIServer 创建Web UI服务器
func NewWebUIServer(port string) *WebUIServer {
	server := &WebUIServer{
		port: port,
	}

	// 初始化服务层
	server.initServices()

	// 初始化处理器层
	server.initHandlers()

	return server
}

// initServices 初始化服务层
func (s *WebUIServer) initServices() {
	// 创建服务实例
	s.subscriptionService = services.NewSubscriptionService()
	s.proxyService = services.NewProxyService()
	s.nodeService = services.NewNodeService(s.subscriptionService, s.proxyService)
	s.systemService = services.NewSystemService()
	s.templateService = services.NewTemplateService("cmd/web-ui/templates")
}

// initHandlers 初始化处理器层
func (s *WebUIServer) initHandlers() {
	s.subscriptionHandler = handlers.NewSubscriptionHandler(s.subscriptionService)
	s.nodeHandler = handlers.NewNodeHandler(s.nodeService)
	s.proxyHandler = handlers.NewProxyHandler(s.proxyService)
	s.statusHandler = handlers.NewStatusHandler(s.systemService)
}

// setupRoutes 设置路由
func (s *WebUIServer) setupRoutes() {
	// 静态文件服务 - 尝试多个可能的路径
	staticPaths := []string{
		"web/static/",       // 从项目根目录运行时
		"../../web/static/", // 从 cmd/web-ui 目录运行时
	}

	var staticPath string
	for _, path := range staticPaths {
		if _, err := http.Dir(path).Open("/"); err == nil {
			staticPath = path
			break
		}
	}

	if staticPath != "" {
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
		fmt.Printf("DEBUG: 静态文件路径设置为: %s\n", staticPath)
	} else {
		fmt.Printf("WARNING: 未找到静态文件目录\n")
	}

	// API路由 - 最具体的路径先注册
	http.HandleFunc("/api/status", s.statusHandler.GetStatus)

	// 订阅管理API - 更具体的路径先注册
	http.HandleFunc("/api/subscriptions/parse", s.subscriptionHandler.ParseSubscription)
	http.HandleFunc("/api/subscriptions/delete", s.subscriptionHandler.DeleteSubscription)
	http.HandleFunc("/api/subscriptions/test", s.subscriptionHandler.TestSubscription)
	http.HandleFunc("/api/subscriptions/", s.handleSubscriptionDetails)
	http.HandleFunc("/api/subscriptions", s.handleSubscriptions)

	// 节点管理API - SSE路由必须在普通路由之前
	http.HandleFunc("/api/nodes/batch-test-sse", s.nodeHandler.BatchTestNodesSSE)
	http.HandleFunc("/api/nodes/batch-test", s.nodeHandler.BatchTestNodes)
	http.HandleFunc("/api/nodes/cancel-batch-test", s.nodeHandler.CancelBatchTest)
	http.HandleFunc("/api/nodes/connect", s.nodeHandler.ConnectNode)
	http.HandleFunc("/api/nodes/test", s.nodeHandler.TestNode)
	http.HandleFunc("/api/nodes/speedtest", s.nodeHandler.SpeedTestNode)

	// 代理管理API
	http.HandleFunc("/api/proxy/status", s.proxyHandler.GetProxyStatus)
	http.HandleFunc("/api/proxy/stop", s.proxyHandler.StopProxy)

	// 主页 - 最后注册catch-all路由
	http.HandleFunc("/", s.statusHandler.RenderIndex)
}

// handleSubscriptions 处理订阅列表请求
func (s *WebUIServer) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.subscriptionHandler.GetSubscriptions(w, r)
	case "POST":
		s.subscriptionHandler.AddSubscription(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSubscriptionDetails 处理订阅详情请求
func (s *WebUIServer) handleSubscriptionDetails(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: handleSubscriptionDetails called with path: %s\n", r.URL.Path)

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查是否是获取节点列表的请求
	if len(r.URL.Path) > 6 && r.URL.Path[len(r.URL.Path)-6:] == "/nodes" {
		fmt.Printf("DEBUG: Routing to GetSubscriptionNodes\n")
		s.subscriptionHandler.GetSubscriptionNodes(w, r)
	} else {
		fmt.Printf("DEBUG: Path does not end with /nodes, returning 404\n")
		http.NotFound(w, r)
	}
}

// Start 启动服务器
func (s *WebUIServer) Start() error {
	s.setupRoutes()

	fmt.Printf("🚀 Web UI服务器启动成功！\n")
	fmt.Printf("📱 访问地址: http://localhost%s\n", s.port)
	fmt.Printf("📝 管理界面: http://localhost%s\n", s.port)
	fmt.Printf("🔗 API文档: http://localhost%s/api/status\n", s.port)

	return http.ListenAndServe(s.port, nil)
}

func main() {
	// 获取当前工作目录
	workDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("获取工作目录失败: %v", err)
	}

	fmt.Printf("📁 工作目录: %s\n", workDir)
	fmt.Printf("🌟 V2Ray 订阅管理器 Web UI\n")
	fmt.Printf("🔧 版本: v1.0.0\n")

	// 创建并启动服务器
	server := NewWebUIServer(":8888")
	if err := server.Start(); err != nil {
		log.Fatalf("启动Web UI服务器失败: %v", err)
	}
}
