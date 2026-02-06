package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bitfantasy/nimo-plm/internal/model/entity"
	"github.com/bitfantasy/nimo-plm/internal/repository"
	"github.com/google/uuid"
)

// TemplateService 模板服务
type TemplateService struct {
	templateRepo *repository.TemplateRepository
	projectRepo  *repository.ProjectRepository
}

// NewTemplateService 创建模板服务
func NewTemplateService(templateRepo *repository.TemplateRepository, projectRepo *repository.ProjectRepository) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		projectRepo:  projectRepo,
	}
}

// ListTemplates 获取模板列表
func (s *TemplateService) ListTemplates(ctx context.Context, templateType, productType string, activeOnly bool) ([]entity.ProjectTemplate, error) {
	return s.templateRepo.List(ctx, templateType, productType, activeOnly)
}

// GetTemplate 获取模板详情
func (s *TemplateService) GetTemplate(ctx context.Context, id string) (*entity.ProjectTemplate, error) {
	return s.templateRepo.GetWithTasks(ctx, id)
}

// CreateTemplate 创建模板
func (s *TemplateService) CreateTemplate(ctx context.Context, template *entity.ProjectTemplate) error {
	template.ID = uuid.New().String()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	if template.Phases == nil {
		template.Phases = json.RawMessage(`["CONCEPT","EVT","DVT","PVT","MP"]`)
	}
	return s.templateRepo.Create(ctx, template)
}

// UpdateTemplate 更新模板
func (s *TemplateService) UpdateTemplate(ctx context.Context, template *entity.ProjectTemplate) error {
	template.UpdatedAt = time.Now()
	return s.templateRepo.Update(ctx, template)
}

// DeleteTemplate 删除模板
func (s *TemplateService) DeleteTemplate(ctx context.Context, id string) error {
	// 检查是否系统模板
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if template.TemplateType == "SYSTEM" {
		return fmt.Errorf("cannot delete system template")
	}
	return s.templateRepo.Delete(ctx, id)
}

// DuplicateTemplate 复制模板
func (s *TemplateService) DuplicateTemplate(ctx context.Context, id string, newCode, newName, createdBy string) (*entity.ProjectTemplate, error) {
	// 获取原模板
	original, err := s.templateRepo.GetWithTasks(ctx, id)
	if err != nil {
		return nil, err
	}

	// 创建新模板
	newTemplate := &entity.ProjectTemplate{
		ID:               uuid.New().String(),
		Code:             newCode,
		Name:             newName,
		Description:      original.Description,
		TemplateType:     "CUSTOM",
		ProductType:      original.ProductType,
		Phases:           original.Phases,
		EstimatedDays:    original.EstimatedDays,
		IsActive:         true,
		ParentTemplateID: &original.ID,
		Version:          1,
		CreatedBy:        createdBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.templateRepo.Create(ctx, newTemplate); err != nil {
		return nil, err
	}

	// 复制任务
	for _, task := range original.Tasks {
		newTask := &entity.TemplateTask{
			ID:                  uuid.New().String(),
			TemplateID:          newTemplate.ID,
			TaskCode:            task.TaskCode,
			Name:                task.Name,
			Description:         task.Description,
			Phase:               task.Phase,
			ParentTaskCode:      task.ParentTaskCode,
			TaskType:            task.TaskType,
			DefaultAssigneeRole: task.DefaultAssigneeRole,
			EstimatedDays:       task.EstimatedDays,
			IsCritical:          task.IsCritical,
			Deliverables:        task.Deliverables,
			Checklist:           task.Checklist,
			RequiresApproval:    task.RequiresApproval,
			ApprovalType:        task.ApprovalType,
			SortOrder:           task.SortOrder,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		if err := s.templateRepo.CreateTask(ctx, newTask); err != nil {
			return nil, err
		}
	}

	// 复制依赖
	for _, dep := range original.Dependencies {
		newDep := &entity.TemplateTaskDependency{
			ID:                uuid.New().String(),
			TemplateID:        newTemplate.ID,
			TaskCode:          dep.TaskCode,
			DependsOnTaskCode: dep.DependsOnTaskCode,
			DependencyType:    dep.DependencyType,
			LagDays:           dep.LagDays,
		}
		if err := s.templateRepo.CreateDependency(ctx, newDep); err != nil {
			return nil, err
		}
	}

	return newTemplate, nil
}

// CreateTaskFromTemplate 模板任务操作
func (s *TemplateService) CreateTemplateTask(ctx context.Context, task *entity.TemplateTask) error {
	task.ID = uuid.New().String()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	return s.templateRepo.CreateTask(ctx, task)
}

// UpdateTemplateTask 更新模板任务
func (s *TemplateService) UpdateTemplateTask(ctx context.Context, task *entity.TemplateTask) error {
	task.UpdatedAt = time.Now()
	return s.templateRepo.UpdateTask(ctx, task)
}

// DeleteTemplateTask 删除模板任务
func (s *TemplateService) DeleteTemplateTask(ctx context.Context, templateID, taskCode string) error {
	return s.templateRepo.DeleteTask(ctx, templateID, taskCode)
}

// CreateProjectFromTemplateInput 从模板创建项目的输入
type CreateProjectFromTemplateInput struct {
	TemplateID      string            `json:"template_id"`
	ProjectName     string            `json:"project_name"`
	ProjectCode     string            `json:"project_code"`
	ProductID       string            `json:"product_id"`
	StartDate       time.Time         `json:"start_date"`
	PMID            string            `json:"pm_user_id"`
	SkipWeekends    bool              `json:"skip_weekends"`
	RoleAssignments map[string]string `json:"role_assignments"` // role -> user_id
}

// CreateProjectFromTemplate 从模板创建项目
func (s *TemplateService) CreateProjectFromTemplate(ctx context.Context, input *CreateProjectFromTemplateInput, createdBy string) (*entity.Project, error) {
	// 获取模板
	template, err := s.templateRepo.GetWithTasks(ctx, input.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// 创建项目
	var productID *string
	if input.ProductID != "" {
		productID = &input.ProductID
	}
	
	project := &entity.Project{
		ID:          uuid.New().String()[:32],
		Code:        input.ProjectCode,
		Name:        input.ProjectName,
		ProductID:   productID,
		Phase:       "CONCEPT",
		Status:      "planning",
		StartDate:   &input.StartDate,
		ManagerID:   input.PMID,
		Progress:    0,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	// 构建任务依赖图
	depGraph := buildDependencyGraph(template.Dependencies)
	taskDates := calculateTaskDates(template.Tasks, depGraph, input.StartDate, input.SkipWeekends)

	// 创建任务
	taskMap := make(map[string]string) // task_code -> task_id
	for _, tt := range template.Tasks {
		// 跳过子任务（会在父任务下创建）
		if tt.TaskType == "SUBTASK" {
			continue
		}

		var assigneeID *string
		if tt.DefaultAssigneeRole != "" {
			if userID, ok := input.RoleAssignments[tt.DefaultAssigneeRole]; ok && userID != "" {
				assigneeID = &userID
			}
		}

		dates := taskDates[tt.TaskCode]
		task := &entity.Task{
			ID:           uuid.New().String()[:32],
			ProjectID:    project.ID,
			Code:         tt.TaskCode,
			Title:        tt.Name,
			Description:  tt.Description,
			TaskType:     tt.TaskType,
			Status:       "pending",
			Priority:     "medium",
			AssigneeID:   assigneeID,
			StartDate:    dates.Start,
			DueDate:      dates.End,
			Progress:     0,
			AutoStart:         true,
			RequiresApproval:  tt.RequiresApproval,
			ApprovalType:      tt.ApprovalType,
			CreatedBy:    createdBy,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.projectRepo.CreateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("create task %s: %w", tt.TaskCode, err)
		}
		taskMap[tt.TaskCode] = task.ID
	}

	// 创建子任务
	for _, tt := range template.Tasks {
		if tt.TaskType != "SUBTASK" || tt.ParentTaskCode == "" {
			continue
		}

		parentID, ok := taskMap[tt.ParentTaskCode]
		if !ok {
			continue
		}

		var assigneeID *string
		if tt.DefaultAssigneeRole != "" {
			if userID, ok := input.RoleAssignments[tt.DefaultAssigneeRole]; ok && userID != "" {
				assigneeID = &userID
			}
		}

		dates := taskDates[tt.TaskCode]
		task := &entity.Task{
			ID:           uuid.New().String()[:32],
			ProjectID:    project.ID,
			ParentTaskID: &parentID,
			Code:         tt.TaskCode,
			Title:        tt.Name,
			Description:  tt.Description,
			TaskType:     "SUBTASK",
			Status:       "pending",
			Priority:     "medium",
			AssigneeID:   assigneeID,
			StartDate:    dates.Start,
			DueDate:      dates.End,
			Progress:     0,
			AutoStart:         true,
			RequiresApproval:  tt.RequiresApproval,
			ApprovalType:      tt.ApprovalType,
			CreatedBy:    createdBy,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.projectRepo.CreateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("create subtask %s: %w", tt.TaskCode, err)
		}
		taskMap[tt.TaskCode] = task.ID
	}

	// 创建任务依赖
	for _, dep := range template.Dependencies {
		taskID, ok1 := taskMap[dep.TaskCode]
		depTaskID, ok2 := taskMap[dep.DependsOnTaskCode]
		if !ok1 || !ok2 {
			continue
		}

		taskDep := &entity.TaskDependency{
			ID:             uuid.New().String()[:32],
			TaskID:         taskID,
			DependsOnID:    depTaskID,
			DependencyType: dep.DependencyType,
			LagDays:        dep.LagDays,
		}

		if err := s.projectRepo.CreateTaskDependency(ctx, taskDep); err != nil {
			// 忽略依赖创建错误，继续
			continue
		}
	}

	return project, nil
}

// TaskDates 任务日期
type TaskDates struct {
	Start *time.Time
	End   *time.Time
}

// buildDependencyGraph 构建依赖图
func buildDependencyGraph(deps []entity.TemplateTaskDependency) map[string][]entity.TemplateTaskDependency {
	graph := make(map[string][]entity.TemplateTaskDependency)
	for _, dep := range deps {
		graph[dep.TaskCode] = append(graph[dep.TaskCode], dep)
	}
	return graph
}

// calculateTaskDates 计算任务日期
func calculateTaskDates(tasks []entity.TemplateTask, depGraph map[string][]entity.TemplateTaskDependency, startDate time.Time, skipWeekends bool) map[string]TaskDates {
	dates := make(map[string]TaskDates)
	taskMap := make(map[string]entity.TemplateTask)
	for _, t := range tasks {
		taskMap[t.TaskCode] = t
	}

	// 递归计算每个任务的开始日期
	var calculateStart func(taskCode string) time.Time
	calculateStart = func(taskCode string) time.Time {
		if d, ok := dates[taskCode]; ok && d.Start != nil {
			return *d.Start
		}

		deps := depGraph[taskCode]
		if len(deps) == 0 {
			// 没有依赖，从项目开始日期算
			return startDate
		}

		// 取所有依赖完成后的最大日期
		maxDate := startDate
		for _, dep := range deps {
			depTask, ok := taskMap[dep.DependsOnTaskCode]
			if !ok {
				continue
			}

			depStart := calculateStart(dep.DependsOnTaskCode)
			depEnd := addWorkDays(depStart, depTask.EstimatedDays, skipWeekends)

			switch dep.DependencyType {
			case "FS": // 完成-开始
				candidateStart := addWorkDays(depEnd, dep.LagDays, skipWeekends)
				if candidateStart.After(maxDate) {
					maxDate = candidateStart
				}
			case "SS": // 开始-开始
				candidateStart := addWorkDays(depStart, dep.LagDays, skipWeekends)
				if candidateStart.After(maxDate) {
					maxDate = candidateStart
				}
			default:
				// 默认 FS
				candidateStart := addWorkDays(depEnd, dep.LagDays, skipWeekends)
				if candidateStart.After(maxDate) {
					maxDate = candidateStart
				}
			}
		}

		return maxDate
	}

	// 计算所有任务日期
	for _, task := range tasks {
		start := calculateStart(task.TaskCode)
		end := addWorkDays(start, task.EstimatedDays, skipWeekends)
		dates[task.TaskCode] = TaskDates{Start: &start, End: &end}
	}

	return dates
}

// addWorkDays 添加工作日
func addWorkDays(start time.Time, days int, skipWeekends bool) time.Time {
	if days <= 0 {
		return start
	}

	result := start
	for i := 0; i < days; i++ {
		result = result.AddDate(0, 0, 1)
		if skipWeekends {
			for result.Weekday() == time.Saturday || result.Weekday() == time.Sunday {
				result = result.AddDate(0, 0, 1)
			}
		}
	}
	return result
}
