package engine

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =============================================================================
// Repository — 状态机引擎数据访问层
// =============================================================================

// Repository 状态机引擎的数据访问层
type Repository struct {
	DB *gorm.DB
}

// NewRepository 创建数据访问层实例
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// =============================================================================
// 状态机定义 CRUD
// =============================================================================

// SaveMachineDefinition 保存状态机定义（不存在则创建，已存在则更新）
func (r *Repository) SaveMachineDefinition(def *StateMachineDefinition) error {
	if def.ID == uuid.Nil {
		def.ID = uuid.New()
	}

	// 使用 Upsert：按 name 唯一键冲突时更新
	result := r.DB.Where("name = ?", def.Name).FirstOrCreate(def)
	if result.Error != nil {
		return fmt.Errorf("保存状态机定义失败: %w", result.Error)
	}

	// 如果已存在，更新字段
	if result.RowsAffected == 0 {
		result = r.DB.Model(def).Where("name = ?", def.Name).Updates(map[string]interface{}{
			"description":   def.Description,
			"initial_state": def.InitialState,
			"states":        def.States,
		})
		if result.Error != nil {
			return fmt.Errorf("更新状态机定义失败: %w", result.Error)
		}
	}

	return nil
}

// GetMachineDefinition 按名称获取状态机定义
func (r *Repository) GetMachineDefinition(name string) (*StateMachineDefinition, error) {
	var def StateMachineDefinition
	result := r.DB.Where("name = ?", name).First(&def)
	if result.Error != nil {
		return nil, fmt.Errorf("获取状态机定义失败 [name=%s]: %w", name, result.Error)
	}
	return &def, nil
}

// GetMachineDefinitionByID 按ID获取状态机定义
func (r *Repository) GetMachineDefinitionByID(id uuid.UUID) (*StateMachineDefinition, error) {
	var def StateMachineDefinition
	result := r.DB.Where("id = ?", id).First(&def)
	if result.Error != nil {
		return nil, fmt.Errorf("获取状态机定义失败 [id=%s]: %w", id.String(), result.Error)
	}
	return &def, nil
}

// =============================================================================
// 状态转换规则 CRUD
// =============================================================================

// SaveTransitions 批量保存转换规则
func (r *Repository) SaveTransitions(transitions []StateTransition) error {
	if len(transitions) == 0 {
		return nil
	}

	// 为没有 ID 的转换生成 UUID
	for i := range transitions {
		if transitions[i].ID == uuid.Nil {
			transitions[i].ID = uuid.New()
		}
	}

	// 先删除该状态机的旧规则，再批量插入新规则
	machineID := transitions[0].MachineID
	tx := r.DB.Begin()

	if err := tx.Where("machine_id = ?", machineID).Delete(&StateTransition{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除旧转换规则失败: %w", err)
	}

	if err := tx.Create(&transitions).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存转换规则失败: %w", err)
	}

	return tx.Commit().Error
}

// GetTransitions 获取状态机的所有转换规则
func (r *Repository) GetTransitions(machineID uuid.UUID) ([]StateTransition, error) {
	var transitions []StateTransition
	result := r.DB.Where("machine_id = ?", machineID).Order("priority DESC").Find(&transitions)
	if result.Error != nil {
		return nil, fmt.Errorf("获取转换规则失败: %w", result.Error)
	}
	return transitions, nil
}

// GetMatchingTransitions 获取匹配的转换规则（from_state + event）
func (r *Repository) GetMatchingTransitions(machineID uuid.UUID, fromState string, event string) ([]StateTransition, error) {
	var transitions []StateTransition
	result := r.DB.Where("machine_id = ? AND from_state = ? AND event = ?", machineID, fromState, event).
		Order("priority DESC").
		Find(&transitions)
	if result.Error != nil {
		return nil, fmt.Errorf("获取匹配转换规则失败: %w", result.Error)
	}
	return transitions, nil
}

// =============================================================================
// 实体状态 CRUD
// =============================================================================

// GetEntityState 获取实体当前状态
func (r *Repository) GetEntityState(entityType string, entityID uuid.UUID) (*EntityState, error) {
	var state EntityState
	result := r.DB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).First(&state)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // 未找到不算错误，表示新实体
		}
		return nil, fmt.Errorf("获取实体状态失败: %w", result.Error)
	}
	return &state, nil
}

// SaveEntityState 保存/更新实体状态
func (r *Repository) SaveEntityState(state *EntityState) error {
	if state.ID == uuid.Nil {
		state.ID = uuid.New()
	}

	// Upsert: 按 entity_type + entity_id 唯一索引
	result := r.DB.Where("entity_type = ? AND entity_id = ?", state.EntityType, state.EntityID).
		Assign(map[string]interface{}{
			"current_state": state.CurrentState,
			"machine_id":    state.MachineID,
		}).
		FirstOrCreate(state)

	if result.Error != nil {
		return fmt.Errorf("保存实体状态失败: %w", result.Error)
	}
	return nil
}

// =============================================================================
// 转换日志 CRUD
// =============================================================================

// SaveTransitionLog 保存转换日志
func (r *Repository) SaveTransitionLog(logEntry *TransitionLog) error {
	if logEntry.ID == uuid.Nil {
		logEntry.ID = uuid.New()
	}
	result := r.DB.Create(logEntry)
	if result.Error != nil {
		return fmt.Errorf("保存转换日志失败: %w", result.Error)
	}
	return nil
}

// GetTransitionLogs 获取实体的转换日志（按时间倒序）
func (r *Repository) GetTransitionLogs(entityType string, entityID uuid.UUID) ([]TransitionLog, error) {
	var logs []TransitionLog
	result := r.DB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("获取转换日志失败: %w", result.Error)
	}
	return logs, nil
}
