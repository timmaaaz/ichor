package workflow_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// testHandler is a minimal ActionHandler for testing ActionService.
// It records calls and can be configured to fail.
type testHandler struct {
	actionType     string
	isAsync        bool
	supportsManual bool
	validateErr    error
	executeErr     error
	executeResult  any
	lastExecCtx    workflow.ActionExecutionContext
	executeCalled  bool
}

func (h *testHandler) GetType() string                       { return h.actionType }
func (h *testHandler) IsAsync() bool                         { return h.isAsync }
func (h *testHandler) SupportsManualExecution() bool         { return h.supportsManual }
func (h *testHandler) GetDescription() string                { return "test handler: " + h.actionType }
func (h *testHandler) Validate(config json.RawMessage) error { return h.validateErr }
func (h *testHandler) GetOutputPorts() []workflow.OutputPort { return nil }

func (h *testHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	h.executeCalled = true
	h.lastExecCtx = execCtx
	return h.executeResult, h.executeErr
}

func Test_ActionService(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ActionService")

	sd, err := insertActionServiceSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, executeTests(db, sd), "execute")
	unitest.Run(t, executeForAutomationTests(db, sd), "executeForAutomation")
	unitest.Run(t, listTests(db, sd), "list")
	unitest.Run(t, statusTests(db, sd), "status")
}

type actionServiceSeedData struct {
	Users []userbus.User
}

func insertActionServiceSeedData(busDomain dbtest.BusDomain) (actionServiceSeedData, error) {
	ctx := context.Background()

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return actionServiceSeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	return actionServiceSeedData{Users: users}, nil
}

// newTestService creates an ActionService with the given handlers registered.
func newTestService(db *dbtest.Database, handlers ...*testHandler) *workflow.ActionService {
	registry := workflow.NewActionRegistry()
	for _, h := range handlers {
		registry.Register(h)
	}
	return workflow.NewActionService(db.Log, db.DB, registry)
}

// passDiff is a CmpFunc that treats "" as pass and any other string as a diff.
func passDiff(got any, exp any) string {
	gotStr, ok := got.(string)
	if !ok {
		return fmt.Sprintf("unexpected type %T", got)
	}
	return gotStr
}

// =============================================================================
// Execute Tests
// =============================================================================

func executeTests(db *dbtest.Database, sd actionServiceSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "happy-path",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				handler := &testHandler{
					actionType:     "test_action",
					supportsManual: true,
					executeResult:  map[string]any{"ok": true},
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "test_action",
					Config:     json.RawMessage(`{"key":"value"}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err != nil {
					return fmt.Sprintf("execute error: %s", err)
				}
				if result.Status != "completed" {
					return fmt.Sprintf("expected status completed, got %s", result.Status)
				}
				if result.ExecutionID == uuid.Nil {
					return "execution ID should not be nil"
				}
				if !handler.executeCalled {
					return "handler.Execute was not called"
				}
				return ""
			},
			CmpFunc: passDiff,
		},
		{
			Name:    "unknown-action-type",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				svc := newTestService(db) // empty registry

				_, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "nonexistent",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err == nil {
					return "expected error for unknown action type"
				}
				return ""
			},
			CmpFunc: passDiff,
		},
		{
			Name:    "validation-failure",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				handler := &testHandler{
					actionType:     "test_action",
					supportsManual: true,
					validateErr:    fmt.Errorf("bad config"),
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "test_action",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err == nil {
					return "expected validation error"
				}
				if result.Status != "failed" {
					return fmt.Sprintf("expected status failed, got %s", result.Status)
				}
				return ""
			},
			CmpFunc: passDiff,
		},
		{
			Name:    "handler-execution-failure",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				handler := &testHandler{
					actionType:     "test_fail",
					supportsManual: true,
					executeErr:     fmt.Errorf("handler exploded"),
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "test_fail",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err == nil {
					return "expected execution error"
				}
				if result.Status != "failed" {
					return fmt.Sprintf("expected status failed, got %s", result.Status)
				}
				return ""
			},
			CmpFunc: passDiff,
		},
		{
			Name:    "manual-execution-not-supported",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				handler := &testHandler{
					actionType:     "auto_only",
					supportsManual: false,
				}
				svc := newTestService(db, handler)

				_, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "auto_only",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err == nil {
					return "expected manual execution not supported error"
				}
				return ""
			},
			CmpFunc: passDiff,
		},
		{
			Name:    "async-action-returns-queued",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				handler := &testHandler{
					actionType:     "async_test",
					supportsManual: true,
					isAsync:        true,
					executeResult:  map[string]any{"task_id": "123"},
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "async_test",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err != nil {
					return fmt.Sprintf("unexpected error: %s", err)
				}
				if result.Status != "queued" {
					return fmt.Sprintf("expected status queued for async, got %s", result.Status)
				}
				return ""
			},
			CmpFunc: passDiff,
		},
	}
}

// =============================================================================
// ExecuteForAutomation Tests
// =============================================================================

func executeForAutomationTests(db *dbtest.Database, sd actionServiceSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "automation-context-populated",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				handler := &testHandler{
					actionType:     "auto_action",
					supportsManual: true,
					executeResult:  map[string]any{"done": true},
				}
				svc := newTestService(db, handler)

				ruleID := uuid.New()
				result, err := svc.ExecuteForAutomation(ctx, workflow.ExecuteRequest{
					ActionType: "auto_action",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, ruleID, "My Rule", "on_update")
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				if result.Status != "completed" {
					return fmt.Sprintf("expected completed, got %s", result.Status)
				}

				// Verify the execution context passed to the handler
				if handler.lastExecCtx.RuleID == nil || *handler.lastExecCtx.RuleID != ruleID {
					return fmt.Sprintf("expected RuleID %s, got %v", ruleID, handler.lastExecCtx.RuleID)
				}
				if handler.lastExecCtx.RuleName != "My Rule" {
					return fmt.Sprintf("expected RuleName 'My Rule', got %q", handler.lastExecCtx.RuleName)
				}
				if handler.lastExecCtx.EventType != "on_update" {
					return fmt.Sprintf("expected EventType on_update, got %q", handler.lastExecCtx.EventType)
				}
				if handler.lastExecCtx.TriggerSource != workflow.TriggerSourceAutomation {
					return fmt.Sprintf("expected TriggerSource automation, got %q", handler.lastExecCtx.TriggerSource)
				}
				return ""
			},
			CmpFunc: passDiff,
		},
		{
			Name:    "automation-unknown-action",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				svc := newTestService(db)
				_, err := svc.ExecuteForAutomation(ctx, workflow.ExecuteRequest{
					ActionType: "nonexistent",
					Config:     json.RawMessage(`{}`),
					UserID:     sd.Users[0].ID,
				}, uuid.New(), "rule", "on_create")
				if err == nil {
					return "expected error for unknown action"
				}
				return ""
			},
			CmpFunc: passDiff,
		},
	}
}

// =============================================================================
// List Tests
// =============================================================================

func listTests(db *dbtest.Database, sd actionServiceSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "list-all",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				manual := &testHandler{actionType: "manual_a", supportsManual: true}
				autoOnly := &testHandler{actionType: "auto_b", supportsManual: false}
				svc := newTestService(db, manual, autoOnly)

				all := svc.ListAvailableActions()
				if len(all) != 2 {
					return fmt.Sprintf("expected 2 actions, got %d", len(all))
				}
				return ""
			},
			CmpFunc: passDiff,
		},
		{
			Name:    "list-manual-only",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				manual := &testHandler{actionType: "manual_a", supportsManual: true}
				autoOnly := &testHandler{actionType: "auto_b", supportsManual: false}
				svc := newTestService(db, manual, autoOnly)

				manualActions := svc.ListManuallyExecutableActions()
				if len(manualActions) != 1 {
					return fmt.Sprintf("expected 1 manual action, got %d", len(manualActions))
				}
				if manualActions[0].Type != "manual_a" {
					return fmt.Sprintf("expected manual_a, got %s", manualActions[0].Type)
				}
				return ""
			},
			CmpFunc: passDiff,
		},
	}
}

// =============================================================================
// Status Tests
// =============================================================================

func statusTests(db *dbtest.Database, sd actionServiceSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "get-execution-status-after-execute",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				handler := &testHandler{
					actionType:     "status_test",
					supportsManual: true,
					executeResult:  map[string]any{"ok": true},
				}
				svc := newTestService(db, handler)

				result, err := svc.Execute(ctx, workflow.ExecuteRequest{
					ActionType: "status_test",
					Config:     json.RawMessage(`{}`),
					EntityName: "test_entity",
					UserID:     sd.Users[0].ID,
				}, workflow.TriggerSourceManual)
				if err != nil {
					return fmt.Sprintf("execute error: %s", err)
				}

				status, err := svc.GetExecutionStatus(ctx, result.ExecutionID)
				if err != nil {
					return fmt.Sprintf("status error: %s", err)
				}
				if status.Status != "completed" {
					return fmt.Sprintf("expected completed, got %s", status.Status)
				}
				if status.ExecutionID != result.ExecutionID {
					return "execution IDs don't match"
				}
				return ""
			},
			CmpFunc: passDiff,
		},
		{
			Name:    "get-execution-status-not-found",
			ExpResp: "",
			ExcFunc: func(ctx context.Context) any {
				svc := newTestService(db)
				_, err := svc.GetExecutionStatus(ctx, uuid.New())
				if err == nil {
					return "expected error for nonexistent execution"
				}
				return ""
			},
			CmpFunc: passDiff,
		},
	}
}
