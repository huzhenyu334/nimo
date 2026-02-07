package engine

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// JSONB — 本地 JSONB 类型，避免与 internal/plm/entity 循环引用
// =============================================================================

// JSONB 用于 PostgreSQL JSONB 类型的序列化/反序列化
type JSONB map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// =============================================================================
// StateMachineDefinition — 状态机定义
// =============================================================================

// StateMachineDefinition 状态机定义，可复用于 PLM/ERP/WMS
type StateMachineDefinition struct {
	ID           uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey"`
	Name         string           `json:"name" gorm:"size:100;not null;uniqueIndex"`          // 如: plm_task, purchase_order
	Description  string           `json:"description" gorm:"type:text"`                       // 描述
	InitialState string           `json:"initial_state" gorm:"size:50;not null"`              // 初始状态
	States       json.RawMessage  `json:"states" gorm:"type:jsonb"`                           // 状态列表及属性
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`

	// 关联
	Transitions []StateTransition `json:"transitions,omitempty" gorm:"foreignKey:MachineID"`
}

// TableName 指定表名
func (StateMachineDefinition) TableName() string {
	return "state_machine_definitions"
}

// StateDefinition 单个状态的定义（用于 States JSONB 字段解析）
type StateDefinition struct {
	Name        string `json:"name"`         // 状态标识: unassigned, pending, in_progress...
	Label       string `json:"label"`        // 显示名称: 待指派, 待处理, 进行中...
	Description string `json:"description"`  // 描述
	IsFinal     bool   `json:"is_final"`     // 是否终态
}

// =============================================================================
// StateTransition — 状态转换规则
// =============================================================================

// StateTransition 状态转换规则
type StateTransition struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	MachineID   uuid.UUID       `json:"machine_id" gorm:"type:uuid;not null;index"`           // 所属状态机
	FromState   string          `json:"from_state" gorm:"size:50;not null"`                    // 起始状态
	ToState     string          `json:"to_state" gorm:"size:50;not null"`                      // 目标状态
	Event       string          `json:"event" gorm:"size:100;not null"`                        // 触发事件名
	Condition   json.RawMessage `json:"condition" gorm:"type:jsonb"`                           // 条件表达式（可选）
	Actions     json.RawMessage `json:"actions" gorm:"type:jsonb"`                             // 触发的动作列表
	Priority    int             `json:"priority" gorm:"default:0"`                             // 优先级（同事件多条规则时，值大者优先）
	Description string          `json:"description" gorm:"type:text"`                          // 描述
}

// TableName 指定表名
func (StateTransition) TableName() string {
	return "state_transitions"
}

// =============================================================================
// TransitionAction — 转换动作（JSON 结构，非独立表）
// =============================================================================

// TransitionAction 转换时执行的动作配置
type TransitionAction struct {
	Type   string                 `json:"type"`   // 动作类型: feishu_create_task, notify_users, etc.
	Config map[string]interface{} `json:"config"` // 动作配置参数
}

// =============================================================================
// TransitionLog — 状态转换日志（审计追溯）
// =============================================================================

// TransitionLog 状态转换日志，用于审计追溯
type TransitionLog struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	EntityType      string          `json:"entity_type" gorm:"size:50;not null;index"`          // plm_task, purchase_order 等
	EntityID        uuid.UUID       `json:"entity_id" gorm:"type:uuid;not null;index"`          // 实体ID
	FromState       string          `json:"from_state" gorm:"size:50"`                           // 转换前状态（首次可为空）
	ToState         string          `json:"to_state" gorm:"size:50;not null"`                    // 转换后状态
	Event           string          `json:"event" gorm:"size:100;not null"`                      // 触发事件
	EventData       json.RawMessage `json:"event_data" gorm:"type:jsonb"`                        // 事件携带的数据
	TriggeredBy     string          `json:"triggered_by" gorm:"size:64"`                         // 操作人（user_id 或 agent_id）
	TriggeredByType string          `json:"triggered_by_type" gorm:"size:20"`                    // user | agent | system
	ActionsExecuted json.RawMessage `json:"actions_executed" gorm:"type:jsonb"`                  // 执行了哪些动作
	CreatedAt       time.Time       `json:"created_at"`
}

// TableName 指定表名
func (TransitionLog) TableName() string {
	return "state_transition_logs"
}

// =============================================================================
// EntityState — 实体当前状态追踪
// =============================================================================

// EntityState 追踪实体当前所处状态
type EntityState struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	EntityType   string    `json:"entity_type" gorm:"size:50;not null;uniqueIndex:idx_entity_type_id"` // 实体类型
	EntityID     uuid.UUID `json:"entity_id" gorm:"type:uuid;not null;uniqueIndex:idx_entity_type_id"` // 实体ID
	CurrentState string    `json:"current_state" gorm:"size:50;not null"`                               // 当前状态
	MachineID    uuid.UUID `json:"machine_id" gorm:"type:uuid"`                                         // 所属状态机
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (EntityState) TableName() string {
	return "entity_states"
}
