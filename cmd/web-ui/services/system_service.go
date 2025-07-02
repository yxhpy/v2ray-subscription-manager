package services

import (
	"time"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
)

// SystemServiceImpl 系统服务实现
type SystemServiceImpl struct {
	settings *models.Settings
}

// NewSystemService 创建系统服务
func NewSystemService() SystemService {
	return &SystemServiceImpl{
		settings: &models.Settings{
			HTTPPort:  8888,
			SOCKSPort: 1080,
			TestURL:   "https://www.google.com",
		},
	}
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
	if settings.HTTPPort > 0 {
		s.settings.HTTPPort = settings.HTTPPort
	}
	if settings.SOCKSPort > 0 {
		s.settings.SOCKSPort = settings.SOCKSPort
	}
	if settings.TestURL != "" {
		s.settings.TestURL = settings.TestURL
	}
	return nil
}
