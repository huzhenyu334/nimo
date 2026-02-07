package engine

import (
	"fmt"
	"log"

	"github.com/google/uuid"
)

// =============================================================================
// 动作执行器接口 — Phase 1 仅定义接口和日志实现
// Phase 2 将实现飞书对接的 ActionExecutor
// =============================================================================

// ActionContext 动作执行上下文，包含状态转换的完整信息
type ActionContext struct {
	EntityType string                 `json:"entity_type"` // 实体类型: plm_task, purchase_order...
	EntityID   uuid.UUID              `json:"entity_id"`   // 实体ID
	FromState  string                 `json:"from_state"`  // 转换前状态
	ToState    string                 `json:"to_state"`    // 转换后状态
	Event      string                 `json:"event"`       // 触发事件
	EventData  map[string]interface{} `json:"event_data"`  // 事件数据
}

// ActionExecutor 动作执行器接口
// Phase 2 将实现飞书相关的执行器（创建任务、发起审批、发送通知等）
type ActionExecutor interface {
	// Execute 执行单个动作
	// action: 动作定义（类型+配置）
	// ctx: 动作执行上下文
	Execute(action TransitionAction, ctx ActionContext) error
}

// =============================================================================
// LoggingActionExecutor — 默认实现，仅记录日志
// =============================================================================

// LoggingActionExecutor 日志记录动作执行器（默认实现）
// 仅将动作信息输出到日志，不执行实际操作
// 用于 Phase 1 开发和测试阶段
type LoggingActionExecutor struct{}

// NewLoggingActionExecutor 创建日志记录动作执行器
func NewLoggingActionExecutor() *LoggingActionExecutor {
	return &LoggingActionExecutor{}
}

// Execute 记录动作日志（不执行实际操作）
func (l *LoggingActionExecutor) Execute(action TransitionAction, ctx ActionContext) error {
	log.Printf("[StateEngine] 执行动作: type=%s entity=%s/%s transition=%s→%s event=%s config=%v",
		action.Type,
		ctx.EntityType,
		ctx.EntityID.String(),
		ctx.FromState,
		ctx.ToState,
		ctx.Event,
		action.Config,
	)
	return nil
}

// =============================================================================
// CompositeActionExecutor — 组合执行器，支持按动作类型分发
// =============================================================================

// CompositeActionExecutor 组合动作执行器
// 根据动作类型分发到不同的处理器，未注册的类型回退到默认执行器
type CompositeActionExecutor struct {
	handlers map[string]ActionExecutor // 按动作类型注册的处理器
	fallback ActionExecutor            // 未匹配时的回退执行器
}

// NewCompositeActionExecutor 创建组合动作执行器
func NewCompositeActionExecutor(fallback ActionExecutor) *CompositeActionExecutor {
	return &CompositeActionExecutor{
		handlers: make(map[string]ActionExecutor),
		fallback: fallback,
	}
}

// Register 注册特定动作类型的处理器
func (c *CompositeActionExecutor) Register(actionType string, handler ActionExecutor) {
	c.handlers[actionType] = handler
}

// Execute 根据动作类型分发执行
func (c *CompositeActionExecutor) Execute(action TransitionAction, ctx ActionContext) error {
	if handler, ok := c.handlers[action.Type]; ok {
		return handler.Execute(action, ctx)
	}
	if c.fallback != nil {
		return c.fallback.Execute(action, ctx)
	}
	return fmt.Errorf("no handler registered for action type: %s", action.Type)
}
