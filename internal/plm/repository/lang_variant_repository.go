package repository

import (
	"context"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"gorm.io/gorm"
)

type LangVariantRepository struct {
	db *gorm.DB
}

func NewLangVariantRepository(db *gorm.DB) *LangVariantRepository {
	return &LangVariantRepository{db: db}
}

func (r *LangVariantRepository) ListByBOMItem(ctx context.Context, bomItemID string) ([]entity.BOMItemLangVariant, error) {
	var variants []entity.BOMItemLangVariant
	err := r.db.WithContext(ctx).
		Where("bom_item_id = ?", bomItemID).
		Order("variant_index ASC").
		Find(&variants).Error
	return variants, err
}

func (r *LangVariantRepository) ListByBOMItems(ctx context.Context, bomItemIDs []string) ([]entity.BOMItemLangVariant, error) {
	var variants []entity.BOMItemLangVariant
	if len(bomItemIDs) == 0 {
		return variants, nil
	}
	err := r.db.WithContext(ctx).
		Where("bom_item_id IN ?", bomItemIDs).
		Order("bom_item_id, variant_index ASC").
		Find(&variants).Error
	return variants, err
}

func (r *LangVariantRepository) FindByID(ctx context.Context, id string) (*entity.BOMItemLangVariant, error) {
	var v entity.BOMItemLangVariant
	err := r.db.WithContext(ctx).First(&v, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *LangVariantRepository) Create(ctx context.Context, v *entity.BOMItemLangVariant) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *LangVariantRepository) Update(ctx context.Context, v *entity.BOMItemLangVariant) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *LangVariantRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.BOMItemLangVariant{}, "id = ?", id).Error
}

func (r *LangVariantRepository) GetNextVariantIndex(ctx context.Context, bomItemID string) (int, error) {
	var maxIndex int
	err := r.db.WithContext(ctx).
		Model(&entity.BOMItemLangVariant{}).
		Where("bom_item_id = ?", bomItemID).
		Select("COALESCE(MAX(variant_index), 0)").
		Scan(&maxIndex).Error
	if err != nil {
		return 0, err
	}
	return maxIndex + 1, nil
}
