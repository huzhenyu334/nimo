package entity

import "time"

// Routing channel constants
const (
	RoutingChannelFeishu = "feishu" // 人工审批（飞书）
	RoutingChannelAgent  = "agent"  // Agent自动处理
	RoutingChannelAuto   = "auto"   // 智能判断
)

// RoutingRule 路由规则
type RoutingRule struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	Name         string    `json:"name" gorm:"size:100;not null"`
	EntityType   string    `json:"entity_type" gorm:"size:50;not null"`
	Event        string    `json:"event" gorm:"size:100;not null"`
	Conditions   JSONB     `json:"conditions" gorm:"type:jsonb;not null"`
	Channel      string    `json:"channel" gorm:"size:20;not null"`
	Priority     int       `json:"priority" gorm:"default:0"`
	ActionConfig JSONB     `json:"action_config" gorm:"type:jsonb"`
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	Description  string    `json:"description" gorm:"type:text"`
	CreatedBy    string    `json:"created_by" gorm:"size:32"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (RoutingRule) TableName() string { return "routing_rules" }

// RoutingDecision 路由决策结果
type RoutingDecision struct {
	Channel      string                 `json:"channel"`       // feishu | agent
	RuleID       string                 `json:"rule_id"`       // 匹配的规则ID（空=默认规则）
	RuleName     string                 `json:"rule_name"`     // 匹配的规则名称
	ActionConfig map[string]interface{} `json:"action_config"` // 动作配置
	Reason       string                 `json:"reason"`        // 决策原因
}

// RoutingLog 路由日志
type RoutingLog struct {
	ID         string    `json:"id" gorm:"primaryKey;size:36"`
	RuleID     string    `json:"rule_id" gorm:"size:36"`
	RuleName   string    `json:"rule_name" gorm:"size:100"`
	EntityType string    `json:"entity_type" gorm:"size:50;not null"`
	EntityID   string    `json:"entity_id" gorm:"size:50"`
	Event      string    `json:"event" gorm:"size:100;not null"`
	Channel    string    `json:"channel" gorm:"size:20;not null"`
	Context    JSONB     `json:"context" gorm:"type:jsonb"`
	Reason     string    `json:"reason" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (RoutingLog) TableName() string { return "routing_logs" }
