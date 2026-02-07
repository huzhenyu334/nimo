package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/service"

	"github.com/gin-gonic/gin"
)

type BOMHandler struct {
	svc *service.ProjectBOMService
}

func NewBOMHandler(svc *service.ProjectBOMService) *BOMHandler {
	return &BOMHandler{svc: svc}
}

// ListBOMs GET /projects/:id/boms
func (h *BOMHandler) ListBOMs(c *gin.Context) {
	projectID := c.Param("id")
	bomType := c.Query("bom_type")
	status := c.Query("status")

	boms, err := h.svc.ListBOMs(c.Request.Context(), projectID, bomType, status)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, boms)
}

// GetBOM GET /projects/:id/boms/:bomId
func (h *BOMHandler) GetBOM(c *gin.Context) {
	bomID := c.Param("bomId")

	bom, err := h.svc.GetBOM(c.Request.Context(), bomID)
	if err != nil {
		NotFound(c, "BOM not found")
		return
	}

	Success(c, bom)
}

// CreateBOM POST /projects/:id/boms
func (h *BOMHandler) CreateBOM(c *gin.Context) {
	projectID := c.Param("id")
	var input service.CreateBOMInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	userID := c.GetString("user_id")
	bom, err := h.svc.CreateBOM(c.Request.Context(), projectID, &input, userID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Created(c, bom)
}

// UpdateBOM PUT /projects/:id/boms/:bomId
func (h *BOMHandler) UpdateBOM(c *gin.Context) {
	bomID := c.Param("bomId")
	var input service.UpdateBOMInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	bom, err := h.svc.UpdateBOM(c.Request.Context(), bomID, &input)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, bom)
}

// SubmitBOM POST /projects/:id/boms/:bomId/submit
func (h *BOMHandler) SubmitBOM(c *gin.Context) {
	bomID := c.Param("bomId")
	userID := c.GetString("user_id")

	bom, err := h.svc.SubmitBOM(c.Request.Context(), bomID, userID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, bom)
}

// ApproveBOM POST /projects/:id/boms/:bomId/approve
func (h *BOMHandler) ApproveBOM(c *gin.Context) {
	bomID := c.Param("bomId")
	userID := c.GetString("user_id")

	var input struct {
		Comment string `json:"comment"`
	}
	c.ShouldBindJSON(&input)

	bom, err := h.svc.ApproveBOM(c.Request.Context(), bomID, userID, input.Comment)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, bom)
}

// RejectBOM POST /projects/:id/boms/:bomId/reject
func (h *BOMHandler) RejectBOM(c *gin.Context) {
	bomID := c.Param("bomId")
	userID := c.GetString("user_id")

	var input struct {
		Comment string `json:"comment" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "请填写驳回原因")
		return
	}

	bom, err := h.svc.RejectBOM(c.Request.Context(), bomID, userID, input.Comment)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, bom)
}

// FreezeBOM POST /projects/:id/boms/:bomId/freeze
func (h *BOMHandler) FreezeBOM(c *gin.Context) {
	bomID := c.Param("bomId")
	userID := c.GetString("user_id")

	bom, err := h.svc.FreezeBOM(c.Request.Context(), bomID, userID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, bom)
}

// AddItem POST /projects/:id/boms/:bomId/items
func (h *BOMHandler) AddItem(c *gin.Context) {
	bomID := c.Param("bomId")
	var input service.BOMItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	item, err := h.svc.AddItem(c.Request.Context(), bomID, &input)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Created(c, item)
}

// BatchAddItems POST /projects/:id/boms/:bomId/items/batch
func (h *BOMHandler) BatchAddItems(c *gin.Context) {
	bomID := c.Param("bomId")
	var input struct {
		Items []service.BOMItemInput `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	count, err := h.svc.BatchAddItems(c.Request.Context(), bomID, input.Items)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"created": count})
}

// DeleteItem DELETE /projects/:id/boms/:bomId/items/:itemId
func (h *BOMHandler) DeleteItem(c *gin.Context) {
	bomID := c.Param("bomId")
	itemID := c.Param("itemId")

	if err := h.svc.DeleteItem(c.Request.Context(), bomID, itemID); err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{"deleted": true})
}
