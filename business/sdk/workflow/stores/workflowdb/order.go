package workflowdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

var automationRuleOrderByFields = map[string]string{
	workflow.OrderByID:          "ar.id",
	workflow.OrderByName:        "ar.name",
	workflow.OrderByCreatedDate: "ar.created_date",
	workflow.OrderByUpdatedDate: "ar.updated_date",
	workflow.OrderByIsActive:    "ar.is_active",
}

var executionOrderByFields = map[string]string{
	workflow.ExecutionOrderByID:         "ae.id",
	workflow.ExecutionOrderByExecutedAt: "ae.executed_at",
	workflow.ExecutionOrderByStatus:     "ae.status",
	workflow.ExecutionOrderByRuleID:     "ae.automation_rules_id",
	workflow.ExecutionOrderByEntityType: "ae.entity_type",
}

func orderByClauseAutomationRule(orderBy order.By) (string, error) {
	byField, exists := automationRuleOrderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + byField + " " + orderBy.Direction, nil
}

func orderByClauseExecution(orderBy order.By) (string, error) {
	byField, exists := executionOrderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + byField + " " + orderBy.Direction, nil
}
