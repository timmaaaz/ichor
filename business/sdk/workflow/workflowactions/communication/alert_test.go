package communication_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func Test_CreateAlertHandler(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_CreateAlertHandler")

	sd, err := insertAlertSeedData(db.BusDomain, db.Log)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, validateTests(sd), "validate")
	unitest.Run(t, executeTests(db.BusDomain, sd), "execute")
}

// =============================================================================

type alertSeedData struct {
	unitest.SeedData
	Handler *communication.CreateAlertHandler
	UserID  uuid.UUID
	RoleID  uuid.UUID
}

func insertAlertSeedData(busDomain dbtest.BusDomain, log *logger.Logger) (alertSeedData, error) {
	ctx := context.Background()

	// Seed a user (required for recipients)
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return alertSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	testUser := users[0]

	// Create handler with the test logger (nil workflowQueue for unit tests)
	handler := communication.NewCreateAlertHandler(log, busDomain.Alert, nil)

	return alertSeedData{
		SeedData: unitest.SeedData{
			Users: []unitest.User{{User: testUser}},
		},
		Handler: handler,
		UserID:  testUser.ID,
		RoleID:  uuid.New(), // Roles don't have FK validation in recipients
	}, nil
}

// =============================================================================
// Validate Tests

func validateTests(sd alertSeedData) []unitest.Table {
	return []unitest.Table{
		validateValidConfig(sd),
		validateMissingMessage(sd),
		validateMissingRecipients(sd),
		validateInvalidSeverity(sd),
		validateInvalidJSON(sd),
	}
}

func validateValidConfig(sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "valid_config",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "test_alert",
				"severity": "high",
				"title": "Test Alert",
				"message": "This is a test alert message",
				"recipients": {
					"users": ["%s"],
					"roles": ["%s"]
				}
			}`, sd.UserID, sd.RoleID))

			return sd.Handler.Validate(config)
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("expected nil error, got: %v", got)
			}
			return ""
		},
	}
}

func validateMissingMessage(sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_message",
		ExpResp: "alert message is required",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "test_alert",
				"severity": "high",
				"title": "Test Alert",
				"message": "",
				"recipients": {
					"users": ["%s"]
				}
			}`, sd.UserID))

			err := sd.Handler.Validate(config)
			if err != nil {
				return err.Error()
			}
			return nil
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func validateMissingRecipients(sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_recipients",
		ExpResp: "at least one recipient is required",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"alert_type": "test_alert",
				"severity": "medium",
				"title": "Test Alert",
				"message": "This is a test alert",
				"recipients": {
					"users": [],
					"roles": []
				}
			}`)

			err := sd.Handler.Validate(config)
			if err != nil {
				return err.Error()
			}
			return nil
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func validateInvalidSeverity(sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_severity",
		ExpResp: "invalid severity level: extreme",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "test_alert",
				"severity": "extreme",
				"title": "Test Alert",
				"message": "This is a test alert",
				"recipients": {
					"users": ["%s"]
				}
			}`, sd.UserID))

			err := sd.Handler.Validate(config)
			if err != nil {
				return err.Error()
			}
			return nil
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func validateInvalidJSON(sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_json",
		ExpResp: true, // We expect an error containing "invalid configuration format"
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{not valid json}`)

			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return containsString(err.Error(), "invalid configuration format")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

// =============================================================================
// Execute Tests

func executeTests(busDomain dbtest.BusDomain, sd alertSeedData) []unitest.Table {
	return []unitest.Table{
		executeCreatesAlert(busDomain, sd),
		executeCreatesRecipients(busDomain, sd),
		executeTemplateSubstitution(busDomain, sd),
		executeDefaultSeverity(busDomain, sd),
		executeInvalidUserUUID(sd),
		executeInvalidRoleUUID(sd),
		executeResolvePrior(busDomain, sd),
	}
}

func executeCreatesAlert(busDomain dbtest.BusDomain, sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "creates_alert",
		ExpResp: "created",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "order_failed",
				"severity": "high",
				"title": "Order Processing Failed",
				"message": "Failed to process order",
				"recipients": {
					"users": ["%s"]
				}
			}`, sd.UserID))

			// Use uuid.Nil for RuleID to avoid FK constraint on automation_rules
			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				UserID:      sd.UserID,
				RuleID:      nil, // nil for test executions
				RuleName:    "Test Rule",
				ExecutionID: uuid.New(),
			}

			result, err := sd.Handler.Execute(ctx, config, execCtx)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				return "unexpected result type"
			}

			return resultMap["status"]
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeCreatesRecipients(busDomain dbtest.BusDomain, sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "creates_recipients",
		ExpResp: "created",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "multi_recipient",
				"severity": "medium",
				"title": "Multi-recipient Alert",
				"message": "Alert for both user and role",
				"recipients": {
					"users": ["%s"],
					"roles": ["%s"]
				}
			}`, sd.UserID, sd.RoleID))

			// Use uuid.Nil for RuleID to avoid FK constraint on automation_rules
			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_update",
				UserID:      sd.UserID,
				RuleID:      nil, // nil for test executions
				RuleName:    "Multi-recipient Rule",
				ExecutionID: uuid.New(),
			}

			result, err := sd.Handler.Execute(ctx, config, execCtx)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				return "unexpected result type"
			}

			alertIDStr, ok := resultMap["alert_id"].(string)
			if !ok {
				return "missing alert_id in result"
			}

			alertID, err := uuid.Parse(alertIDStr)
			if err != nil {
				return fmt.Errorf("invalid alert_id: %w", err)
			}

			// Query the alert to verify it was created
			alert, err := busDomain.Alert.QueryByID(ctx, alertID)
			if err != nil {
				return err
			}

			// Verify the alert is the one we created
			if alert.Title != "Multi-recipient Alert" {
				return fmt.Sprintf("wrong alert title: %s", alert.Title)
			}

			// Verify this user can see the alert via QueryMine (recipient was created)
			userAlerts, err := busDomain.Alert.QueryMine(ctx, sd.UserID, []uuid.UUID{sd.RoleID}, alertbus.QueryFilter{ID: &alertID}, order.NewBy(alertbus.OrderByCreatedDate, order.DESC), page.MustParse("1", "10"))
			if err != nil {
				return err
			}

			if len(userAlerts) == 0 {
				return "user cannot see their own alert - recipient not created"
			}

			return resultMap["status"]
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeTemplateSubstitution(busDomain dbtest.BusDomain, sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "template_substitution",
		ExpResp: "Order ORD-12345 failed for customer CUST-001",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "template_test",
				"severity": "low",
				"title": "Order {{order_number}} Issue",
				"message": "Order {{order_number}} failed for customer {{customer_name}}",
				"recipients": {
					"users": ["%s"]
				}
			}`, sd.UserID))

			// Use uuid.Nil for RuleID to avoid FK constraint on automation_rules
			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				UserID:      sd.UserID,
				RuleID:      nil, // nil for test executions
				RuleName:    "Template Test Rule",
				ExecutionID: uuid.New(),
				RawData: map[string]interface{}{
					"order_number":  "ORD-12345",
					"customer_name": "CUST-001",
				},
			}

			result, err := sd.Handler.Execute(ctx, config, execCtx)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				return "unexpected result type"
			}

			alertIDStr, ok := resultMap["alert_id"].(string)
			if !ok {
				return "missing alert_id"
			}

			alertID, err := uuid.Parse(alertIDStr)
			if err != nil {
				return err
			}

			// Query the alert to verify template substitution
			alert, err := busDomain.Alert.QueryByID(ctx, alertID)
			if err != nil {
				return err
			}

			return alert.Message
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeDefaultSeverity(busDomain dbtest.BusDomain, sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "default_severity",
		ExpResp: alertbus.SeverityMedium,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "no_severity",
				"title": "No Severity Alert",
				"message": "Alert without severity specified",
				"recipients": {
					"users": ["%s"]
				}
			}`, sd.UserID))

			// Use uuid.Nil for RuleID to avoid FK constraint on automation_rules
			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				UserID:      sd.UserID,
				RuleID:      nil, // nil for test executions
				RuleName:    "Default Severity Rule",
				ExecutionID: uuid.New(),
			}

			result, err := sd.Handler.Execute(ctx, config, execCtx)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				return "unexpected result type"
			}

			alertIDStr, ok := resultMap["alert_id"].(string)
			if !ok {
				return "missing alert_id"
			}

			alertID, err := uuid.Parse(alertIDStr)
			if err != nil {
				return err
			}

			alert, err := busDomain.Alert.QueryByID(ctx, alertID)
			if err != nil {
				return err
			}

			return alert.Severity
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeInvalidUserUUID(sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_user_uuid",
		ExpResp: true, // Expect error containing "invalid user UUID"
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"alert_type": "test",
				"message": "Test message",
				"recipients": {
					"users": ["not-a-valid-uuid"]
				}
			}`)

			invalidUserRuleID := uuid.New()
			execCtx := workflow.ActionExecutionContext{
				EntityID:      uuid.New(),
				EntityName:    "orders",
				EventType:     "on_create",
				UserID:        sd.UserID,
				RuleID:        &invalidUserRuleID,
				RuleName:      "Invalid UUID Rule",
				ExecutionID:   uuid.New(),
				TriggerSource: workflow.TriggerSourceAutomation,
			}

			_, err := sd.Handler.Execute(ctx, config, execCtx)
			if err == nil {
				return false
			}
			return containsString(err.Error(), "invalid user UUID")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeInvalidRoleUUID(sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_role_uuid",
		ExpResp: true, // Expect error containing "invalid role UUID"
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "test",
				"message": "Test message",
				"recipients": {
					"users": ["%s"],
					"roles": ["not-a-valid-uuid"]
				}
			}`, sd.UserID))

			invalidRoleRuleID := uuid.New()
			execCtx := workflow.ActionExecutionContext{
				EntityID:      uuid.New(),
				EntityName:    "orders",
				EventType:     "on_create",
				UserID:        sd.UserID,
				RuleID:        &invalidRoleRuleID,
				RuleName:      "Invalid Role UUID Rule",
				ExecutionID:   uuid.New(),
				TriggerSource: workflow.TriggerSourceAutomation,
			}

			_, err := sd.Handler.Execute(ctx, config, execCtx)
			if err == nil {
				return false
			}
			return containsString(err.Error(), "invalid role UUID")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeResolvePrior(busDomain dbtest.BusDomain, sd alertSeedData) unitest.Table {
	return unitest.Table{
		Name:    "resolve_prior",
		ExpResp: 2, // Expect 2 prior alerts to be resolved
		ExcFunc: func(ctx context.Context) any {
			// Use a consistent entity ID for all alerts in this test
			sourceEntityID := uuid.New()
			alertType := "allocation_status"

			// Create two "failure" alerts first (these should get resolved)
			failureConfig := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "%s",
				"severity": "high",
				"title": "Allocation Failed",
				"message": "Failed to allocate inventory",
				"recipients": {
					"users": ["%s"]
				},
				"resolve_prior": false
			}`, alertType, sd.UserID))

			failureExecCtx := workflow.ActionExecutionContext{
				EntityID:    sourceEntityID,
				EntityName:  "sales_order",
				EventType:   "on_update",
				UserID:      sd.UserID,
				RuleID:      nil, // nil for test executions
				RuleName:    "Allocation Failure Rule",
				ExecutionID: uuid.New(),
			}

			// Create first failure alert
			_, err := sd.Handler.Execute(ctx, failureConfig, failureExecCtx)
			if err != nil {
				return fmt.Errorf("create failure alert 1: %w", err)
			}

			// Create second failure alert
			failureExecCtx.ExecutionID = uuid.New()
			_, err = sd.Handler.Execute(ctx, failureConfig, failureExecCtx)
			if err != nil {
				return fmt.Errorf("create failure alert 2: %w", err)
			}

			// Now create a "success" alert with resolve_prior: true
			successConfig := json.RawMessage(fmt.Sprintf(`{
				"alert_type": "%s",
				"severity": "low",
				"title": "Allocation Successful",
				"message": "Successfully allocated inventory",
				"recipients": {
					"users": ["%s"]
				},
				"resolve_prior": true
			}`, alertType, sd.UserID))

			successExecCtx := workflow.ActionExecutionContext{
				EntityID:    sourceEntityID,
				EntityName:  "sales_order",
				EventType:   "on_update",
				UserID:      sd.UserID,
				RuleID:      nil, // nil for test executions
				RuleName:    "Allocation Success Rule",
				ExecutionID: uuid.New(),
			}

			result, err := sd.Handler.Execute(ctx, successConfig, successExecCtx)
			if err != nil {
				return fmt.Errorf("create success alert: %w", err)
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				return "unexpected result type"
			}

			// The handler returns resolved_count in its result
			resolvedCount, ok := resultMap["resolved_count"].(int)
			if !ok {
				return fmt.Sprintf("resolved_count not found or wrong type: %v", resultMap["resolved_count"])
			}

			return resolvedCount
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v resolved alerts, want %v", got, exp)
			}
			return ""
		},
	}
}

// =============================================================================
// Helpers

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
