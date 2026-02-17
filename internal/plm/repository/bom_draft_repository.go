package repository

import (
	"context"
	"github.com/bitfantasy/nimo/internal/plm/entity"
	"gorm.io/gorm"
)

type BOMDraftRepository struct {
	db *gorm.DB
}

func NewBOMDraftRepository(db *gorm.DB) *BOMDraftRepository {
	return &BOMDraftRepository{db: db}
}

// FindByBOMID 根据BOM ID查找草稿
func (r *BOMDraftRepository) FindByBOMID(ctx context.Context, bomID string) (*entity.BOMDraft, error) {
	var draft entity.BOMDraft
	err := r.db.WithContext(ctx).
		Preload("Creator").
		Where("bom_id = ?", bomID).
		First(&draft).Error
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

// Create 创建草稿
func (r *BOMDraftRepository) Create(ctx context.Context, draft *entity.BOMDraft) error {
	return r.db.WithContext(ctx).Create(draft).Error
}

// Update 更新草稿
func (r *BOMDraftRepository) Update(ctx context.Context, draft *entity.BOMDraft) error {
	return r.db.WithContext(ctx).Save(draft).Error
}

// Upsert 创建或更新草稿
func (r *BOMDraftRepository) Upsert(ctx context.Context, draft *entity.BOMDraft) error {
	return r.db.WithContext(ctx).
		Where("bom_id = ?", draft.BOMID).
		Assign(draft).
		FirstOrCreate(draft).Error
}

// Delete 删除草稿
func (r *BOMDraftRepository) Delete(ctx context.Context, bomID string) error {
	return r.db.WithContext(ctx).
		Where("bom_id = ?", bomID).
		Delete(&entity.BOMDraft{}).Error
}
