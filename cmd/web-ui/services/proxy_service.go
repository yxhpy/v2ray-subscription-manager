package services

import (
	"fmt"
	"sync"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/internal/core/proxy"
	"github.com/yxhpy/v2ray-subscription-manager/pkg/types"
)

// ProxyServiceImpl 代理服务实现
type ProxyServiceImpl struct {
	v2rayManager     *proxy.ProxyManager
	hysteria2Manager *proxy.ProxyManager // 暂时使用同一个管理器类型
	httpPort         int
	socksPort        int
	systemService    SystemService  // 添加系统服务依赖
	mutex            sync.RWMutex
}

// NewProxyService 创建代理服务
func NewProxyService() ProxyService {
	return &ProxyServiceImpl{
		v2rayManager:     proxy.NewProxyManager(),
		hysteria2Manager: proxy.NewProxyManager(),
		httpPort:         8888, // 默认HTTP端口
		socksPort:        1080, // 默认SOCKS端口
	}
}

// NewProxyServiceWithSystemService 创建带系统服务的代理服务
func NewProxyServiceWithSystemService(systemService SystemService) ProxyService {
	service := &ProxyServiceImpl{
		v2rayManager:     proxy.NewProxyManager(),
		hysteria2Manager: proxy.NewProxyManager(),
		systemService:    systemService,
		httpPort:         8888, // 默认HTTP端口
		socksPort:        1080, // 默认SOCKS端口
	}
	
	// 从系统设置加载端口配置
	service.loadPortsFromSettings()
	
	return service
}

// loadPortsFromSettings 从系统设置加载端口配置
func (p *ProxyServiceImpl) loadPortsFromSettings() {
	if p.systemService == nil {
		return
	}
	
	settings, err := p.systemService.GetSettings()
	if err != nil {
		fmt.Printf("⚠️  加载端口设置失败，使用默认值: %v\n", err)
		return
	}
	
	if settings.HTTPPort > 0 {
		p.httpPort = settings.HTTPPort
		fmt.Printf("📡 使用设置中的HTTP端口: %d\n", p.httpPort)
	}
	
	if settings.SOCKSPort > 0 {
		p.socksPort = settings.SOCKSPort
		fmt.Printf("📡 使用设置中的SOCKS端口: %d\n", p.socksPort)
	}
}

// GetProxyStatus 获取代理状态
func (p *ProxyServiceImpl) GetProxyStatus() (*models.ProxyStatus, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	status := &models.ProxyStatus{
		V2RayRunning:     p.v2rayManager.IsRunning(),
		Hysteria2Running: p.hysteria2Manager.IsRunning(),
		HTTPPort:         p.httpPort,
		SOCKSPort:        p.socksPort,
	}

	// 获取当前连接的节点
	if status.V2RayRunning {
		if currentNode := p.v2rayManager.GetCurrentNode(); currentNode != nil {
			status.CurrentNode = currentNode.Name
		}
	} else if status.Hysteria2Running {
		if currentNode := p.hysteria2Manager.GetCurrentNode(); currentNode != nil {
			status.CurrentNode = currentNode.Name
		}
	}

	return status, nil
}

// StopAllProxies 停止所有代理
func (p *ProxyServiceImpl) StopAllProxies() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var errors []string

	// 停止V2Ray代理
	if err := p.v2rayManager.StopProxy(); err != nil {
		errors = append(errors, fmt.Sprintf("停止V2Ray失败: %v", err))
	}

	// 停止Hysteria2代理
	if err := p.hysteria2Manager.StopProxy(); err != nil {
		errors = append(errors, fmt.Sprintf("停止Hysteria2失败: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("停止代理时发生错误: %v", errors)
	}

	return nil
}

// StartV2RayProxy 启动V2Ray代理
func (p *ProxyServiceImpl) StartV2RayProxy(node *types.Node) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 先停止其他代理
	if p.hysteria2Manager.IsRunning() {
		if err := p.hysteria2Manager.StopProxy(); err != nil {
			return fmt.Errorf("停止Hysteria2代理失败: %v", err)
		}
	}

	// 设置固定端口
	p.v2rayManager.SetFixedPorts(p.httpPort, p.socksPort)

	// 启动V2Ray代理
	if err := p.v2rayManager.StartProxy(node); err != nil {
		return fmt.Errorf("启动V2Ray代理失败: %v", err)
	}

	return nil
}

// StopV2RayProxy 停止V2Ray代理
func (p *ProxyServiceImpl) StopV2RayProxy() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.v2rayManager.StopProxy()
}

// StartHysteria2Proxy 启动Hysteria2代理
func (p *ProxyServiceImpl) StartHysteria2Proxy(node *types.Node) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 先停止其他代理
	if p.v2rayManager.IsRunning() {
		if err := p.v2rayManager.StopProxy(); err != nil {
			return fmt.Errorf("停止V2Ray代理失败: %v", err)
		}
	}

	// 设置固定端口
	p.hysteria2Manager.SetFixedPorts(p.httpPort, p.socksPort)

	// 启动Hysteria2代理
	if err := p.hysteria2Manager.StartProxy(node); err != nil {
		return fmt.Errorf("启动Hysteria2代理失败: %v", err)
	}

	return nil
}

// StopHysteria2Proxy 停止Hysteria2代理
func (p *ProxyServiceImpl) StopHysteria2Proxy() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.hysteria2Manager.StopProxy()
}

// SetFixedPorts 设置固定端口
func (p *ProxyServiceImpl) SetFixedPorts(httpPort, socksPort int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if httpPort > 0 {
		p.httpPort = httpPort
	}
	if socksPort > 0 {
		p.socksPort = socksPort
	}

	// 更新管理器的端口设置
	p.v2rayManager.SetFixedPorts(p.httpPort, p.socksPort)
	p.hysteria2Manager.SetFixedPorts(p.httpPort, p.socksPort)
}

// StopAllConnections 停止所有连接
func (p *ProxyServiceImpl) StopAllConnections() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	var errs []error
	
	// 停止V2Ray代理
	if p.v2rayManager.IsRunning() {
		if err := p.v2rayManager.StopProxy(); err != nil {
			errs = append(errs, fmt.Errorf("停止V2Ray代理失败: %v", err))
		} else {
			fmt.Printf("🛑 V2Ray代理已停止\n")
		}
	}
	
	// 停止Hysteria2代理
	if p.hysteria2Manager.IsRunning() {
		if err := p.hysteria2Manager.StopProxy(); err != nil {
			errs = append(errs, fmt.Errorf("停止Hysteria2代理失败: %v", err))
		} else {
			fmt.Printf("🛑 Hysteria2代理已停止\n")
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("停止连接时发生错误: %v", errs)
	}
	
	fmt.Printf("✅ 所有代理连接已停止\n")
	return nil
}
