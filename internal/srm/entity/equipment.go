package entity

import "time"

// Equipment 设备
type Equipment struct {
	ID        string    `json:"id" gorm:"primaryKey;size:32"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
