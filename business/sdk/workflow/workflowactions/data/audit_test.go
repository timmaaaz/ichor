package data_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func Test_AuditLogAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_AuditLogAction")

	sd, err := insertAuditSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.AuditHandler = data.NewAuditLogHandler(log, db.DB)

	unitest.Run(t, auditLogActionTests(sd), "auditLogAction")
}

// =============================================================================

type auditSeedData struct {
	unitest.SeedData
	AuditHandler *data.AuditLogHandler
}

func insertAuditSeedData(busDomain dbtest.BusDomain) (auditSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return auditSeedData{}, fmt.Errorf("seeding users : %w", err)
	}
	adminUser := admins[0]

	return auditSeedData{
		SeedData: unitest.SeedData{
			Users:  []unitest.User{{User: adminUser}},
			Admins: []unitest.User{{User: adminUser}},
		},
	}, nil
}

// =============================================================================

func auditLogActionTests(sd auditSeedData) []unitest.Table {
	return []unitest.Table{
		logBasicEntry(sd),
		logWithTemplates(sd),
		logWithMetadata(sd),
		logWithRuleContext(sd),
		logWithoutMetadata(sd),
		auditValidateEmptyEntityName(sd),
		auditValidateEmptyAction(sd),
		auditValidateEmptyMessage(sd),
	}
}

func logBasicEntry(sd auditSeedData) unitest.Table {
	return unitest.Table{
		Name: "log_basic_entry",
		ExpResp: map[string]any{
			"status": "logged",
		},
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity_name": "sales.orders",
				"entity_id": "{{entity_id}}",
				"action": "status_change",
				"message": "Order status changed"
			}`)

			ruleID := uuid.New()
			execContext := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "sales.orders",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				RuleID:      &ruleID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.AuditHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["status"] != "logged" {
				return fmt.Sprintf("status mismatch: got %v, want logged", gotResp["status"])
			}

			if gotResp["audit_id"] == nil {
				return "expected audit_id, got nil"
			}

			return ""
		},
	}
}

func logWithTemplates(sd auditSeedData) unitest.Table {
	return unitest.Table{
		Name: "log_with_templates",
		ExpResp: map[string]any{
			"status": "logged",
		},
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity_name": "sales.orders",
				"entity_id": "{{entity_id}}",
				"action": "status_change",
				"message": "Order updated by {{user_id}}"
			}`)

			ruleID := uuid.New()
			execContext := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "sales.orders",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				RuleID:      &ruleID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.AuditHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["status"] != "logged" {
				return fmt.Sprintf("status mismatch: got %v, want logged", gotResp["status"])
			}

			return ""
		},
	}
}

func logWithMetadata(sd auditSeedData) unitest.Table {
	return unitest.Table{
		Name: "log_with_metadata",
		ExpResp: map[string]any{
			"status": "logged",
		},
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity_name": "sales.orders",
				"entity_id": "{{entity_id}}",
				"action": "approval",
				"message": "Order approved",
				"metadata": {
					"reason": "meets criteria",
					"approved_by": "{{user_id}}"
				}
			}`)

			ruleID := uuid.New()
			execContext := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "sales.orders",
				EventType:   "on_update",
				UserID:      sd.Admins[0].ID,
				RuleID:      &ruleID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.AuditHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["status"] != "logged" {
				return fmt.Sprintf("status mismatch: got %v, want logged", gotResp["status"])
			}

			return ""
		},
	}
}

func logWithRuleContext(sd auditSeedData) unitest.Table {
	return unitest.Table{
		Name: "log_with_rule_context",
		ExpResp: map[string]any{
			"status": "logged",
		},
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity_name": "sales.orders",
				"entity_id": "{{entity_id}}",
				"action": "automated_check",
				"message": "Automated compliance check passed"
			}`)

			ruleID := uuid.New()
			execContext := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "sales.orders",
				EventType:   "on_create",
				UserID:      sd.Admins[0].ID,
				RuleID:      &ruleID,
				RuleName:    "compliance-check-rule",
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.AuditHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["status"] != "logged" {
				return fmt.Sprintf("status mismatch: got %v, want logged", gotResp["status"])
			}

			return ""
		},
	}
}

func logWithoutMetadata(sd auditSeedData) unitest.Table {
	return unitest.Table{
		Name: "log_without_metadata",
		ExpResp: map[string]any{
			"status": "logged",
		},
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity_name": "sales.orders",
				"entity_id": "{{entity_id}}",
				"action": "simple_log",
				"message": "Simple audit entry"
			}`)

			ruleID := uuid.New()
			execContext := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "sales.orders",
				EventType:   "on_create",
				UserID:      sd.Admins[0].ID,
				RuleID:      &ruleID,
				ExecutionID: uuid.New(),
				Timestamp:   time.Now().UTC(),
			}

			result, err := sd.AuditHandler.Execute(ctx, config, execContext)
			if err != nil {
				return err
			}

			return result
		},
		CmpFunc: func(got any, exp any) string {
			gotResp, ok := got.(map[string]any)
			if !ok {
				return fmt.Sprintf("error occurred or wrong type: %v", got)
			}

			if gotResp["status"] != "logged" {
				return fmt.Sprintf("status mismatch: got %v, want logged", gotResp["status"])
			}

			return ""
		},
	}
}

func auditValidateEmptyEntityName(sd auditSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_empty_entity_name",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity_name": "",
				"action": "test",
				"message": "test"
			}`)

			err := sd.AuditHandler.Validate(config)
			if err == nil {
				return "expected validation error, got nil"
			}

			return "error"
		},
		CmpFunc: func(got any, exp any) string {
			if got != "error" {
				return fmt.Sprintf("expected validation error: %v", got)
			}
			return ""
		},
	}
}

func auditValidateEmptyAction(sd auditSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_empty_action",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity_name": "sales.orders",
				"action": "",
				"message": "test"
			}`)

			err := sd.AuditHandler.Validate(config)
			if err == nil {
				return "expected validation error, got nil"
			}

			return "error"
		},
		CmpFunc: func(got any, exp any) string {
			if got != "error" {
				return fmt.Sprintf("expected validation error: %v", got)
			}
			return ""
		},
	}
}

func auditValidateEmptyMessage(sd auditSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_empty_message",
		ExpResp: "error",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"entity_name": "sales.orders",
				"action": "test",
				"message": ""
			}`)

			err := sd.AuditHandler.Validate(config)
			if err == nil {
				return "expected validation error, got nil"
			}

			return "error"
		},
		CmpFunc: func(got any, exp any) string {
			if got != "error" {
				return fmt.Sprintf("expected validation error: %v", got)
			}
			return ""
		},
	}
}
