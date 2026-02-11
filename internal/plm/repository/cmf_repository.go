package repository

import (
	"gorm.io/gorm"
)

// CMFRepository CMF仓库
type CMFRepository struct {
	db *gorm.DB
}

func NewCMFRepository(db *gorm.DB) *CMFRepository {
	return &CMFRepository{db: db}
}
