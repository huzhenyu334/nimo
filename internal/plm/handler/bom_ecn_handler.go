package handler

import (
	"net/http"

	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
)

type BOMECNHandler struct {
	svc *service.BOMECNService
}

func NewBOMECNHandler(svc *service.BOMECNService) *BOMECNHandler {
	return &BOMECNHandler{svc: svc}
}

// SaveDraft POST /api/v1/bom/:id/draft
func (h *BOMECNHandler) SaveDraft(c *gin.Context) {
	bomID := c.Param("id")
	userID := c.GetString("user_id")

	var input service.DraftData
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	draft, err := h.svc.SaveDraft(c.Request.Context(), bomID, &input, userID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    draft,
	})
}

// GetDraft GET /api/v1/bom/:id/draft
func (h *BOMECNHandler) GetDraft(c *gin.Context) {
	bomID := c.Param("id")

	draft, err := h.svc.GetDraft(c.Request.Context(), bomID)
	if err != nil {
		NotFound(c, "未找到草稿")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    draft,
	})
}

// DiscardDraft DELETE /api/v1/bom/:id/draft
func (h *BOMECNHandler) DiscardDraft(c *gin.Context) {
	bomID := c.Param("id")

	if err := h.svc.DiscardDraft(c.Request.Context(), bomID); err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "草稿已撤销",
		"data":    gin.H{"discarded": true},
	})
}

// StartEditing POST /api/v1/bom/:id/edit
func (h *BOMECNHandler) StartEditing(c *gin.Context) {
	bomID := c.Param("id")

	bom, err := h.svc.StartEditing(c.Request.Context(), bomID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "已进入编辑模式",
		"data":    bom,
	})
}

// SubmitECN POST /api/v1/bom/:id/ecn
func (h *BOMECNHandler) SubmitECN(c *gin.Context) {
	bomID := c.Param("id")
	userID := c.GetString("user_id")

	var input struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "请提供ECN标题")
		return
	}

	ecn, err := h.svc.SubmitECN(c.Request.Context(), bomID, input.Title, userID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "ECN已提交",
		"data":    ecn,
	})
}

// ListECNs GET /api/v1/ecn
func (h *BOMECNHandler) ListECNs(c *gin.Context) {
	bomID := c.Query("bom_id")
	status := c.Query("status")

	ecns, err := h.svc.ListECNs(c.Request.Context(), bomID, status)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    gin.H{"items": ecns, "total": len(ecns)},
	})
}

// GetECN GET /api/v1/ecn/:id
func (h *BOMECNHandler) GetECN(c *gin.Context) {
	ecnID := c.Param("id")

	ecn, err := h.svc.GetECN(c.Request.Context(), ecnID)
	if err != nil {
		NotFound(c, "ECN not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    ecn,
	})
}

// ApproveECN POST /api/v1/ecn/:id/approve
func (h *BOMECNHandler) ApproveECN(c *gin.Context) {
	ecnID := c.Param("id")
	userID := c.GetString("user_id")

	ecn, err := h.svc.ApproveECN(c.Request.Context(), ecnID, userID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ECN已批准",
		"data":    ecn,
	})
}

// RejectECN POST /api/v1/ecn/:id/reject
func (h *BOMECNHandler) RejectECN(c *gin.Context) {
	ecnID := c.Param("id")
	userID := c.GetString("user_id")

	var input struct {
		Note string `json:"note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "请提供拒绝原因")
		return
	}

	ecn, err := h.svc.RejectECN(c.Request.Context(), ecnID, userID, input.Note)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ECN已拒绝",
		"data":    ecn,
	})
}
