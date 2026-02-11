package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/repository"
	"github.com/gin-gonic/gin"
)

// CMFHandler CMF处理器
type CMFHandler struct {
	cmfRepo        *repository.CMFRepository
	projectBOMRepo *repository.ProjectBOMRepository
	taskRepo       *repository.TaskRepository
}

func NewCMFHandler(cmfRepo *repository.CMFRepository, projectBOMRepo *repository.ProjectBOMRepository, taskRepo *repository.TaskRepository) *CMFHandler {
	return &CMFHandler{cmfRepo: cmfRepo, projectBOMRepo: projectBOMRepo, taskRepo: taskRepo}
}

func (h *CMFHandler) ListSpecsByProject(c *gin.Context) {
	// TODO: implement
	Success(c, nil)
}

func (h *CMFHandler) ListSpecsByTask(c *gin.Context) {
	// TODO: implement
	Success(c, nil)
}

func (h *CMFHandler) CreateSpec(c *gin.Context) {
	// TODO: implement
	Created(c, nil)
}

func (h *CMFHandler) ResolveSource(c *gin.Context) {
	// TODO: implement
	Success(c, nil)
}

func (h *CMFHandler) GetSpec(c *gin.Context) {
	// TODO: implement
	Success(c, nil)
}

func (h *CMFHandler) UpdateSpec(c *gin.Context) {
	// TODO: implement
	Success(c, nil)
}

func (h *CMFHandler) DeleteSpec(c *gin.Context) {
	// TODO: implement
	Success(c, nil)
}

func (h *CMFHandler) ListAppearanceParts(c *gin.Context) {
	// TODO: implement
	Success(c, nil)
}
