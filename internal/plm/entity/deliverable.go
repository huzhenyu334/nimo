package entity

import "time"

// PhaseDeliverable 阶段交付物
type PhaseDeliverable struct {
	ID              string     `json:"id" gorm:"primaryKey;size:32"`
	PhaseID         string     `json:"phase_id" gorm:"size:32;not null"`
	Name            string     `json:"name" gorm:"size:128;not null"`
	DeliverableType string     `json:"deliverable_type" gorm:"size:16;not null;default:document"` // document/bom/review
	ResponsibleRole string     `json:"responsible_role,omitempty" gorm:"size:32"`
	IsRequired      bool       `json:"is_required" gorm:"default:true"`
	Status          string     `json:"status" gorm:"size:16;not null;default:pending"` // pending/submitted/approved
	DocumentID      *string    `json:"document_id,omitempty" gorm:"size:32"`
	BOMID           *string    `json:"bom_id,omitempty" gorm:"size:32"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
	SubmittedBy     *string    `json:"submitted_by,omitempty" gorm:"size:32"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	ApprovedBy      *string    `json:"approved_by,omitempty" gorm:"size:32"`
	SortOrder       int        `json:"sort_order" gorm:"default:0"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	Phase    *ProjectPhase `json:"phase,omitempty" gorm:"foreignKey:PhaseID"`
	Document *Document     `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	BOM      *ProjectBOM   `json:"bom,omitempty" gorm:"foreignKey:BOMID"`
}

func (PhaseDeliverable) TableName() string {
	return "phase_deliverables"
}
