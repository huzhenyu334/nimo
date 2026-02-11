package service

import "github.com/bitfantasy/nimo/internal/srm/repository"

// EquipmentService 设备服务
type EquipmentService struct {
	repo *repository.EquipmentRepository
}

func NewEquipmentService(repo *repository.EquipmentRepository) *EquipmentService {
	return &EquipmentService{repo: repo}
}
