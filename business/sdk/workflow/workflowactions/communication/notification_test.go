package communication_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func Test_SendNotificationHandler(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "notification_test", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })

	handler := communication.NewSendNotificationHandler(log, nil, nil)

	unitest.Run(t, notificationValidateTests(handler), "validate")
	unitest.Run(t, notificationExecuteTests(handler), "execute")
}

func Test_SendNotification_NilBus_Skips(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "notif", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })
	handler := communication.NewSendNotificationHandler(log, nil, nil)
	cfg := []byte(`{"recipients":["` + uuid.NewString() + `"],"priority":"low","message":"hi"}`)
	out, err := handler.Execute(context.Background(), cfg, workflow.ActionExecutionContext{RawData: map[string]any{}})
	require.NoError(t, err)
	require.Equal(t, "skipped", out.(map[string]interface{})["status"])
}

// =============================================================================
// Validate Tests

func notificationValidateTests(handler *communication.SendNotificationHandler) []unitest.Table {
	return []unitest.Table{
		notifyValidateValidConfig(handler),
		notifyValidateMissingMessage(handler),
		notifyValidateMissingRecipients(handler),
		notifyValidateInvalidPriority(handler),
		notifyValidateInvalidJSON(handler),
		notifyValidateInvalidRecipientUUID(handler),
	}
}

func notifyValidateValidConfig(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "valid_config",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"recipients": ["%s"],
				"priority": "high",
				"message": "Test notification message"
			}`, uuid.New()))

			return handler.Validate(config)
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("expected nil error, got: %v", got)
			}
			return ""
		},
	}
}

func notifyValidateMissingMessage(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "missing_message",
		ExpResp: "notification message is required",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"recipients": ["%s"],
				"priority": "high",
				"message": ""
			}`, uuid.New()))

			err := handler.Validate(config)
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

func notifyValidateMissingRecipients(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "missing_recipients",
		ExpResp: "recipients list is required and must not be empty",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"recipients": [],
				"priority": "medium",
				"message": "Test notification"
			}`)

			err := handler.Validate(config)
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

func notifyValidateInvalidPriority(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "invalid_priority",
		ExpResp: "invalid priority level",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"recipients": ["%s"],
				"priority": "extreme",
				"message": "Test notification"
			}`, uuid.New()))

			err := handler.Validate(config)
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

func notifyValidateInvalidJSON(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "invalid_json",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{not valid json}`)

			err := handler.Validate(config)
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

func notifyValidateInvalidRecipientUUID(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "invalid_recipient_uuid",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"recipients": ["not-a-uuid"],
				"priority": "medium",
				"message": "Test notification"
			}`)

			err := handler.Validate(config)
			if err == nil {
				return false
			}
			return containsString(err.Error(), "invalid recipient UUID")
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
//
// All execute tests use a nil-bus handler (no alertBus wired), so Execute always
// returns status "skipped" via the graceful-degradation path. The DB-backed
// integration test (status "sent" + real alert row) lives in a separate package
// and is added in the next task once call sites are updated.

func notificationExecuteTests(handler *communication.SendNotificationHandler) []unitest.Table {
	return []unitest.Table{
		notifyExecuteBasic(handler),
		notifyExecuteTemplateSubstitution(handler),
		notifyExecuteResultFields(handler),
		notifyExecuteInvalidRecipientUUID(handler),
	}
}

func notifyExecuteBasic(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "basic_execution",
		ExpResp: "skipped",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"recipients": ["%s"],
				"priority": "medium",
				"message": "Order has been processed",
				"title": "Order Update"
			}`, uuid.New()))

			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				UserID:      uuid.New(),
				ExecutionID: uuid.New(),
			}

			result, err := handler.Execute(ctx, config, execCtx)
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

func notifyExecuteTemplateSubstitution(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "template_substitution",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(fmt.Sprintf(`{
				"recipients": ["%s"],
				"priority": "high",
				"message": "Order {{order_number}} is ready",
				"title": "Order {{order_number}} Status"
			}`, uuid.New()))

			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_update",
				UserID:      uuid.New(),
				ExecutionID: uuid.New(),
				RawData: map[string]interface{}{
					"order_number": "ORD-99999",
				},
			}

			result, err := handler.Execute(ctx, config, execCtx)
			if err != nil {
				return false
			}

			// With nil alertBus the handler returns graceful-degradation skipped output.
			// Verify the result has the expected fields.
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				return false
			}

			_, hasID := resultMap["notification_id"]
			_, hasStatus := resultMap["status"]
			return hasID && hasStatus
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func notifyExecuteResultFields(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "result_fields",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			recipientA := uuid.New()
			recipientB := uuid.New()

			config := json.RawMessage(fmt.Sprintf(`{
				"recipients": ["%s", "%s"],
				"priority": "low",
				"message": "Bulk notification"
			}`, recipientA, recipientB))

			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "inventory",
				EventType:   "on_update",
				UserID:      uuid.New(),
				ExecutionID: uuid.New(),
			}

			result, err := handler.Execute(ctx, config, execCtx)
			if err != nil {
				return false
			}

			resultMap, ok := result.(map[string]interface{})
			if !ok {
				return false
			}

			// With nil alertBus, degraded output: notification_id=uuid.Nil, status="skipped", recipients=0.
			idStr, ok := resultMap["notification_id"].(string)
			if !ok {
				return false
			}
			if _, err := uuid.Parse(idStr); err != nil {
				return false
			}

			if resultMap["status"] != "skipped" {
				return false
			}

			// recipients count is 0 in skipped mode (no alertBus to persist them).
			if resultMap["recipients"] != 0 {
				return false
			}

			return true
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func notifyExecuteInvalidRecipientUUID(handler *communication.SendNotificationHandler) unitest.Table {
	return unitest.Table{
		Name:    "invalid_recipient_uuid_errors",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			// The new alertbus-based handler validates all UUIDs before persisting
			// (fail-fast). An invalid UUID now returns an error rather than skipping.
			config := json.RawMessage(fmt.Sprintf(`{
				"recipients": ["not-a-uuid", "%s"],
				"priority": "medium",
				"message": "Notification with mixed recipients"
			}`, uuid.New()))

			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				UserID:      uuid.New(),
				ExecutionID: uuid.New(),
			}

			_, err := handler.Execute(ctx, config, execCtx)
			return err != nil
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}
