package entity

import "time"

// BOMECN BOM工程变更通知
type BOMECN struct {
	ID            string     `json:"id" gorm:"primaryKey;size:32"`
	ECNNumber     string     `json:"ecn_number" gorm:"size:32;not null;uniqueIndex"` // ECN-2026-0001
	BOMID         string     `json:"bom_id" gorm:"size:32;not null;index"`
	Title         string     `json:"title" gorm:"size:256;not null"`
	Description   string     `json:"description" gorm:"type:text"`
	ChangeSummary JSONB      `json:"change_summary" gorm:"type:jsonb;not null"` // 变更diff
	Status        string     `json:"status" gorm:"size:16;not null;default:pending"` // pending/approved/rejected
	CreatedBy     string     `json:"created_by" gorm:"size:32;not null"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ApprovedBy    *string    `json:"approved_by,omitempty" gorm:"size:32"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	RejectedBy    *string    `json:"rejected_by,omitempty" gorm:"size:32"`
	RejectedAt    *time.Time `json:"rejected_at,omitempty"`
	RejectionNote string     `json:"rejection_note,omitempty" gorm:"type:text"`

	// Relations
	BOM      *ProjectBOM `json:"bom,omitempty" gorm:"foreignKey:BOMID"`
	Creator  *User       `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Approver *User       `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
	Rejecter *User       `json:"rejecter,omitempty" gorm:"foreignKey:RejectedBy"`
}

func (BOMECN) TableName() string {
	return "bom_ecns"
}

// BOM ECN状态常量
const (
	BOMECNStatusPending  = "pending"
	BOMECNStatusApproved = "approved"
	BOMECNStatusRejected = "rejected"
)
