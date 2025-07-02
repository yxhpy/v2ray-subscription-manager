package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/services"
)

// ProxyHandler 代理处理器
type ProxyHandler struct {
	proxyService services.ProxyService
}

// NewProxyHandler 创建代理处理器
func NewProxyHandler(proxyService services.ProxyService) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
	}
}

// GetProxyStatus 获取代理状态
func (h *ProxyHandler) GetProxyStatus(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	status, err := h.proxyService.GetProxyStatus()
	if err != nil {
		response.SetError(err, "获取代理状态失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(status, "获取代理状态成功")
	h.writeJSONResponse(w, response)
}

// StopProxy 停止代理
func (h *ProxyHandler) StopProxy(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	err := h.proxyService.StopAllProxies()
	if err != nil {
		response.SetError(err, "停止代理失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(nil, "代理已停止")
	h.writeJSONResponse(w, response)
}

// writeJSONResponse 写入JSON响应
func (h *ProxyHandler) writeJSONResponse(w http.ResponseWriter, response *models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
