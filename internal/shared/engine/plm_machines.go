package engine

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// =============================================================================
// PLM 预设状态机定义
// =============================================================================

// NewPLMTaskMachine 创建 PLM 任务状态机定义
//
// 状态流转图:
//
//	                   ┌──────────┐
//	                   │ unassigned│ (待指派)
//	                   └────┬─────┘
//	                     assign
//	                   ┌────▼─────┐
//	                   │  pending  │ (待处理) ←──── reject (回退)
//	                   └────┬─────┘                    │
//	                     start                         │
//	                   ┌────▼──────┐                   │
//	                   │in_progress│ (进行中)          │
//	                   └────┬──────┘                   │
//	               ┌────────┴────────┐                 │
//	           complete         submit_review          │
//	          (无需审批)          (需审批)              │
//	               │                 │                 │
//	        ┌──────▼──┐      ┌───────▼──┐             │
//	        │completed │      │reviewing │ (待审批)    │
//	        │(已完成)  │      └────┬─────┘             │
//	        └─────────┘      ┌────┴────┐              │
//	                      approve    reject            │
//	                    ┌────▼───┐  ┌──▼──────────────┘
//	                    │completed│  │(回退到进行中)
//	                    │(已完成) │  └─────────────────
//	                    └────────┘
func NewPLMTaskMachine() *StateMachineDefinition {
	// 状态定义
	states := []StateDefinition{
		{Name: "unassigned", Label: "待指派", Description: "任务已创建，尚未指派执行人"},
		{Name: "pending", Label: "待处理", Description: "已指派执行人，等待开始"},
		{Name: "in_progress", Label: "进行中", Description: "执行人正在处理任务"},
		{Name: "reviewing", Label: "待审批", Description: "任务已提交，等待审批"},
		{Name: "completed", Label: "已完成", Description: "任务已完成"},
		{Name: "rejected", Label: "已驳回", Description: "任务被驳回，需要修改"},
	}
	statesJSON, _ := json.Marshal(states)

	// 转换规则
	transitions := []StateTransition{
		// unassigned + assign → pending
		{
			ID:        uuid.New(),
			FromState: "unassigned",
			ToState:   "pending",
			Event:     "assign",
			Actions: mustMarshalJSON([]TransitionAction{
				{Type: "feishu_create_task", Config: map[string]interface{}{"description": "为执行人创建飞书任务"}},
				{Type: "notify_users", Config: map[string]interface{}{"message": "您有新的任务待处理"}},
			}),
			Priority:    0,
			Description: "指派任务给执行人",
		},

		// pending + start → in_progress
		{
			ID:        uuid.New(),
			FromState: "pending",
			ToState:   "in_progress",
			Event:     "start",
			Actions: mustMarshalJSON([]TransitionAction{
				{Type: "feishu_update_task", Config: map[string]interface{}{"status": "in_progress"}},
			}),
			Priority:    0,
			Description: "开始执行任务",
		},

		// in_progress + complete → completed（无需审批时）
		{
			ID:        uuid.New(),
			FromState: "in_progress",
			ToState:   "completed",
			Event:     "complete",
			Condition: mustMarshalJSON(map[string]interface{}{
				"field": "needs_review",
				"op":    "eq",
				"value": false,
			}),
			Actions: mustMarshalJSON([]TransitionAction{
				{Type: "feishu_update_task", Config: map[string]interface{}{"status": "completed"}},
				{Type: "start_dependent_tasks", Config: map[string]interface{}{"description": "检查并启动依赖任务"}},
			}),
			Priority:    10, // 优先级高于 submit_review
			Description: "完成任务（无需审批）",
		},

		// in_progress + complete → reviewing（需审批时，same event, lower priority）
		{
			ID:        uuid.New(),
			FromState: "in_progress",
			ToState:   "reviewing",
			Event:     "complete",
			Condition: mustMarshalJSON(map[string]interface{}{
				"field": "needs_review",
				"op":    "eq",
				"value": true,
			}),
			Actions: mustMarshalJSON([]TransitionAction{
				{Type: "feishu_create_approval", Config: map[string]interface{}{"description": "发起飞书审批"}},
			}),
			Priority:    0,
			Description: "提交审批（需审批时）",
		},

		// in_progress + submit_review → reviewing（显式提交审批）
		{
			ID:        uuid.New(),
			FromState: "in_progress",
			ToState:   "reviewing",
			Event:     "submit_review",
			Actions: mustMarshalJSON([]TransitionAction{
				{Type: "feishu_create_approval", Config: map[string]interface{}{"description": "发起飞书审批"}},
			}),
			Priority:    0,
			Description: "显式提交审批",
		},

		// reviewing + approve → completed
		{
			ID:        uuid.New(),
			FromState: "reviewing",
			ToState:   "completed",
			Event:     "approve",
			Actions: mustMarshalJSON([]TransitionAction{
				{Type: "feishu_update_task", Config: map[string]interface{}{"status": "completed"}},
				{Type: "start_dependent_tasks", Config: map[string]interface{}{"description": "审批通过，启动依赖任务"}},
				{Type: "notify_users", Config: map[string]interface{}{"message": "您的任务已审批通过"}},
			}),
			Priority:    0,
			Description: "审批通过",
		},

		// reviewing + reject → in_progress（回退到进行中）
		{
			ID:        uuid.New(),
			FromState: "reviewing",
			ToState:   "in_progress",
			Event:     "reject",
			Actions: mustMarshalJSON([]TransitionAction{
				{Type: "feishu_update_task", Config: map[string]interface{}{"status": "in_progress"}},
				{Type: "notify_users", Config: map[string]interface{}{"message": "您的任务审批未通过，请修改后重新提交"}},
			}),
			Priority:    0,
			Description: "审批驳回，回退到进行中",
		},
	}

	return &StateMachineDefinition{
		ID:           uuid.New(),
		Name:         "plm_task",
		Description:  "PLM 任务状态机 — 管理任务从指派到完成的全生命周期",
		InitialState: "unassigned",
		States:       statesJSON,
		Transitions:  transitions,
	}
}

// mustMarshalJSON 将对象序列化为 json.RawMessage，出错则 panic
func mustMarshalJSON(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("json.Marshal failed: %v", err))
	}
	return data
}
