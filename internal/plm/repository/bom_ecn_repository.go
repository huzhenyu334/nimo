package repository

import (
	"context"
	"fmt"
	"github.com/bitfantasy/nimo/internal/plm/entity"
	"gorm.io/gorm"
	"time"
)

type BOMECNRepository struct {
	db *gorm.DB
}

func NewBOMECNRepository(db *gorm.DB) *BOMECNRepository {
	return &BOMECNRepository{db: db}
}

// GenerateECNNumber 生成ECN编号 ECN-YYYY-NNNN
func (r *BOMECNRepository) GenerateECNNumber(ctx context.Context) (string, error) {
	year := time.Now().Year()
	prefix := fmt.Sprintf("ECN-%d-", year)

	var ecns []entity.BOMECN
	err := r.db.WithContext(ctx).
		Model(&entity.BOMECN{}).
		Where("ecn_number LIKE ?", prefix+"%").
		Select("ecn_number").
		Find(&ecns).Error
	if err != nil {
		return "", err
	}

	maxNumber := 0
	for _, ecn := range ecns {
		// 提取编号部分 ECN-2026-0001 → 0001
		parts := ecn.ECNNumber[len(prefix):]
		var num int
		if _, err := fmt.Sscanf(parts, "%d", &num); err == nil && num > maxNumber {
			maxNumber = num
		}
	}

	return fmt.Sprintf("%s%04d", prefix, maxNumber+1), nil
}

// Create 创建ECN
func (r *BOMECNRepository) Create(ctx context.Context, ecn *entity.BOMECN) error {
	return r.db.WithContext(ctx).Create(ecn).Error
}

// FindByID 根据ID查找ECN
func (r *BOMECNRepository) FindByID(ctx context.Context, id string) (*entity.BOMECN, error) {
	var ecn entity.BOMECN
	err := r.db.WithContext(ctx).
		Preload("BOM").
		Preload("Creator").
		Preload("Approver").
		Preload("Rejecter").
		First(&ecn, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ecn, nil
}

// List 获取ECN列表
func (r *BOMECNRepository) List(ctx context.Context, bomID string, status string) ([]entity.BOMECN, error) {
	var ecns []entity.BOMECN
	query := r.db.WithContext(ctx).
		Preload("BOM").
		Preload("Creator").
		Preload("Approver").
		Preload("Rejecter")

	if bomID != "" {
		query = query.Where("bom_id = ?", bomID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Find(&ecns).Error
	return ecns, err
}

// Update 更新ECN
func (r *BOMECNRepository) Update(ctx context.Context, ecn *entity.BOMECN) error {
	return r.db.WithContext(ctx).Save(ecn).Error
}
