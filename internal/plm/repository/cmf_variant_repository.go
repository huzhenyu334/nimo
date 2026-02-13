package repository

import (
	"context"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"gorm.io/gorm"
)

type CMFVariantRepository struct {
	db *gorm.DB
}

func NewCMFVariantRepository(db *gorm.DB) *CMFVariantRepository {
	return &CMFVariantRepository{db: db}
}

// ListByBOMItem 获取某零件的所有CMF变体
func (r *CMFVariantRepository) ListByBOMItem(ctx context.Context, bomItemID string) ([]entity.BOMItemCMFVariant, error) {
	var variants []entity.BOMItemCMFVariant
	err := r.db.WithContext(ctx).
		Where("bom_item_id = ?", bomItemID).
		Order("variant_index ASC").
		Find(&variants).Error
	return variants, err
}

// ListByBOMItems 批量获取多个零件的CMF变体
func (r *CMFVariantRepository) ListByBOMItems(ctx context.Context, bomItemIDs []string) ([]entity.BOMItemCMFVariant, error) {
	var variants []entity.BOMItemCMFVariant
	if len(bomItemIDs) == 0 {
		return variants, nil
	}
	err := r.db.WithContext(ctx).
		Where("bom_item_id IN ?", bomItemIDs).
		Order("bom_item_id, variant_index ASC").
		Find(&variants).Error
	return variants, err
}

// FindByID 根据ID查找CMF变体
func (r *CMFVariantRepository) FindByID(ctx context.Context, id string) (*entity.BOMItemCMFVariant, error) {
	var variant entity.BOMItemCMFVariant
	err := r.db.WithContext(ctx).First(&variant, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

// Create 创建CMF变体
func (r *CMFVariantRepository) Create(ctx context.Context, variant *entity.BOMItemCMFVariant) error {
	return r.db.WithContext(ctx).Create(variant).Error
}

// Update 更新CMF变体
func (r *CMFVariantRepository) Update(ctx context.Context, variant *entity.BOMItemCMFVariant) error {
	return r.db.WithContext(ctx).Save(variant).Error
}

// Delete 删除CMF变体
func (r *CMFVariantRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.BOMItemCMFVariant{}, "id = ?", id).Error
}

// GetNextVariantIndex 获取下一个变体序号
func (r *CMFVariantRepository) GetNextVariantIndex(ctx context.Context, bomItemID string) (int, error) {
	var maxIndex int
	err := r.db.WithContext(ctx).
		Model(&entity.BOMItemCMFVariant{}).
		Where("bom_item_id = ?", bomItemID).
		Select("COALESCE(MAX(variant_index), 0)").
		Scan(&maxIndex).Error
	if err != nil {
		return 0, err
	}
	return maxIndex + 1, nil
}

