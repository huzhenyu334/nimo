package entity

import "time"

// BOMDraft BOM草稿临时保存
type BOMDraft struct {
	ID        string    `json:"id" gorm:"primaryKey;size:32"`
	BOMID     string    `json:"bom_id" gorm:"size:32;not null;uniqueIndex"`
	DraftData JSONB     `json:"draft_data" gorm:"type:jsonb;not null"` // 临时修改的BOM数据
	CreatedBy string    `json:"created_by" gorm:"size:32;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	BOM     *ProjectBOM `json:"bom,omitempty" gorm:"foreignKey:BOMID"`
	Creator *User       `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

func (BOMDraft) TableName() string {
	return "bom_drafts"
}
