package handler

import (
	"github.com/bitfantasy/nimo/internal/srm/service"
	"github.com/gin-gonic/gin"
)

// PRItemHandler 采购申请行项处理器
type PRItemHandler struct {
	svc *service.PRItemService
}

func NewPRItemHandler(svc *service.PRItemService) *PRItemHandler {
	return &PRItemHandler{svc: svc}
}

func (h *PRItemHandler) UpdatePRItemStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "缺少行项ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "缺少目标状态")
		return
	}

	userID := GetUserID(c)
	item, err := h.svc.UpdatePRItemStatus(c.Request.Context(), id, req.Status, userID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, item)
}
