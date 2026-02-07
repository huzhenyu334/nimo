package repository

import (
	"context"
	"github.com/bitfantasy/nimo/internal/plm/entity"

	"gorm.io/gorm"
)

type DeliverableRepository struct {
	db *gorm.DB
}

func NewDeliverableRepository(db *gorm.DB) *DeliverableRepository {
	return &DeliverableRepository{db: db}
}

// Create 创建交付物
func (r *DeliverableRepository) Create(ctx context.Context, d *entity.PhaseDeliverable) error {
	return r.db.WithContext(ctx).Create(d).Error
}

// FindByID 根据ID查找
func (r *DeliverableRepository) FindByID(ctx context.Context, id string) (*entity.PhaseDeliverable, error) {
	var d entity.PhaseDeliverable
	err := r.db.WithContext(ctx).First(&d, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListByPhase 获取阶段的交付物列表
func (r *DeliverableRepository) ListByPhase(ctx context.Context, phaseID string) ([]entity.PhaseDeliverable, error) {
	var deliverables []entity.PhaseDeliverable
	err := r.db.WithContext(ctx).
		Where("phase_id = ?", phaseID).
		Order("sort_order ASC").
		Find(&deliverables).Error
	return deliverables, err
}

// ListByProject 获取项目所有交付物
func (r *DeliverableRepository) ListByProject(ctx context.Context, projectID string) ([]entity.PhaseDeliverable, error) {
	var deliverables []entity.PhaseDeliverable
	err := r.db.WithContext(ctx).
		Joins("JOIN project_phases ON project_phases.id = phase_deliverables.phase_id").
		Where("project_phases.project_id = ?", projectID).
		Preload("Phase").
		Order("project_phases.sequence ASC, phase_deliverables.sort_order ASC").
		Find(&deliverables).Error
	return deliverables, err
}

// Update 更新交付物
func (r *DeliverableRepository) Update(ctx context.Context, d *entity.PhaseDeliverable) error {
	return r.db.WithContext(ctx).Save(d).Error
}

// BatchCreate 批量创建交付物
func (r *DeliverableRepository) BatchCreate(ctx context.Context, deliverables []entity.PhaseDeliverable) error {
	if len(deliverables) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&deliverables).Error
}

// CountByPhase 统计阶段交付物完成情况
func (r *DeliverableRepository) CountByPhase(ctx context.Context, phaseID string) (total int64, completed int64, err error) {
	err = r.db.WithContext(ctx).Model(&entity.PhaseDeliverable{}).
		Where("phase_id = ? AND is_required = true", phaseID).
		Count(&total).Error
	if err != nil {
		return
	}
	err = r.db.WithContext(ctx).Model(&entity.PhaseDeliverable{}).
		Where("phase_id = ? AND is_required = true AND status IN ('submitted','approved')", phaseID).
		Count(&completed).Error
	return
}
