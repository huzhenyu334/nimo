package repository

import (
	"context"
	"errors"

	"github.com/bitfantasy/nimo/internal/srm/entity"
	"gorm.io/gorm"
)

// InventoryRepository 库存仓库
type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// FindAll 库存列表
func (r *InventoryRepository) FindAll(ctx context.Context, page, pageSize int, filters map[string]string) ([]entity.InventoryRecord, int64, error) {
	var items []entity.InventoryRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.InventoryRecord{})

	if search := filters["search"]; search != "" {
		query = query.Where("material_name ILIKE ? OR material_code ILIKE ? OR mpn ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if warehouse := filters["warehouse"]; warehouse != "" {
		query = query.Where("warehouse = ?", warehouse)
	}
	if supplierID := filters["supplier_id"]; supplierID != "" {
		query = query.Where("supplier_id = ?", supplierID)
	}
	if lowStock := filters["low_stock"]; lowStock == "true" {
		query = query.Where("quantity < safety_stock AND safety_stock > 0")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("Supplier").
		Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error

	return items, total, err
}

// FindByID 根据ID查找
func (r *InventoryRepository) FindByID(ctx context.Context, id string) (*entity.InventoryRecord, error) {
	var record entity.InventoryRecord
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Where("id = ?", id).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &record, nil
}

// FindByMaterialAndSupplier 根据物料和供应商查找
func (r *InventoryRepository) FindByMaterialAndSupplier(ctx context.Context, materialCode string, supplierID *string) (*entity.InventoryRecord, error) {
	var record entity.InventoryRecord
	query := r.db.WithContext(ctx).Where("material_code = ?", materialCode)
	if supplierID != nil && *supplierID != "" {
		query = query.Where("supplier_id = ?", *supplierID)
	} else {
		query = query.Where("supplier_id IS NULL OR supplier_id = ''")
	}
	err := query.First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // not found is ok, will create
		}
		return nil, err
	}
	return &record, nil
}

// Create 创建
func (r *InventoryRepository) Create(ctx context.Context, record *entity.InventoryRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// Update 更新
func (r *InventoryRepository) Update(ctx context.Context, record *entity.InventoryRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// CreateTransaction 创建流水
func (r *InventoryRepository) CreateTransaction(ctx context.Context, tx *entity.InventoryTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// FindTransactions 查询流水
func (r *InventoryRepository) FindTransactions(ctx context.Context, inventoryID string, page, pageSize int) ([]entity.InventoryTransaction, int64, error) {
	var items []entity.InventoryTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.InventoryTransaction{}).Where("inventory_id = ?", inventoryID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error

	return items, total, err
}
