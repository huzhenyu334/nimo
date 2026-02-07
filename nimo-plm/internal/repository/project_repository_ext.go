package repository

import (
	"context"

	"github.com/bitfantasy/nimo-plm/internal/model/entity"
)

// CreateTask 创建任务
// Omit association fields to prevent GORM from nullifying foreign keys
// when the associated struct pointer (e.g. ParentTask) is nil.
func (r *ProjectRepository) CreateTask(ctx context.Context, task *entity.Task) error {
	return r.db.WithContext(ctx).
		Omit("ParentTask", "SubTasks", "Project", "Phase", "Assignee", "Reviewer", "Creator").
		Create(task).Error
}

// GetTask 获取任务
func (r *ProjectRepository) GetTask(ctx context.Context, taskID string) (*entity.Task, error) {
	var task entity.Task
	err := r.db.WithContext(ctx).First(&task, "id = ?", taskID).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask 更新任务
func (r *ProjectRepository) UpdateTask(ctx context.Context, task *entity.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// UpdateTaskStatus 更新任务状态
func (r *ProjectRepository) UpdateTaskStatus(ctx context.Context, taskID string, status string) error {
	return r.db.WithContext(ctx).Model(&entity.Task{}).Where("id = ?", taskID).Update("status", status).Error
}

// GetSubTasks 获取子任务
func (r *ProjectRepository) GetSubTasks(ctx context.Context, parentTaskID string) ([]entity.Task, error) {
	var tasks []entity.Task
	err := r.db.WithContext(ctx).Where("parent_task_id = ?", parentTaskID).Find(&tasks).Error
	return tasks, err
}

// CreateTaskDependency 创建任务依赖
func (r *ProjectRepository) CreateTaskDependency(ctx context.Context, dep *entity.TaskDependency) error {
	return r.db.WithContext(ctx).Create(dep).Error
}

// GetTaskDependencies 获取任务的依赖
func (r *ProjectRepository) GetTaskDependencies(ctx context.Context, taskID string) ([]entity.TaskDependency, error) {
	var deps []entity.TaskDependency
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Find(&deps).Error
	return deps, err
}

// GetDependentTasks 获取依赖于指定任务的所有任务
func (r *ProjectRepository) GetDependentTasks(ctx context.Context, taskID string) ([]entity.Task, error) {
	var tasks []entity.Task
	err := r.db.WithContext(ctx).
		Joins("JOIN task_dependencies ON tasks.id = task_dependencies.task_id").
		Where("task_dependencies.depends_on_task_id = ?", taskID).
		Find(&tasks).Error
	return tasks, err
}

// GetPhaseMilestoneTasks 获取阶段的里程碑任务
func (r *ProjectRepository) GetPhaseMilestoneTasks(ctx context.Context, projectID string, phase string) ([]entity.Task, error) {
	var tasks []entity.Task
	err := r.db.WithContext(ctx).
		Joins("JOIN project_phases ON tasks.phase_id = project_phases.id").
		Where("project_phases.project_id = ? AND project_phases.phase = ? AND tasks.task_type = ?", projectID, phase, "milestone").
		Find(&tasks).Error
	return tasks, err
}

// CreateAutomationLog 创建自动化日志
func (r *ProjectRepository) CreateAutomationLog(ctx context.Context, log *entity.AutomationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// Note: GetByID, Create, Update methods are in project_repository.go
