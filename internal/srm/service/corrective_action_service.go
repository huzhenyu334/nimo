package service

import (
	"github.com/bitfantasy/nimo/internal/shared/feishu"
	"github.com/bitfantasy/nimo/internal/srm/repository"
)

// CorrectiveActionService 纠正措施服务
type CorrectiveActionService struct {
	repo           *repository.CorrectiveActionRepository
	inspectionRepo *repository.InspectionRepository
	feishuClient   *feishu.FeishuClient
}

func NewCorrectiveActionService(repo *repository.CorrectiveActionRepository, inspectionRepo *repository.InspectionRepository) *CorrectiveActionService {
	return &CorrectiveActionService{repo: repo, inspectionRepo: inspectionRepo}
}

// SetFeishuClient 注入飞书客户端
func (s *CorrectiveActionService) SetFeishuClient(fc *feishu.FeishuClient) {
	s.feishuClient = fc
}
