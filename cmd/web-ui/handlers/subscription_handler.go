package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/models"
	"github.com/yxhpy/v2ray-subscription-manager/cmd/web-ui/services"
)

// SubscriptionHandler 订阅处理器
type SubscriptionHandler struct {
	subscriptionService services.SubscriptionService
}

// NewSubscriptionHandler 创建订阅处理器
func NewSubscriptionHandler(subscriptionService services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// GetSubscriptions 获取订阅列表
func (h *SubscriptionHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	subscriptions := h.subscriptionService.GetAllSubscriptions()
	response.SetSuccess(subscriptions, "获取订阅列表成功")

	h.writeJSONResponse(w, response)
}

// AddSubscription 添加订阅
func (h *SubscriptionHandler) AddSubscription(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var req models.AddSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	if req.URL == "" {
		response.SetError(nil, "订阅URL不能为空")
		response.Error = "订阅URL不能为空"
		h.writeJSONResponse(w, response)
		return
	}

	subscription, err := h.subscriptionService.AddSubscription(req.URL, req.Name)
	if err != nil {
		response.SetError(err, "添加订阅失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(subscription, "订阅添加成功")
	h.writeJSONResponse(w, response)
}

// ParseSubscription 解析订阅
func (h *SubscriptionHandler) ParseSubscription(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var req models.ParseSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	subscription, err := h.subscriptionService.ParseSubscription(req.ID)
	if err != nil {
		response.SetError(err, "解析订阅失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(subscription, "订阅解析成功")
	h.writeJSONResponse(w, response)
}

// DeleteSubscription 删除订阅
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var req models.DeleteSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	err := h.subscriptionService.DeleteSubscription(req.ID)
	if err != nil {
		response.SetError(err, "删除订阅失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(nil, "订阅删除成功")
	h.writeJSONResponse(w, response)
}

// TestSubscription 测试订阅
func (h *SubscriptionHandler) TestSubscription(w http.ResponseWriter, r *http.Request) {
	response := models.NewAPIResponse()

	var req models.TestSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SetError(err, "请求参数错误")
		h.writeJSONResponse(w, response)
		return
	}

	results, err := h.subscriptionService.TestSubscription(req.ID)
	if err != nil {
		response.SetError(err, "测试订阅失败")
		h.writeJSONResponse(w, response)
		return
	}

	response.SetSuccess(results, "订阅测试完成")
	h.writeJSONResponse(w, response)
}

// GetSubscriptionNodes 获取订阅的节点列表
func (h *SubscriptionHandler) GetSubscriptionNodes(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("DEBUG: GetSubscriptionNodes called with path: %s\n", r.URL.Path)

	response := models.NewAPIResponse()

	// 从URL路径中提取订阅ID
	subscriptionID := r.URL.Path[len("/api/subscriptions/"):]
	if len(subscriptionID) > 6 && subscriptionID[len(subscriptionID)-6:] == "/nodes" {
		subscriptionID = subscriptionID[:len(subscriptionID)-6]
	}

	fmt.Printf("DEBUG: Extracted subscription ID: %s\n", subscriptionID)

	subscription, err := h.subscriptionService.GetSubscriptionByID(subscriptionID)
	if err != nil {
		fmt.Printf("DEBUG: Error getting subscription: %v\n", err)
		response.SetError(err, "获取订阅失败")
		h.writeJSONResponse(w, response)
		return
	}

	fmt.Printf("DEBUG: Found subscription with %d nodes\n", len(subscription.Nodes))
	response.SetSuccess(subscription, "获取节点列表成功")
	h.writeJSONResponse(w, response)
}

// writeJSONResponse 写入JSON响应
func (h *SubscriptionHandler) writeJSONResponse(w http.ResponseWriter, response *models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
