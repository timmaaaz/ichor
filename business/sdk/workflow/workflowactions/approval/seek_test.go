package approval_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func Test_SeekApprovalHandler(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_SeekApprovalHandler")

	sd, err := insertSeekSeedData(db)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, validateTests(sd), "validate")
	unitest.Run(t, startAsyncTests(sd), "start-async")
	unitest.Run(t, executeTests(sd), "execute")
	unitest.Run(t, nilBusTests(), "nil-bus")
}

// =============================================================================

type seekSeedData struct {
	unitest.SeedData
	Handler            *approval.SeekApprovalHandler
	ApprovalRequestBus *approvalrequestbus.Business
	AlertBus           *alertbus.Business
	Approvers          []uuid.UUID
}

func insertSeekSeedData(db *dbtest.Database) (seekSeedData, error) {
	approverA := uuid.New()
	approverB := uuid.New()

	return seekSeedData{
		Handler:            approval.NewSeekApprovalHandler(db.Log, db.DB, nil, db.BusDomain.Alert),
		ApprovalRequestBus: nil,
		AlertBus:           db.BusDomain.Alert,
		Approvers:          []uuid.UUID{approverA, approverB},
	}, nil
}

// =============================================================================
// Validate tests

func validateTests(sd seekSeedData) []unitest.Table {
	return []unitest.Table{
		validateValidConfig(sd),
		validateMissingApprovers(sd),
		validateEmptyApprovers(sd),
		validateInvalidApprovalType(sd),
		validateInvalidJSON(sd),
	}
}

func validateValidConfig(_ seekSeedData) unitest.Table {
	return unitest.Table{
		Name:    "valid_config",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := approval.NewSeekApprovalHandler(log, nil, nil, nil)

			config := json.RawMessage(`{
				"approvers": ["` + uuid.New().String() + `"],
				"approval_type": "any",
				"approval_message": "Please approve this"
			}`)
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

func validateMissingApprovers(_ seekSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_approvers",
		ExpResp: "approvers list is required and must not be empty",
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := approval.NewSeekApprovalHandler(log, nil, nil, nil)

			config := json.RawMessage(`{
				"approval_type": "any",
				"approval_message": "Please approve"
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

func validateEmptyApprovers(_ seekSeedData) unitest.Table {
	return unitest.Table{
		Name:    "empty_approvers",
		ExpResp: "approvers list is required and must not be empty",
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := approval.NewSeekApprovalHandler(log, nil, nil, nil)

			config := json.RawMessage(`{
				"approvers": [],
				"approval_type": "any",
				"approval_message": "Please approve"
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

func validateInvalidApprovalType(_ seekSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_approval_type",
		ExpResp: "invalid approval_type, must be: any, all, or majority",
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := approval.NewSeekApprovalHandler(log, nil, nil, nil)

			config := json.RawMessage(`{
				"approvers": ["` + uuid.New().String() + `"],
				"approval_type": "bogus",
				"approval_message": "Please approve"
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

func validateInvalidJSON(_ seekSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_json",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := approval.NewSeekApprovalHandler(log, nil, nil, nil)

			config := json.RawMessage(`{not valid json}`)
			err := handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "invalid configuration format")
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
// StartAsync tests

func startAsyncTests(sd seekSeedData) []unitest.Table {
	return []unitest.Table{
		startAsyncNilBusReturnsError(sd),
		startAsyncInvalidApproverUUID(sd),
	}
}

func startAsyncNilBusReturnsError(_ seekSeedData) unitest.Table {
	return unitest.Table{
		Name:    "nil_bus_returns_error",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := approval.NewSeekApprovalHandler(log, nil, nil, nil)

			config := json.RawMessage(`{
				"approvers": ["` + uuid.New().String() + `"],
				"approval_type": "any",
				"approval_message": "Please approve"
			}`)

			ruleID := uuid.New()
			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				RuleID:      &ruleID,
				RuleName:    "Test Rule",
				ActionName:  "seek_approval_0",
				ExecutionID: uuid.New(),
			}

			taskToken := []byte("test-task-token")
			err := handler.StartAsync(ctx, config, execCtx, taskToken)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "not available in core registration")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func startAsyncInvalidApproverUUID(_ seekSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_approver_uuid",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			// Need non-nil buses to get past the nil check, but they can be real.
			// For this test, we just need to verify UUID parsing fails first.
			// Use a handler with non-nil bus stubs — but since we can't easily mock,
			// we use a real handler that will fail on UUID parse before DB call.
			mockApprovalBus := approvalrequestbus.NewBusiness(log, &noopApprovalStorer{})
			mockAlertBus := alertbus.NewBusiness(log, &noopAlertStorer{})
			handler := approval.NewSeekApprovalHandler(log, nil, mockApprovalBus, mockAlertBus)

			config := json.RawMessage(`{
				"approvers": ["not-a-uuid"],
				"approval_type": "any",
				"approval_message": "Please approve"
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				ActionName:  "seek_approval_0",
				ExecutionID: uuid.New(),
			}

			err := handler.StartAsync(ctx, config, execCtx, []byte("token"))
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "invalid approver UUID")
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

func executeTests(sd seekSeedData) []unitest.Table {
	return []unitest.Table{
		executeNilBusGracefulDegradation(sd),
		executeInvalidApproverUUID(sd),
	}
}

func executeNilBusGracefulDegradation(_ seekSeedData) unitest.Table {
	return unitest.Table{
		Name:    "nil_bus_graceful_degradation",
		ExpResp: "approved",
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := approval.NewSeekApprovalHandler(log, nil, nil, nil)

			config := json.RawMessage(`{
				"approvers": ["` + uuid.New().String() + `"],
				"approval_type": "any",
				"approval_message": "Please approve"
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				RuleName:    "Test Rule",
				ActionName:  "seek_approval_0",
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
			return resultMap["output"]
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeInvalidApproverUUID(_ seekSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_approver_uuid",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			mockApprovalBus := approvalrequestbus.NewBusiness(log, &noopApprovalStorer{})
			handler := approval.NewSeekApprovalHandler(log, nil, mockApprovalBus, nil)

			config := json.RawMessage(`{
				"approvers": ["not-a-uuid"],
				"approval_type": "any",
				"approval_message": "Please approve"
			}`)

			execCtx := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				ActionName:  "seek_approval_0",
				ExecutionID: uuid.New(),
			}

			_, err := handler.Execute(ctx, config, execCtx)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "invalid approver UUID")
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
// Nil bus tests — verify handler properties and GetType

func nilBusTests() []unitest.Table {
	return []unitest.Table{
		handlerProperties(),
		handlerOutputPorts(),
	}
}

func handlerProperties() unitest.Table {
	return unitest.Table{
		Name:    "handler_properties",
		ExpResp: "seek_approval|true|true",
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := approval.NewSeekApprovalHandler(log, nil, nil, nil)

			return fmt.Sprintf("%s|%v|%v", handler.GetType(), handler.IsAsync(), handler.SupportsManualExecution())
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func handlerOutputPorts() unitest.Table {
	return unitest.Table{
		Name:    "output_ports",
		ExpResp: 3,
		ExcFunc: func(ctx context.Context) any {
			log := logger.New(io.Discard, logger.LevelInfo, "test", nil)
			handler := approval.NewSeekApprovalHandler(log, nil, nil, nil)
			ports := handler.GetOutputPorts()
			return len(ports)
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v ports, want %v", got, exp)
			}
			return ""
		},
	}
}

// =============================================================================
// StartAsync with real DB — tests that actually create approval requests

func Test_SeekApprovalHandler_StartAsync_WithDB(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_SeekApprovalHandler_StartAsync")
	ctx := context.Background()

	// Seed a user for FK references.
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		t.Fatalf("Seeding users: %s", err)
	}
	userID := users[0].ID

	// Seed full workflow data for FK constraints (execution_id, rule_id).
	wfData, err := workflow.TestSeedFullWorkflow(ctx, userID, db.BusDomain.Workflow)
	if err != nil {
		t.Fatalf("Seeding workflow data: %s", err)
	}

	ruleID := wfData.AutomationRules[0].ID

	approverA := uuid.New()
	approverB := uuid.New()

	approvalStore := approvalrequestbus.NewBusiness(db.Log, newApprovalRequestStore(db))
	handler := approval.NewSeekApprovalHandler(db.Log, db.DB, approvalStore, db.BusDomain.Alert)

	t.Run("creates_approval_request_with_task_token", func(t *testing.T) {
		executionID := wfData.AutomationExecutions[0].ID

		config := json.RawMessage(`{
			"approvers": ["` + approverA.String() + `", "` + approverB.String() + `"],
			"approval_type": "any",
			"timeout_hours": 48,
			"approval_message": "Please review and approve this order"
		}`)

		execCtx := workflow.ActionExecutionContext{
			EntityID:    uuid.New(),
			EntityName:  "orders",
			EventType:   "on_create",
			RuleID:      &ruleID,
			RuleName:    "Order Approval",
			ActionName:  "seek_approval_0",
			ExecutionID: executionID,
		}

		taskToken := []byte("temporal-task-token-12345")
		err := handler.StartAsync(ctx, config, execCtx, taskToken)
		if err != nil {
			t.Fatalf("StartAsync failed: %s", err)
		}
	})

	t.Run("execute_creates_request_with_empty_token", func(t *testing.T) {
		executionID := wfData.AutomationExecutions[1].ID

		config := json.RawMessage(`{
			"approvers": ["` + approverA.String() + `"],
			"approval_type": "any",
			"approval_message": "Manual execution test"
		}`)

		execCtx := workflow.ActionExecutionContext{
			EntityID:    uuid.New(),
			EntityName:  "orders",
			EventType:   "manual_trigger",
			RuleID:      &ruleID,
			RuleName:    "Manual Rule",
			ActionName:  "seek_approval_1",
			ExecutionID: executionID,
		}

		result, err := handler.Execute(ctx, config, execCtx)
		if err != nil {
			t.Fatalf("Execute failed: %s", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map result, got %T", result)
		}

		if resultMap["status"] != "pending" {
			t.Fatalf("expected status=pending, got %v", resultMap["status"])
		}

		if resultMap["approval_id"] == nil || resultMap["approval_id"] == "" {
			t.Fatal("expected non-empty approval_id")
		}
	})

	t.Run("start_async_base64_encodes_task_token", func(t *testing.T) {
		executionID := wfData.AutomationExecutions[2].ID

		config := json.RawMessage(`{
			"approvers": ["` + approverA.String() + `"],
			"approval_type": "any",
			"approval_message": "Token encoding test"
		}`)

		taskToken := []byte("raw-binary-token-\x00\x01\x02")
		expectedEncoded := base64.StdEncoding.EncodeToString(taskToken)

		execCtx := workflow.ActionExecutionContext{
			EntityID:    uuid.New(),
			EntityName:  "orders",
			EventType:   "on_create",
			RuleID:      &ruleID,
			RuleName:    "Token Test",
			ActionName:  "seek_approval_2",
			ExecutionID: executionID,
		}

		err := handler.StartAsync(ctx, config, execCtx, taskToken)
		if err != nil {
			t.Fatalf("StartAsync failed: %s", err)
		}

		// Verify the token was base64 encoded by querying the approval request.
		filter := approvalrequestbus.QueryFilter{ExecutionID: &executionID}
		reqs, err := approvalStore.Query(ctx, filter, approvalrequestbus.DefaultOrderBy, pageOne())
		if err != nil {
			t.Fatalf("Query failed: %s", err)
		}

		if len(reqs) == 0 {
			t.Fatal("expected at least 1 approval request")
		}

		if reqs[0].TaskToken != expectedEncoded {
			t.Fatalf("expected base64 token %q, got %q", expectedEncoded, reqs[0].TaskToken)
		}
	})

	t.Run("start_async_default_timeout", func(t *testing.T) {
		executionID := wfData.AutomationExecutions[3].ID

		config := json.RawMessage(`{
			"approvers": ["` + approverA.String() + `"],
			"approval_type": "any",
			"approval_message": "Default timeout test"
		}`)

		execCtx := workflow.ActionExecutionContext{
			EntityID:    uuid.New(),
			EntityName:  "orders",
			EventType:   "on_create",
			RuleID:      &ruleID,
			RuleName:    "Timeout Test",
			ActionName:  "seek_approval_3",
			ExecutionID: executionID,
		}

		err := handler.StartAsync(ctx, config, execCtx, []byte("token"))
		if err != nil {
			t.Fatalf("StartAsync failed: %s", err)
		}

		filter := approvalrequestbus.QueryFilter{ExecutionID: &executionID}
		reqs, err := approvalStore.Query(ctx, filter, approvalrequestbus.DefaultOrderBy, pageOne())
		if err != nil {
			t.Fatalf("Query failed: %s", err)
		}

		if len(reqs) == 0 {
			t.Fatal("expected at least 1 approval request")
		}

		if reqs[0].TimeoutHours != 72 {
			t.Fatalf("expected default timeout 72h, got %d", reqs[0].TimeoutHours)
		}
	})
}

// =============================================================================
// Helpers

func pageOne() page.Page {
	return page.MustParse("1", "10")
}

// noopApprovalStorer is a minimal mock that satisfies the Storer interface
// for unit tests that need non-nil buses but don't hit the DB.
type noopApprovalStorer struct{}

func (s *noopApprovalStorer) NewWithTx(_ sqldb.CommitRollbacker) (approvalrequestbus.Storer, error) {
	return s, nil
}
func (s *noopApprovalStorer) Create(_ context.Context, _ approvalrequestbus.ApprovalRequest) error {
	return nil
}
func (s *noopApprovalStorer) QueryByID(_ context.Context, _ uuid.UUID) (approvalrequestbus.ApprovalRequest, error) {
	return approvalrequestbus.ApprovalRequest{}, nil
}
func (s *noopApprovalStorer) Resolve(_ context.Context, _, _ uuid.UUID, _, _ string) (approvalrequestbus.ApprovalRequest, error) {
	return approvalrequestbus.ApprovalRequest{}, nil
}
func (s *noopApprovalStorer) Query(_ context.Context, _ approvalrequestbus.QueryFilter, _ order.By, _ page.Page) ([]approvalrequestbus.ApprovalRequest, error) {
	return nil, nil
}
func (s *noopApprovalStorer) Count(_ context.Context, _ approvalrequestbus.QueryFilter) (int, error) {
	return 0, nil
}
func (s *noopApprovalStorer) IsApprover(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

// noopAlertStorer satisfies alertbus.Storer for unit tests.
type noopAlertStorer struct{}

func (s *noopAlertStorer) NewWithTx(_ sqldb.CommitRollbacker) (alertbus.Storer, error) {
	return s, nil
}
func (s *noopAlertStorer) Create(_ context.Context, _ alertbus.Alert) error                 { return nil }
func (s *noopAlertStorer) CreateRecipients(_ context.Context, _ []alertbus.AlertRecipient) error {
	return nil
}
func (s *noopAlertStorer) CreateAcknowledgment(_ context.Context, _ alertbus.AlertAcknowledgment) error {
	return nil
}
func (s *noopAlertStorer) QueryByID(_ context.Context, _ uuid.UUID) (alertbus.Alert, error) {
	return alertbus.Alert{}, nil
}
func (s *noopAlertStorer) Query(_ context.Context, _ alertbus.QueryFilter, _ order.By, _ page.Page) ([]alertbus.Alert, error) {
	return nil, nil
}
func (s *noopAlertStorer) QueryByUserID(_ context.Context, _ uuid.UUID, _ []uuid.UUID, _ alertbus.QueryFilter, _ order.By, _ page.Page) ([]alertbus.Alert, error) {
	return nil, nil
}
func (s *noopAlertStorer) QueryRecipientsByAlertID(_ context.Context, _ uuid.UUID) ([]alertbus.AlertRecipient, error) {
	return nil, nil
}
func (s *noopAlertStorer) QueryRecipientsByAlertIDs(_ context.Context, _ []uuid.UUID) (map[uuid.UUID][]alertbus.AlertRecipient, error) {
	return nil, nil
}
func (s *noopAlertStorer) Count(_ context.Context, _ alertbus.QueryFilter) (int, error) {
	return 0, nil
}
func (s *noopAlertStorer) CountByUserID(_ context.Context, _ uuid.UUID, _ []uuid.UUID, _ alertbus.QueryFilter) (int, error) {
	return 0, nil
}
func (s *noopAlertStorer) UpdateStatus(_ context.Context, _ uuid.UUID, _ string, _ time.Time) error {
	return nil
}
func (s *noopAlertStorer) IsRecipient(_ context.Context, _, _ uuid.UUID, _ []uuid.UUID) (bool, error) {
	return false, nil
}
func (s *noopAlertStorer) FilterRecipientAlerts(_ context.Context, _ []uuid.UUID, _ uuid.UUID, _ []uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s *noopAlertStorer) QueryActiveByUserID(_ context.Context, _ uuid.UUID, _ []uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s *noopAlertStorer) AcknowledgeMultiple(_ context.Context, _ []uuid.UUID, _ uuid.UUID, _ string, _ time.Time) (int, error) {
	return 0, nil
}
func (s *noopAlertStorer) DismissMultiple(_ context.Context, _ []uuid.UUID, _ time.Time) (int, error) {
	return 0, nil
}
func (s *noopAlertStorer) ResolveRelatedAlerts(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID, _ time.Time) (int, error) {
	return 0, nil
}

// newApprovalRequestStore creates a real approvalrequestbus.Business using the test DB.
func newApprovalRequestStore(db *dbtest.Database) *approvalrequestdb.Store {
	return approvalrequestdb.NewStore(db.Log, db.DB)
}
