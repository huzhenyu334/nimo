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
	Success(c, nil)
}
