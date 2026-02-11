package handler

import (
	"github.com/bitfantasy/nimo/internal/srm/service"
	"github.com/gin-gonic/gin"
)

// EquipmentHandler 设备处理器
type EquipmentHandler struct {
	svc *service.EquipmentService
}

func NewEquipmentHandler(svc *service.EquipmentService) *EquipmentHandler {
	return &EquipmentHandler{svc: svc}
}

func (h *EquipmentHandler) List(c *gin.Context) {
	Success(c, nil)
}

func (h *EquipmentHandler) Create(c *gin.Context) {
	Created(c, nil)
}

func (h *EquipmentHandler) Get(c *gin.Context) {
	Success(c, nil)
}

func (h *EquipmentHandler) Update(c *gin.Context) {
	Success(c, nil)
}

func (h *EquipmentHandler) Delete(c *gin.Context) {
	Success(c, nil)
}
