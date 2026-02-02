package executionapi

import "github.com/timmaaaz/ichor/business/sdk/workflow"

// orderByFields maps API field names to business layer order constants.
// Both "rule_id" (API-friendly) and "automation_rules_id" (DB column) are accepted for consistency.
var orderByFields = map[string]string{
	"id":                  workflow.ExecutionOrderByID,
	"executed_at":         workflow.ExecutionOrderByExecutedAt,
	"status":              workflow.ExecutionOrderByStatus,
	"rule_id":             workflow.ExecutionOrderByRuleID, // API-friendly alias
	"automation_rules_id": workflow.ExecutionOrderByRuleID, // DB column name (backward compat)
	"entity_type":         workflow.ExecutionOrderByEntityType,
}
