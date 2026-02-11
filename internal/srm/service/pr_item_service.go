package service

import (
	"github.com/bitfantasy/nimo/internal/srm/repository"
	"gorm.io/gorm"
)

// PRItemService 采购申请行项服务
type PRItemService struct {
	prRepo          *repository.PRRepository
	projectRepo     *repository.ProjectRepository
	activityLogRepo *repository.ActivityLogRepository
	db              *gorm.DB
}

func NewPRItemService(prRepo *repository.PRRepository, projectRepo *repository.ProjectRepository, activityLogRepo *repository.ActivityLogRepository, db *gorm.DB) *PRItemService {
	return &PRItemService{prRepo: prRepo, projectRepo: projectRepo, activityLogRepo: activityLogRepo, db: db}
}
