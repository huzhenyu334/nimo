package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// =============================================================================
// 条件评估器 — 支持简单条件和 AND/OR 组合
// =============================================================================

// EvaluateCondition 评估条件表达式
// condition 为 nil 或空时返回 true（无条件限制）
// 支持格式:
//   - 简单条件: {"field": "review_result", "op": "eq", "value": "pass"}
//   - AND 组合: {"and": [condition1, condition2, ...]}
//   - OR 组合:  {"or":  [condition1, condition2, ...]}
func EvaluateCondition(condition json.RawMessage, context map[string]interface{}) bool {
	// 空条件 → 直接通过
	if len(condition) == 0 || string(condition) == "null" || string(condition) == "{}" {
		return true
	}

	// 解析条件 JSON
	var condMap map[string]interface{}
	if err := json.Unmarshal(condition, &condMap); err != nil {
		// 解析失败，按不满足处理
		return false
	}

	return evaluateConditionMap(condMap, context)
}

// evaluateConditionMap 递归评估条件
func evaluateConditionMap(condMap map[string]interface{}, context map[string]interface{}) bool {
	// 检查 AND 组合
	if andConds, ok := condMap["and"]; ok {
		return evaluateCompound(andConds, context, true)
	}

	// 检查 OR 组合
	if orConds, ok := condMap["or"]; ok {
		return evaluateCompound(orConds, context, false)
	}

	// 简单条件: {"field": ..., "op": ..., "value": ...}
	return evaluateSimpleCondition(condMap, context)
}

// evaluateCompound 评估复合条件（AND/OR）
// isAnd=true 时所有条件必须满足；isAnd=false 时任一满足即可
func evaluateCompound(conditions interface{}, context map[string]interface{}, isAnd bool) bool {
	condList, ok := conditions.([]interface{})
	if !ok {
		return false
	}

	for _, cond := range condList {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}
		result := evaluateConditionMap(condMap, context)
		if isAnd && !result {
			return false // AND: 任一不满足则失败
		}
		if !isAnd && result {
			return true // OR: 任一满足则成功
		}
	}

	if isAnd {
		return true  // AND: 全部满足
	}
	return false // OR: 全部不满足
}

// evaluateSimpleCondition 评估简单条件
// 格式: {"field": "xxx", "op": "eq", "value": "yyy"}
func evaluateSimpleCondition(condMap map[string]interface{}, context map[string]interface{}) bool {
	fieldName, ok := condMap["field"].(string)
	if !ok {
		return false
	}

	op, ok := condMap["op"].(string)
	if !ok {
		return false
	}

	expectedValue := condMap["value"]

	// 从 context 中获取实际值（支持嵌套字段用 . 分隔）
	actualValue := getNestedValue(context, fieldName)

	// 执行比较
	return compareValues(actualValue, op, expectedValue)
}

// getNestedValue 从 map 中获取嵌套值，支持 "a.b.c" 格式
func getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = m[part]
	}

	return current
}

// compareValues 比较两个值
func compareValues(actual interface{}, op string, expected interface{}) bool {
	switch op {
	case "eq":
		return isEqual(actual, expected)
	case "ne":
		return !isEqual(actual, expected)
	case "gt":
		return compareNumeric(actual, expected) > 0
	case "gte":
		return compareNumeric(actual, expected) >= 0
	case "lt":
		return compareNumeric(actual, expected) < 0
	case "lte":
		return compareNumeric(actual, expected) <= 0
	case "in":
		return isIn(actual, expected)
	case "contains":
		return containsValue(actual, expected)
	default:
		return false
	}
}

// isEqual 判断两值是否相等（处理类型转换）
func isEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// 尝试统一为字符串比较
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	if aStr == bStr {
		return true
	}

	// 尝试数值比较（JSON 数值默认为 float64）
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if aOk && bOk {
		return aFloat == bFloat
	}

	// 布尔比较
	aBool, aOk := toBool(a)
	bBool, bOk := toBool(b)
	if aOk && bOk {
		return aBool == bBool
	}

	return reflect.DeepEqual(a, b)
}

// compareNumeric 数值比较，返回 -1, 0, 1
func compareNumeric(a, b interface{}) int {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if !aOk || !bOk {
		return 0
	}

	if aFloat < bFloat {
		return -1
	}
	if aFloat > bFloat {
		return 1
	}
	return 0
}

// isIn 判断值是否在列表中
func isIn(actual, expected interface{}) bool {
	list, ok := expected.([]interface{})
	if !ok {
		return false
	}

	for _, item := range list {
		if isEqual(actual, item) {
			return true
		}
	}
	return false
}

// containsValue 判断字符串是否包含子串
func containsValue(actual, expected interface{}) bool {
	aStr, ok := actual.(string)
	if !ok {
		return false
	}
	bStr, ok := expected.(string)
	if !ok {
		return false
	}
	return strings.Contains(aStr, bStr)
}

// toFloat64 尝试将值转换为 float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// toBool 尝试将值转换为 bool
func toBool(v interface{}) (bool, bool) {
	switch val := v.(type) {
	case bool:
		return val, true
	case string:
		switch strings.ToLower(val) {
		case "true", "1", "yes":
			return true, true
		case "false", "0", "no":
			return false, true
		}
	case float64:
		return val != 0, true
	}
	return false, false
}
