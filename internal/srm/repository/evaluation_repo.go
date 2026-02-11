package repository

import (
	"gorm.io/gorm"
)

// EvaluationRepository 评估仓库
type EvaluationRepository struct {
	db *gorm.DB
}

func NewEvaluationRepository(db *gorm.DB) *EvaluationRepository {
	return &EvaluationRepository{db: db}
}
