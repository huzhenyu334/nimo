package entity

import (
	"time"
)

// 审批状态常量
const (
	PLMApprovalStatusPending  = "pending"
	PLMApprovalStatusApproved = "approved"
	PLMApprovalStatusRejected = "rejected"
	PLMApprovalStatusCanceled = "canceled"
)

// ApprovalRequest 审批请求
type ApprovalRequest struct {
	ID            string     `json:"id" gorm:"primaryKey;size:36"`
	ProjectID     string     `json:"project_id" gorm:"size:32;not null"`
	TaskID        string     `json:"task_id" gorm:"size:32;not null"`
	Title         string     `json:"title" gorm:"size:200;not null"`
	Description   string     `json:"description" gorm:"type:text"`
	Type          string     `json:"type" gorm:"size:50;not null;default:'task_review'"`
	Status        string     `json:"status" gorm:"size:20;not null;default:'pending'"`
	FormData      JSONB      `json:"form_data" gorm:"type:jsonb"`
	Result        string     `json:"result" gorm:"size:20"`
	ResultComment string     `json:"result_comment" gorm:"type:text"`
	RequestedBy   string     `json:"requested_by" gorm:"size:32;not null"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// 关联
	Reviewers []ApprovalReviewer `json:"reviewers,omitempty" gorm:"foreignKey:ApprovalID"`
	Requester *User              `json:"requester,omitempty" gorm:"foreignKey:RequestedBy"`
	Task      *Task              `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	Project   *Project           `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

func (ApprovalRequest) TableName() string {
	return "approval_requests"
}

// ApprovalReviewer 审批审核人
type ApprovalReviewer struct {
	ID         string     `json:"id" gorm:"primaryKey;size:36"`
	ApprovalID string     `json:"approval_id" gorm:"size:36;not null"`
	UserID     string     `json:"user_id" gorm:"size:32;not null"`
	Status     string     `json:"status" gorm:"size:20;not null;default:'pending'"`
	Comment    string     `json:"comment" gorm:"type:text"`
	DecidedAt  *time.Time `json:"decided_at"`
	Sequence   int        `json:"sequence" gorm:"default:0"`

	// 关联
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (ApprovalReviewer) TableName() string {
	return "approval_reviewers"
}
