package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
)

type CMFVariantHandler struct {
	svc *service.CMFVariantService
}

func NewCMFVariantHandler(svc *service.CMFVariantService) *CMFVariantHandler {
	return &CMFVariantHandler{svc: svc}
}

// ListVariants GET /projects/:id/bom-items/:itemId/cmf-variants
func (h *CMFVariantHandler) ListVariants(c *gin.Context) {
	itemID := c.Param("itemId")
	variants, err := h.svc.ListByBOMItem(c.Request.Context(), itemID)
	if err != nil {
		InternalError(c, "获取CMF变体列表失败: "+err.Error())
		return
	}
	Success(c, variants)
}

// CreateVariant POST /projects/:id/bom-items/:itemId/cmf-variants
func (h *CMFVariantHandler) CreateVariant(c *gin.Context) {
	itemID := c.Param("itemId")
	var input service.CreateVariantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	variant, err := h.svc.Create(c.Request.Context(), itemID, &input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Created(c, variant)
}

// UpdateVariant PUT /projects/:id/cmf-variants/:variantId
func (h *CMFVariantHandler) UpdateVariant(c *gin.Context) {
	variantID := c.Param("variantId")
	var input service.UpdateVariantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	variant, err := h.svc.Update(c.Request.Context(), variantID, &input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, variant)
}

// DeleteVariant DELETE /projects/:id/cmf-variants/:variantId
func (h *CMFVariantHandler) DeleteVariant(c *gin.Context) {
	variantID := c.Param("variantId")
	if err := h.svc.Delete(c.Request.Context(), variantID); err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, gin.H{"message": "删除成功"})
}

// GetAppearanceParts GET /projects/:id/appearance-parts
func (h *CMFVariantHandler) GetAppearanceParts(c *gin.Context) {
	projectID := c.Param("id")
	parts, err := h.svc.GetAppearanceParts(c.Request.Context(), projectID)
	if err != nil {
		InternalError(c, "获取外观件列表失败: "+err.Error())
		return
	}
	Success(c, parts)
}

// GetSRMItems GET /projects/:id/srm/items
func (h *CMFVariantHandler) GetSRMItems(c *gin.Context) {
	projectID := c.Param("id")
	items, err := h.svc.GetSRMItems(c.Request.Context(), projectID)
	if err != nil {
		InternalError(c, "获取SRM采购项失败: "+err.Error())
		return
	}
	Success(c, items)
}
