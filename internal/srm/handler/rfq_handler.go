package handler

import (
	"github.com/bitfantasy/nimo/internal/srm/service"
	"github.com/gin-gonic/gin"
)

// RFQHandler 询价处理器
type RFQHandler struct {
	svc *service.RFQService
}

func NewRFQHandler(svc *service.RFQService) *RFQHandler {
	return &RFQHandler{svc: svc}
}

func (h *RFQHandler) ListRFQs(c *gin.Context) {
	Success(c, nil)
}

func (h *RFQHandler) CreateRFQ(c *gin.Context) {
	Created(c, nil)
}

func (h *RFQHandler) GetRFQ(c *gin.Context) {
	Success(c, nil)
}

func (h *RFQHandler) AddQuote(c *gin.Context) {
	Created(c, nil)
}

func (h *RFQHandler) UpdateQuote(c *gin.Context) {
	Success(c, nil)
}

func (h *RFQHandler) SelectQuote(c *gin.Context) {
	Success(c, nil)
}

func (h *RFQHandler) ConvertToPO(c *gin.Context) {
	Created(c, nil)
}

func (h *RFQHandler) GetComparison(c *gin.Context) {
	Success(c, nil)
}
