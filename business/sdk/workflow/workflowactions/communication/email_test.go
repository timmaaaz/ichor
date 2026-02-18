package communication_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func Test_SendEmailHandler(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_SendEmailHandler")

	sd, err := insertEmailSeedData(db.BusDomain, db.Log)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, emailValidateTests(sd), "validate")
	unitest.Run(t, emailExecuteTests(sd), "execute")
}

// =============================================================================

type emailSeedData struct {
	unitest.SeedData
	Handler *communication.SendEmailHandler
	Mock    *communication.MockEmailClient
}

func insertEmailSeedData(_ dbtest.BusDomain, log *logger.Logger) (emailSeedData, error) {
	mock := &communication.MockEmailClient{}
	handler := communication.NewSendEmailHandler(log, nil, mock, "noreply@ichor.test")

	return emailSeedData{
		Handler: handler,
		Mock:    mock,
	}, nil
}

// =============================================================================
// Validate tests

func emailValidateTests(sd emailSeedData) []unitest.Table {
	return []unitest.Table{
		emailValidateValidConfig(sd),
		emailValidateMissingRecipients(sd),
		emailValidateMissingSubject(sd),
		emailValidateSimulateFailureSkipsValidation(sd),
		emailValidateInvalidJSON(sd),
	}
}

func emailValidateValidConfig(sd emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "valid_config",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{
				"recipients": ["user@example.com"],
				"subject": "Hello",
				"body": "World"
			}`)
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

func emailValidateMissingRecipients(sd emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_recipients",
		ExpResp: "email recipients list is required and must not be empty",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"recipients": [], "subject": "Hello"}`)
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

func emailValidateMissingSubject(sd emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_subject",
		ExpResp: "email subject is required",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"recipients": ["user@example.com"], "subject": ""}`)
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

func emailValidateSimulateFailureSkipsValidation(sd emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "simulate_failure_skips_validation",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			// simulate_failure=true should pass validation even with empty recipients/subject.
			config := json.RawMessage(`{"simulate_failure": true}`)
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

func emailValidateInvalidJSON(sd emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_json",
		ExpResp: true,
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
// Execute tests

func emailExecuteTests(sd emailSeedData) []unitest.Table {
	return []unitest.Table{
		emailExecuteSendsEmail(sd),
		emailExecuteTemplateSubstitution(sd),
		emailExecuteSimulateFailure(sd),
		emailExecuteNilClient(sd),
		emailExecuteClientError(sd),
	}
}

func emailExecuteSendsEmail(sd emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "sends_email",
		ExpResp: "sent",
		ExcFunc: func(ctx context.Context) any {
			sd.Mock.Reset()

			config := json.RawMessage(`{
				"recipients": ["alice@example.com", "bob@example.com"],
				"subject": "Order Confirmed",
				"body": "Your order is ready."
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				RuleName:    "Order Notify",
				ExecutionID: uuid.New(),
			}

			result, err := sd.Handler.Execute(ctx, config, execCtx)
			if err != nil {
				return err
			}

			// Verify the mock was called with correct args.
			if sd.Mock.CallCount() != 1 {
				return fmt.Sprintf("expected 1 Send call, got %d", sd.Mock.CallCount())
			}
			call := sd.Mock.SendCalls[0]
			if len(call.To) != 2 {
				return fmt.Sprintf("expected 2 recipients, got %d", len(call.To))
			}
			if call.Subject != "Order Confirmed" {
				return fmt.Sprintf("wrong subject: %s", call.Subject)
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

func emailExecuteTemplateSubstitution(sd emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "template_substitution",
		ExpResp: "Order ORD-999 is ready for customer CUST-42",
		ExcFunc: func(ctx context.Context) any {
			sd.Mock.Reset()

			config := json.RawMessage(`{
				"recipients": ["buyer@example.com"],
				"subject": "Order {{order_number}} Ready",
				"body": "Order {{order_number}} is ready for customer {{customer_id}}"
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityID:  uuid.New(),
				RuleName:  "Order Ready",
				RawData: map[string]interface{}{
					"order_number": "ORD-999",
					"customer_id":  "CUST-42",
				},
			}

			_, err := sd.Handler.Execute(ctx, config, execCtx)
			if err != nil {
				return err
			}

			if sd.Mock.CallCount() != 1 {
				return fmt.Sprintf("expected 1 Send call, got %d", sd.Mock.CallCount())
			}

			return sd.Mock.SendCalls[0].Body
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %q, want %q", got, exp)
			}
			return ""
		},
	}
}

func emailExecuteSimulateFailure(sd emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "simulate_failure",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			sd.Mock.Reset()

			config := json.RawMessage(`{
				"recipients": ["user@example.com"],
				"subject": "Test",
				"simulate_failure": true,
				"failure_message": "SMTP server unavailable"
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityID: uuid.New(),
				RuleName: "Failure Test",
			}

			_, err := sd.Handler.Execute(ctx, config, execCtx)
			if err == nil {
				return false
			}
			// Mock should NOT have been called — simulate_failure exits early.
			if sd.Mock.CallCount() != 0 {
				return false
			}
			return containsString(err.Error(), "SMTP server unavailable")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func emailExecuteNilClient(_ emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "nil_client_graceful_degradation",
		ExpResp: "sent",
		ExcFunc: func(ctx context.Context) any {
			// Handler with nil email client should still return success (log + skip).
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := communication.NewSendEmailHandler(log, nil, nil, "")

			config := json.RawMessage(`{
				"recipients": ["user@example.com"],
				"subject": "Hello",
				"body": "World"
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityID: uuid.New(),
				RuleName: "Nil Client Test",
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

func emailExecuteClientError(sd emailSeedData) unitest.Table {
	return unitest.Table{
		Name:    "client_error_propagated",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			sd.Mock.Reset()
			sd.Mock.SendErr = errors.New("API rate limit exceeded")

			config := json.RawMessage(`{
				"recipients": ["user@example.com"],
				"subject": "Hello",
				"body": "World"
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityID: uuid.New(),
				RuleName: "Error Test",
			}

			_, err := sd.Handler.Execute(ctx, config, execCtx)

			// Reset the error for subsequent tests.
			sd.Mock.SendErr = nil

			if err == nil {
				return false
			}
			return containsString(err.Error(), "API rate limit exceeded")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}
