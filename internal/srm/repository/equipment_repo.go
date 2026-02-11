package repository

import (
	"gorm.io/gorm"
)

// EquipmentRepository 设备仓库
type EquipmentRepository struct {
	db *gorm.DB
}

func NewEquipmentRepository(db *gorm.DB) *EquipmentRepository {
	return &EquipmentRepository{db: db}
}
