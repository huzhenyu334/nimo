package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =============================================================================
// Engine — 状态机核心引擎
// =============================================================================

// Engine 状态机引擎
// 管理状态机定义、执行状态转换、记录审计日志
type Engine struct {
	DB             *gorm.DB
	repo           *Repository
	actionExecutor ActionExecutor
	machines       map[string]*StateMachineDefinition // name -> definition 内存缓存
	mu             sync.RWMutex                       // 保护 machines map
}

// NewEngine 创建状态机引擎实例
// db: 数据库连接
// executor: 动作执行器（传 nil 则使用默认的日志执行器）
func NewEngine(db *gorm.DB, executor ActionExecutor) *Engine {
	if executor == nil {
		executor = NewLoggingActionExecutor()
	}

	return &Engine{
		DB:             db,
		repo:           NewRepository(db),
		actionExecutor: executor,
		machines:       make(map[string]*StateMachineDefinition),
	}
}

// =============================================================================
// 状态机注册
// =============================================================================

// RegisterMachine 注册状态机定义
// 将定义存入数据库并缓存到内存
func (e *Engine) RegisterMachine(def *StateMachineDefinition) error {
	if def == nil {
		return fmt.Errorf("状态机定义不能为空")
	}
	if def.Name == "" {
		return fmt.Errorf("状态机名称不能为空")
	}
	if def.InitialState == "" {
		return fmt.Errorf("初始状态不能为空")
	}

	// 生成 ID（如果没有）
	if def.ID == uuid.Nil {
		def.ID = uuid.New()
	}

	// 保存到数据库
	if err := e.repo.SaveMachineDefinition(def); err != nil {
		return fmt.Errorf("注册状态机失败: %w", err)
	}

	// 从数据库重新读取（确保 ID 一致）
	saved, err := e.repo.GetMachineDefinition(def.Name)
	if err != nil {
		return fmt.Errorf("读取已保存的状态机失败: %w", err)
	}

	// 保存转换规则
	if len(def.Transitions) > 0 {
		for i := range def.Transitions {
			def.Transitions[i].MachineID = saved.ID
		}
		if err := e.repo.SaveTransitions(def.Transitions); err != nil {
			return fmt.Errorf("保存转换规则失败: %w", err)
		}
	}

	// 更新内存缓存
	e.mu.Lock()
	saved.Transitions = def.Transitions
	e.machines[def.Name] = saved
	e.mu.Unlock()

	log.Printf("[StateEngine] 注册状态机: name=%s initial_state=%s transitions=%d",
		def.Name, def.InitialState, len(def.Transitions))

	return nil
}

// GetMachine 获取已注册的状态机定义
func (e *Engine) GetMachine(name string) (*StateMachineDefinition, error) {
	// 先从内存缓存查找
	e.mu.RLock()
	if m, ok := e.machines[name]; ok {
		e.mu.RUnlock()
		return m, nil
	}
	e.mu.RUnlock()

	// 缓存未命中，从数据库加载
	def, err := e.repo.GetMachineDefinition(name)
	if err != nil {
		return nil, err
	}

	// 加载转换规则
	transitions, err := e.repo.GetTransitions(def.ID)
	if err != nil {
		return nil, err
	}
	def.Transitions = transitions

	// 放入缓存
	e.mu.Lock()
	e.machines[name] = def
	e.mu.Unlock()

	return def, nil
}

// =============================================================================
// 状态转换 — 核心方法
// =============================================================================

// Fire 触发状态转换
// entityType: 实体类型（如 "plm_task"）
// entityID: 实体ID
// event: 触发事件（如 "assign", "start", "complete"）
// eventData: 事件携带的数据（用于条件评估和动作执行）
// triggeredBy: 操作人ID
// triggeredByType: 操作人类型（user/agent/system）
func (e *Engine) Fire(entityType string, entityID uuid.UUID, event string, eventData map[string]interface{}, triggeredBy string, triggeredByType string) (*TransitionLog, error) {
	// 1. 获取状态机定义
	machine, err := e.findMachineForEntity(entityType)
	if err != nil {
		return nil, fmt.Errorf("未找到实体类型 [%s] 对应的状态机: %w", entityType, err)
	}

	// 2. 获取当前状态
	currentState, err := e.getCurrentStateOrInitial(entityType, entityID, machine)
	if err != nil {
		return nil, fmt.Errorf("获取当前状态失败: %w", err)
	}

	// 3. 查找匹配的转换规则
	transition, err := e.findMatchingTransition(machine.ID, currentState, event, eventData)
	if err != nil {
		return nil, err
	}

	// 4. 在事务中执行状态转换
	var transitionLog *TransitionLog
	err = e.DB.Transaction(func(tx *gorm.DB) error {
		txRepo := NewRepository(tx)

		// 更新实体状态
		entityState := &EntityState{
			EntityType:   entityType,
			EntityID:     entityID,
			CurrentState: transition.ToState,
			MachineID:    machine.ID,
			UpdatedAt:    time.Now(),
		}
		if err := txRepo.SaveEntityState(entityState); err != nil {
			return fmt.Errorf("更新实体状态失败: %w", err)
		}

		// 解析并执行动作
		actions, actionsExecuted := e.executeActions(transition, entityType, entityID, currentState, event, eventData)

		// 序列化事件数据和执行的动作
		eventDataJSON, _ := json.Marshal(eventData)
		actionsJSON, _ := json.Marshal(actionsExecuted)

		// 记录转换日志
		transitionLog = &TransitionLog{
			ID:              uuid.New(),
			EntityType:      entityType,
			EntityID:        entityID,
			FromState:       currentState,
			ToState:         transition.ToState,
			Event:           event,
			EventData:       eventDataJSON,
			TriggeredBy:     triggeredBy,
			TriggeredByType: triggeredByType,
			ActionsExecuted: actionsJSON,
			CreatedAt:       time.Now(),
		}
		if err := txRepo.SaveTransitionLog(transitionLog); err != nil {
			return fmt.Errorf("保存转换日志失败: %w", err)
		}

		log.Printf("[StateEngine] 状态转换: entity=%s/%s %s→%s event=%s actions=%d",
			entityType, entityID.String(), currentState, transition.ToState, event, len(actions))

		return nil
	})

	if err != nil {
		return nil, err
	}

	return transitionLog, nil
}

// =============================================================================
// 查询方法
// =============================================================================

// GetCurrentState 获取实体当前状态
func (e *Engine) GetCurrentState(entityType string, entityID uuid.UUID) (string, error) {
	state, err := e.repo.GetEntityState(entityType, entityID)
	if err != nil {
		return "", err
	}
	if state == nil {
		// 未找到记录，返回状态机的初始状态
		machine, err := e.findMachineForEntity(entityType)
		if err != nil {
			return "", fmt.Errorf("未找到实体类型 [%s] 对应的状态机: %w", entityType, err)
		}
		return machine.InitialState, nil
	}
	return state.CurrentState, nil
}

// GetHistory 获取实体的状态转换历史
func (e *Engine) GetHistory(entityType string, entityID uuid.UUID) ([]TransitionLog, error) {
	return e.repo.GetTransitionLogs(entityType, entityID)
}

// =============================================================================
// 内部辅助方法
// =============================================================================

// findMachineForEntity 根据实体类型查找对应的状态机
// 约定: 实体类型名与状态机名一致（如 plm_task → plm_task 状态机）
func (e *Engine) findMachineForEntity(entityType string) (*StateMachineDefinition, error) {
	return e.GetMachine(entityType)
}

// getCurrentStateOrInitial 获取实体当前状态，如果是新实体则返回初始状态
func (e *Engine) getCurrentStateOrInitial(entityType string, entityID uuid.UUID, machine *StateMachineDefinition) (string, error) {
	state, err := e.repo.GetEntityState(entityType, entityID)
	if err != nil {
		return "", err
	}
	if state == nil {
		return machine.InitialState, nil
	}
	return state.CurrentState, nil
}

// findMatchingTransition 查找匹配的转换规则
// 按优先级排序，找到第一个条件满足的规则
func (e *Engine) findMatchingTransition(machineID uuid.UUID, fromState string, event string, eventData map[string]interface{}) (*StateTransition, error) {
	transitions, err := e.repo.GetMatchingTransitions(machineID, fromState, event)
	if err != nil {
		return nil, fmt.Errorf("查找转换规则失败: %w", err)
	}

	if len(transitions) == 0 {
		return nil, fmt.Errorf("无效的状态转换: state=%s event=%s（没有匹配的转换规则）", fromState, event)
	}

	// 按优先级排序（已在 SQL 中排序），找第一个条件满足的
	for _, t := range transitions {
		if EvaluateCondition(t.Condition, eventData) {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("无效的状态转换: state=%s event=%s（条件不满足）", fromState, event)
}

// executeActions 解析并执行转换动作
// 返回解析出的动作列表和执行结果
func (e *Engine) executeActions(transition *StateTransition, entityType string, entityID uuid.UUID, fromState string, event string, eventData map[string]interface{}) ([]TransitionAction, []map[string]interface{}) {
	var actions []TransitionAction
	var actionsExecuted []map[string]interface{}

	// 解析动作列表
	if len(transition.Actions) > 0 && string(transition.Actions) != "null" {
		if err := json.Unmarshal(transition.Actions, &actions); err != nil {
			log.Printf("[StateEngine] 解析动作失败: %v", err)
			return actions, actionsExecuted
		}
	}

	// 构建动作上下文
	ctx := ActionContext{
		EntityType: entityType,
		EntityID:   entityID,
		FromState:  fromState,
		ToState:    transition.ToState,
		Event:      event,
		EventData:  eventData,
	}

	// 执行每个动作
	for _, action := range actions {
		result := map[string]interface{}{
			"type":   action.Type,
			"status": "success",
		}

		if err := e.actionExecutor.Execute(action, ctx); err != nil {
			log.Printf("[StateEngine] 动作执行失败: type=%s error=%v", action.Type, err)
			result["status"] = "error"
			result["error"] = err.Error()
		}

		actionsExecuted = append(actionsExecuted, result)
	}

	return actions, actionsExecuted
}
