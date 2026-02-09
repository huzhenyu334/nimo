package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
)

// RoutingHandler 路由规则处理器
type RoutingHandler struct {
	svc *service.RoutingService
}

// NewRoutingHandler 创建路由规则处理器
func NewRoutingHandler(svc *service.RoutingService) *RoutingHandler {
	return &RoutingHandler{svc: svc}
}

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	Name         string      `json:"name" binding:"required"`
	EntityType   string      `json:"entity_type" binding:"required"`
	Event        string      `json:"event" binding:"required"`
	Conditions   entity.JSONB `json:"conditions" binding:"required"`
	Channel      string      `json:"channel" binding:"required"`
	Priority     int         `json:"priority"`
	ActionConfig entity.JSONB `json:"action_config"`
	Enabled      *bool       `json:"enabled"`
	Description  string      `json:"description"`
}

// UpdateRuleRequest 更新规则请求
type UpdateRuleRequest struct {
	Name         *string      `json:"name"`
	EntityType   *string      `json:"entity_type"`
	Event        *string      `json:"event"`
	Conditions   entity.JSONB `json:"conditions"`
	Channel      *string      `json:"channel"`
	Priority     *int         `json:"priority"`
	ActionConfig entity.JSONB `json:"action_config"`
	Enabled      *bool        `json:"enabled"`
	Description  *string      `json:"description"`
}

// TestRouteRequest 测试路由请求
type TestRouteRequest struct {
	EntityType string                 `json:"entity_type" binding:"required"`
	Event      string                 `json:"event" binding:"required"`
	Context    map[string]interface{} `json:"context" binding:"required"`
}

// ListRules 查询路由规则列表
// GET /api/v1/routing-rules
func (h *RoutingHandler) ListRules(c *gin.Context) {
	page, pageSize := GetPagination(c)
	entityType := c.Query("entity_type")
	event := c.Query("event")

	rules, total, err := h.svc.ListRules(c.Request.Context(), entityType, event, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	Success(c, ListResponse{
		Items: rules,
		Pagination: &Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      int(total),
			TotalPages: totalPages,
		},
	})
}

// CreateRule 创建路由规则
// POST /api/v1/routing-rules
func (h *RoutingHandler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证 channel
	if req.Channel != entity.RoutingChannelFeishu && req.Channel != entity.RoutingChannelAgent && req.Channel != entity.RoutingChannelAuto {
		BadRequest(c, "channel 必须为 feishu、agent 或 auto")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	rule := &entity.RoutingRule{
		Name:         req.Name,
		EntityType:   req.EntityType,
		Event:        req.Event,
		Conditions:   req.Conditions,
		Channel:      req.Channel,
		Priority:     req.Priority,
		ActionConfig: req.ActionConfig,
		Enabled:      enabled,
		Description:  req.Description,
		CreatedBy:    GetUserID(c),
	}

	if err := h.svc.CreateRule(c.Request.Context(), rule); err != nil {
		InternalError(c, err.Error())
		return
	}

	Created(c, rule)
}

// GetRule 获取路由规则详情
// GET /api/v1/routing-rules/:id
func (h *RoutingHandler) GetRule(c *gin.Context) {
	id := c.Param("id")

	rule, err := h.svc.GetRule(c.Request.Context(), id)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	Success(c, rule)
}

// UpdateRule 更新路由规则
// PUT /api/v1/routing-rules/:id
func (h *RoutingHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.EntityType != nil {
		updates["entity_type"] = *req.EntityType
	}
	if req.Event != nil {
		updates["event"] = *req.Event
	}
	if req.Conditions != nil {
		updates["conditions"] = req.Conditions
	}
	if req.Channel != nil {
		if *req.Channel != entity.RoutingChannelFeishu && *req.Channel != entity.RoutingChannelAgent && *req.Channel != entity.RoutingChannelAuto {
			BadRequest(c, "channel 必须为 feishu、agent 或 auto")
			return
		}
		updates["channel"] = *req.Channel
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.ActionConfig != nil {
		updates["action_config"] = req.ActionConfig
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if len(updates) == 0 {
		BadRequest(c, "没有需要更新的字段")
		return
	}

	if err := h.svc.UpdateRule(c.Request.Context(), id, updates); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "规则已更新"})
}

// DeleteRule 删除路由规则
// DELETE /api/v1/routing-rules/:id
func (h *RoutingHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.DeleteRule(c.Request.Context(), id); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "规则已删除"})
}

// TestRoute 测试路由决策
// POST /api/v1/routing-rules/test
func (h *RoutingHandler) TestRoute(c *gin.Context) {
	var req TestRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	decision, err := h.svc.EvaluateRoute(c.Request.Context(), req.EntityType, req.Event, req.Context)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, decision)
}

// ListLogs 查询路由日志
// GET /api/v1/routing-logs
func (h *RoutingHandler) ListLogs(c *gin.Context) {
	page, pageSize := GetPagination(c)
	entityType := c.Query("entity_type")
	event := c.Query("event")

	logs, total, err := h.svc.ListLogs(c.Request.Context(), entityType, event, page, pageSize)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	Success(c, ListResponse{
		Items: logs,
		Pagination: &Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      int(total),
			TotalPages: totalPages,
		},
	})
}
