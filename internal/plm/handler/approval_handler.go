package handler

import (
	"github.com/bitfantasy/nimo/internal/plm/service"
	"github.com/gin-gonic/gin"
)

// ApprovalHandler 审批处理器
type ApprovalHandler struct {
	svc *service.ApprovalService
}

// NewApprovalHandler 创建审批处理器
func NewApprovalHandler(svc *service.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{svc: svc}
}

// Create 创建审批请求
// POST /api/v1/approvals
func (h *ApprovalHandler) Create(c *gin.Context) {
	var req service.CreateApprovalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := GetUserID(c)
	if userID == "" {
		Unauthorized(c, "未登录")
		return
	}

	approval, err := h.svc.CreateApproval(c.Request.Context(), req, userID)
	if err != nil {
		InternalError(c, "创建审批失败: "+err.Error())
		return
	}

	Created(c, approval)
}

// List 审批列表
// GET /api/v1/approvals?status=pending&my_pending=true
func (h *ApprovalHandler) List(c *gin.Context) {
	status := c.Query("status")
	myPending := c.Query("my_pending") == "true"
	userID := GetUserID(c)

	approvals, err := h.svc.ListApprovals(c.Request.Context(), status, userID, myPending)
	if err != nil {
		InternalError(c, "获取审批列表失败: "+err.Error())
		return
	}

	Success(c, gin.H{"items": approvals})
}

// Get 审批详情
// GET /api/v1/approvals/:id
func (h *ApprovalHandler) Get(c *gin.Context) {
	approvalID := c.Param("id")

	approval, err := h.svc.GetApproval(c.Request.Context(), approvalID)
	if err != nil {
		NotFound(c, "审批请求不存在")
		return
	}

	Success(c, approval)
}

// ApproveRejectRequest 通过/驳回请求体
type ApproveRejectRequest struct {
	Comment string `json:"comment"`
}

// Approve 通过审批
// POST /api/v1/approvals/:id/approve
func (h *ApprovalHandler) Approve(c *gin.Context) {
	approvalID := c.Param("id")
	userID := GetUserID(c)
	if userID == "" {
		Unauthorized(c, "未登录")
		return
	}

	var req ApproveRejectRequest
	c.ShouldBindJSON(&req)

	if err := h.svc.Approve(c.Request.Context(), approvalID, userID, req.Comment); err != nil {
		InternalError(c, "审批通过操作失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "审批已通过"})
}

// Reject 驳回审批
// POST /api/v1/approvals/:id/reject
func (h *ApprovalHandler) Reject(c *gin.Context) {
	approvalID := c.Param("id")
	userID := GetUserID(c)
	if userID == "" {
		Unauthorized(c, "未登录")
		return
	}

	var req ApproveRejectRequest
	c.ShouldBindJSON(&req)

	if err := h.svc.Reject(c.Request.Context(), approvalID, userID, req.Comment); err != nil {
		InternalError(c, "审批驳回操作失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "审批已驳回"})
}
