package service

import (
	"context"
	"fmt"

	"github.com/bitfantasy/nimo/internal/srm/entity"
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

// UpdatePRItemStatus 更新PR行项状态（带状态流转校验和日志）
func (s *PRItemService) UpdatePRItemStatus(ctx context.Context, itemID, toStatus, operatorID string) (*entity.PRItem, error) {
	item, err := s.prRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("物料不存在")
	}

	// 验证状态流转合法性
	allowedStatuses, ok := entity.ValidPRItemTransitions[item.Status]
	if !ok {
		return nil, fmt.Errorf("当前状态 %s 不允许流转", item.Status)
	}
	valid := false
	for _, s := range allowedStatuses {
		if s == toStatus {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("不允许从 %s 流转到 %s", item.Status, toStatus)
	}

	fromStatus := item.Status
	item.Status = toStatus

	if err := s.prRepo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("更新状态失败: %w", err)
	}

	// 记录ActivityLog
	if s.activityLogRepo != nil {
		content := fmt.Sprintf("状态变更: %s → %s", fromStatus, toStatus)
		s.activityLogRepo.LogActivity(ctx, "pr_item", item.ID, item.MaterialCode,
			"status_change", fromStatus, toStatus, content, operatorID, "")
	}

	return item, nil
}
