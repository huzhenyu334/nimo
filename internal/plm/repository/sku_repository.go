package repository

import (
	"context"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"gorm.io/gorm"
)

type SKURepository struct {
	db *gorm.DB
}

func NewSKURepository(db *gorm.DB) *SKURepository {
	return &SKURepository{db: db}
}

func (r *SKURepository) DB() *gorm.DB {
	return r.db
}

// ========== ProductSKU ==========

func (r *SKURepository) Create(ctx context.Context, sku *entity.ProductSKU) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

func (r *SKURepository) FindByID(ctx context.Context, id string) (*entity.ProductSKU, error) {
	var sku entity.ProductSKU
	err := r.db.WithContext(ctx).
		Preload("CMFConfigs").
		Preload("CMFConfigs.BOMItem").
		Preload("BOMItems").
		Preload("BOMItems.BOMItem").
		Preload("BOMItems.CMFVariant").
		First(&sku, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *SKURepository) ListByProject(ctx context.Context, projectID string) ([]entity.ProductSKU, error) {
	var skus []entity.ProductSKU
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("sort_order ASC, created_at ASC").
		Find(&skus).Error
	return skus, err
}

func (r *SKURepository) Update(ctx context.Context, sku *entity.ProductSKU) error {
	return r.db.WithContext(ctx).Save(sku).Error
}

func (r *SKURepository) Delete(ctx context.Context, id string) error {
	r.db.WithContext(ctx).Where("sku_id = ?", id).Delete(&entity.SKUCMFConfig{})
	r.db.WithContext(ctx).Where("sku_id = ?", id).Delete(&entity.SKUBOMItem{})
	return r.db.WithContext(ctx).Delete(&entity.ProductSKU{}, "id = ?", id).Error
}

// ========== SKUCMFConfig ==========

func (r *SKURepository) ListCMFConfigs(ctx context.Context, skuID string) ([]entity.SKUCMFConfig, error) {
	var configs []entity.SKUCMFConfig
	err := r.db.WithContext(ctx).
		Preload("BOMItem").
		Where("sku_id = ?", skuID).
		Find(&configs).Error
	return configs, err
}

func (r *SKURepository) BatchSaveCMFConfigs(ctx context.Context, skuID string, configs []entity.SKUCMFConfig) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Where("sku_id = ?", skuID).Delete(&entity.SKUCMFConfig{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(configs) > 0 {
		if err := tx.Create(&configs).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// ========== SKUBOMItem（SKU零件勾选） ==========

func (r *SKURepository) ListBOMItems(ctx context.Context, skuID string) ([]entity.SKUBOMItem, error) {
	var items []entity.SKUBOMItem
	err := r.db.WithContext(ctx).
		Preload("BOMItem").
		Where("sku_id = ?", skuID).
		Find(&items).Error
	return items, err
}

// BatchSaveBOMItems 批量保存SKU的BOM零件勾选（先删后插）
func (r *SKURepository) BatchSaveBOMItems(ctx context.Context, skuID string, items []entity.SKUBOMItem) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Where("sku_id = ?", skuID).Delete(&entity.SKUBOMItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *SKURepository) CreateBOMItem(ctx context.Context, item *entity.SKUBOMItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *SKURepository) DeleteBOMItem(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.SKUBOMItem{}, "id = ?", id).Error
}
