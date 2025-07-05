package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/database"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/handlers"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/services"
)

// WebUIServer Web UI服务器
type WebUIServer struct {
	// 服务层
	subscriptionService     services.SubscriptionService
	nodeService            services.NodeService
	proxyService           services.ProxyService
	systemService          services.SystemService
	templateService        services.TemplateService
	intelligentProxyService services.IntelligentProxyService

	// 处理器层
	subscriptionHandler      *handlers.SubscriptionHandler
	nodeHandler             *handlers.NodeHandler
	proxyHandler            *handlers.ProxyHandler
	statusHandler           *handlers.StatusHandler
	intelligentProxyHandler *handlers.IntelligentProxyHandler
	intelligentProxyPageHandler *handlers.IntelligentProxyPageHandler

	// 服务器配置
	port       string
	httpServer *http.Server
}

// NewWebUIServer 创建Web UI服务器
func NewWebUIServer(port string) *WebUIServer {
	server := &WebUIServer{
		port: port,
	}

	// 初始化HTTP服务器
	server.httpServer = &http.Server{
		Addr:         port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 初始化服务层
	server.initServices()

	// 初始化处理器层
	server.initHandlers()

	return server
}

// initServices 初始化服务层
func (s *WebUIServer) initServices() {
	// 先创建系统服务
	s.systemService = services.NewSystemService()
	s.templateService = services.NewTemplateService("cmd/web-ui/templates")
	
	// 创建订阅服务（使用系统设置）
	s.subscriptionService = services.NewSubscriptionServiceWithSystemService(s.systemService)
	
	// 创建代理服务（使用系统设置）
	s.proxyService = services.NewProxyServiceWithSystemService(s.systemService)
	
	// 创建节点服务（传入系统服务以使用设置）
	s.nodeService = services.NewNodeServiceWithSystemService(s.subscriptionService, s.proxyService, s.systemService)
	
	// 创建智能代理服务
	s.intelligentProxyService = services.NewIntelligentProxyService(database.GetDB(), s.subscriptionService, s.proxyService)
	
	// 设置系统服务的服务依赖（用于设置变更时重启）
	if systemServiceImpl, ok := s.systemService.(*services.SystemServiceImpl); ok {
		systemServiceImpl.SetServiceDependencies(s.proxyService, s.nodeService)
	}
}

// initHandlers 初始化处理器层
func (s *WebUIServer) initHandlers() {
	s.subscriptionHandler = handlers.NewSubscriptionHandler(s.subscriptionService)
	s.nodeHandler = handlers.NewNodeHandler(s.nodeService)
	s.proxyHandler = handlers.NewProxyHandler(s.proxyService, s.nodeService)
	s.statusHandler = handlers.NewStatusHandler(s.systemService)
	s.intelligentProxyHandler = handlers.NewIntelligentProxyHandler(s.intelligentProxyService)
	s.intelligentProxyPageHandler = handlers.NewIntelligentProxyPageHandler(s.subscriptionService)
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
	http.HandleFunc("/api/settings", s.handleSettings)

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
	http.HandleFunc("/api/nodes/delete", s.nodeHandler.DeleteNodes)
	http.HandleFunc("/api/nodes/connect", s.nodeHandler.ConnectNode)
	http.HandleFunc("/api/nodes/test", s.nodeHandler.TestNode)
	http.HandleFunc("/api/nodes/speedtest", s.nodeHandler.SpeedTestNode)
	http.HandleFunc("/api/nodes/check-port-conflict", s.nodeHandler.CheckPortConflict)

	// 代理管理API
	http.HandleFunc("/api/proxy/status", s.proxyHandler.GetProxyStatus)
	http.HandleFunc("/api/proxy/stop", s.proxyHandler.StopProxy)
	http.HandleFunc("/api/proxy/connections", s.proxyHandler.GetActiveConnections)
	http.HandleFunc("/api/proxy/stop-all", s.proxyHandler.StopAllConnections)

	// 智能代理API - 注册智能代理路由
	s.intelligentProxyHandler.RegisterRoutes(http.DefaultServeMux)
	
	// 智能代理页面
	s.intelligentProxyPageHandler.RegisterPageRoutes(http.DefaultServeMux)

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

// handleSettings 处理设置请求
func (s *WebUIServer) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.statusHandler.GetSettings(w, r)
	case "POST":
		s.statusHandler.SaveSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Start 启动服务器
func (s *WebUIServer) Start() error {
	s.setupRoutes()

	fmt.Printf("🚀 Web UI服务器启动成功！\n")
	fmt.Printf("📱 访问地址: http://localhost%s\n", s.port)
	fmt.Printf("📝 管理界面: http://localhost%s\n", s.port)
	fmt.Printf("🔗 API文档: http://localhost%s/api/status\n", s.port)

	return s.httpServer.ListenAndServe()
}

// Shutdown 优雅关闭服务器
func (s *WebUIServer) Shutdown(ctx context.Context) error {
	fmt.Printf("🛑 正在优雅关闭Web UI服务器...\n")
	
	// 关闭HTTP服务器
	if err := s.httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("❌ HTTP服务器关闭失败: %v\n", err)
		return err
	}
	
	// 清理服务层资源
	s.cleanup()
	
	fmt.Printf("✅ Web UI服务器已优雅关闭\n")
	return nil
}

// cleanup 清理所有资源
func (s *WebUIServer) cleanup() {
	fmt.Printf("🧹 正在清理系统资源...\n")
	
	// 停止智能代理服务
	if s.intelligentProxyService != nil {
		fmt.Printf("🤖 停止智能代理服务...\n")
		if err := s.intelligentProxyService.StopIntelligentProxy(); err != nil {
			fmt.Printf("⚠️ 停止智能代理服务失败: %v\n", err)
		}
	}
	
	// 停止所有活跃的代理连接
	if s.proxyService != nil {
		fmt.Printf("🔌 停止所有代理连接...\n")
		s.proxyService.StopAllConnections()
	}
	
	// 停止所有节点连接
	if s.nodeService != nil {
		fmt.Printf("🔌 停止所有节点连接...\n")
		s.nodeService.StopAllNodeConnections()
	}
	
	// 关闭服务层资源
	if s.subscriptionService != nil {
		s.subscriptionService.Close()
	}
	
	// 关闭全局数据库连接
	database.CloseGlobalDB()
	
	// 清理临时文件
	fmt.Printf("🗑️ 清理临时文件...\n")
	s.cleanupTempFiles()
	
	fmt.Printf("✅ 资源清理完成\n")
}

// cleanupTempFiles 清理临时文件
func (s *WebUIServer) cleanupTempFiles() {
	fmt.Printf("🗑️ 开始清理临时文件...\n")
	
	// 清理V2Ray临时配置文件
	tempFiles := []string{
		"temp_v2ray_config_*.json",
		"test_v2ray_config_*.json", 
		"temp_*.yaml",
		"test_config_*.yaml",
		"temp_*.json",
		"test_*.json",
		"*.tmp",
		"*.temp",
		"auto_proxy_*.json",
		"proxy_state.json",
		"valid_nodes.json",
		"mvp_best_node.json",
	}
	
	totalCleaned := 0
	for _, pattern := range tempFiles {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("⚠️ 匹配模式失败 %s: %v\n", pattern, err)
			continue
		}
		
		for _, file := range matches {
			// 跳过测试脚本和正常文件
			if strings.Contains(file, "test_all_features.sh") || 
			   strings.Contains(file, "test_batch_cancel") ||
			   strings.Contains(file, "test_frontend.html") {
				continue
			}
			
			if err := os.Remove(file); err == nil {
				fmt.Printf("🗑️ 已删除临时文件: %s\n", file)
				totalCleaned++
			} else {
				fmt.Printf("⚠️ 删除文件失败 %s: %v\n", file, err)
			}
		}
	}
	
	fmt.Printf("✅ 临时文件清理完成，共清理 %d 个文件\n", totalCleaned)
}

// setupSignalHandler 设置信号处理器
func (s *WebUIServer) setupSignalHandler() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	
	go func() {
		sig := <-signalChan
		fmt.Printf("\n🛑 接收到退出信号: %v\n", sig)
		fmt.Printf("🔄 开始优雅关闭流程...\n")
		
		// 立即执行资源清理
		fmt.Printf("🧹 执行资源清理...\n")
		s.cleanup()
		
		// 创建超时上下文，给服务器5秒时间优雅关闭
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		// 优雅关闭HTTP服务器
		if s.httpServer != nil {
			fmt.Printf("🛑 关闭HTTP服务器...\n")
			if err := s.httpServer.Shutdown(ctx); err != nil {
				fmt.Printf("❌ HTTP服务器关闭失败: %v\n", err)
			} else {
				fmt.Printf("✅ HTTP服务器已关闭\n")
			}
		}
		
		fmt.Printf("👋 程序已安全退出\n")
		os.Exit(0)
	}()
}

// validateAndCleanupOnStartup 启动时验证和清理状态
func (s *WebUIServer) validateAndCleanupOnStartup() {
	fmt.Printf("🔍 正在验证系统状态...\n")
	
	// 检查实际进程状态
	v2rayRunning := s.checkV2RayProcess()
	hysteria2Running := s.checkHysteria2Process()
	
	fmt.Printf("📊 实际进程状态: V2Ray=%v, Hysteria2=%v\n", v2rayRunning, hysteria2Running)
	
	// 重置代理服务状态
	if s.proxyService != nil {
		fmt.Printf("🔧 重置代理服务状态...\n")
		s.proxyService.StopAllConnections()
	}
	
	// 清理数据库中的运行状态
	s.cleanupDatabaseStatus()
	
	fmt.Printf("✅ 系统状态验证完成\n")
}

// checkV2RayProcess 检查V2Ray进程是否真实运行
func (s *WebUIServer) checkV2RayProcess() bool {
	// TODO: 实现进程检查逻辑
	// 这里可以通过检查进程名、PID文件或端口占用来判断
	return false
}

// checkHysteria2Process 检查Hysteria2进程是否真实运行
func (s *WebUIServer) checkHysteria2Process() bool {
	// TODO: 实现进程检查逻辑
	return false
}

// cleanupDatabaseStatus 清理数据库中的运行状态
func (s *WebUIServer) cleanupDatabaseStatus() {
	fmt.Printf("🗄️ 正在清理数据库状态...\n")
	
	// 获取数据库实例
	db := database.GetDB()
	if db == nil {
		fmt.Printf("⚠️ 无法获取数据库实例\n")
		return
	}
	
	// 重置所有节点的运行状态
	resetNodesSQL := `
	UPDATE nodes 
	SET is_running = FALSE, 
	    status = 'idle',
	    http_port = 0,
	    socks_port = 0,
	    connect_time = '',
	    updated_at = CURRENT_TIMESTAMP
	WHERE is_running = TRUE;`
	
	if _, err := db.DB.Exec(resetNodesSQL); err != nil {
		fmt.Printf("❌ 重置节点状态失败: %v\n", err)
	} else {
		fmt.Printf("✅ 节点运行状态已重置\n")
	}
	
	// 重置代理状态
	resetProxySQL := `
	UPDATE proxy_status 
	SET v2ray_running = FALSE,
	    hysteria2_running = FALSE,
	    current_node_id = NULL,
	    total_connections = 0,
	    updated_at = CURRENT_TIMESTAMP
	WHERE id = 1;`
	
	if _, err := db.DB.Exec(resetProxySQL); err != nil {
		fmt.Printf("❌ 重置代理状态失败: %v\n", err)
	} else {
		fmt.Printf("✅ 代理运行状态已重置\n")
	}
	
	// 重置智能代理状态
	resetIntelligentProxySQL := `
	UPDATE intelligent_proxy_config 
	SET is_running = FALSE,
	    last_update = CURRENT_TIMESTAMP
	WHERE id = 1;`
	
	if _, err := db.DB.Exec(resetIntelligentProxySQL); err != nil {
		fmt.Printf("❌ 重置智能代理状态失败: %v\n", err)
	} else {
		fmt.Printf("✅ 智能代理运行状态已重置\n")
	}
	
	// 清理智能代理队列中的激活状态
	resetQueueSQL := `
	UPDATE intelligent_proxy_queue 
	SET is_active = FALSE,
	    status = 'queued',
	    updated_at = CURRENT_TIMESTAMP
	WHERE is_active = TRUE;`
	
	if _, err := db.DB.Exec(resetQueueSQL); err != nil {
		fmt.Printf("❌ 重置智能代理队列状态失败: %v\n", err)
	} else {
		fmt.Printf("✅ 智能代理队列状态已重置\n")
	}
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

	// 创建服务器实例
	server := NewWebUIServer(":8888")
	
	// 设置信号处理器（必须在启动服务器之前）
	server.setupSignalHandler()
	
	// 启动时验证和清理状态
	server.validateAndCleanupOnStartup()
	
	// 设置defer确保资源清理
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("🚨 程序异常退出: %v\n", r)
			fmt.Printf("🧹 执行紧急资源清理...\n")
			server.cleanup()
		}
	}()

	// 启动服务器
	fmt.Printf("🔄 启动服务器中...\n")
	if err := server.Start(); err != nil {
		if err == http.ErrServerClosed {
			fmt.Printf("✅ 服务器已正常关闭\n")
		} else {
			fmt.Printf("❌ 启动Web UI服务器失败: %v\n", err)
			// 确保在异常退出时也执行清理
			server.cleanup()
			os.Exit(1)
		}
	}
}
