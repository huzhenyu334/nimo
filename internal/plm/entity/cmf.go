package entity

import "time"

// CMFSpec CMF规格
type CMFSpec struct {
	ID        string    `json:"id" gorm:"primaryKey;size:32"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CMFItem CMF项目
type CMFItem struct {
	ID        string    `json:"id" gorm:"primaryKey;size:32"`
	SpecID    string    `json:"spec_id" gorm:"size:32"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
