package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"gorm.io/gorm"
)

// MaterialRepository 物料仓库
type MaterialRepository struct {
	db *gorm.DB
}

// NewMaterialRepository 创建物料仓库
func NewMaterialRepository(db *gorm.DB) *MaterialRepository {
	return &MaterialRepository{db: db}
}

// MaterialCategoryRepository 物料分类仓库
type MaterialCategoryRepository struct {
	db *gorm.DB
}

// NewMaterialCategoryRepository 创建物料分类仓库
func NewMaterialCategoryRepository(db *gorm.DB) *MaterialCategoryRepository {
	return &MaterialCategoryRepository{db: db}
}

// FindByID 根据ID查找物料
func (r *MaterialRepository) FindByID(ctx context.Context, id string) (*entity.Material, error) {
	var material entity.Material
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&material).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &material, nil
}

// FindByCode 根据编码查找物料
func (r *MaterialRepository) FindByCode(ctx context.Context, code string) (*entity.Material, error) {
	var material entity.Material
	err := r.db.WithContext(ctx).
		Where("code = ? AND deleted_at IS NULL", code).
		First(&material).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &material, nil
}

// Create 创建物料
func (r *MaterialRepository) Create(ctx context.Context, material *entity.Material) error {
	return r.db.WithContext(ctx).Create(material).Error
}

// Update 更新物料
func (r *MaterialRepository) Update(ctx context.Context, material *entity.Material) error {
	return r.db.WithContext(ctx).Save(material).Error
}

// Delete 软删除物料
func (r *MaterialRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Material{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

// List 获取物料列表
func (r *MaterialRepository) List(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]entity.Material, int64, error) {
	var materials []entity.Material
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Material{}).Where("deleted_at IS NULL")

	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if categoryID, ok := filters["category_id"].(string); ok && categoryID != "" {
		// 检查是否为一级分类，若是则查询其下所有二级分类的物料
		var cat entity.MaterialCategory
		if err := r.db.WithContext(ctx).Where("id = ?", categoryID).First(&cat).Error; err == nil && cat.Level == 1 {
			query = query.Where("category_id = ? OR category_id IN (SELECT id FROM material_categories WHERE parent_id = ?)", categoryID, categoryID)
		} else {
			query = query.Where("category_id = ?", categoryID)
		}
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("Category").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&materials).Error

	return materials, total, err
}

// GetCategories 获取物料类别树形结构（一级分类带Children）
func (r *MaterialRepository) GetCategories(ctx context.Context) ([]entity.MaterialCategory, error) {
	var all []entity.MaterialCategory
	err := r.db.WithContext(ctx).
		Order("level ASC, sort_order ASC, name ASC").
		Find(&all).Error
	if err != nil {
		return nil, err
	}

	// 按 parent_id 分组
	childrenMap := make(map[string][]entity.MaterialCategory)
	var roots []entity.MaterialCategory
	for _, cat := range all {
		if cat.Level == 1 || cat.ParentID == "" {
			roots = append(roots, cat)
		} else {
			childrenMap[cat.ParentID] = append(childrenMap[cat.ParentID], cat)
		}
	}

	// 组装树
	for i := range roots {
		if children, ok := childrenMap[roots[i].ID]; ok {
			roots[i].Children = children
		}
	}

	return roots, nil
}

// FindCategoryByID 根据ID查找分类
func (r *MaterialRepository) FindCategoryByID(ctx context.Context, id string) (*entity.MaterialCategory, error) {
	var cat entity.MaterialCategory
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&cat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

// GenerateCode 生成物料编码
// categoryCode 为二级分类code（如 "EL-CAP"），若传入一级code（如 "EL"）则自动补 -OTH
func (r *MaterialRepository) GenerateCode(ctx context.Context, categoryCode string) (string, error) {
	// 如果是一级code（不含-），补 -OTH
	if !strings.Contains(categoryCode, "-") {
		categoryCode = categoryCode + "-OTH"
	}
	var seq int64
	err := r.db.WithContext(ctx).Raw("SELECT nextval('material_code_seq')").Scan(&seq).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%06d", categoryCode, seq), nil
}
