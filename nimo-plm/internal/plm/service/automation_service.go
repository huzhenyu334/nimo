package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/bitfantasy/nimo/internal/plm/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AutomationService 自动化服务
type AutomationService struct {
	projectRepo *repository.ProjectRepository
	rdb         *redis.Client
	feishuSvc   *FeishuIntegrationService
	logger      *zap.Logger
}

// NewAutomationService 创建自动化服务
func NewAutomationService(projectRepo *repository.ProjectRepository, rdb *redis.Client, feishuSvc *FeishuIntegrationService, logger *zap.Logger) *AutomationService {
	return &AutomationService{
		projectRepo: projectRepo,
		rdb:         rdb,
		feishuSvc:   feishuSvc,
		logger:      logger,
	}
}

// OnTaskCompleted 任务完成时触发
func (s *AutomationService) OnTaskCompleted(ctx context.Context, taskID string) error {
	s.logger.Info("Task completed, triggering automation", zap.String("task_id", taskID))

	// 获取任务信息
	task, err := s.projectRepo.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	// 1. 检查依赖此任务的后续任务
	dependentTasks, err := s.projectRepo.GetDependentTasks(ctx, taskID)
	if err != nil {
		s.logger.Error("Failed to get dependent tasks", zap.Error(err))
	} else {
		for _, depTask := range dependentTasks {
			// 检查该任务的所有前置依赖是否满足
			allMet, err := s.checkAllDependenciesMet(ctx, depTask.ID)
			if err != nil {
				s.logger.Error("Failed to check dependencies", zap.String("task_id", depTask.ID), zap.Error(err))
				continue
			}

			if allMet {
				// 更新状态为就绪或进行中
				newStatus := "ready"
				if depTask.AutoStart {
					newStatus = "in_progress"
				}

				if err := s.projectRepo.UpdateTaskStatus(ctx, depTask.ID, newStatus); err != nil {
					s.logger.Error("Failed to update task status", zap.String("task_id", depTask.ID), zap.Error(err))
					continue
				}

				s.logger.Info("Task auto-started", zap.String("task_id", depTask.ID), zap.String("status", newStatus))

				// 发送通知
				if depTask.AssigneeID != nil && *depTask.AssigneeID != "" {
					s.sendTaskNotification(ctx, &depTask, "TASK_READY")
				}
			}
		}
	}

	// 2. 检查是否所有子任务完成，自动完成父任务
	if task.ParentTaskID != nil && *task.ParentTaskID != "" {
		allSubTasksComplete, err := s.checkAllSubTasksComplete(ctx, *task.ParentTaskID)
		if err != nil {
			s.logger.Error("Failed to check subtasks", zap.Error(err))
		} else if allSubTasksComplete {
			if err := s.projectRepo.UpdateTaskStatus(ctx, *task.ParentTaskID, "completed"); err != nil {
				s.logger.Error("Failed to complete parent task", zap.Error(err))
			} else {
				s.logger.Info("Parent task auto-completed", zap.String("parent_task_id", *task.ParentTaskID))
				// 递归触发父任务完成
				s.OnTaskCompleted(ctx, *task.ParentTaskID)
			}
		}
	}

	// 3. 检查阶段是否完成
	if err := s.checkPhaseCompletion(ctx, task.ProjectID); err != nil {
		s.logger.Error("Failed to check phase completion", zap.Error(err))
	}

	return nil
}

// checkAllDependenciesMet 检查所有依赖是否满足
func (s *AutomationService) checkAllDependenciesMet(ctx context.Context, taskID string) (bool, error) {
	dependencies, err := s.projectRepo.GetTaskDependencies(ctx, taskID)
	if err != nil {
		return false, err
	}

	if len(dependencies) == 0 {
		return true, nil
	}

	for _, dep := range dependencies {
		depTask, err := s.projectRepo.GetTask(ctx, dep.DependsOnID)
		if err != nil {
			return false, err
		}

		switch dep.DependencyType {
		case "FS": // 完成-开始
			if depTask.Status != "completed" && depTask.Status != "approved" {
				return false, nil
			}
		case "SS": // 开始-开始
			if depTask.Status == "pending" || depTask.Status == "ready" {
				return false, nil
			}
		case "FF": // 完成-完成
			// 允许开始，但完成需要等依赖完成
		case "SF": // 开始-完成
			if depTask.Status == "pending" || depTask.Status == "ready" {
				return false, nil
			}
		default:
			// 默认 FS
			if depTask.Status != "completed" && depTask.Status != "approved" {
				return false, nil
			}
		}
	}

	return true, nil
}

// checkAllSubTasksComplete 检查所有子任务是否完成
func (s *AutomationService) checkAllSubTasksComplete(ctx context.Context, parentTaskID string) (bool, error) {
	subTasks, err := s.projectRepo.GetSubTasks(ctx, parentTaskID)
	if err != nil {
		return false, err
	}

	if len(subTasks) == 0 {
		return false, nil // 没有子任务，不自动完成
	}

	for _, st := range subTasks {
		if st.Status != "completed" && st.Status != "approved" && st.Status != "cancelled" {
			return false, nil
		}
	}

	return true, nil
}

// checkPhaseCompletion 检查阶段是否完成
func (s *AutomationService) checkPhaseCompletion(ctx context.Context, projectID string) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}

	// 获取当前阶段的所有里程碑任务
	milestoneTasks, err := s.projectRepo.GetPhaseMilestoneTasks(ctx, projectID, project.Phase)
	if err != nil {
		return err
	}

	allMilestonesComplete := true
	for _, mt := range milestoneTasks {
		if mt.Status != "completed" && mt.Status != "approved" {
			allMilestonesComplete = false
			break
		}
	}

	if allMilestonesComplete && len(milestoneTasks) > 0 {
		s.logger.Info("Phase completed", zap.String("project_id", projectID), zap.String("phase", project.Phase))
		// 可以触发阶段评审创建等后续动作
	}

	return nil
}

// sendTaskNotification 发送任务通知
func (s *AutomationService) sendTaskNotification(ctx context.Context, task *entity.Task, notificationType string) {
	if s.feishuSvc == nil || task.AssigneeID == nil || *task.AssigneeID == "" {
		return
	}

	var content string
	switch notificationType {
	case "TASK_READY":
		content = fmt.Sprintf("📋 任务就绪通知\n\n任务【%s】的前置任务已完成，现在可以开始工作了。", task.Title)
	case "TASK_OVERDUE":
		content = fmt.Sprintf("⚠️ 任务逾期提醒\n\n任务【%s】已逾期，请尽快处理。", task.Title)
	default:
		content = fmt.Sprintf("📌 任务通知\n\n任务【%s】状态已更新。", task.Title)
	}

	// 发送飞书消息
	if err := s.feishuSvc.SendMessage(ctx, *task.AssigneeID, content); err != nil {
		s.logger.Error("Failed to send notification", zap.Error(err))
	}
}

// StartTask 开始任务
func (s *AutomationService) StartTask(ctx context.Context, taskID string) error {
	// 检查依赖是否满足
	allMet, err := s.checkAllDependenciesMet(ctx, taskID)
	if err != nil {
		return fmt.Errorf("check dependencies: %w", err)
	}

	if !allMet {
		return fmt.Errorf("task dependencies not met")
	}

	// 更新状态
	if err := s.projectRepo.UpdateTaskStatus(ctx, taskID, "in_progress"); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

// CompleteTask 完成任务
func (s *AutomationService) CompleteTask(ctx context.Context, taskID string, userID string) error {
	task, err := s.projectRepo.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	// 检查是否需要审批
	if task.RequiresApproval {
		// 更新状态为待审批
		if err := s.projectRepo.UpdateTaskStatus(ctx, taskID, "needs_review"); err != nil {
			return fmt.Errorf("update status: %w", err)
		}
		return nil
	}

	// 直接完成
	now := time.Now()
	task.Status = "completed"
	task.CompletedAt = &now
	task.Progress = 100

	if err := s.projectRepo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// 触发自动化
	go s.OnTaskCompleted(context.Background(), taskID)

	return nil
}

// ApproveTask 审批通过任务
func (s *AutomationService) ApproveTask(ctx context.Context, taskID string, approverID string, comment string) error {
	task, err := s.projectRepo.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	if task.Status != "needs_review" {
		return fmt.Errorf("task is not in review status")
	}

	now := time.Now()
	task.Status = "completed"
	task.CompletedAt = &now
	task.Progress = 100
	task.ApprovalStatus = "APPROVED"

	if err := s.projectRepo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// 记录审批
	s.logApproval(ctx, taskID, approverID, "APPROVED", comment)

	// 触发自动化
	go s.OnTaskCompleted(context.Background(), taskID)

	return nil
}

// RejectTask 审批驳回任务
func (s *AutomationService) RejectTask(ctx context.Context, taskID string, approverID string, comment string) error {
	task, err := s.projectRepo.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	if task.Status != "needs_review" {
		return fmt.Errorf("task is not in review status")
	}

	task.Status = "in_progress"
	task.ApprovalStatus = "REJECTED"

	if err := s.projectRepo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// 记录审批
	s.logApproval(ctx, taskID, approverID, "REJECTED", comment)

	// 通知负责人
	if task.AssigneeID != nil && *task.AssigneeID != "" {
		s.sendTaskNotification(ctx, task, "TASK_REJECTED")
	}

	return nil
}

func (s *AutomationService) logApproval(ctx context.Context, taskID, approverID, status, comment string) {
	log := &entity.AutomationLog{
		ID: uuid.New().String(),
		TriggerEvent: json.RawMessage(fmt.Sprintf(`{"type":"APPROVAL","task_id":"%s","approver":"%s"}`, taskID, approverID)),
		ActionResult: json.RawMessage(fmt.Sprintf(`{"status":"%s","comment":"%s"}`, status, comment)),
		Status:       "SUCCESS",
		ExecutedAt:   time.Now(),
	}
	// 存储日志（忽略错误）
	_ = s.projectRepo.CreateAutomationLog(ctx, log)
}
