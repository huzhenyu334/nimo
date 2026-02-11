package handler

import (
	"github.com/bitfantasy/nimo/internal/srm/service"
	"github.com/gin-gonic/gin"
)

// SettlementHandler 结算处理器
type SettlementHandler struct {
	svc *service.SettlementService
}

func NewSettlementHandler(svc *service.SettlementService) *SettlementHandler {
	return &SettlementHandler{svc: svc}
}

func (h *SettlementHandler) ListSettlements(c *gin.Context) {
	Success(c, nil)
}

func (h *SettlementHandler) ExportSettlements(c *gin.Context) {
	Success(c, nil)
}

func (h *SettlementHandler) CreateSettlement(c *gin.Context) {
	Created(c, nil)
}

func (h *SettlementHandler) GenerateSettlement(c *gin.Context) {
	Created(c, nil)
}

func (h *SettlementHandler) GetSettlement(c *gin.Context) {
	Success(c, nil)
}

func (h *SettlementHandler) UpdateSettlement(c *gin.Context) {
	Success(c, nil)
}

func (h *SettlementHandler) DeleteSettlement(c *gin.Context) {
	Success(c, nil)
}

func (h *SettlementHandler) ConfirmByBuyer(c *gin.Context) {
	Success(c, nil)
}

func (h *SettlementHandler) ConfirmBySupplier(c *gin.Context) {
	Success(c, nil)
}

func (h *SettlementHandler) AddDispute(c *gin.Context) {
	Created(c, nil)
}

func (h *SettlementHandler) UpdateDispute(c *gin.Context) {
	Success(c, nil)
}
