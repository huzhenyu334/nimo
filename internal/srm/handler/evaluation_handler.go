package handler

import (
	"github.com/bitfantasy/nimo/internal/srm/service"
	"github.com/gin-gonic/gin"
)

// EvaluationHandler 评估处理器
type EvaluationHandler struct {
	svc *service.EvaluationService
}

func NewEvaluationHandler(svc *service.EvaluationService) *EvaluationHandler {
	return &EvaluationHandler{svc: svc}
}

func (h *EvaluationHandler) ListEvaluations(c *gin.Context) {
	Success(c, nil)
}

func (h *EvaluationHandler) CreateEvaluation(c *gin.Context) {
	Created(c, nil)
}

func (h *EvaluationHandler) AutoGenerate(c *gin.Context) {
	Created(c, nil)
}

func (h *EvaluationHandler) GetSupplierHistory(c *gin.Context) {
	Success(c, nil)
}

func (h *EvaluationHandler) GetEvaluation(c *gin.Context) {
	Success(c, nil)
}

func (h *EvaluationHandler) UpdateEvaluation(c *gin.Context) {
	Success(c, nil)
}

func (h *EvaluationHandler) Submit(c *gin.Context) {
	Success(c, nil)
}

func (h *EvaluationHandler) Approve(c *gin.Context) {
	Success(c, nil)
}
