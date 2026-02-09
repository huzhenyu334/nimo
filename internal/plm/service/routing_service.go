package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bitfantasy/nimo/internal/plm/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoutingService 智能路由服务
type RoutingService struct {
	db *gorm.DB
}

// NewRoutingService 创建路由服务
func NewRoutingService(db *gorm.DB) *RoutingService {
	return &RoutingService{db: db}
}

// =============================================================================
// 核心路由决策
// =============================================================================

// EvaluateRoute 评估路由决策
// entityType: plm_task, purchase_order, bom_review 等
// event: task_complete, bom_submit, approval_needed 等
// routeCtx: 路由上下文（包含业务数据用于条件匹配）
func (s *RoutingService) EvaluateRoute(ctx context.Context, entityType, event string, routeCtx map[string]interface{}) (*entity.RoutingDecision, error) {
	// 查找匹配的启用规则，按优先级降序
	var rules []entity.RoutingRule
	if err := s.db.WithContext(ctx).
		Where("entity_type = ? AND event = ? AND enabled = true", entityType, event).
		Order("priority DESC").
		Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("查询路由规则失败: %w", err)
	}

	// 逐条评估规则
	for _, rule := range rules {
		matched, err := evaluateConditions(rule.Conditions, routeCtx)
		if err != nil {
			log.Printf("[RoutingService] 规则[%s]条件评估失败: %v", rule.Name, err)
			continue
		}
		if matched {
			channel := rule.Channel
			// auto 通道需要进一步判断（简单策略：默认走 agent）
			if channel == entity.RoutingChannelAuto {
				channel = entity.RoutingChannelAgent
			}

			decision := &entity.RoutingDecision{
				Channel:      channel,
				RuleID:       rule.ID,
				RuleName:     rule.Name,
				ActionConfig: map[string]interface{}(rule.ActionConfig),
				Reason:       fmt.Sprintf("匹配规则[%s]", rule.Name),
			}

			// 记录路由日志
			s.logRouting(ctx, &rule, entityType, "", event, channel, routeCtx, decision.Reason)

			return decision, nil
		}
	}

	// 没有匹配规则，默认走人工（feishu）
	decision := &entity.RoutingDecision{
		Channel: entity.RoutingChannelFeishu,
		Reason:  "无匹配规则，默认走人工审批",
	}

	s.logRouting(ctx, nil, entityType, "", event, entity.RoutingChannelFeishu, routeCtx, decision.Reason)

	return decision, nil
}

// =============================================================================
// 条件评估引擎
// =============================================================================

// evaluateConditions 评估条件组合
// 支持格式:
//
//	{"operator": "and", "conditions": [...]}
//	{"operator": "or", "conditions": [...]}
//	{"field": "xxx", "op": "eq", "value": xxx}  （单条件也可以直接放在顶层）
func evaluateConditions(conditions entity.JSONB, routeCtx map[string]interface{}) (bool, error) {
	if conditions == nil || len(conditions) == 0 {
		return true, nil // 空条件默认匹配
	}

	// 检查是否有 operator 字段（组合条件）
	if op, ok := conditions["operator"]; ok {
		operator, _ := op.(string)
		subConds, ok := conditions["conditions"].([]interface{})
		if !ok {
			return false, fmt.Errorf("conditions 字段格式错误")
		}
		return evaluateGroup(operator, subConds, routeCtx)
	}

	// 单条件
	return evaluateSingleCondition(conditions, routeCtx)
}

// evaluateGroup 评估 and/or 组合
func evaluateGroup(operator string, conditions []interface{}, routeCtx map[string]interface{}) (bool, error) {
	operator = strings.ToLower(operator)

	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		// 检查是否为嵌套组合
		if _, hasOp := condMap["operator"]; hasOp {
			matched, err := evaluateConditions(entity.JSONB(condMap), routeCtx)
			if err != nil {
				return false, err
			}
			if operator == "or" && matched {
				return true, nil
			}
			if operator == "and" && !matched {
				return false, nil
			}
		} else {
			// 单条件
			matched, err := evaluateSingleCondition(condMap, routeCtx)
			if err != nil {
				return false, err
			}
			if operator == "or" && matched {
				return true, nil
			}
			if operator == "and" && !matched {
				return false, nil
			}
		}
	}

	// and: 全部通过返回 true; or: 全部不通过返回 false
	return operator == "and", nil
}

// evaluateSingleCondition 评估单个条件 {"field": "xxx", "op": "eq", "value": xxx}
func evaluateSingleCondition(cond map[string]interface{}, routeCtx map[string]interface{}) (bool, error) {
	field, _ := cond["field"].(string)
	op, _ := cond["op"].(string)
	expected := cond["value"]

	if field == "" || op == "" {
		return false, fmt.Errorf("条件缺少 field 或 op")
	}

	actual, exists := routeCtx[field]
	if !exists {
		// 字段不存在，视为不匹配（除了 neq/not_in/not_contains）
		switch op {
		case "neq", "not_in", "not_contains":
			return true, nil
		default:
			return false, nil
		}
	}

	return compareValues(op, actual, expected)
}

// compareValues 比较值
func compareValues(op string, actual, expected interface{}) (bool, error) {
	switch op {
	case "eq":
		return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected), nil
	case "neq":
		return fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected), nil
	case "gt":
		return toFloat64(actual) > toFloat64(expected), nil
	case "gte":
		return toFloat64(actual) >= toFloat64(expected), nil
	case "lt":
		return toFloat64(actual) < toFloat64(expected), nil
	case "lte":
		return toFloat64(actual) <= toFloat64(expected), nil
	case "in":
		return valueInSlice(actual, expected), nil
	case "not_in":
		return !valueInSlice(actual, expected), nil
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected)), nil
	case "not_contains":
		return !strings.Contains(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected)), nil
	default:
		return false, fmt.Errorf("不支持的操作符: %s", op)
	}
}

// toFloat64 将 interface{} 转为 float64
func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	case string:
		var f float64
		fmt.Sscanf(n, "%f", &f)
		return f
	default:
		return 0
	}
}

// valueInSlice 检查值是否在切片中
func valueInSlice(actual, expected interface{}) bool {
	actualStr := fmt.Sprintf("%v", actual)
	switch arr := expected.(type) {
	case []interface{}:
		for _, v := range arr {
			if fmt.Sprintf("%v", v) == actualStr {
				return true
			}
		}
	case []string:
		for _, v := range arr {
			if v == actualStr {
				return true
			}
		}
	}
	return false
}

// =============================================================================
// CRUD 方法
// =============================================================================

// CreateRule 创建路由规则
func (s *RoutingService) CreateRule(ctx context.Context, rule *entity.RoutingRule) error {
	rule.ID = uuid.New().String()
	if err := s.db.WithContext(ctx).Create(rule).Error; err != nil {
		return fmt.Errorf("创建路由规则失败: %w", err)
	}
	return nil
}

// UpdateRule 更新路由规则
func (s *RoutingService) UpdateRule(ctx context.Context, id string, updates map[string]interface{}) error {
	result := s.db.WithContext(ctx).Model(&entity.RoutingRule{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新路由规则失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("路由规则不存在")
	}
	return nil
}

// DeleteRule 删除路由规则
func (s *RoutingService) DeleteRule(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.RoutingRule{})
	if result.Error != nil {
		return fmt.Errorf("删除路由规则失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("路由规则不存在")
	}
	return nil
}

// GetRule 获取路由规则
func (s *RoutingService) GetRule(ctx context.Context, id string) (*entity.RoutingRule, error) {
	var rule entity.RoutingRule
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error; err != nil {
		return nil, fmt.Errorf("路由规则不存在")
	}
	return &rule, nil
}

// ListRules 查询路由规则列表
func (s *RoutingService) ListRules(ctx context.Context, entityType, event string, page, pageSize int) ([]entity.RoutingRule, int64, error) {
	query := s.db.WithContext(ctx).Model(&entity.RoutingRule{})

	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if event != "" {
		query = query.Where("event = ?", event)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询规则总数失败: %w", err)
	}

	var rules []entity.RoutingRule
	if err := query.Order("priority DESC, created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&rules).Error; err != nil {
		return nil, 0, fmt.Errorf("查询规则列表失败: %w", err)
	}

	return rules, total, nil
}

// =============================================================================
// 路由日志
// =============================================================================

// ListLogs 查询路由日志
func (s *RoutingService) ListLogs(ctx context.Context, entityType, event string, page, pageSize int) ([]entity.RoutingLog, int64, error) {
	query := s.db.WithContext(ctx).Model(&entity.RoutingLog{})

	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	if event != "" {
		query = query.Where("event = ?", event)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询路由日志总数失败: %w", err)
	}

	var logs []entity.RoutingLog
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询路由日志失败: %w", err)
	}

	return logs, total, nil
}

// logRouting 记录路由日志
func (s *RoutingService) logRouting(ctx context.Context, rule *entity.RoutingRule, entityType, entityID, event, channel string, routeCtx map[string]interface{}, reason string) {
	logEntry := entity.RoutingLog{
		ID:         uuid.New().String(),
		EntityType: entityType,
		EntityID:   entityID,
		Event:      event,
		Channel:    channel,
		Context:    entity.JSONB(routeCtx),
		Reason:     reason,
	}
	if rule != nil {
		logEntry.RuleID = rule.ID
		logEntry.RuleName = rule.Name
	}
	if err := s.db.WithContext(ctx).Create(&logEntry).Error; err != nil {
		log.Printf("[RoutingService] 记录路由日志失败: %v", err)
	}
}
