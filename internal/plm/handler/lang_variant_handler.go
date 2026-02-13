package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
)

type LangVariantHandler struct {
	svc *service.LangVariantService
}

func NewLangVariantHandler(svc *service.LangVariantService) *LangVariantHandler {
	return &LangVariantHandler{svc: svc}
}

// ListVariants GET /projects/:id/bom-items/:itemId/lang-variants
func (h *LangVariantHandler) ListVariants(c *gin.Context) {
	itemID := c.Param("itemId")
	variants, err := h.svc.ListByBOMItem(c.Request.Context(), itemID)
	if err != nil {
		InternalError(c, "获取语言变体列表失败: "+err.Error())
		return
	}
	Success(c, variants)
}

// CreateVariant POST /projects/:id/bom-items/:itemId/lang-variants
func (h *LangVariantHandler) CreateVariant(c *gin.Context) {
	itemID := c.Param("itemId")
	var input service.CreateLangVariantInput
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

// UpdateVariant PUT /projects/:id/lang-variants/:variantId
func (h *LangVariantHandler) UpdateVariant(c *gin.Context) {
	variantID := c.Param("variantId")
	var input service.UpdateLangVariantInput
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

// DeleteVariant DELETE /projects/:id/lang-variants/:variantId
func (h *LangVariantHandler) DeleteVariant(c *gin.Context) {
	variantID := c.Param("variantId")
	if err := h.svc.Delete(c.Request.Context(), variantID); err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, gin.H{"message": "删除成功"})
}

// GetMultilangParts GET /projects/:id/multilang-parts
func (h *LangVariantHandler) GetMultilangParts(c *gin.Context) {
	projectID := c.Param("id")
	parts, err := h.svc.GetMultilangParts(c.Request.Context(), projectID)
	if err != nil {
		InternalError(c, "获取多语言件列表失败: "+err.Error())
		return
	}
	Success(c, parts)
}
