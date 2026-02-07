# Phase 3: PLM Workflow Integration Spec

## Overview
Integrate the state machine engine (internal/shared/engine/) and Feishu client (internal/shared/feishu/) into PLM business logic by creating workflow entities, services, handlers, and routes.

## Files to Create

### 1. internal/plm/entity/workflow.go

Define 4 GORM structs + constants. Use `entity.JSONB` type (already defined in product.go as `type JSONB map[string]interface{}`).

```go
package entity

import "time"

// Workflow task status constants (extend existing ones in project.go)
const (
	TaskStatusUnassigned = "unassigned"
	// TaskStatusPending already exists in project.go
	// TaskStatusInProgress already exists in project.go 
	TaskStatusReviewing  = "reviewing"
	// TaskStatusCompleted already exists in project.go
	TaskStatusRejected   = "rejected"
)

// Task action type constants
const (
	TaskActionAssign       = "assign"
	TaskActionStart        = "start"
	TaskActionComplete     = "complete"
	TaskActionSubmitReview = "submit_review"
	TaskActionApprove      = "approve"
	TaskActionReject       = "reject"
	TaskActionRollback     = "rollback"
)

// TemplatePhaseRole 阶段角色配置（模板级别）
type TemplatePhaseRole struct {
	ID              string `json:"id" gorm:"primaryKey;size:36"`
	TemplateID      string `json:"template_id" gorm:"size:36;not null"`
	Phase           string `json:"phase" gorm:"size:20;not null"`
	RoleCode        string `json:"role_code" gorm:"size:50;not null"`
	RoleName        string `json:"role_name" gorm:"size:100;not null"`
	IsRequired      bool   `json:"is_required" gorm:"default:true"`
	TriggerTaskCode string `json:"trigger_task_code" gorm:"size:50"`
}

func (TemplatePhaseRole) TableName() string { return "template_phase_roles" }

// TemplateTaskOutcome 模板任务评审结果选项
type TemplateTaskOutcome struct {
	ID                  string `json:"id" gorm:"primaryKey;size:36"`
	TemplateID          string `json:"template_id" gorm:"size:36;not null"`
	TaskCode            string `json:"task_code" gorm:"size:50;not null"`
	OutcomeCode         string `json:"outcome_code" gorm:"size:50;not null"`
	OutcomeName         string `json:"outcome_name" gorm:"size:100;not null"`
	OutcomeType         string `json:"outcome_type" gorm:"size:20;not null;default:'pass'"` // pass / fail_rollback
	RollbackToTaskCode  string `json:"rollback_to_task_code" gorm:"size:50"`
	RollbackCascade     bool   `json:"rollback_cascade" gorm:"default:false"`
	SortOrder           int    `json:"sort_order" gorm:"default:0"`
}

func (TemplateTaskOutcome) TableName() string { return "template_task_outcomes" }

// ProjectRoleAssignment 项目角色指派（项目级别）
type ProjectRoleAssignment struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	ProjectID    string    `json:"project_id" gorm:"size:32;not null"`
	Phase        string    `json:"phase" gorm:"size:20;not null"`
	RoleCode     string    `json:"role_code" gorm:"size:50;not null"`
	UserID       string    `json:"user_id" gorm:"size:32;not null"`
	FeishuUserID string    `json:"feishu_user_id" gorm:"size:64"`
	AssignedBy   string    `json:"assigned_by" gorm:"size:32;not null"`
	AssignedAt   time.Time `json:"assigned_at" gorm:"autoCreateTime"`
}

func (ProjectRoleAssignment) TableName() string { return "project_role_assignments" }

// TaskActionLog 任务操作历史
type TaskActionLog struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	ProjectID    string    `json:"project_id" gorm:"size:32;not null"`
	TaskID       string    `json:"task_id" gorm:"size:32;not null"`
	Action       string    `json:"action" gorm:"size:50;not null"`
	FromStatus   string    `json:"from_status" gorm:"size:20"`
	ToStatus     string    `json:"to_status" gorm:"size:20;not null"`
	OperatorID   string    `json:"operator_id" gorm:"size:64;not null"`
	OperatorType string    `json:"operator_type" gorm:"size:20;default:'user'"`
	EventData    JSONB     `json:"event_data" gorm:"type:jsonb"`
	Comment      string    `json:"comment" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (TaskActionLog) TableName() string { return "task_action_logs" }
```

**IMPORTANT**: project.go already defines `TaskStatusPending = "pending"`, `TaskStatusInProgress = "in_progress"`, `TaskStatusCompleted = "completed"`. Do NOT redefine those. Only add `TaskStatusUnassigned`, `TaskStatusReviewing`, `TaskStatusRejected` as new constants.

### 2. internal/plm/service/workflow_service.go

```go
package service

// WorkflowService connects the state machine engine with Feishu integration
type WorkflowService struct {
    db           *gorm.DB
    engine       *engine.Engine
    feishuClient *feishu.FeishuClient  // may be nil if Feishu is not configured
    projectRepo  *repository.ProjectRepository
    taskRepo     *repository.TaskRepository
}
```

Core methods:

1. **AssignTask(projectID, taskID, assigneeID, feishuUserID, operatorID string) error**
   - Load task from DB, verify status is "unassigned" or "pending" (for reassignment)
   - Update task: set AssigneeID, Status = "pending"
   - Save task_action_log
   - If task.AutoCreateFeishuTask && feishuClient != nil: create feishu task (log error but don't block)
   
2. **StartTask(projectID, taskID, operatorID string) error**
   - Load task, verify status = "pending"
   - Check all dependencies completed (query task_dependencies where task_id=taskID, check depends_on tasks are all completed)
   - Update status to "in_progress", set actual_start=now
   - Save task_action_log

3. **CompleteTask(projectID, taskID, operatorID string) error**
   - Load task, verify status = "in_progress"
   - If task.RequiresApproval: status → "reviewing"
   - Else: status → "completed", set completed_at=now, progress=100
   - Save task_action_log
   - If completed: call checkAndStartDependentTasks

4. **SubmitReview(projectID, taskID, outcomeCode, comment, operatorID string) error**
   - Load task, verify status = "reviewing"
   - Look up TemplateTaskOutcome by template_id + task code + outcome_code
   - If outcome not found, use simple logic: outcomeCode=="approve" → pass, else reject
   - If pass: status → "completed", call checkAndStartDependentTasks
   - If fail_rollback: call RollbackTask with the configured rollback target
   - Save task_action_log

5. **RollbackTask(projectID, taskID, rollbackToTaskCode string, cascade bool, operatorID string) error**
   - Find target task by project_id + code=rollbackToTaskCode
   - Reset target task status to "in_progress"
   - If cascade: find all tasks with sort_order > target's sort_order in same phase, reset to "pending"
   - Save task_action_log for each reset task

6. **AssignPhaseRoles(projectID, phase string, assignments []RoleAssignment, operatorID string) error**
   - Save each assignment to project_role_assignments (upsert: ON CONFLICT(project_id, phase, role_code) DO UPDATE)
   - For each role, find unassigned tasks in that phase where default_assignee_role matches role_code
   - Assign those tasks to the role's user

7. **GetTaskHistory(projectID, taskID string) ([]TaskActionLog, error)**
   - Query task_action_logs WHERE project_id AND task_id ORDER BY created_at DESC

8. **checkAndStartDependentTasks(projectID, completedTaskID string)** (private)
   - Find task_dependencies WHERE depends_on_task_id = completedTaskID
   - For each dependent task, check if ALL its dependencies are completed
   - If all completed: update status to "pending" (or "in_progress" if task.AutoStart)

**CRITICAL**: All Feishu calls must be wrapped in error handling that logs but doesn't return error. Pattern:
```go
if s.feishuClient != nil && task.AutoCreateFeishuTask {
    go func() {
        if _, err := s.feishuClient.CreateTask(ctx, req); err != nil {
            log.Printf("[WorkflowService] 飞书任务创建失败: %v", err)
        }
    }()
}
```

### 3. internal/plm/handler/workflow_handler.go

```go
package handler

type WorkflowHandler struct {
    svc *service.WorkflowService
}

func NewWorkflowHandler(svc *service.WorkflowService) *WorkflowHandler {
    return &WorkflowHandler{svc: svc}
}
```

Methods: AssignTask, StartTask, CompleteTask, SubmitReview, AssignPhaseRoles, GetTaskHistory

Each extracts params from gin.Context (path params `:id` for projectId, `:taskId`, `:phase`), gets operatorID from JWT context via `GetUserID(c)`, binds request body, calls service, returns JSON response using the existing Success/Error helpers from handler.go.

### 4. Modify existing files

#### cmd/plm/main.go changes:

1. Add 4 new CREATE TABLE SQL statements to the migrationSQL slice (after the existing Phase 1 tables):
```sql
CREATE TABLE IF NOT EXISTS template_phase_roles (...)
CREATE TABLE IF NOT EXISTS template_task_outcomes (...)
CREATE TABLE IF NOT EXISTS project_role_assignments (...)
CREATE TABLE IF NOT EXISTS task_action_logs (...)
```

2. After `services := service.NewServices(...)`, initialize:
```go
// Initialize state machine engine
stateEngine := engine.NewEngine(db, nil)
plmTaskMachine := engine.NewPLMTaskMachine()
if err := stateEngine.RegisterMachine(plmTaskMachine); err != nil {
    zapLogger.Warn("Failed to register PLM task state machine", zap.Error(err))
}

// Initialize Feishu client for workflow
var feishuWorkflowClient *feishu.FeishuClient
feishuAppID := cfg.Feishu.AppID
feishuAppSecret := cfg.Feishu.AppSecret
if envID := os.Getenv("FEISHU_APP_ID"); envID != "" {
    feishuAppID = envID
}
if envSecret := os.Getenv("FEISHU_APP_SECRET"); envSecret != "" {
    feishuAppSecret = envSecret
}
if feishuAppID != "" && feishuAppSecret != "" {
    feishuWorkflowClient = feishu.NewClient(feishuAppID, feishuAppSecret)
}

// Initialize WorkflowService
workflowSvc := service.NewWorkflowService(db, stateEngine, feishuWorkflowClient, repos.Project, repos.Task)
```

3. Pass workflowSvc to handlers. Since NewHandlers currently takes specific args, we need to add WorkflowHandler to the Handlers struct:

In handler.go, add `Workflow *WorkflowHandler` to Handlers struct.
In NewHandlers, accept workflowSvc and create the handler.

Alternative (simpler): Create WorkflowHandler separately in main.go and add routes directly.

**Best approach**: Modify Handlers struct to include `Workflow *WorkflowHandler`, and modify NewHandlers to also accept `*service.WorkflowService`.

4. Add routes inside the `projects` group in registerRoutes:
```go
// Workflow operations
projects.POST("/:id/tasks/:taskId/assign", h.Workflow.AssignTask)
projects.POST("/:id/tasks/:taskId/start", h.Workflow.StartTask)
projects.POST("/:id/tasks/:taskId/complete", h.Workflow.CompleteTask)
projects.POST("/:id/tasks/:taskId/review", h.Workflow.SubmitReview)
projects.POST("/:id/phases/:phase/assign-roles", h.Workflow.AssignPhaseRoles)
projects.GET("/:id/tasks/:taskId/history", h.Workflow.GetTaskHistory)
```

## Important Notes

- Do NOT modify entity/project.go — Task struct already has Status, RequiresApproval, FeishuTaskID, AutoCreateFeishuTask etc.
- The existing TaskStatusPending/TaskStatusInProgress/TaskStatusCompleted in project.go remain. workflow.go only adds Unassigned/Reviewing/Rejected.
- Use `github.com/google/uuid` for generating IDs in workflow.go entities (uuid.New().String())
- The engine.Engine.Fire() method uses uuid.UUID type for entityID. Since Task.ID is a string (32 char), you'll need to either: (a) parse it as UUID, or (b) not use engine.Fire directly in workflow_service but just manage status transitions manually in the service. **Recommended**: manage transitions directly in workflow_service (the engine's Fire is more suited for auto-transition scenarios). The state machine is registered for future use but WorkflowService handles the actual status changes directly via DB updates for now.
- All DB operations should use `context.Background()` or the gin context.
- Use log.Printf for logging (consistent with existing codebase).
