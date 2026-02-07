package repository

import (
	"context"
	"github.com/bitfantasy/nimo-plm/internal/model/entity"

	"gorm.io/gorm"
)

type CodenameRepository struct {
	db *gorm.DB
}

func NewCodenameRepository(db *gorm.DB) *CodenameRepository {
	return &CodenameRepository{db: db}
}

// List 获取代号列表
func (r *CodenameRepository) List(ctx context.Context, codenameType string, availableOnly bool) ([]entity.ProjectCodename, error) {
	var codenames []entity.ProjectCodename
	query := r.db.WithContext(ctx)

	if codenameType != "" {
		query = query.Where("codename_type = ?", codenameType)
	}
	if availableOnly {
		query = query.Where("is_used = false")
	}

	err := query.Order("generation ASC NULLS LAST, codename ASC").Find(&codenames).Error
	return codenames, err
}

// MarkUsed 标记代号为已使用
func (r *CodenameRepository) MarkUsed(ctx context.Context, id string, projectID string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ProjectCodename{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_used":            true,
			"used_by_project_id": projectID,
		}).Error
}

// MarkAvailable 标记代号为可用
func (r *CodenameRepository) MarkAvailable(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entity.ProjectCodename{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_used":            false,
			"used_by_project_id": nil,
		}).Error
}

// FindByCodename 根据代号查找
func (r *CodenameRepository) FindByCodename(ctx context.Context, codename, codenameType string) (*entity.ProjectCodename, error) {
	var cn entity.ProjectCodename
	err := r.db.WithContext(ctx).
		Where("codename = ? AND codename_type = ?", codename, codenameType).
		First(&cn).Error
	if err != nil {
		return nil, err
	}
	return &cn, nil
}
