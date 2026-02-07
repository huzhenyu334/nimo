package entity

import "time"

// ProjectCodename 项目代号
type ProjectCodename struct {
	ID               string  `json:"id" gorm:"primaryKey;size:32"`
	Codename         string  `json:"codename" gorm:"size:32;not null"`
	CodenameType     string  `json:"codename_type" gorm:"size:16;not null"` // platform/product
	Generation       *int    `json:"generation,omitempty"`
	Theme            string  `json:"theme,omitempty" gorm:"size:64"`
	Description      string  `json:"description,omitempty" gorm:"size:256"`
	IsUsed           bool    `json:"is_used" gorm:"default:false"`
	UsedByProjectID  *string `json:"used_by_project_id,omitempty" gorm:"size:32"`
	CreatedAt        time.Time `json:"created_at"`

	// Relations
	UsedByProject *Project `json:"used_by_project,omitempty" gorm:"foreignKey:UsedByProjectID"`
}

func (ProjectCodename) TableName() string {
	return "project_codenames"
}
