package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
)

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	svc *service.WorkflowService
}

// NewWorkflowHandler 创建工作流处理器
func NewWorkflowHandler(svc *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{svc: svc}
}

// AssignTaskRequest 指派任务请求
type AssignTaskRequest struct {
	AssigneeID   string `json:"assignee_id" binding:"required"`
	FeishuUserID string `json:"feishu_user_id"`
}

// SubmitReviewRequest 提交评审请求
type SubmitReviewRequest struct {
	OutcomeCode string `json:"outcome_code" binding:"required"`
	Comment     string `json:"comment"`
}

// AssignPhaseRolesRequest 指派阶段角色请求
type AssignPhaseRolesRequest struct {
	Assignments []service.RoleAssignment `json:"assignments" binding:"required"`
}

// AssignTask 指派任务
// POST /api/v1/projects/:id/tasks/:taskId/assign
func (h *WorkflowHandler) AssignTask(c *gin.Context) {
	projectID := c.Param("id")
	taskID := c.Param("taskId")
	operatorID := GetUserID(c)

	var req AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.AssignTask(c.Request.Context(), projectID, taskID, req.AssigneeID, req.FeishuUserID, operatorID); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "任务已指派"})
}

// StartTask 开始任务
// POST /api/v1/projects/:id/tasks/:taskId/start
func (h *WorkflowHandler) StartTask(c *gin.Context) {
	projectID := c.Param("id")
	taskID := c.Param("taskId")
	operatorID := GetUserID(c)

	if err := h.svc.StartTask(c.Request.Context(), projectID, taskID, operatorID); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "任务已开始"})
}

// CompleteTask 完成任务
// POST /api/v1/projects/:id/tasks/:taskId/complete
func (h *WorkflowHandler) CompleteTask(c *gin.Context) {
	projectID := c.Param("id")
	taskID := c.Param("taskId")
	operatorID := GetUserID(c)

	if err := h.svc.CompleteTask(c.Request.Context(), projectID, taskID, operatorID); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "任务已完成"})
}

// SubmitReview 提交评审结果
// POST /api/v1/projects/:id/tasks/:taskId/review
func (h *WorkflowHandler) SubmitReview(c *gin.Context) {
	projectID := c.Param("id")
	taskID := c.Param("taskId")
	operatorID := GetUserID(c)

	var req SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.SubmitReview(c.Request.Context(), projectID, taskID, req.OutcomeCode, req.Comment, operatorID); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "评审结果已提交"})
}

// AssignPhaseRoles 指派阶段角色
// POST /api/v1/projects/:id/phases/:phase/assign-roles
func (h *WorkflowHandler) AssignPhaseRoles(c *gin.Context) {
	projectID := c.Param("id")
	phase := c.Param("phase")
	operatorID := GetUserID(c)

	var req AssignPhaseRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.AssignPhaseRoles(c.Request.Context(), projectID, phase, req.Assignments, operatorID); err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"message": "角色已指派"})
}

// GetTaskHistory 获取任务操作历史
// GET /api/v1/projects/:id/tasks/:taskId/history
func (h *WorkflowHandler) GetTaskHistory(c *gin.Context) {
	projectID := c.Param("id")
	taskID := c.Param("taskId")

	logs, err := h.svc.GetTaskHistory(c.Request.Context(), projectID, taskID)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{"items": logs})
}
