// Package approvalapi_test contains end-to-end integration tests for the
// seek_approval workflow action using a real Temporal container.
//
// These tests verify that:
//   - seek_approval correctly pauses a Temporal workflow and creates a DB record
//   - Completing the async activity with "approved" routes the workflow to the approved port
//   - Completing the async activity with "rejected" routes the workflow to the rejected port
package approvalapi_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

// TestSeekApproval_Approved verifies the full seek_approval → approved routing path:
//  1. Workflow fires with seek_approval action
//  2. Temporal pauses at the async activity; approval_request record created in DB
//  3. We resolve the request as "approved" and complete the Temporal activity
//  4. Workflow continues on the "approved" output port without error
func TestSeekApproval_Approved(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_SeekApproval_Approved")
	wf := apitest.InitWorkflowInfra(t, db)

	testSeekApprovalFlow(t, db, wf, "approved")
}

// TestSeekApproval_Rejected verifies the full seek_approval → rejected routing path.
func TestSeekApproval_Rejected(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_SeekApproval_Rejected")
	wf := apitest.InitWorkflowInfra(t, db)

	testSeekApprovalFlow(t, db, wf, "rejected")
}

// testSeekApprovalFlow runs the seek_approval end-to-end test for a given resolution.
// resolution must be "approved" or "rejected".
func testSeekApprovalFlow(t *testing.T, db *dbtest.Database, wf *apitest.WorkflowInfra, resolution string) {
	t.Helper()
	ctx := context.Background()

	// ── 1. Get a seeded user (needed for rule CreatedBy FK and as approver) ──────
	users, err := db.BusDomain.User.Query(ctx, userbus.QueryFilter{}, userbus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil || len(users) == 0 {
		t.Fatalf("querying seeded users: %v", err)
	}
	createdBy := users[0].ID
	approverID := users[0].ID

	// ── 2. Build a workflow: start → seek_approval ────────────────────────────────
	// We query the seeded "customers" entity and "on_create" trigger type so the
	// TriggerProcessor can match our manually-fired TriggerEvent.
	customerEntity, err := wf.WorkflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("querying customers entity: %v", err)
	}
	entityType, err := wf.WorkflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %v", err)
	}
	triggerTypeCreate, err := wf.WorkflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %v", err)
	}

	// Create the automation rule.
	rule, err := wf.WorkflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "SeekApproval Test - " + resolution + " - " + uuid.New().String()[:8],
		Description:   "End-to-end test for seek_approval " + resolution + " path",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerTypeCreate.ID,
		IsActive:      true,
		CreatedBy:     createdBy,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	// Create the seek_approval template (needed so the edge store resolves ActionType).
	seekTemplate, err := wf.WorkflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:        "Seek Approval Template",
		Description: "Template for seek_approval test",
		ActionType:  "seek_approval",
		DefaultConfig: json.RawMessage(`{
			"approvers":       ["` + approverID.String() + `"],
			"approval_type":   "any",
			"approval_message": "Please review"
		}`),
		CreatedBy: createdBy,
	})
	if err != nil {
		t.Fatalf("creating seek_approval template: %v", err)
	}

	// Create the seek_approval action referencing the template.
	seekAction, err := wf.WorkflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Seek Approval",
		Description:      "Pause workflow pending human approval",
		ActionConfig: json.RawMessage(`{
			"approvers":        ["` + approverID.String() + `"],
			"approval_type":    "any",
			"timeout_hours":    72,
			"approval_message": "Integration test approval request"
		}`),
		IsActive:   true,
		TemplateID: &seekTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating seek_approval action: %v", err)
	}

	// Wire start edge: nil source → seek_approval action.
	_, err = wf.WorkflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: seekAction.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	// Refresh TriggerProcessor so the new rule is matched against incoming events.
	if err := wf.TriggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing trigger processor: %v", err)
	}

	// ── 3. Fire a TriggerEvent directly at the WorkflowTrigger ───────────────────
	// This bypasses the delegate handler and simulates what it would produce on
	// a real business-layer entity creation.
	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"test": true},
		UserID:     createdBy,
	}
	if err := wf.WorkflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger event: %v", err)
	}

	// ── 4. Poll the DB for the approval_request created by seek_approval ─────────
	// The Temporal activity runs asynchronously, so we poll until the record appears
	// or we time out. Typically appears within 2–3 seconds.
	var req approvalrequestbus.ApprovalRequest
	statusPending := approvalrequestbus.StatusPending
	for i := 0; i < 30; i++ {
		reqs, err := wf.ApprovalRequestBus.Query(
			ctx,
			approvalrequestbus.QueryFilter{Status: &statusPending},
			approvalrequestbus.DefaultOrderBy,
			page.MustParse("1", "10"),
		)
		if err != nil {
			t.Fatalf("polling approval requests: %v", err)
		}
		if len(reqs) > 0 {
			req = reqs[0]
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if req.ID == uuid.Nil {
		t.Fatal("timeout: no pending approval request found after 15s — seek_approval may have failed to create DB record")
	}

	t.Logf("approval_request created: id=%s execution_id=%s action=%s", req.ID, req.ExecutionID, req.ActionName)

	// Verify the record has the expected fields.
	if req.Status != approvalrequestbus.StatusPending {
		t.Errorf("expected status %q, got %q", approvalrequestbus.StatusPending, req.Status)
	}
	if len(req.Approvers) == 0 {
		t.Error("expected approvers list to be populated")
	}
	if req.TaskToken == "" {
		t.Error("expected task token to be set (needed for Temporal async completion)")
	}

	// ── 5. Resolve the approval request in the DB ─────────────────────────────────
	// This is what the HTTP POST /v1/workflow/approvals/{id}/resolve endpoint does.
	resolvedReq, err := wf.ApprovalRequestBus.Resolve(ctx, req.ID, approverID, resolution, "Integration test "+resolution)
	if err != nil {
		t.Fatalf("resolving approval request as %q: %v", resolution, err)
	}
	if resolvedReq.Status != resolution {
		t.Errorf("expected resolved status %q, got %q", resolution, resolvedReq.Status)
	}

	// ── 6. Complete the Temporal async activity to resume the workflow ─────────────
	// Decode the base64 task token stored in the approval_request record.
	taskToken, err := base64.StdEncoding.DecodeString(req.TaskToken)
	if err != nil {
		t.Fatalf("decoding task token: %v", err)
	}

	// Build the output that routes the workflow to the correct output port.
	// The workflow's edge routing uses Result["output"] to select the next action.
	output := workflowtemporal.ActionActivityOutput{
		ActionName: req.ActionName,
		Result: map[string]any{
			"output":      resolution,
			"approval_id": req.ID.String(),
			"resolved_by": approverID.String(),
			"reason":      "Integration test " + resolution,
		},
		Success: true,
	}
	if err := wf.TemporalClient.CompleteActivity(ctx, taskToken, output, nil); err != nil {
		t.Fatalf("completing temporal activity (resolution=%q): %v", resolution, err)
	}

	// ── 7. Wait for the workflow to process the completion ────────────────────────
	time.Sleep(3 * time.Second)

	// ── 8. Final verification ──────────────────────────────────────────────────────
	// Re-query the approval request to confirm the status persisted correctly.
	final, err := wf.ApprovalRequestBus.QueryByID(ctx, req.ID)
	if err != nil {
		t.Fatalf("final query of approval request: %v", err)
	}
	if final.Status != resolution {
		t.Errorf("expected final status %q, got %q", resolution, final.Status)
	}
	if final.ResolvedBy == nil || *final.ResolvedBy != approverID {
		t.Errorf("expected ResolvedBy=%s, got %v", approverID, final.ResolvedBy)
	}

	t.Logf("SUCCESS: seek_approval → %q path verified end-to-end (request: %s)", resolution, req.ID)
}

