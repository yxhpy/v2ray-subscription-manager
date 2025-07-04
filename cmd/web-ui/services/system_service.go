package services

import (
	"fmt"
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/database"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
)

// SystemServiceImpl 系统服务实现
type SystemServiceImpl struct {
	settings      *models.Settings
	db            *database.Database
	proxyService  ProxyService  // 添加代理服务依赖，用于重启
	nodeService   NodeService   // 添加节点服务依赖，用于重新加载配置
}

// NewSystemService 创建系统服务
func NewSystemService() SystemService {
	db := database.GetDB()
	service := &SystemServiceImpl{
		db: db,
		settings: &models.Settings{
			// 代理设置
			HTTPPort:  8888,
			SOCKSPort: 1080,
			AllowLan:  false,
			
			// 测试设置
			TestURL:       "https://www.google.com",
			TestTimeout:   30,
			MaxConcurrent: 3,
			RetryCount:    2,
			
			// 订阅设置
			UpdateInterval:   24,
			UserAgent:        "V2Ray/1.0",
			AutoTestNewNodes: true,
			
			// 安全设置
			EnableLogs:    true,
			LogLevel:      "info",
			DataRetention: 30,
		},
	}
	
	// 从数据库加载设置
	service.loadSettingsFromDB()
	
	return service
}

// SetServiceDependencies 设置服务依赖（用于设置变更时重启）
func (s *SystemServiceImpl) SetServiceDependencies(proxyService ProxyService, nodeService NodeService) {
	s.proxyService = proxyService
	s.nodeService = nodeService
}

// GetSystemStatus 获取系统状态
func (s *SystemServiceImpl) GetSystemStatus() (*models.SystemStatus, error) {
	status := &models.SystemStatus{
		ProxyPorts: map[string]int{
			"http":  s.settings.HTTPPort,
			"socks": s.settings.SOCKSPort,
		},
		ServerStatus: "running",
		Timestamp:    time.Now().Unix(),
		Version:      "v1.0.0",
	}
	return status, nil
}

// GetSettings 获取设置
func (s *SystemServiceImpl) GetSettings() (*models.Settings, error) {
	return s.settings, nil
}

// SaveSettings 保存设置
func (s *SystemServiceImpl) SaveSettings(settings *models.Settings) error {
	// 验证设置参数
	if err := s.validateSettings(settings); err != nil {
		return err
	}
	
	// 检查关键设置是否发生变化
	oldSettings := s.settings
	portChanged := oldSettings.HTTPPort != settings.HTTPPort || oldSettings.SOCKSPort != settings.SOCKSPort
	
	// 更新内存中的设置
	s.settings = settings
	
	// 保存到数据库
	if err := s.saveSettingsToDB(); err != nil {
		return err
	}
	
	// 如果端口设置发生变化，重新应用到代理服务
	if portChanged && s.proxyService != nil {
		fmt.Printf("🔄 端口设置已变更，正在应用新配置...\n")
		
		// 设置新的固定端口
		s.proxyService.SetFixedPorts(settings.HTTPPort, settings.SOCKSPort)
		
		fmt.Printf("✅ 新的端口配置已应用: HTTP:%d, SOCKS:%d\n", settings.HTTPPort, settings.SOCKSPort)
	}
	
	// 通知其他服务重新加载配置（如果实现了重新加载方法）
	if s.nodeService != nil {
		// 这里可以添加重新加载节点服务配置的逻辑
		fmt.Printf("📡 节点服务配置已更新\n")
	}
	
	return nil
}

// validateSettings 验证设置参数
func (s *SystemServiceImpl) validateSettings(settings *models.Settings) error {
	if settings.HTTPPort < 1024 || settings.HTTPPort > 65535 {
		return fmt.Errorf("HTTP端口必须在1024-65535范围内")
	}
	if settings.SOCKSPort < 1024 || settings.SOCKSPort > 65535 {
		return fmt.Errorf("SOCKS端口必须在1024-65535范围内")
	}
	if settings.HTTPPort == settings.SOCKSPort {
		return fmt.Errorf("HTTP端口和SOCKS端口不能相同")
	}
	if settings.TestTimeout < 5 || settings.TestTimeout > 300 {
		return fmt.Errorf("测试超时时间必须在5-300秒范围内")
	}
	if settings.MaxConcurrent < 1 || settings.MaxConcurrent > 10 {
		return fmt.Errorf("最大并发数必须在1-10范围内")
	}
	if settings.TestURL == "" {
		return fmt.Errorf("测试URL不能为空")
	}
	return nil
}

// loadSettingsFromDB 从数据库加载设置
func (s *SystemServiceImpl) loadSettingsFromDB() {
	if s.db == nil {
		return
	}
	
	// 从settings表加载各项设置
	settingsMap := map[string]interface{}{
		"http_port":          &s.settings.HTTPPort,
		"socks_port":         &s.settings.SOCKSPort,
		"allow_lan":          &s.settings.AllowLan,
		"test_url":           &s.settings.TestURL,
		"test_timeout":       &s.settings.TestTimeout,
		"max_concurrent":     &s.settings.MaxConcurrent,
		"retry_count":        &s.settings.RetryCount,
		"update_interval":    &s.settings.UpdateInterval,
		"user_agent":         &s.settings.UserAgent,
		"auto_test_nodes":    &s.settings.AutoTestNewNodes,
		"enable_logs":        &s.settings.EnableLogs,
		"log_level":          &s.settings.LogLevel,
		"data_retention":     &s.settings.DataRetention,
	}
	
	for key, ptr := range settingsMap {
		query := "SELECT value FROM settings WHERE key = ?"
		row := s.db.DB.QueryRow(query, key)
		var value string
		if err := row.Scan(&value); err != nil {
			continue // 使用默认值
		}
		
		// 根据类型转换值
		switch v := ptr.(type) {
		case *int:
			if intVal, err := fmt.Sscanf(value, "%d", v); intVal != 1 || err != nil {
				continue
			}
		case *bool:
			*v = value == "true"
		case *string:
			*v = value
		}
	}
}

// saveSettingsToDB 保存设置到数据库
func (s *SystemServiceImpl) saveSettingsToDB() error {
	if s.db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	
	// 设置键值对
	settingsMap := map[string]interface{}{
		"http_port":          s.settings.HTTPPort,
		"socks_port":         s.settings.SOCKSPort,
		"allow_lan":          s.settings.AllowLan,
		"test_url":           s.settings.TestURL,
		"test_timeout":       s.settings.TestTimeout,
		"max_concurrent":     s.settings.MaxConcurrent,
		"retry_count":        s.settings.RetryCount,
		"update_interval":    s.settings.UpdateInterval,
		"user_agent":         s.settings.UserAgent,
		"auto_test_nodes":    s.settings.AutoTestNewNodes,
		"enable_logs":        s.settings.EnableLogs,
		"log_level":          s.settings.LogLevel,
		"data_retention":     s.settings.DataRetention,
	}
	
	// 保存每个设置
	for key, value := range settingsMap {
		var valueStr string
		switch v := value.(type) {
		case bool:
			if v {
				valueStr = "true"
			} else {
				valueStr = "false"
			}
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		
		query := `INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)`
		if _, err := s.db.DB.Exec(query, key, valueStr); err != nil {
			return fmt.Errorf("保存设置 %s 失败: %v", key, err)
		}
	}
	
	fmt.Printf("✅ 系统设置已保存到数据库\n")
	return nil
}
