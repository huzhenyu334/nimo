package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
)

// CMFHandler CMF处理器
type CMFHandler struct {
	svc *service.CMFService
}

func NewCMFHandler(svc *service.CMFService) *CMFHandler {
	return &CMFHandler{svc: svc}
}

// GetAppearanceParts 获取外观件列表
// GET /api/v1/projects/:id/tasks/:taskId/cmf/appearance-parts
func (h *CMFHandler) GetAppearanceParts(c *gin.Context) {
	projectID := c.Param("id")
	taskID := c.Param("taskId")

	parts, err := h.svc.GetAppearanceParts(c.Request.Context(), projectID, taskID)
	if err != nil {
		InternalError(c, "获取外观件失败: "+err.Error())
		return
	}
	Success(c, gin.H{"items": parts})
}

// ListDesigns 列出所有CMF方案
// GET /api/v1/projects/:id/tasks/:taskId/cmf/designs
func (h *CMFHandler) ListDesigns(c *gin.Context) {
	projectID := c.Param("id")
	taskID := c.Param("taskId")

	designs, err := h.svc.ListDesigns(c.Request.Context(), projectID, taskID)
	if err != nil {
		InternalError(c, "获取CMF方案失败: "+err.Error())
		return
	}
	Success(c, gin.H{"items": designs})
}

// ListDesignsByProject 列出项目所有CMF方案
// GET /api/v1/projects/:id/cmf/designs
func (h *CMFHandler) ListDesignsByProject(c *gin.Context) {
	projectID := c.Param("id")

	designs, err := h.svc.ListDesignsByProject(c.Request.Context(), projectID)
	if err != nil {
		InternalError(c, "获取CMF方案失败: "+err.Error())
		return
	}
	Success(c, gin.H{"items": designs})
}

// CreateDesign 创建CMF方案
// POST /api/v1/projects/:id/tasks/:taskId/cmf/designs
func (h *CMFHandler) CreateDesign(c *gin.Context) {
	projectID := c.Param("id")
	taskID := c.Param("taskId")

	var input service.CreateDesignInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	design, err := h.svc.CreateDesign(c.Request.Context(), projectID, taskID, &input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Created(c, design)
}

// UpdateDesign 更新CMF方案
// PUT /api/v1/projects/:id/tasks/:taskId/cmf/designs/:designId
func (h *CMFHandler) UpdateDesign(c *gin.Context) {
	designID := c.Param("designId")

	var input service.UpdateDesignInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	design, err := h.svc.UpdateDesign(c.Request.Context(), designID, &input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Success(c, design)
}

// DeleteDesign 删除CMF方案
// DELETE /api/v1/projects/:id/tasks/:taskId/cmf/designs/:designId
func (h *CMFHandler) DeleteDesign(c *gin.Context) {
	designID := c.Param("designId")

	if err := h.svc.DeleteDesign(c.Request.Context(), designID); err != nil {
		InternalError(c, "删除CMF方案失败: "+err.Error())
		return
	}
	Success(c, nil)
}

// AddDrawing 添加图纸
// POST /api/v1/cmf-designs/:designId/drawings
func (h *CMFHandler) AddDrawing(c *gin.Context) {
	designID := c.Param("designId")

	var input service.AddDrawingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	drawing, err := h.svc.AddDrawing(c.Request.Context(), designID, &input)
	if err != nil {
		InternalError(c, err.Error())
		return
	}
	Created(c, drawing)
}

// RemoveDrawing 删除图纸
// DELETE /api/v1/cmf-designs/:designId/drawings/:drawingId
func (h *CMFHandler) RemoveDrawing(c *gin.Context) {
	drawingID := c.Param("drawingId")

	if err := h.svc.RemoveDrawing(c.Request.Context(), drawingID); err != nil {
		InternalError(c, "删除图纸失败: "+err.Error())
		return
	}
	Success(c, nil)
}
