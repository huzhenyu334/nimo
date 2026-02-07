package entity

import "time"

// Workflow task status constants (extending those in project.go)
const (
	TaskStatusUnassigned = "unassigned"
	TaskStatusReviewing  = "reviewing"
	TaskStatusRejected   = "rejected"
)

// Task action type constants
const (
	TaskActionAssign       = "assign"
	TaskActionStart        = "start"
	TaskActionComplete     = "complete"
	TaskActionSubmitReview = "submit_review"
	TaskActionApprove      = "approve"
	TaskActionReject       = "reject"
	TaskActionRollback     = "rollback"
)

// TemplatePhaseRole 阶段角色配置（模板级别）
type TemplatePhaseRole struct {
	ID              string `json:"id" gorm:"primaryKey;size:36"`
	TemplateID      string `json:"template_id" gorm:"size:36;not null;index"`
	Phase           string `json:"phase" gorm:"size:20;not null"`
	RoleCode        string `json:"role_code" gorm:"size:50;not null"`
	RoleName        string `json:"role_name" gorm:"size:100;not null"`
	IsRequired      bool   `json:"is_required" gorm:"default:true"`
	TriggerTaskCode string `json:"trigger_task_code" gorm:"size:50"`
}

func (TemplatePhaseRole) TableName() string { return "template_phase_roles" }

// TemplateTaskOutcome 模板任务评审结果选项
type TemplateTaskOutcome struct {
	ID                 string `json:"id" gorm:"primaryKey;size:36"`
	TemplateID         string `json:"template_id" gorm:"size:36;not null;index"`
	TaskCode           string `json:"task_code" gorm:"size:50;not null"`
	OutcomeCode        string `json:"outcome_code" gorm:"size:50;not null"`
	OutcomeName        string `json:"outcome_name" gorm:"size:100;not null"`
	OutcomeType        string `json:"outcome_type" gorm:"size:20;not null;default:'pass'"`
	RollbackToTaskCode string `json:"rollback_to_task_code" gorm:"size:50"`
	RollbackCascade    bool   `json:"rollback_cascade" gorm:"default:false"`
	SortOrder          int    `json:"sort_order" gorm:"default:0"`
}

func (TemplateTaskOutcome) TableName() string { return "template_task_outcomes" }

// ProjectRoleAssignment 项目角色指派（项目级别）
type ProjectRoleAssignment struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	ProjectID    string    `json:"project_id" gorm:"size:32;not null;index"`
	Phase        string    `json:"phase" gorm:"size:20;not null"`
	RoleCode     string    `json:"role_code" gorm:"size:50;not null"`
	UserID       string    `json:"user_id" gorm:"size:32;not null"`
	FeishuUserID string    `json:"feishu_user_id" gorm:"size:64"`
	AssignedBy   string    `json:"assigned_by" gorm:"size:32;not null"`
	AssignedAt   time.Time `json:"assigned_at" gorm:"autoCreateTime"`
}

func (ProjectRoleAssignment) TableName() string { return "project_role_assignments" }

// TaskActionLog 任务操作历史
type TaskActionLog struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	ProjectID    string    `json:"project_id" gorm:"size:32;not null;index"`
	TaskID       string    `json:"task_id" gorm:"size:32;not null;index"`
	Action       string    `json:"action" gorm:"size:50;not null"`
	FromStatus   string    `json:"from_status" gorm:"size:20"`
	ToStatus     string    `json:"to_status" gorm:"size:20;not null"`
	OperatorID   string    `json:"operator_id" gorm:"size:64;not null"`
	OperatorType string    `json:"operator_type" gorm:"size:20;default:'user'"`
	EventData    JSONB     `json:"event_data" gorm:"type:jsonb"`
	Comment      string    `json:"comment" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (TaskActionLog) TableName() string { return "task_action_logs" }
