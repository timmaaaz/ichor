package approvalrequestbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus/stores/approvalrequestdb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

func Test_ApprovalRequest(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ApprovalRequest")

	sd, err := insertApprovalSeedData(db)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, createAndQuery(sd), "create-and-query")
	unitest.Run(t, queryWithFilters(sd), "query-with-filters")
	unitest.Run(t, resolveTests(sd), "resolve")
	unitest.Run(t, isApproverTests(sd), "is-approver")
}

// =============================================================================

type approvalSeedData struct {
	unitest.SeedData
	ApprovalBus *approvalrequestbus.Business
	WFData      *workflow.TestWorkflowData
	UserID      uuid.UUID
	ApproverA   uuid.UUID
	ApproverB   uuid.UUID
}

func insertApprovalSeedData(db *dbtest.Database) (approvalSeedData, error) {
	ctx := context.Background()

	// Seed a user for FK references (resolved_by).
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		return approvalSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	userID := users[0].ID

	// Seed full workflow data for FK constraints (execution_id, rule_id).
	wfData, err := workflow.TestSeedFullWorkflow(ctx, userID, db.BusDomain.Workflow)
	if err != nil {
		return approvalSeedData{}, fmt.Errorf("seeding workflow data: %w", err)
	}

	store := approvalrequestdb.NewStore(db.Log, db.DB)
	approvalBus := approvalrequestbus.NewBusiness(db.Log, store)

	return approvalSeedData{
		ApprovalBus: approvalBus,
		WFData:      wfData,
		UserID:      userID,
		ApproverA:   uuid.New(),
		ApproverB:   uuid.New(),
	}, nil
}

// =============================================================================
// Create + QueryByID round-trip

func createAndQuery(sd approvalSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "create_and_query_by_id",
			ExpResp: "pending",
			ExcFunc: func(ctx context.Context) any {
				ruleID := sd.WFData.AutomationRules[0].ID
				executionID := sd.WFData.AutomationExecutions[0].ID

				req, err := sd.ApprovalBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
					ExecutionID:     executionID,
					RuleID:          ruleID,
					ActionName:      "seek_approval_0",
					Approvers:       []uuid.UUID{sd.ApproverA, sd.ApproverB},
					ApprovalType:    "any",
					TimeoutHours:    48,
					TaskToken:       "dGVzdC10b2tlbg==",
					ApprovalMessage: "Please approve this",
				})
				if err != nil {
					return err
				}

				// Query it back.
				got, err := sd.ApprovalBus.QueryByID(ctx, req.ID)
				if err != nil {
					return err
				}

				if got.ID != req.ID {
					return fmt.Sprintf("ID mismatch: got %s, want %s", got.ID, req.ID)
				}
				if got.ActionName != "seek_approval_0" {
					return fmt.Sprintf("ActionName mismatch: got %s", got.ActionName)
				}
				if got.ApprovalMessage != "Please approve this" {
					return fmt.Sprintf("Message mismatch: got %s", got.ApprovalMessage)
				}
				if got.TaskToken != "dGVzdC10b2tlbg==" {
					return fmt.Sprintf("TaskToken mismatch: got %s", got.TaskToken)
				}
				if len(got.Approvers) != 2 {
					return fmt.Sprintf("Approvers count: got %d, want 2", len(got.Approvers))
				}
				if got.TimeoutHours != 48 {
					return fmt.Sprintf("TimeoutHours: got %d, want 48", got.TimeoutHours)
				}

				return got.Status
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "query_by_id_not_found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := sd.ApprovalBus.QueryByID(ctx, uuid.New())
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
			Name:    "count",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				count, err := sd.ApprovalBus.Count(ctx, approvalrequestbus.QueryFilter{})
				if err != nil {
					return err
				}
				// At least 1 from the create test above.
				return count >= 1
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
// Query with filters

func queryWithFilters(sd approvalSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "query_with_status_filter",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				ruleID := sd.WFData.AutomationRules[1].ID
				executionID := sd.WFData.AutomationExecutions[1].ID

				_, err := sd.ApprovalBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
					ExecutionID:     executionID,
					RuleID:          ruleID,
					ActionName:      "seek_approval_filter",
					Approvers:       []uuid.UUID{sd.ApproverA},
					ApprovalType:    "any",
					TimeoutHours:    72,
					TaskToken:       "filter-test-token",
					ApprovalMessage: "Filter test",
				})
				if err != nil {
					return err
				}

				status := approvalrequestbus.StatusPending
				filter := approvalrequestbus.QueryFilter{Status: &status}
				reqs, err := sd.ApprovalBus.Query(ctx, filter, approvalrequestbus.DefaultOrderBy, page.MustParse("1", "50"))
				if err != nil {
					return err
				}

				// All returned should be pending.
				for _, r := range reqs {
					if r.Status != approvalrequestbus.StatusPending {
						return fmt.Sprintf("expected all pending, got status=%s", r.Status)
					}
				}
				return len(reqs) >= 1
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "query_with_approver_id_filter",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				ruleID := sd.WFData.AutomationRules[2].ID
				executionID := sd.WFData.AutomationExecutions[2].ID
				uniqueApprover := uuid.New()

				_, err := sd.ApprovalBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
					ExecutionID:     executionID,
					RuleID:          ruleID,
					ActionName:      "seek_approval_approver",
					Approvers:       []uuid.UUID{uniqueApprover},
					ApprovalType:    "any",
					TimeoutHours:    72,
					TaskToken:       "approver-filter-token",
					ApprovalMessage: "Approver filter test",
				})
				if err != nil {
					return err
				}

				filter := approvalrequestbus.QueryFilter{ApproverID: &uniqueApprover}
				reqs, err := sd.ApprovalBus.Query(ctx, filter, approvalrequestbus.DefaultOrderBy, page.MustParse("1", "50"))
				if err != nil {
					return err
				}

				return len(reqs) == 1
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
// Resolve tests

func resolveTests(sd approvalSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "resolve_approve",
			ExpResp: approvalrequestbus.StatusApproved,
			ExcFunc: func(ctx context.Context) any {
				ruleID := sd.WFData.AutomationRules[3].ID
				executionID := sd.WFData.AutomationExecutions[3].ID

				req, err := sd.ApprovalBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
					ExecutionID:     executionID,
					RuleID:          ruleID,
					ActionName:      "seek_approval_resolve_approve",
					Approvers:       []uuid.UUID{sd.ApproverA},
					ApprovalType:    "any",
					TimeoutHours:    72,
					TaskToken:       "resolve-approve-token",
					ApprovalMessage: "Resolve approve test",
				})
				if err != nil {
					return err
				}

				resolved, err := sd.ApprovalBus.Resolve(ctx, req.ID, sd.UserID, approvalrequestbus.StatusApproved, "Looks good")
				if err != nil {
					return err
				}

				if resolved.ResolvedBy == nil || *resolved.ResolvedBy != sd.UserID {
					return fmt.Sprintf("expected resolved_by=%s, got %v", sd.UserID, resolved.ResolvedBy)
				}
				if resolved.ResolutionReason != "Looks good" {
					return fmt.Sprintf("expected reason='Looks good', got %s", resolved.ResolutionReason)
				}
				if resolved.ResolvedDate == nil {
					return "expected resolved_date to be set"
				}

				return resolved.Status
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "resolve_reject",
			ExpResp: approvalrequestbus.StatusRejected,
			ExcFunc: func(ctx context.Context) any {
				ruleID := sd.WFData.AutomationRules[4].ID
				executionID := sd.WFData.AutomationExecutions[4].ID

				req, err := sd.ApprovalBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
					ExecutionID:     executionID,
					RuleID:          ruleID,
					ActionName:      "seek_approval_resolve_reject",
					Approvers:       []uuid.UUID{sd.ApproverA},
					ApprovalType:    "any",
					TimeoutHours:    72,
					TaskToken:       "resolve-reject-token",
					ApprovalMessage: "Resolve reject test",
				})
				if err != nil {
					return err
				}

				resolved, err := sd.ApprovalBus.Resolve(ctx, req.ID, sd.UserID, approvalrequestbus.StatusRejected, "Not acceptable")
				if err != nil {
					return err
				}

				return resolved.Status
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "resolve_already_resolved",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				ruleID := sd.WFData.AutomationRules[0].ID
				executionID := sd.WFData.AutomationExecutions[5].ID

				req, err := sd.ApprovalBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
					ExecutionID:     executionID,
					RuleID:          ruleID,
					ActionName:      "seek_approval_double_resolve",
					Approvers:       []uuid.UUID{sd.ApproverA},
					ApprovalType:    "any",
					TimeoutHours:    72,
					TaskToken:       "double-resolve-token",
					ApprovalMessage: "Double resolve test",
				})
				if err != nil {
					return err
				}

				// First resolve succeeds.
				_, err = sd.ApprovalBus.Resolve(ctx, req.ID, sd.UserID, approvalrequestbus.StatusApproved, "First")
				if err != nil {
					return fmt.Sprintf("first resolve failed: %s", err)
				}

				// Second resolve should fail with ErrAlreadyResolved.
				_, err = sd.ApprovalBus.Resolve(ctx, req.ID, sd.UserID, approvalrequestbus.StatusRejected, "Second")
				return err != nil
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
// IsApprover tests

func isApproverTests(sd approvalSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "is_approver_true",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				ruleID := sd.WFData.AutomationRules[1].ID
				executionID := sd.WFData.AutomationExecutions[6].ID

				req, err := sd.ApprovalBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
					ExecutionID:     executionID,
					RuleID:          ruleID,
					ActionName:      "seek_approval_is_approver",
					Approvers:       []uuid.UUID{sd.ApproverA, sd.ApproverB},
					ApprovalType:    "any",
					TimeoutHours:    72,
					TaskToken:       "is-approver-token",
					ApprovalMessage: "Is approver test",
				})
				if err != nil {
					return err
				}

				isApprover, err := sd.ApprovalBus.IsApprover(ctx, req.ID, sd.ApproverA)
				if err != nil {
					return err
				}
				return isApprover
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "is_approver_false",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				ruleID := sd.WFData.AutomationRules[2].ID
				executionID := sd.WFData.AutomationExecutions[7].ID

				req, err := sd.ApprovalBus.Create(ctx, approvalrequestbus.NewApprovalRequest{
					ExecutionID:     executionID,
					RuleID:          ruleID,
					ActionName:      "seek_approval_not_approver",
					Approvers:       []uuid.UUID{sd.ApproverA},
					ApprovalType:    "any",
					TimeoutHours:    72,
					TaskToken:       "not-approver-token",
					ApprovalMessage: "Not approver test",
				})
				if err != nil {
					return err
				}

				randomUser := uuid.New()
				isApprover, err := sd.ApprovalBus.IsApprover(ctx, req.ID, randomUser)
				if err != nil {
					return err
				}
				return isApprover
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
