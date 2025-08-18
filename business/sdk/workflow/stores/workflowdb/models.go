package workflowdb

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Table name constants
const (
	TableAutomationRules  = "automation_rules"
	TableRuleActions      = "rule_actions"
	TableActionTemplates  = "action_templates"
	TableRuleDependencies = "rule_dependencies"
	TableTriggerTypes     = "trigger_types"
	TableEntityTypes      = "entity_types"
	TableEntities         = "entities"
)

// AutomationRule represents a workflow automation rule
type AutomationRule struct {
	ID                string          `db:"id"`
	Name              string          `db:"name"`
	Description       sql.NullString  `db:"description"`
	EntityName        string          `db:"entity_name"`
	EntityTypeID      string          `db:"entity_type_id"`
	TriggerTypeID     string          `db:"trigger_type_id"`
	TriggerConditions json.RawMessage `db:"trigger_conditions"`
	IsActive          bool            `db:"is_active"`
	CreatedDate       time.Time       `db:"created_date"`
	UpdatedDate       time.Time       `db:"updated_date"`
	CreatedBy         string          `db:"created_by"`
	UpdatedBy         string          `db:"updated_by"`
}

// RuleAction represents an action within a rule
type RuleAction struct {
	ID                string          `db:"id"`
	AutomationRulesID string          `db:"automation_rules_id"`
	TemplateID        sql.NullString  `db:"template_id"`
	Name              string          `db:"name"`
	Description       sql.NullString  `db:"description"`
	ActionConfig      json.RawMessage `db:"action_config"`
	ExecutionOrder    int             `db:"execution_order"`
	IsActive          bool            `db:"is_active"`
	CreatedDate       time.Time       `db:"created_date"`
	UpdatedDate       time.Time       `db:"updated_date"`
}

// ActionTemplate represents a reusable action configuration
type ActionTemplate struct {
	ID            string          `db:"id"`
	Name          string          `db:"name"`
	Description   sql.NullString  `db:"description"`
	ActionType    string          `db:"action_type"`
	DefaultConfig json.RawMessage `db:"default_config"`
	CreatedDate   time.Time       `db:"created_date"`
	UpdatedDate   time.Time       `db:"updated_date"`
	CreatedBy     string          `db:"created_by"`
	UpdatedBy     string          `db:"updated_by"`
}

// RuleDependency represents a dependency between rules
type RuleDependency struct {
	ParentRuleID string       `db:"parent_rule_id"`
	ChildRuleID  string       `db:"child_rule_id"`
	CreatedDate  sql.NullTime `db:"created_date"`
	UpdatedDate  sql.NullTime `db:"updated_date"`
}

// TriggerType represents types of triggers (on_create, on_update, etc.)
type TriggerType struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
}

// EntityType represents types of entities (table, view, etc.)
type EntityType struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
}

// Entity represents a monitored database entity
type Entity struct {
	ID           string `db:"id"`
	Name         string `db:"name"`
	EntityTypeID string `db:"entity_type_id"`
	IsActive     bool   `db:"is_active"`
}

// AutomationRulesView is a flattened view with joined data
type AutomationRulesView struct {
	ID                string          `db:"id"`
	Name              string          `db:"name"`
	Description       sql.NullString  `db:"description"`
	EntityID          sql.NullString  `db:"entity_id"`
	TriggerConditions json.RawMessage `db:"trigger_conditions"`
	Actions           json.RawMessage `db:"actions"`
	IsActive          bool            `db:"is_active"`
	CreatedDate       time.Time       `db:"created_date"`
	UpdatedDate       time.Time       `db:"updated_date"`
	CreatedBy         string          `db:"created_by"`
	UpdatedBy         string          `db:"updated_by"`
	// Trigger type information
	TriggerTypeID          sql.NullString `db:"trigger_type_id"`
	TriggerTypeName        sql.NullString `db:"trigger_type_name"`
	TriggerTypeDescription sql.NullString `db:"trigger_type_description"`
	// Entity type information
	EntityTypeID          sql.NullString `db:"entity_type_id"`
	EntityTypeName        sql.NullString `db:"entity_type_name"`
	EntityTypeDescription sql.NullString `db:"entity_type_description"`
	// Entity information
	EntityName       sql.NullString `db:"entity_name"`
	EntitySchemaName sql.NullString `db:"entity_schema_name"`
}

// RuleActionView is a view with template information
type RuleActionView struct {
	ID                string          `db:"id"`
	AutomationRulesID sql.NullString  `db:"automation_rules_id"`
	Name              sql.NullString  `db:"name"`
	Description       sql.NullString  `db:"description"`
	ActionConfig      json.RawMessage `db:"action_config"`
	ExecutionOrder    sql.NullInt32   `db:"execution_order"`
	IsActive          sql.NullBool    `db:"is_active"`
	TemplateID        sql.NullString  `db:"template_id"`
	// Template information
	TemplateName          sql.NullString  `db:"template_name"`
	TemplateActionType    sql.NullString  `db:"template_action_type"`
	TemplateDefaultConfig json.RawMessage `db:"template_default_config"`
}
