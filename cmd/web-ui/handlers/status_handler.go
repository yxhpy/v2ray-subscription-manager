package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/services"
)

// StatusHandler 状态处理器
type StatusHandler struct {
	systemService services.SystemService
}

// NewStatusHandler 创建状态处理器
func NewStatusHandler(systemService services.SystemService) *StatusHandler {
	return &StatusHandler{
		systemService: systemService,
	}
}

// GetStatus 获取系统状态
func (h *StatusHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	status, err := h.systemService.GetSystemStatus()
	if err != nil {
		response.SetError(err, "获取系统状态失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(status, "获取系统状态成功")
	h.writeJSONResponse(w, response)
}

// RenderIndex 渲染主页
func (h *StatusHandler) RenderIndex(w http.ResponseWriter, r *http.Request) {
	// 尝试多个可能的模板路径
	templatePaths := []string{
		"cmd/web-ui/templates/index.html", // 从项目根目录运行时
		"templates/index.html",            // 从 cmd/web-ui 目录运行时
	}

	var templatePath string
	for _, path := range templatePaths {
		if info, err := http.Dir(".").Open(path); err == nil {
			info.Close()
			templatePath = path
			break
		}
	}

	if templatePath == "" {
		// 如果找不到模板文件，返回简单的 HTML 页面
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>V2Ray 订阅管理器</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .status { padding: 20px; margin: 20px 0; border-radius: 5px; }
        .success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .info { background: #d1ecf1; color: #0c5460; border: 1px solid #bee5eb; }
        .warning { background: #fff3cd; color: #856404; border: 1px solid #ffeaa7; }
        a { color: #007bff; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌟 V2Ray 订阅管理器 Web UI</h1>
        
        <div class="status success">
            <h3>✅ 服务状态</h3>
            <p>Web UI 服务器正在运行中...</p>
            <p>版本: v1.0.0</p>
        </div>
        
        <div class="status info">
            <h3>📡 API 接口</h3>
            <p>系统状态: <a href="/api/status" target="_blank">/api/status</a></p>
            <p>订阅管理: <a href="/api/subscriptions" target="_blank">/api/subscriptions</a></p>
            <p>节点管理: /api/nodes/*</p>
        </div>
        
        <div class="status warning">
            <h3>⚠️ 注意事项</h3>
            <p>模板文件未找到，显示简化版本界面</p>
            <p>请确保模板文件位于正确路径：templates/index.html</p>
            <p>批量测试功能已修复，支持 SSE 实时进度和取消操作</p>
        </div>
        
        <div class="status info">
            <h3>🔧 修复内容</h3>
            <ul>
                <li>增强了 SSE 连接稳定性</li>
                <li>添加了连接测试和心跳机制</li>
                <li>改进了错误处理和取消机制</li>
                <li>增加了超时保护 (20分钟总体，3分钟进度)</li>
                <li>修复了批量测试的取消逻辑</li>
            </ul>
        </div>
    </div>
    
    <script>
        // 简单的状态检查
        fetch('/api/status')
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    console.log('API状态正常:', data);
                }
            })
            .catch(error => {
                console.error('API连接失败:', error);
            });
    </script>
</body>
</html>`))
		return
	}

	// 使用找到的模板文件
	http.ServeFile(w, r, templatePath)
}

// writeJSONResponse 写入JSON响应
func (h *StatusHandler) writeJSONResponse(w http.ResponseWriter, response *models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSettings 获取系统设置
func (h *StatusHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	settings, err := h.systemService.GetSettings()
	if err != nil {
		response.SetError(err, "获取系统设置失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(settings, "获取系统设置成功")
	h.writeJSONResponse(w, response)
}

// SaveSettings 保存系统设置
func (h *StatusHandler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var settings models.Settings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	if err := h.systemService.SaveSettings(&settings); err != nil {
		response.SetError(err, "保存系统设置失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(settings, "系统设置保存成功")
	h.writeJSONResponse(w, response)
}
