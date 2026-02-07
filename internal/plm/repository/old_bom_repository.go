package repository

import "gorm.io/gorm"

// BOMRepository 旧BOM仓库（产品关联，保留兼容）
type BOMRepository struct {
	db *gorm.DB
}

func NewBOMRepository(db *gorm.DB) *BOMRepository {
	return &BOMRepository{db: db}
}
