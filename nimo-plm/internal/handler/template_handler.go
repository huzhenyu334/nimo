package handler

import (
	"time"

	"github.com/bitfantasy/nimo-plm/internal/model/entity"
	"github.com/bitfantasy/nimo-plm/internal/service"
	"github.com/gin-gonic/gin"
)

// TemplateHandler 模板处理器
type TemplateHandler struct {
	svc *service.TemplateService
}

// NewTemplateHandler 创建模板处理器
func NewTemplateHandler(svc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{svc: svc}
}

// List 获取模板列表
func (h *TemplateHandler) List(c *gin.Context) {
	templateType := c.Query("type")
	productType := c.Query("product_type")
	activeOnly := c.Query("active_only") != "false"

	templates, err := h.svc.ListTemplates(c.Request.Context(), templateType, productType, activeOnly)
	if err != nil {
		InternalError(c, "Failed to list templates")
		return
	}

	Success(c, templates)
}

// Get 获取模板详情
func (h *TemplateHandler) Get(c *gin.Context) {
	id := c.Param("id")

	template, err := h.svc.GetTemplate(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "Template not found")
		return
	}

	Success(c, template)
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Code          string   `json:"code" binding:"required"`
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	ProductType   string   `json:"product_type"`
	Phases        []string `json:"phases"`
	EstimatedDays int      `json:"estimated_days"`
}

// Create 创建模板
func (h *TemplateHandler) Create(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	userID := GetUserID(c)

	template := &entity.ProjectTemplate{
		Code:          req.Code,
		Name:          req.Name,
		Description:   req.Description,
		TemplateType:  "CUSTOM",
		ProductType:   req.ProductType,
		EstimatedDays: req.EstimatedDays,
		IsActive:      true,
		CreatedBy:     userID,
	}

	if err := h.svc.CreateTemplate(c.Request.Context(), template); err != nil {
		InternalError(c, "Failed to create template")
		return
	}

	Created(c, template)
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ProductType   string `json:"product_type"`
	EstimatedDays int    `json:"estimated_days"`
	IsActive      *bool  `json:"is_active"`
}

// Update 更新模板
func (h *TemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	template, err := h.svc.GetTemplate(c.Request.Context(), id)
	if err != nil {
		NotFound(c, "Template not found")
		return
	}

	// 不能修改系统模板
	if template.TemplateType == "SYSTEM" {
		Forbidden(c, "Cannot modify system template")
		return
	}

	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.ProductType != "" {
		template.ProductType = req.ProductType
	}
	if req.EstimatedDays > 0 {
		template.EstimatedDays = req.EstimatedDays
	}
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}

	if err := h.svc.UpdateTemplate(c.Request.Context(), template); err != nil {
		InternalError(c, "Failed to update template")
		return
	}

	Success(c, template)
}

// Delete 删除模板
func (h *TemplateHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.DeleteTemplate(c.Request.Context(), id); err != nil {
		if err.Error() == "cannot delete system template" {
			Forbidden(c, err.Error())
			return
		}
		InternalError(c, "Failed to delete template")
		return
	}

	Success(c, nil)
}

// DuplicateTemplateRequest 复制模板请求
type DuplicateTemplateRequest struct {
	NewCode string `json:"new_code" binding:"required"`
	NewName string `json:"new_name" binding:"required"`
}

// Duplicate 复制模板
func (h *TemplateHandler) Duplicate(c *gin.Context) {
	id := c.Param("id")

	var req DuplicateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	userID := GetUserID(c)

	newTemplate, err := h.svc.DuplicateTemplate(c.Request.Context(), id, req.NewCode, req.NewName, userID)
	if err != nil {
		InternalError(c, "Failed to duplicate template: "+err.Error())
		return
	}

	Created(c, newTemplate)
}

// CreateTemplateTaskRequest 创建模板任务请求
type CreateTemplateTaskRequest struct {
	TaskCode            string `json:"task_code" binding:"required"`
	Name                string `json:"name" binding:"required"`
	Description         string `json:"description"`
	Phase               string `json:"phase" binding:"required"`
	ParentTaskCode      string `json:"parent_task_code"`
	TaskType            string `json:"task_type"`
	DefaultAssigneeRole string `json:"default_assignee_role"`
	EstimatedDays       int    `json:"estimated_days"`
	IsCritical          bool   `json:"is_critical"`
	RequiresApproval    bool   `json:"requires_approval"`
	ApprovalType        string `json:"approval_type"`
	SortOrder           int    `json:"sort_order"`
}

// CreateTask 创建模板任务
func (h *TemplateHandler) CreateTask(c *gin.Context) {
	templateID := c.Param("id")

	var req CreateTemplateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	task := &entity.TemplateTask{
		TemplateID:          templateID,
		TaskCode:            req.TaskCode,
		Name:                req.Name,
		Description:         req.Description,
		Phase:               req.Phase,
		ParentTaskCode:      req.ParentTaskCode,
		TaskType:            req.TaskType,
		DefaultAssigneeRole: req.DefaultAssigneeRole,
		EstimatedDays:       req.EstimatedDays,
		IsCritical:          req.IsCritical,
		RequiresApproval:    req.RequiresApproval,
		ApprovalType:        req.ApprovalType,
		SortOrder:           req.SortOrder,
	}

	if task.TaskType == "" {
		task.TaskType = "TASK"
	}
	if task.EstimatedDays == 0 {
		task.EstimatedDays = 1
	}

	if err := h.svc.CreateTemplateTask(c.Request.Context(), task); err != nil {
		InternalError(c, "Failed to create task")
		return
	}

	Created(c, task)
}

// UpdateTask 更新模板任务
func (h *TemplateHandler) UpdateTask(c *gin.Context) {
	templateID := c.Param("id")
	taskCode := c.Param("taskCode")

	var req CreateTemplateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body")
		return
	}

	task := &entity.TemplateTask{
		TemplateID:          templateID,
		TaskCode:            taskCode,
		Name:                req.Name,
		Description:         req.Description,
		Phase:               req.Phase,
		ParentTaskCode:      req.ParentTaskCode,
		TaskType:            req.TaskType,
		DefaultAssigneeRole: req.DefaultAssigneeRole,
		EstimatedDays:       req.EstimatedDays,
		IsCritical:          req.IsCritical,
		RequiresApproval:    req.RequiresApproval,
		ApprovalType:        req.ApprovalType,
		SortOrder:           req.SortOrder,
	}

	if err := h.svc.UpdateTemplateTask(c.Request.Context(), task); err != nil {
		InternalError(c, "Failed to update task")
		return
	}

	Success(c, task)
}

// DeleteTask 删除模板任务
func (h *TemplateHandler) DeleteTask(c *gin.Context) {
	templateID := c.Param("id")
	taskCode := c.Param("taskCode")

	if err := h.svc.DeleteTemplateTask(c.Request.Context(), templateID, taskCode); err != nil {
		InternalError(c, "Failed to delete task")
		return
	}

	Success(c, nil)
}

// CreateProjectFromTemplateRequest 从模板创建项目请求
type CreateProjectFromTemplateRequest struct {
	TemplateID      string            `json:"template_id" binding:"required"`
	ProjectName     string            `json:"project_name" binding:"required"`
	ProjectCode     string            `json:"project_code" binding:"required"`
	ProductID       string            `json:"product_id"`
	StartDate       string            `json:"start_date" binding:"required"`
	PMID            string            `json:"pm_user_id" binding:"required"`
	SkipWeekends    bool              `json:"skip_weekends"`
	RoleAssignments map[string]string `json:"role_assignments"`
}

// CreateProjectFromTemplate 从模板创建项目
func (h *TemplateHandler) CreateProjectFromTemplate(c *gin.Context) {
	var req CreateProjectFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		BadRequest(c, "Invalid start_date format, use YYYY-MM-DD")
		return
	}

	userID := GetUserID(c)

	input := &service.CreateProjectFromTemplateInput{
		TemplateID:      req.TemplateID,
		ProjectName:     req.ProjectName,
		ProjectCode:     req.ProjectCode,
		ProductID:       req.ProductID,
		StartDate:       startDate,
		PMID:            req.PMID,
		SkipWeekends:    req.SkipWeekends,
		RoleAssignments: req.RoleAssignments,
	}

	project, err := h.svc.CreateProjectFromTemplate(c.Request.Context(), input, userID)
	if err != nil {
		InternalError(c, "Failed to create project: "+err.Error())
		return
	}

	Created(c, project)
}
