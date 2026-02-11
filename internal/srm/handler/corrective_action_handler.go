package handler

import (
	"github.com/bitfantasy/nimo/internal/srm/service"
	"github.com/gin-gonic/gin"
)

// CorrectiveActionHandler 纠正措施处理器
type CorrectiveActionHandler struct {
	svc *service.CorrectiveActionService
}

func NewCorrectiveActionHandler(svc *service.CorrectiveActionService) *CorrectiveActionHandler {
	return &CorrectiveActionHandler{svc: svc}
}

func (h *CorrectiveActionHandler) ListCorrectiveActions(c *gin.Context) {
	Success(c, nil)
}

func (h *CorrectiveActionHandler) CreateCorrectiveAction(c *gin.Context) {
	Created(c, nil)
}

func (h *CorrectiveActionHandler) GetCorrectiveAction(c *gin.Context) {
	Success(c, nil)
}

func (h *CorrectiveActionHandler) UpdateCorrectiveAction(c *gin.Context) {
	Success(c, nil)
}

func (h *CorrectiveActionHandler) SupplierRespond(c *gin.Context) {
	Success(c, nil)
}

func (h *CorrectiveActionHandler) Verify(c *gin.Context) {
	Success(c, nil)
}

func (h *CorrectiveActionHandler) Close(c *gin.Context) {
	Success(c, nil)
}
