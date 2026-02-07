package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/plm/repository"
	"github.com/bitfantasy/nimo/internal/shared/engine"
	"github.com/bitfantasy/nimo/internal/shared/feishu"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleAssignment 角色指派信息
type RoleAssignment struct {
	RoleCode     string `json:"role_code"`
	UserID       string `json:"user_id"`
	FeishuUserID string `json:"feishu_user_id"`
}

// WorkflowService 工作流服务 —— 连接状态机引擎和飞书集成
type WorkflowService struct {
	db           *gorm.DB
	engine       *engine.Engine
	feishuClient *feishu.FeishuClient
	projectRepo  *repository.ProjectRepository
	taskRepo     *repository.TaskRepository
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService(db *gorm.DB, eng *engine.Engine, fc *feishu.FeishuClient, projectRepo *repository.ProjectRepository, taskRepo *repository.TaskRepository) *WorkflowService {
	return &WorkflowService{
		db:           db,
		engine:       eng,
		feishuClient: fc,
		projectRepo:  projectRepo,
		taskRepo:     taskRepo,
	}
}

// AssignTask 指派任务
// 把任务状态从 unassigned → pending，记录操作日志，可选创建飞书任务
func (s *WorkflowService) AssignTask(ctx context.Context, projectID, taskID, assigneeID, feishuUserID, operatorID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("查找任务失败: %w", err)
	}
	if task.ProjectID != projectID {
		return fmt.Errorf("任务不属于该项目")
	}

	// 允许从 unassigned 或 pending 指派/重新指派
	if task.Status != entity.TaskStatusUnassigned && task.Status != entity.TaskStatusPending {
		return fmt.Errorf("任务当前状态[%s]不允许指派，需要处于 unassigned 或 pending 状态", task.Status)
	}

	fromStatus := task.Status

	// 更新任务
	task.AssigneeID = &assigneeID
	task.Status = entity.TaskStatusPending
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("更新任务失败: %w", err)
	}

	// 记录操作日志
	s.logAction(ctx, projectID, taskID, entity.TaskActionAssign, fromStatus, entity.TaskStatusPending, operatorID, map[string]interface{}{
		"assignee_id":   assigneeID,
		"feishu_user_id": feishuUserID,
	}, "")

	// 异步创建飞书任务（不阻断主流程）
	if s.feishuClient != nil && task.AutoCreateFeishuTask && feishuUserID != "" {
		go func() {
			bgCtx := context.Background()
			taskGUID, err := s.feishuClient.CreateTask(bgCtx, feishu.CreateTaskReq{
				Summary:     task.Title,
				Description: task.Description,
				Members: []feishu.TaskMember{
					{ID: feishuUserID, Role: "assignee"},
				},
			})
			if err != nil {
				log.Printf("[WorkflowService] 飞书任务创建失败 (task=%s): %v", taskID, err)
				return
			}
			// 保存飞书任务ID到task
			s.db.WithContext(bgCtx).Model(&entity.Task{}).Where("id = ?", taskID).Update("feishu_task_id", taskGUID)
			log.Printf("[WorkflowService] 飞书任务创建成功 task=%s feishu_guid=%s", taskID, taskGUID)
		}()
	}

	// 异步发飞书卡片通知给被指派人
	if s.feishuClient != nil {
		go s.notifyTaskAssigned(context.Background(), task, assigneeID, projectID)
	}

	return nil
}

// StartTask 开始任务
// 检查前置依赖是否完成，状态 pending → in_progress
func (s *WorkflowService) StartTask(ctx context.Context, projectID, taskID, operatorID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("查找任务失败: %w", err)
	}
	if task.ProjectID != projectID {
		return fmt.Errorf("任务不属于该项目")
	}
	if task.Status != entity.TaskStatusPending {
		return fmt.Errorf("任务当前状态[%s]不允许启动，需要处于 pending 状态", task.Status)
	}

	// 检查前置依赖是否全部完成
	if err := s.checkDependenciesCompleted(ctx, taskID); err != nil {
		return err
	}

	// 更新状态
	now := time.Now()
	task.Status = entity.TaskStatusInProgress
	task.ActualStart = &now
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("更新任务失败: %w", err)
	}

	// 记录操作日志
	s.logAction(ctx, projectID, taskID, entity.TaskActionStart, entity.TaskStatusPending, entity.TaskStatusInProgress, operatorID, nil, "")

	return nil
}

// CompleteTask 完成任务
// 如果 requires_approval → reviewing，否则 → completed
func (s *WorkflowService) CompleteTask(ctx context.Context, projectID, taskID, operatorID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("查找任务失败: %w", err)
	}
	if task.ProjectID != projectID {
		return fmt.Errorf("任务不属于该项目")
	}
	if task.Status != entity.TaskStatusInProgress {
		return fmt.Errorf("任务当前状态[%s]不允许完成，需要处于 in_progress 状态", task.Status)
	}

	if task.RequiresApproval {
		// 需要审批 → reviewing
		task.Status = entity.TaskStatusReviewing
		if err := s.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("更新任务失败: %w", err)
		}
		s.logAction(ctx, projectID, taskID, entity.TaskActionSubmitReview, entity.TaskStatusInProgress, entity.TaskStatusReviewing, operatorID, nil, "")
	} else {
		// 不需审批 → completed
		now := time.Now()
		task.Status = entity.TaskStatusCompleted
		task.CompletedAt = &now
		task.Progress = 100
		if err := s.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("更新任务失败: %w", err)
		}
		s.logAction(ctx, projectID, taskID, entity.TaskActionComplete, entity.TaskStatusInProgress, entity.TaskStatusCompleted, operatorID, nil, "")

		// 检查并启动依赖任务
		s.checkAndStartDependentTasks(ctx, projectID, taskID)

		// 异步完成飞书任务
		if s.feishuClient != nil && task.FeishuTaskID != "" {
			go func() {
				if err := s.feishuClient.CompleteTask(context.Background(), task.FeishuTaskID); err != nil {
					log.Printf("[WorkflowService] 飞书任务完成失败 (feishu_task=%s): %v", task.FeishuTaskID, err)
				}
			}()
		}
	}

	return nil
}

// SubmitReview 提交评审结果
func (s *WorkflowService) SubmitReview(ctx context.Context, projectID, taskID, outcomeCode, comment, operatorID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("查找任务失败: %w", err)
	}
	if task.ProjectID != projectID {
		return fmt.Errorf("任务不属于该项目")
	}
	if task.Status != entity.TaskStatusReviewing {
		return fmt.Errorf("任务当前状态[%s]不允许评审，需要处于 reviewing 状态", task.Status)
	}

	// 查找评审结果配置
	var outcome entity.TemplateTaskOutcome
	outcomeFound := false

	// 获取项目模板ID
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err == nil && project.TemplateID != nil {
		if err := s.db.WithContext(ctx).
			Where("template_id = ? AND task_code = ? AND outcome_code = ?", *project.TemplateID, task.Code, outcomeCode).
			First(&outcome).Error; err == nil {
			outcomeFound = true
		}
	}

	// 根据结果类型处理
	if outcomeFound && outcome.OutcomeType == "fail_rollback" {
		// 评审不通过，需要回退
		task.Status = entity.TaskStatusRejected
		if err := s.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("更新任务失败: %w", err)
		}
		s.logAction(ctx, projectID, taskID, entity.TaskActionReject, entity.TaskStatusReviewing, entity.TaskStatusRejected, operatorID, map[string]interface{}{
			"outcome_code": outcomeCode,
		}, comment)

		// 执行回退
		if outcome.RollbackToTaskCode != "" {
			if err := s.RollbackTask(ctx, projectID, taskID, outcome.RollbackToTaskCode, outcome.RollbackCascade, operatorID); err != nil {
				log.Printf("[WorkflowService] 回退失败 (task=%s rollback_to=%s): %v", taskID, outcome.RollbackToTaskCode, err)
			}
		}
	} else if outcomeCode == "reject" || outcomeCode == "rejected" {
		// 简单驳回逻辑（无模板配置时）
		task.Status = entity.TaskStatusInProgress
		if err := s.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("更新任务失败: %w", err)
		}
		s.logAction(ctx, projectID, taskID, entity.TaskActionReject, entity.TaskStatusReviewing, entity.TaskStatusInProgress, operatorID, map[string]interface{}{
			"outcome_code": outcomeCode,
		}, comment)
	} else {
		// 审批通过 → completed
		now := time.Now()
		task.Status = entity.TaskStatusCompleted
		task.CompletedAt = &now
		task.Progress = 100
		if err := s.taskRepo.Update(ctx, task); err != nil {
			return fmt.Errorf("更新任务失败: %w", err)
		}
		s.logAction(ctx, projectID, taskID, entity.TaskActionApprove, entity.TaskStatusReviewing, entity.TaskStatusCompleted, operatorID, map[string]interface{}{
			"outcome_code": outcomeCode,
		}, comment)

		// 检查并启动依赖任务
		s.checkAndStartDependentTasks(ctx, projectID, taskID)

		// 异步完成飞书任务
		if s.feishuClient != nil && task.FeishuTaskID != "" {
			go func() {
				if err := s.feishuClient.CompleteTask(context.Background(), task.FeishuTaskID); err != nil {
					log.Printf("[WorkflowService] 飞书任务完成失败 (feishu_task=%s): %v", task.FeishuTaskID, err)
				}
			}()
		}
	}

	return nil
}

// RollbackTask 回退任务
func (s *WorkflowService) RollbackTask(ctx context.Context, projectID, taskID, rollbackToTaskCode string, cascade bool, operatorID string) error {
	// 查找目标任务
	var targetTask entity.Task
	if err := s.db.WithContext(ctx).
		Where("project_id = ? AND code = ?", projectID, rollbackToTaskCode).
		First(&targetTask).Error; err != nil {
		return fmt.Errorf("查找回退目标任务失败 (code=%s): %w", rollbackToTaskCode, err)
	}

	// 重置目标任务为 in_progress
	fromStatus := targetTask.Status
	targetTask.Status = entity.TaskStatusInProgress
	targetTask.CompletedAt = nil
	targetTask.Progress = 0
	if err := s.db.WithContext(ctx).Save(&targetTask).Error; err != nil {
		return fmt.Errorf("重置目标任务失败: %w", err)
	}
	s.logAction(ctx, projectID, targetTask.ID, entity.TaskActionRollback, fromStatus, entity.TaskStatusInProgress, operatorID, map[string]interface{}{
		"triggered_by_task": taskID,
		"cascade":           cascade,
	}, "")

	if cascade {
		// 获取目标任务所在阶段的后续任务
		var subsequentTasks []entity.Task
		if err := s.db.WithContext(ctx).
			Where("project_id = ? AND phase_id = ? AND sequence > ? AND id != ?",
				projectID, targetTask.PhaseID, targetTask.Sequence, targetTask.ID).
			Find(&subsequentTasks).Error; err != nil {
			log.Printf("[WorkflowService] 查找后续任务失败: %v", err)
			return nil
		}

		for _, t := range subsequentTasks {
			if t.Status == entity.TaskStatusCompleted || t.Status == entity.TaskStatusInProgress || t.Status == entity.TaskStatusReviewing {
				oldStatus := t.Status
				t.Status = entity.TaskStatusPending
				t.CompletedAt = nil
				t.Progress = 0
				if err := s.db.WithContext(ctx).Save(&t).Error; err != nil {
					log.Printf("[WorkflowService] 重置后续任务失败 (task=%s): %v", t.ID, err)
					continue
				}
				s.logAction(ctx, projectID, t.ID, entity.TaskActionRollback, oldStatus, entity.TaskStatusPending, operatorID, map[string]interface{}{
					"triggered_by_task": taskID,
					"cascade":           true,
				}, "级联回退")
			}
		}
	}

	return nil
}

// AssignPhaseRoles 指派阶段角色
func (s *WorkflowService) AssignPhaseRoles(ctx context.Context, projectID, phase string, assignments []RoleAssignment, operatorID string) error {
	for _, a := range assignments {
		// 保存角色指派（upsert）
		assignment := entity.ProjectRoleAssignment{
			ID:           uuid.New().String(),
			ProjectID:    projectID,
			Phase:        phase,
			RoleCode:     a.RoleCode,
			UserID:       a.UserID,
			FeishuUserID: a.FeishuUserID,
			AssignedBy:   operatorID,
			AssignedAt:   time.Now(),
		}

		result := s.db.WithContext(ctx).
			Where("project_id = ? AND phase = ? AND role_code = ?", projectID, phase, a.RoleCode).
			Assign(map[string]interface{}{
				"user_id":        a.UserID,
				"feishu_user_id": a.FeishuUserID,
				"assigned_by":    operatorID,
				"assigned_at":    time.Now(),
			}).
			FirstOrCreate(&assignment)
		if result.Error != nil {
			return fmt.Errorf("保存角色指派失败 (role=%s): %w", a.RoleCode, result.Error)
		}

		// 查找该阶段中默认角色匹配的未指派任务
		var unassignedTasks []entity.Task
		s.db.WithContext(ctx).
			Joins("JOIN project_phases ON tasks.phase_id = project_phases.id").
			Where("tasks.project_id = ? AND project_phases.phase = ? AND (tasks.assignee_id IS NULL OR tasks.assignee_id = '') AND tasks.status = ?",
				projectID, phase, entity.TaskStatusUnassigned).
			Find(&unassignedTasks)

		// 还需要检查模板任务的 default_assignee_role
		for _, task := range unassignedTasks {
			// 查模板任务的默认角色
			var templateTask entity.TemplateTask
			if err := s.db.WithContext(ctx).
				Where("task_code = ? AND default_assignee_role = ?", task.Code, a.RoleCode).
				First(&templateTask).Error; err != nil {
				continue // 不匹配则跳过
			}

			// 指派任务
			if err := s.AssignTask(ctx, projectID, task.ID, a.UserID, a.FeishuUserID, operatorID); err != nil {
				log.Printf("[WorkflowService] 自动指派任务失败 (task=%s role=%s): %v", task.ID, a.RoleCode, err)
			}
		}
	}

	return nil
}

// GetTaskHistory 获取任务操作历史
func (s *WorkflowService) GetTaskHistory(ctx context.Context, projectID, taskID string) ([]entity.TaskActionLog, error) {
	var logs []entity.TaskActionLog
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND task_id = ?", projectID, taskID).
		Order("created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, fmt.Errorf("查询操作历史失败: %w", err)
	}
	return logs, nil
}

// checkDependenciesCompleted 检查任务的所有前置依赖是否已完成
func (s *WorkflowService) checkDependenciesCompleted(ctx context.Context, taskID string) error {
	var deps []entity.TaskDependency
	if err := s.db.WithContext(ctx).Where("task_id = ?", taskID).Find(&deps).Error; err != nil {
		return fmt.Errorf("查询任务依赖失败: %w", err)
	}

	for _, dep := range deps {
		var depTask entity.Task
		if err := s.db.WithContext(ctx).Where("id = ?", dep.DependsOnID).First(&depTask).Error; err != nil {
			return fmt.Errorf("查找依赖任务失败 (id=%s): %w", dep.DependsOnID, err)
		}
		if depTask.Status != entity.TaskStatusCompleted {
			return fmt.Errorf("前置任务[%s]尚未完成（当前状态: %s），无法启动", depTask.Title, depTask.Status)
		}
	}

	return nil
}

// checkAndStartDependentTasks 检查并启动依赖当前任务的后续任务
func (s *WorkflowService) checkAndStartDependentTasks(ctx context.Context, projectID, completedTaskID string) {
	// 查找依赖于已完成任务的任务
	var deps []entity.TaskDependency
	if err := s.db.WithContext(ctx).Where("depends_on_task_id = ?", completedTaskID).Find(&deps).Error; err != nil {
		log.Printf("[WorkflowService] 查找依赖任务失败: %v", err)
		return
	}

	for _, dep := range deps {
		// 获取依赖任务
		var task entity.Task
		if err := s.db.WithContext(ctx).Where("id = ?", dep.TaskID).First(&task).Error; err != nil {
			log.Printf("[WorkflowService] 查找任务失败 (id=%s): %v", dep.TaskID, err)
			continue
		}

		// 只处理 pending 状态的任务
		if task.Status != entity.TaskStatusPending {
			continue
		}

		// 检查该任务的所有依赖是否都已完成
		allCompleted := true
		var allDeps []entity.TaskDependency
		if err := s.db.WithContext(ctx).Where("task_id = ?", task.ID).Find(&allDeps).Error; err != nil {
			log.Printf("[WorkflowService] 查找任务所有依赖失败: %v", err)
			continue
		}

		for _, d := range allDeps {
			var depTask entity.Task
			if err := s.db.WithContext(ctx).Where("id = ?", d.DependsOnID).First(&depTask).Error; err != nil {
				allCompleted = false
				break
			}
			if depTask.Status != entity.TaskStatusCompleted {
				allCompleted = false
				break
			}
		}

		if allCompleted {
			// 自动启动任务
			now := time.Now()
			task.Status = entity.TaskStatusInProgress
			task.ActualStart = &now
			if err := s.db.WithContext(ctx).Save(&task).Error; err != nil {
				log.Printf("[WorkflowService] 自动启动任务失败 (task=%s): %v", task.ID, err)
				continue
			}
			s.logAction(ctx, projectID, task.ID, entity.TaskActionStart, entity.TaskStatusPending, entity.TaskStatusInProgress, "system", map[string]interface{}{
				"auto_started":       true,
				"completed_dep_task": completedTaskID,
			}, "依赖任务完成，自动启动")
			log.Printf("[WorkflowService] 自动启动任务 task=%s (依赖任务 %s 完成)", task.ID, completedTaskID)
		}
	}
}

// =============================================================================
// 飞书通知辅助方法
// =============================================================================

// notifyTaskAssigned 通知任务被指派
func (s *WorkflowService) notifyTaskAssigned(ctx context.Context, task *entity.Task, assigneeID, projectID string) {
	// 查找被指派人
	var assignee entity.User
	if err := s.db.WithContext(ctx).Where("id = ?", assigneeID).First(&assignee).Error; err != nil {
		log.Printf("[WorkflowNotify] 查找被指派人失败 (user_id=%s): %v", assigneeID, err)
		return
	}
	if assignee.FeishuOpenID == "" {
		log.Printf("[WorkflowNotify] 被指派人[%s]没有飞书 open_id，跳过通知", assignee.Name)
		return
	}

	// 查找项目名
	projectName := "未知项目"
	var project entity.Project
	if err := s.db.WithContext(ctx).Where("id = ?", projectID).First(&project).Error; err == nil {
		projectName = project.Name
	}

	dueDate := "无"
	if task.DueDate != nil {
		dueDate = task.DueDate.Format("2006-01-02")
	}

	card := feishu.NewTaskAssignmentCard(task.Title, projectName, assignee.Name, dueDate)
	if err := s.feishuClient.SendUserCard(ctx, assignee.FeishuOpenID, card); err != nil {
		log.Printf("[WorkflowNotify] 发送任务指派通知给[%s]失败: %v", assignee.Name, err)
	} else {
		log.Printf("[WorkflowNotify] 已通知[%s]任务指派: %s", assignee.Name, task.Title)
	}
}

// notifyTaskStatusChange 通知任务状态变更
func (s *WorkflowService) notifyTaskStatusChange(ctx context.Context, task *entity.Task, fromStatus, toStatus, projectID string) {
	if task.AssigneeID == nil || *task.AssigneeID == "" {
		return
	}

	var assignee entity.User
	if err := s.db.WithContext(ctx).Where("id = ?", *task.AssigneeID).First(&assignee).Error; err != nil {
		return
	}
	if assignee.FeishuOpenID == "" {
		return
	}

	projectName := "未知项目"
	var project entity.Project
	if err := s.db.WithContext(ctx).Where("id = ?", projectID).First(&project).Error; err == nil {
		projectName = project.Name
	}

	card := feishu.InteractiveCard{
		Config: &feishu.CardConfig{WideScreenMode: true},
		Header: &feishu.CardHeader{
			Title:    feishu.CardText{Tag: "plain_text", Content: "📝 任务状态变更"},
			Template: "blue",
		},
		Elements: []feishu.CardElement{
			{
				Tag: "div",
				Fields: []feishu.CardField{
					{IsShort: true, Text: feishu.CardText{Tag: "lark_md", Content: fmt.Sprintf("**任务**\n%s", task.Title)}},
					{IsShort: true, Text: feishu.CardText{Tag: "lark_md", Content: fmt.Sprintf("**项目**\n%s", projectName)}},
					{IsShort: true, Text: feishu.CardText{Tag: "lark_md", Content: fmt.Sprintf("**状态变更**\n%s → %s", fromStatus, toStatus)}},
				},
			},
		},
	}

	if err := s.feishuClient.SendUserCard(ctx, assignee.FeishuOpenID, card); err != nil {
		log.Printf("[WorkflowNotify] 发送状态变更通知失败: %v", err)
	}
}

// logAction 记录任务操作日志
func (s *WorkflowService) logAction(ctx context.Context, projectID, taskID, action, fromStatus, toStatus, operatorID string, eventData map[string]interface{}, comment string) {
	actionLog := entity.TaskActionLog{
		ID:           uuid.New().String(),
		ProjectID:    projectID,
		TaskID:       taskID,
		Action:       action,
		FromStatus:   fromStatus,
		ToStatus:     toStatus,
		OperatorID:   operatorID,
		OperatorType: "user",
		Comment:      comment,
	}

	if operatorID == "system" {
		actionLog.OperatorType = "system"
	}

	if eventData != nil {
		actionLog.EventData = entity.JSONB(eventData)
	}

	if err := s.db.WithContext(ctx).Create(&actionLog).Error; err != nil {
		log.Printf("[WorkflowService] 记录操作日志失败: %v", err)
	}
}
