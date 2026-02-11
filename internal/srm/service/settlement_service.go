package service

import "github.com/bitfantasy/nimo/internal/srm/repository"

// SettlementService 结算服务
type SettlementService struct {
	repo *repository.SettlementRepository
}

func NewSettlementService(repo *repository.SettlementRepository) *SettlementService {
	return &SettlementService{repo: repo}
}
