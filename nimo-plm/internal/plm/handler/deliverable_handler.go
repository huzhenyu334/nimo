package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/repository"

	"github.com/gin-gonic/gin"
)

type DeliverableHandler struct {
	repo *repository.DeliverableRepository
}

func NewDeliverableHandler(repo *repository.DeliverableRepository) *DeliverableHandler {
	return &DeliverableHandler{repo: repo}
}

// ListByProject GET /projects/:id/deliverables
func (h *DeliverableHandler) ListByProject(c *gin.Context) {
	projectID := c.Param("id")

	deliverables, err := h.repo.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, deliverables)
}

// ListByPhase GET /projects/:id/phases/:phaseId/deliverables
func (h *DeliverableHandler) ListByPhase(c *gin.Context) {
	phaseID := c.Param("phaseId")

	deliverables, err := h.repo.ListByPhase(c.Request.Context(), phaseID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	// 统计完成情况
	total, completed, _ := h.repo.CountByPhase(c.Request.Context(), phaseID)

	Success(c, gin.H{
		"deliverables": deliverables,
		"total":        total,
		"completed":    completed,
	})
}

// Update PUT /projects/:id/deliverables/:deliverableId
func (h *DeliverableHandler) Update(c *gin.Context) {
	deliverableID := c.Param("deliverableId")

	d, err := h.repo.FindByID(c.Request.Context(), deliverableID)
	if err != nil {
		NotFound(c, "Deliverable not found")
		return
	}

	var input struct {
		Status     string  `json:"status"`
		DocumentID *string `json:"document_id"`
		BOMID      *string `json:"bom_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if input.Status != "" {
		d.Status = input.Status
	}
	if input.DocumentID != nil {
		d.DocumentID = input.DocumentID
	}
	if input.BOMID != nil {
		d.BOMID = input.BOMID
	}

	if err := h.repo.Update(c.Request.Context(), d); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, d)
}
