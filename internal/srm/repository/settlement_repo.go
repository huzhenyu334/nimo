package repository

import (
	"gorm.io/gorm"
)

// SettlementRepository 结算仓库
type SettlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) *SettlementRepository {
	return &SettlementRepository{db: db}
}
