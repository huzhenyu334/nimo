package entity

import "time"

// SupplierEvaluation 供应商评估
type SupplierEvaluation struct {
	ID        string    `json:"id" gorm:"primaryKey;size:32"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
