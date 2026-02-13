package repository

import (
	"context"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"gorm.io/gorm"
)

// CMFRepository CMF仓库
type CMFRepository struct {
	db *gorm.DB
}

func NewCMFRepository(db *gorm.DB) *CMFRepository {
	return &CMFRepository{db: db}
}

// ListDesignsByTask 按任务查询所有CMF方案（预加载BOMItem和Drawings）
func (r *CMFRepository) ListDesignsByTask(ctx context.Context, projectID, taskID string) ([]entity.CMFDesign, error) {
	var designs []entity.CMFDesign
	err := r.db.WithContext(ctx).
		Preload("BOMItem").
		Preload("Drawings").
		Where("project_id = ? AND task_id = ?", projectID, taskID).
		Order("bom_item_id ASC, sort_order ASC, created_at ASC").
		Find(&designs).Error
	return designs, err
}

// ListDesignsByProject 按项目查询所有CMF方案
func (r *CMFRepository) ListDesignsByProject(ctx context.Context, projectID string) ([]entity.CMFDesign, error) {
	var designs []entity.CMFDesign
	err := r.db.WithContext(ctx).
		Preload("BOMItem").
		Preload("Drawings").
		Where("project_id = ?", projectID).
		Order("task_id ASC, bom_item_id ASC, sort_order ASC").
		Find(&designs).Error
	return designs, err
}

// ListDesignsByBOMItem 按BOM行项查询CMF方案
func (r *CMFRepository) ListDesignsByBOMItem(ctx context.Context, bomItemID string) ([]entity.CMFDesign, error) {
	var designs []entity.CMFDesign
	err := r.db.WithContext(ctx).
		Preload("Drawings").
		Where("bom_item_id = ?", bomItemID).
		Order("sort_order ASC, created_at ASC").
		Find(&designs).Error
	return designs, err
}

// FindDesignByID 根据ID查找CMF方案
func (r *CMFRepository) FindDesignByID(ctx context.Context, id string) (*entity.CMFDesign, error) {
	var design entity.CMFDesign
	err := r.db.WithContext(ctx).
		Preload("BOMItem").
		Preload("Drawings").
		First(&design, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &design, nil
}

// CreateDesign 创建CMF方案
func (r *CMFRepository) CreateDesign(ctx context.Context, design *entity.CMFDesign) error {
	return r.db.WithContext(ctx).Create(design).Error
}

// UpdateDesign 更新CMF方案
func (r *CMFRepository) UpdateDesign(ctx context.Context, design *entity.CMFDesign) error {
	return r.db.WithContext(ctx).Save(design).Error
}

// DeleteDesign 删除CMF方案及其图纸
func (r *CMFRepository) DeleteDesign(ctx context.Context, id string) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Where("cmf_design_id = ?", id).Delete(&entity.CMFDrawing{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&entity.CMFDesign{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// AddDrawing 添加图纸
func (r *CMFRepository) AddDrawing(ctx context.Context, drawing *entity.CMFDrawing) error {
	return r.db.WithContext(ctx).Create(drawing).Error
}

// RemoveDrawing 删除图纸
func (r *CMFRepository) RemoveDrawing(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.CMFDrawing{}, "id = ?", id).Error
}

// FindDrawingByID 根据ID查找图纸
func (r *CMFRepository) FindDrawingByID(ctx context.Context, id string) (*entity.CMFDrawing, error) {
	var drawing entity.CMFDrawing
	err := r.db.WithContext(ctx).First(&drawing, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &drawing, nil
}
