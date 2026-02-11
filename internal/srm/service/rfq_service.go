package service

import (
	"github.com/bitfantasy/nimo/internal/srm/repository"
	"gorm.io/gorm"
)

// RFQService 询价服务
type RFQService struct {
	rfqRepo         *repository.RFQRepository
	poRepo          *repository.PORepository
	prRepo          *repository.PRRepository
	activityLogRepo *repository.ActivityLogRepository
	db              *gorm.DB
}

func NewRFQService(rfqRepo *repository.RFQRepository, poRepo *repository.PORepository, prRepo *repository.PRRepository, activityLogRepo *repository.ActivityLogRepository, db *gorm.DB) *RFQService {
	return &RFQService{rfqRepo: rfqRepo, poRepo: poRepo, prRepo: prRepo, activityLogRepo: activityLogRepo, db: db}
}
