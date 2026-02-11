package service

import "github.com/bitfantasy/nimo/internal/srm/repository"

// EvaluationService 评估服务
type EvaluationService struct {
	repo         *repository.EvaluationRepository
	supplierRepo *repository.SupplierRepository
}

func NewEvaluationService(repo *repository.EvaluationRepository) *EvaluationService {
	return &EvaluationService{repo: repo}
}

func (s *EvaluationService) SetSupplierRepo(repo *repository.SupplierRepository) {
	s.supplierRepo = repo
}
