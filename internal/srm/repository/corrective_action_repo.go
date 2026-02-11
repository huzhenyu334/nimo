package repository

import (
	"gorm.io/gorm"
)

// CorrectiveActionRepository 纠正措施仓库
type CorrectiveActionRepository struct {
	db *gorm.DB
}

func NewCorrectiveActionRepository(db *gorm.DB) *CorrectiveActionRepository {
	return &CorrectiveActionRepository{db: db}
}
