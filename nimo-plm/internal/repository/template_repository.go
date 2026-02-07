package repository

import (
	"context"

	"github.com/bitfantasy/nimo-plm/internal/model/entity"
	"gorm.io/gorm"
)

// TemplateRepository 模板仓库
type TemplateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository 创建模板仓库
func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// DB 暴露数据库连接（用于事务）
func (r *TemplateRepository) DB() *gorm.DB {
	return r.db
}

// List 获取模板列表
func (r *TemplateRepository) List(ctx context.Context, templateType string, productType string, activeOnly bool) ([]entity.ProjectTemplate, error) {
	var templates []entity.ProjectTemplate
	query := r.db.WithContext(ctx)

	if templateType != "" {
		query = query.Where("template_type = ?", templateType)
	}
	if productType != "" {
		query = query.Where("product_type = ?", productType)
	}
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("template_type ASC, name ASC").Find(&templates).Error
	return templates, err
}

// GetByID 根据ID获取模板
func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*entity.ProjectTemplate, error) {
	var template entity.ProjectTemplate
	err := r.db.WithContext(ctx).First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetByCode 根据编码获取模板
func (r *TemplateRepository) GetByCode(ctx context.Context, code string) (*entity.ProjectTemplate, error) {
	var template entity.ProjectTemplate
	err := r.db.WithContext(ctx).First(&template, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetWithTasks 获取模板及其任务
func (r *TemplateRepository) GetWithTasks(ctx context.Context, id string) (*entity.ProjectTemplate, error) {
	var template entity.ProjectTemplate
	err := r.db.WithContext(ctx).
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Dependencies").
		First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	// 构建任务树（父子关系）
	buildTaskTree(&template.Tasks)

	return &template, nil
}

// buildTaskTree 构建任务树
func buildTaskTree(tasks *[]entity.TemplateTask) {
	taskMap := make(map[string]*entity.TemplateTask)
	for i := range *tasks {
		taskMap[(*tasks)[i].TaskCode] = &(*tasks)[i]
	}

	for i := range *tasks {
		if (*tasks)[i].ParentTaskCode != "" {
			if parent, ok := taskMap[(*tasks)[i].ParentTaskCode]; ok {
				parent.SubTasks = append(parent.SubTasks, (*tasks)[i])
			}
		}
	}
}

// Create 创建模板
func (r *TemplateRepository) Create(ctx context.Context, template *entity.ProjectTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// Update 更新模板
func (r *TemplateRepository) Update(ctx context.Context, template *entity.ProjectTemplate) error {
	// 只更新模板本身，不级联关联（Tasks/Dependencies）
	return r.db.WithContext(ctx).Omit("Tasks", "Dependencies").Save(template).Error
}

// Delete 删除模板
func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	// 级联删除任务和依赖（数据库外键已设置 ON DELETE CASCADE）
	return r.db.WithContext(ctx).Delete(&entity.ProjectTemplate{}, "id = ?", id).Error
}

// ListTasks 获取模板任务列表
func (r *TemplateRepository) ListTasks(ctx context.Context, templateID string) ([]entity.TemplateTask, error) {
	var tasks []entity.TemplateTask
	err := r.db.WithContext(ctx).
		Where("template_id = ?", templateID).
		Order("sort_order ASC").
		Find(&tasks).Error
	return tasks, err
}

// GetTask 获取单个任务
func (r *TemplateRepository) GetTask(ctx context.Context, templateID, taskCode string) (*entity.TemplateTask, error) {
	var task entity.TemplateTask
	err := r.db.WithContext(ctx).
		Where("template_id = ? AND task_code = ?", templateID, taskCode).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateTask 创建任务
func (r *TemplateRepository) CreateTask(ctx context.Context, task *entity.TemplateTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// UpdateTask 更新任务
func (r *TemplateRepository) UpdateTask(ctx context.Context, task *entity.TemplateTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// DeleteTask 删除任务
func (r *TemplateRepository) DeleteTask(ctx context.Context, templateID, taskCode string) error {
	return r.db.WithContext(ctx).
		Where("template_id = ? AND task_code = ?", templateID, taskCode).
		Delete(&entity.TemplateTask{}).Error
}

// ListDependencies 获取依赖列表
func (r *TemplateRepository) ListDependencies(ctx context.Context, templateID string) ([]entity.TemplateTaskDependency, error) {
	var deps []entity.TemplateTaskDependency
	err := r.db.WithContext(ctx).
		Where("template_id = ?", templateID).
		Find(&deps).Error
	return deps, err
}

// CreateDependency 创建依赖
func (r *TemplateRepository) CreateDependency(ctx context.Context, dep *entity.TemplateTaskDependency) error {
	return r.db.WithContext(ctx).Create(dep).Error
}

// DeleteDependency 删除依赖
func (r *TemplateRepository) DeleteDependency(ctx context.Context, templateID, taskCode, dependsOnCode string) error {
	return r.db.WithContext(ctx).
		Where("template_id = ? AND task_code = ? AND depends_on_task_code = ?", templateID, taskCode, dependsOnCode).
		Delete(&entity.TemplateTaskDependency{}).Error
}
