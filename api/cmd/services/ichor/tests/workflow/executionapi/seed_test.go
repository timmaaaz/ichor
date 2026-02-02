package execution_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/executionapi"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// ExecutionSeedData holds test data for execution API tests.
type ExecutionSeedData struct {
	apitest.SeedData
	TriggerTypes []workflow.TriggerType
	EntityTypes  []workflow.EntityType
	Entities     []workflow.Entity
	Rules        []workflow.AutomationRule
	Actions      []workflow.RuleAction
	Executions   []workflow.AutomationExecution
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (ExecutionSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Create admin user with full permissions
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return ExecutionSeedData{}, fmt.Errorf("seeding admin users: %w", err)
	}
	adminUser := apitest.User{
		User:  admins[0],
		Token: apitest.Token(db.BusDomain.User, ath, admins[0].Email.Address),
	}

	// Seed trigger types
	triggerTypes, err := workflow.TestSeedTriggerTypes(ctx, 3, busDomain.Workflow)
	if err != nil {
		return ExecutionSeedData{}, fmt.Errorf("seeding trigger types: %w", err)
	}

	// Get entity types from migrations
	entityTypes, err := workflow.GetEntityTypes(ctx, busDomain.Workflow)
	if err != nil {
		return ExecutionSeedData{}, fmt.Errorf("getting entity types: %w", err)
	}

	// Get entities from migrations
	entities, err := workflow.GetEntities(ctx, busDomain.Workflow)
	if err != nil {
		return ExecutionSeedData{}, fmt.Errorf("getting entities: %w", err)
	}

	// Seed automation rules
	entityIDs := make([]uuid.UUID, len(entities))
	for i, e := range entities {
		entityIDs[i] = e.ID
	}
	entityTypeIDs := make([]uuid.UUID, len(entityTypes))
	for i, et := range entityTypes {
		entityTypeIDs[i] = et.ID
	}
	triggerTypeIDs := make([]uuid.UUID, len(triggerTypes))
	for i, tt := range triggerTypes {
		triggerTypeIDs[i] = tt.ID
	}

	rules, err := workflow.TestSeedAutomationRules(ctx, 2, entityIDs, entityTypeIDs, triggerTypeIDs, admins[0].ID, busDomain.Workflow)
	if err != nil {
		return ExecutionSeedData{}, fmt.Errorf("seeding rules: %w", err)
	}

	// Seed rule actions
	ruleIDs := make([]uuid.UUID, len(rules))
	for i, r := range rules {
		ruleIDs[i] = r.ID
	}

	actions, err := workflow.TestSeedRuleActions(ctx, 3, ruleIDs, nil, busDomain.Workflow)
	if err != nil {
		return ExecutionSeedData{}, fmt.Errorf("seeding actions: %w", err)
	}

	// Seed executions for testing
	executions, err := seedExecutions(ctx, rules, busDomain.Workflow)
	if err != nil {
		return ExecutionSeedData{}, fmt.Errorf("seeding executions: %w", err)
	}

	// =========================================================================
	// Table Permissions
	// =========================================================================

	// Create a test role
	roles, err := rolebus.TestSeedRoles(ctx, 1, busDomain.Role)
	if err != nil {
		return ExecutionSeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	// Assign admin user to role
	_, err = userrolebus.TestSeedUserRoles(ctx, uuid.UUIDs{admins[0].ID}, uuid.UUIDs{roles[0].ID}, busDomain.UserRole)
	if err != nil {
		return ExecutionSeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	// Grant access to both rule and execution tables
	tables := []string{
		ruleapi.RouteTable,                     // "workflow.automation_rules"
		executionapi.RouteTable,                // "workflow.automation_executions"
	}
	for _, table := range tables {
		_, err = busDomain.TableAccess.Create(ctx, tableaccessbus.NewTableAccess{
			RoleID:    roles[0].ID,
			TableName: table,
			CanCreate: true,
			CanRead:   true,
			CanUpdate: true,
			CanDelete: true,
		})
		if err != nil {
			return ExecutionSeedData{}, fmt.Errorf("creating table access for %s: %w", table, err)
		}
	}

	return ExecutionSeedData{
		SeedData: apitest.SeedData{
			Users: []apitest.User{adminUser},
		},
		TriggerTypes: triggerTypes,
		EntityTypes:  entityTypes,
		Entities:     entities,
		Rules:        rules,
		Actions:      actions,
		Executions:   executions,
	}, nil
}

// seedExecutions creates test execution records
func seedExecutions(ctx context.Context, rules []workflow.AutomationRule, bus *workflow.Business) ([]workflow.AutomationExecution, error) {
	if len(rules) == 0 {
		return []workflow.AutomationExecution{}, nil
	}

	var executions []workflow.AutomationExecution

	// Create a successful execution for the first rule
	triggerData, _ := json.Marshal(map[string]any{
		"entity_id": "test-entity-123",
		"status":    "shipped",
		"total":     100.50,
	})

	actionsExecuted, _ := json.Marshal([]map[string]any{
		{
			"action_id":   uuid.New().String(),
			"action_name": "Send Email",
			"action_type": "send_email",
			"status":      "completed",
			"duration_ms": 150,
		},
	})

	exec1, err := bus.CreateExecution(ctx, workflow.NewAutomationExecution{
		AutomationRuleID: &rules[0].ID,
		EntityType:       "orders",
		TriggerData:      triggerData,
		ActionsExecuted:  actionsExecuted,
		Status:           workflow.StatusCompleted,
		ErrorMessage:     "",
		ExecutionTimeMs:  250,
		TriggerSource:    workflow.TriggerSourceAutomation,
	})
	if err != nil {
		return nil, fmt.Errorf("creating execution 1: %w", err)
	}
	executions = append(executions, exec1)

	// Create a failed execution for the first rule
	exec2, err := bus.CreateExecution(ctx, workflow.NewAutomationExecution{
		AutomationRuleID: &rules[0].ID,
		EntityType:       "orders",
		TriggerData:      triggerData,
		ActionsExecuted:  json.RawMessage("[]"),
		Status:           workflow.StatusFailed,
		ErrorMessage:     "Email server unavailable",
		ExecutionTimeMs:  50,
		TriggerSource:    workflow.TriggerSourceAutomation,
	})
	if err != nil {
		return nil, fmt.Errorf("creating execution 2: %w", err)
	}
	executions = append(executions, exec2)

	// Create a manual execution (no rule ID)
	exec3, err := bus.CreateExecution(ctx, workflow.NewAutomationExecution{
		AutomationRuleID: nil, // Manual execution - no associated rule
		EntityType:       "manual",
		TriggerData:      json.RawMessage(`{"action": "test_notification"}`),
		ActionsExecuted:  json.RawMessage("[]"),
		Status:           workflow.StatusCompleted,
		ExecutionTimeMs:  100,
		TriggerSource:    workflow.TriggerSourceManual,
		ActionType:       "send_notification",
	})
	if err != nil {
		return nil, fmt.Errorf("creating execution 3: %w", err)
	}
	executions = append(executions, exec3)

	return executions, nil
}
