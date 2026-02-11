package handler

import (
	"time"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/shared/feishu"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleHandler 角色管理处理器
type RoleHandler struct {
	db           *gorm.DB
	feishuClient *feishu.FeishuClient
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(db *gorm.DB, feishuClient *feishu.FeishuClient) *RoleHandler {
	return &RoleHandler{db: db, feishuClient: feishuClient}
}

// List 获取角色列表
// GET /api/v1/roles
func (h *RoleHandler) List(c *gin.Context) {
	var roles []entity.Role
	if err := h.db.Order("created_at ASC").Find(&roles).Error; err != nil {
		InternalError(c, "获取角色列表失败: "+err.Error())
		return
	}
	Success(c, gin.H{"items": roles})
}

// Get 获取角色详情
// GET /api/v1/roles/:id
func (h *RoleHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var role entity.Role
	if err := h.db.Where("id = ?", id).First(&role).Error; err != nil {
		NotFound(c, "角色不存在")
		return
	}
	Success(c, role)
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// Create 创建角色
// POST /api/v1/roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	now := time.Now()
	roleID := uuid.New().String()[:32]
	role := &entity.Role{
		ID:          roleID,
		Code:        "role_" + roleID,
		Name:        req.Name,
		Description: req.Description,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.db.Create(role).Error; err != nil {
		InternalError(c, "创建角色失败: "+err.Error())
		return
	}

	Created(c, role)
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Update 更新角色
// PUT /api/v1/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var role entity.Role
	if err := h.db.Where("id = ?", id).First(&role).Error; err != nil {
		NotFound(c, "角色不存在")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	role.UpdatedAt = time.Now()

	if err := h.db.Save(&role).Error; err != nil {
		InternalError(c, "更新角色失败: "+err.Error())
		return
	}

	Success(c, role)
}

// Delete 删除角色
// DELETE /api/v1/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var role entity.Role
	if err := h.db.Where("id = ?", id).First(&role).Error; err != nil {
		NotFound(c, "角色不存在")
		return
	}

	if role.IsSystem {
		BadRequest(c, "系统角色不可删除")
		return
	}

	// 清理关联表（user_roles、role_permissions），避免外键约束导致删除失败
	h.db.Where("role_id = ?", id).Delete(&entity.UserRole{})
	h.db.Where("role_id = ?", id).Delete(&entity.RolePermission{})

	if err := h.db.Delete(&role).Error; err != nil {
		InternalError(c, "删除角色失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "角色已删除"})
}

// ListTaskRoles 获取任务角色列表
// GET /api/v1/task-roles
func (h *RoleHandler) ListTaskRoles(c *gin.Context) {
	var taskRoles []entity.TaskRole
	if err := h.db.Order("sort_order ASC, created_at ASC").Find(&taskRoles).Error; err != nil {
		InternalError(c, "获取任务角色列表失败: "+err.Error())
		return
	}
	Success(c, gin.H{"items": taskRoles})
}

// ListFeishuRoles 获取飞书部门列表作为角色
// GET /api/v1/feishu/roles
func (h *RoleHandler) ListFeishuRoles(c *gin.Context) {
	if h.feishuClient == nil {
		Success(c, gin.H{"items": []interface{}{}})
		return
	}

	depts, err := h.feishuClient.ListDepartments(c.Request.Context())
	if err != nil {
		InternalError(c, "获取飞书部门列表失败: "+err.Error())
		return
	}

	type FeishuRole struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	items := make([]FeishuRole, 0, len(depts))
	for _, d := range depts {
		code := d.DepartmentID
		if code == "" {
			code = d.OpenDepartmentID
		}
		items = append(items, FeishuRole{
			Code: code,
			Name: d.Name,
		})
	}

	Success(c, gin.H{"items": items})
}

// ============================================================
// 角色成员管理
// ============================================================

// RoleMemberResult 角色成员查询结果
type RoleMemberResult struct {
	UserID         string `json:"user_id" gorm:"column:user_id"`
	Name           string `json:"name" gorm:"column:name"`
	Email          string `json:"email" gorm:"column:email"`
	AvatarURL      string `json:"avatar_url" gorm:"column:avatar_url"`
	DepartmentName string `json:"department_name" gorm:"column:department_name"`
}

// ListMembers 获取角色成员列表
// GET /api/v1/roles/:id/members
func (h *RoleHandler) ListMembers(c *gin.Context) {
	roleID := c.Param("id")

	var members []RoleMemberResult
	err := h.db.Table("user_roles").
		Select("users.id as user_id, users.name, users.email, users.avatar_url, COALESCE(departments.name, '') as department_name").
		Joins("JOIN users ON users.id = user_roles.user_id AND (users.deleted_at IS NULL)").
		Joins("LEFT JOIN departments ON departments.id = users.department_id").
		Where("user_roles.role_id = ?", roleID).
		Scan(&members).Error
	if err != nil {
		InternalError(c, "获取角色成员失败: "+err.Error())
		return
	}
	if members == nil {
		members = []RoleMemberResult{}
	}
	Success(c, gin.H{"items": members})
}

// AddMembersRequest 添加成员请求
type AddMembersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
}

// AddMembers 添加角色成员
// POST /api/v1/roles/:id/members
func (h *RoleHandler) AddMembers(c *gin.Context) {
	roleID := c.Param("id")

	var req AddMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	now := time.Now()
	for _, userID := range req.UserIDs {
		ur := entity.UserRole{UserID: userID, RoleID: roleID, CreatedAt: now}
		h.db.Where("user_id = ? AND role_id = ?", userID, roleID).FirstOrCreate(&ur)
	}

	Success(c, gin.H{"message": "添加成员成功"})
}

// RemoveMembersRequest 移除成员请求
type RemoveMembersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
}

// RemoveMembers 移除角色成员
// DELETE /api/v1/roles/:id/members
func (h *RoleHandler) RemoveMembers(c *gin.Context) {
	roleID := c.Param("id")

	var req RemoveMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.db.Where("role_id = ? AND user_id IN ?", roleID, req.UserIDs).Delete(&entity.UserRole{}).Error; err != nil {
		InternalError(c, "移除成员失败: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "移除成员成功"})
}

// ============================================================
// 部门树（含用户列表，用于添加成员弹窗）
// ============================================================

// DeptUserItem 部门下的用户
type DeptUserItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// DeptTreeNode 部门树节点
type DeptTreeNode struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	ParentID string          `json:"parent_id"`
	Children []*DeptTreeNode `json:"children"`
	Users    []DeptUserItem  `json:"users"`
}

// ListDepartments 获取部门树（含用户）
// GET /api/v1/departments
func (h *RoleHandler) ListDepartments(c *gin.Context) {
	var depts []entity.Department
	h.db.Order("sort_order ASC, name ASC").Find(&depts)

	var users []entity.User
	h.db.Where("deleted_at IS NULL AND status = ?", "active").Order("name ASC").Find(&users)

	// 构建节点映射
	nodeMap := make(map[string]*DeptTreeNode, len(depts))
	for _, d := range depts {
		nodeMap[d.ID] = &DeptTreeNode{
			ID:       d.ID,
			Name:     d.Name,
			ParentID: d.ParentID,
			Children: []*DeptTreeNode{},
			Users:    []DeptUserItem{},
		}
	}

	// 无部门用户归入虚拟"未分配"节点
	unassigned := &DeptTreeNode{
		ID:       "unassigned",
		Name:     "未分配部门",
		ParentID: "",
		Children: []*DeptTreeNode{},
		Users:    []DeptUserItem{},
	}

	for _, u := range users {
		item := DeptUserItem{ID: u.ID, Name: u.Name, Email: u.Email, AvatarURL: u.AvatarURL}
		if node, ok := nodeMap[u.DepartmentID]; ok {
			node.Users = append(node.Users, item)
		} else {
			unassigned.Users = append(unassigned.Users, item)
		}
	}

	// 组装树
	var roots []*DeptTreeNode
	for _, node := range nodeMap {
		if node.ParentID == "" {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	if len(unassigned.Users) > 0 {
		roots = append(roots, unassigned)
	}

	if roots == nil {
		roots = []*DeptTreeNode{}
	}

	Success(c, gin.H{"items": roots})
}
