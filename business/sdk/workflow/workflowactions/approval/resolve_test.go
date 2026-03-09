package approval_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/foundation/logger"
)

func Test_ResolveApprovalRequest(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ResolveApprovalRequest")

	unitest.Run(t, validateResolveTests(), "validate")
	executeResolveTests(t, db)
}

// =============================================================================
// Validate tests (pure unit — no DB required)

func validateResolveTests() []unitest.Table {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "" })
	handler := approval.NewResolveApprovalHandler(log, nil)

	return []unitest.Table{
		{
			Name:    "missing_approval_request_id",
			ExpResp: true,
			ExcFunc: func(_ context.Context) any {
				err := handler.Validate(json.RawMessage(`{"resolution":"approved"}`))
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "invalid_uuid",
			ExpResp: true,
			ExcFunc: func(_ context.Context) any {
				err := handler.Validate(json.RawMessage(`{"approval_request_id":"not-a-uuid","resolution":"approved"}`))
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "invalid_resolution_value",
			ExpResp: true,
			ExcFunc: func(_ context.Context) any {
				validID := uuid.New().String()
				err := handler.Validate(json.RawMessage(`{"approval_request_id":"` + validID + `","resolution":"maybe"}`))
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "valid_approved",
			ExpResp: true,
			ExcFunc: func(_ context.Context) any {
				validID := uuid.New().String()
				err := handler.Validate(json.RawMessage(`{"approval_request_id":"` + validID + `","resolution":"approved"}`))
				return err == nil
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "valid_rejected",
			ExpResp: true,
			ExcFunc: func(_ context.Context) any {
				validID := uuid.New().String()
				err := handler.Validate(json.RawMessage(`{"approval_request_id":"` + validID + `","resolution":"rejected"}`))
				return err == nil
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
	}
}

// =============================================================================
// Execute tests (require real DB)

func executeResolveTests(t *testing.T, db *dbtest.Database) {
	t.Helper()

	ctx := context.Background()

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		t.Fatalf("seeding users: %s", err)
	}
	resolverID := users[0].ID

	wfData, err := workflow.TestSeedFullWorkflow(ctx, resolverID, db.BusDomain.Workflow)
	if err != nil {
		t.Fatalf("seeding workflow data: %s", err)
	}

	store := approvalrequestdb.NewStore(db.Log, db.DB)
	approvalBus := approvalrequestbus.NewBusiness(db.Log, nil, store)
	handler := approval.NewResolveApprovalHandler(db.Log, approvalBus)

	ruleID := wfData.AutomationRules[0].ID
	baseExecCtx := workflow.ActionExecutionContext{
		UserID:    resolverID,
		RuleID:    &ruleID,
		RuleName:  "Test Rule",
		EventType: "on_create",
	}

	createPending := func(execIdx int, actionName string) (approvalrequestbus.ApprovalRequest, error) {
		return approvalBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
			ExecutionID:     wfData.AutomationExecutions[execIdx].ID,
			RuleID:          ruleID,
			ActionName:      actionName,
			Approvers:       []uuid.UUID{resolverID},
			ApprovalType:    approvalrequestbus.ApprovalTypeAny,
			TimeoutHours:    48,
			TaskToken:       "test-token-" + actionName,
			ApprovalMessage: "test",
		})
	}

	t.Run("pending_approved_returns_resolved_approved_port", func(t *testing.T) {
		req, err := createPending(0, "resolve_approve_test")
		if err != nil {
			t.Fatalf("create approval request: %v", err)
		}

		config := json.RawMessage(`{"approval_request_id":"` + req.ID.String() + `","resolution":"approved","reason":"looks good"}`)
		result, err := handler.Execute(ctx, config, baseExecCtx)
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}

		m, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}
		if m["output"] != "resolved_approved" {
			t.Errorf("expected output=resolved_approved, got %v", m["output"])
		}
	})

	t.Run("pending_rejected_returns_resolved_rejected_port", func(t *testing.T) {
		req, err := createPending(1, "resolve_reject_test")
		if err != nil {
			t.Fatalf("create approval request: %v", err)
		}

		config := json.RawMessage(`{"approval_request_id":"` + req.ID.String() + `","resolution":"rejected","reason":"not good"}`)
		result, err := handler.Execute(ctx, config, baseExecCtx)
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}

		m, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}
		if m["output"] != "resolved_rejected" {
			t.Errorf("expected output=resolved_rejected, got %v", m["output"])
		}
	})

	t.Run("unknown_id_returns_not_found_port", func(t *testing.T) {
		unknownID := uuid.New().String()
		config := json.RawMessage(`{"approval_request_id":"` + unknownID + `","resolution":"approved"}`)
		result, err := handler.Execute(ctx, config, baseExecCtx)
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}

		m, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}
		if m["output"] != "not_found" {
			t.Errorf("expected output=not_found, got %v", m["output"])
		}
	})

	t.Run("already_resolved_returns_already_resolved_port", func(t *testing.T) {
		req, err := createPending(2, "resolve_double_test")
		if err != nil {
			t.Fatalf("create approval request: %v", err)
		}

		config := json.RawMessage(`{"approval_request_id":"` + req.ID.String() + `","resolution":"approved"}`)

		// First resolve.
		if _, err := handler.Execute(ctx, config, baseExecCtx); err != nil {
			t.Fatalf("first Execute: %v", err)
		}

		// Second resolve on same request — should be already_resolved.
		result, err := handler.Execute(ctx, config, baseExecCtx)
		if err != nil {
			t.Fatalf("second Execute: %v", err)
		}

		m, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}
		if m["output"] != "already_resolved" {
			t.Errorf("expected output=already_resolved, got %v", m["output"])
		}
	})
}
