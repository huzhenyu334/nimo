package repository

import (
	"gorm.io/gorm"
)

// RFQRepository 询价仓库
type RFQRepository struct {
	db *gorm.DB
}

func NewRFQRepository(db *gorm.DB) *RFQRepository {
	return &RFQRepository{db: db}
}
