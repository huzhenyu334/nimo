package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
)

type SKUHandler struct {
	svc *service.SKUService
}

func NewSKUHandler(svc *service.SKUService) *SKUHandler {
	return &SKUHandler{svc: svc}
}

// ListSKUs GET /projects/:id/skus
func (h *SKUHandler) ListSKUs(c *gin.Context) {
	projectID := c.Param("id")
	skus, err := h.svc.ListSKUs(c.Request.Context(), projectID)
	if err != nil {
		InternalError(c, "获取SKU列表失败: "+err.Error())
		return
	}
	Success(c, gin.H{"items": skus})
}

// GetSKU GET /projects/:id/skus/:skuId
func (h *SKUHandler) GetSKU(c *gin.Context) {
	skuID := c.Param("skuId")
	sku, err := h.svc.GetSKU(c.Request.Context(), skuID)
	if err != nil {
		InternalError(c, "获取SKU详情失败: "+err.Error())
		return
	}
	Success(c, sku)
}

// CreateSKU POST /projects/:id/skus
func (h *SKUHandler) CreateSKU(c *gin.Context) {
	projectID := c.Param("id")
	var input service.CreateSKUInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}
	userID := GetUserID(c)
	sku, err := h.svc.CreateSKU(c.Request.Context(), projectID, &input, userID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Created(c, sku)
}

// UpdateSKU PUT /projects/:id/skus/:skuId
func (h *SKUHandler) UpdateSKU(c *gin.Context) {
	skuID := c.Param("skuId")
	var input service.UpdateSKUInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}
	sku, err := h.svc.UpdateSKU(c.Request.Context(), skuID, &input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, sku)
}

// DeleteSKU DELETE /projects/:id/skus/:skuId
func (h *SKUHandler) DeleteSKU(c *gin.Context) {
	skuID := c.Param("skuId")
	if err := h.svc.DeleteSKU(c.Request.Context(), skuID); err != nil {
		InternalError(c, "删除SKU失败: "+err.Error())
		return
	}
	Success(c, nil)
}

// GetCMFConfigs GET /projects/:id/skus/:skuId/cmf
func (h *SKUHandler) GetCMFConfigs(c *gin.Context) {
	skuID := c.Param("skuId")
	configs, err := h.svc.GetCMFConfigs(c.Request.Context(), skuID)
	if err != nil {
		InternalError(c, "获取CMF配置失败: "+err.Error())
		return
	}
	Success(c, gin.H{"items": configs})
}

// BatchSaveCMFConfigs PUT /projects/:id/skus/:skuId/cmf
func (h *SKUHandler) BatchSaveCMFConfigs(c *gin.Context) {
	skuID := c.Param("skuId")
	var inputs []service.CMFConfigInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}
	configs, err := h.svc.BatchSaveCMFConfigs(c.Request.Context(), skuID, inputs)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, gin.H{"items": configs})
}

// GetBOMItems GET /projects/:id/skus/:skuId/bom-items
func (h *SKUHandler) GetBOMItems(c *gin.Context) {
	skuID := c.Param("skuId")
	items, err := h.svc.GetBOMItems(c.Request.Context(), skuID)
	if err != nil {
		InternalError(c, "获取SKU零件列表失败: "+err.Error())
		return
	}
	Success(c, gin.H{"items": items})
}

// BatchSaveBOMItems PUT /projects/:id/skus/:skuId/bom-items
// Body: [{"bom_item_id": "xxx", "quantity": 0, "notes": ""}]
func (h *SKUHandler) BatchSaveBOMItems(c *gin.Context) {
	skuID := c.Param("skuId")
	var inputs []service.SKUBOMItemInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}
	items, err := h.svc.BatchSaveBOMItems(c.Request.Context(), skuID, inputs)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, gin.H{"items": items})
}

// GetFullBOM GET /projects/:id/skus/:skuId/full-bom
func (h *SKUHandler) GetFullBOM(c *gin.Context) {
	projectID := c.Param("id")
	skuID := c.Param("skuId")
	items, err := h.svc.GetFullBOM(c.Request.Context(), skuID, projectID)
	if err != nil {
		InternalError(c, "获取完整BOM失败: "+err.Error())
		return
	}
	Success(c, gin.H{"items": items})
}
