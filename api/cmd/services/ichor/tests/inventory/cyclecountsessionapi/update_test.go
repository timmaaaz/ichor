package cyclecountsessionapi_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "name-change",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Name: dbtest.StringPointer("Updated Session Name"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			ExpResp: &cyclecountsessionapp.CycleCountSession{
				ID:            sd.CycleCountSessions[0].ID,
				Name:          "Updated Session Name",
				Status:        "draft",
				CreatedBy:     sd.CycleCountSessions[0].CreatedBy,
				CompletedDate: "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountsessionapp.CycleCountSession)
				expResp := exp.(*cyclecountsessionapp.CycleCountSession)
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "draft-to-in-progress",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			ExpResp: &cyclecountsessionapp.CycleCountSession{
				ID:            sd.CycleCountSessions[1].ID,
				Name:          sd.CycleCountSessions[1].Name,
				Status:        "in_progress",
				CreatedBy:     sd.CycleCountSessions[1].CreatedBy,
				CompletedDate: "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountsessionapp.CycleCountSession)
				expResp := exp.(*cyclecountsessionapp.CycleCountSession)
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-status",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("not_a_valid_status"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `parse status: invalid status "not_a_valid_status"`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-update-permission",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusForbidden,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Name: dbtest.StringPointer("Should Fail"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.PermissionDenied, "user does not have permission UPDATE for table: inventory.cycle_count_sessions"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Name: dbtest.StringPointer("Does Not Exist"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "cycle count session not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// TestUpdate200Cancel tests the draft → cancelled status transition.
func TestUpdate200Cancel(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession_Cancel")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Session[2] is in draft — cancel it.
	test.Run(t, []apitest.Table{
		{
			Name:       "draft-to-cancelled",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[2].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("cancelled"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			ExpResp: &cyclecountsessionapp.CycleCountSession{
				ID:            sd.CycleCountSessions[2].ID,
				Name:          sd.CycleCountSessions[2].Name,
				Status:        "cancelled",
				CreatedBy:     sd.CycleCountSessions[2].CreatedBy,
				CompletedDate: "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountsessionapp.CycleCountSession)
				expResp := exp.(*cyclecountsessionapp.CycleCountSession)
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
	}, "cancel")
}

// TestUpdate400TerminalState tests that transitioning from a terminal state returns FailedPrecondition.
func TestUpdate400TerminalState(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession_TerminalState")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	session := sd.CycleCountSessions[2]

	// Step 1: Cancel the session (draft → cancelled).
	test.Run(t, []apitest.Table{
		{
			Name:       "cancel",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("cancelled"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			CmpFunc: func(got, exp any) string { return "" },
		},
	}, "cancel")

	// Step 2: Try to transition from cancelled → in_progress. Should fail.
	test.Run(t, []apitest.Table{
		{
			Name:       "transition-from-cancelled",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.FailedPrecondition, "session is already cancelled and cannot be transitioned"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}, "terminal-state")
}

// TestUpdate400CompleteFromDraft tests that completing a session directly from draft returns FailedPrecondition.
func TestUpdate400CompleteFromDraft(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession_CompleteFromDraft")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	session := sd.CycleCountSessions[2]

	// Try to complete directly from draft — should fail.
	test.Run(t, []apitest.Table{
		{
			Name:       "complete-from-draft",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("completed"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.FailedPrecondition, "session must be in_progress to complete, current status: draft"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}, "complete-from-draft")
}

// TestUpdate200CompleteFlow tests the full cycle count lifecycle:
// create session → add items → count items → approve variance → complete session
// → verify inventory adjustments created AND approved (not pending).
func TestUpdate200CompleteFlow(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession_CompleteFlow")

	sd, err := insertCompleteFlowSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	session := sd.CycleCountSessions[0]

	// -------------------------------------------------------------------------
	// Step 1: Transition session from draft → in_progress
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "to-in-progress",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			CmpFunc: func(got, exp any) string { return "" },
		},
	}, "step1-in-progress")

	// -------------------------------------------------------------------------
	// Step 2: Create a cycle count item for this session
	// -------------------------------------------------------------------------
	var createdItem cyclecountitemapp.CycleCountItem
	test.Run(t, []apitest.Table{
		{
			Name:       "create-item",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      session.ID,
				ProductID:      sd.Products[0].ProductID,
				LocationID:     sd.InventoryLocations[0].LocationID,
				SystemQuantity: "100",
			},
			GotResp: &createdItem,
			CmpFunc: func(got, exp any) string { return "" },
		},
	}, "step2-create-item")

	// -------------------------------------------------------------------------
	// Step 3: Count the item (set counted_quantity, triggering auto-variance)
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "count-item",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", createdItem.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				CountedQuantity: dbtest.StringPointer("95"),
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				// Verify auto-computed variance: 95 - 100 = -5
				if gotResp.Variance != "-5" {
					return fmt.Sprintf("expected variance -5, got %s", gotResp.Variance)
				}
				// Verify counted_by was auto-injected
				if gotResp.CountedBy == "" {
					return "expected counted_by to be auto-injected"
				}
				// Verify counted_date was auto-injected
				if gotResp.CountedDate == "" {
					return "expected counted_date to be auto-injected"
				}
				return ""
			},
		},
	}, "step3-count-item")

	// -------------------------------------------------------------------------
	// Step 4: Approve the variance (pending → variance_approved)
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "approve-variance",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", createdItem.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				Status: dbtest.StringPointer("variance_approved"),
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				if gotResp.Status != "variance_approved" {
					return fmt.Sprintf("expected status variance_approved, got %s", gotResp.Status)
				}
				return ""
			},
		},
	}, "step4-approve-variance")

	// -------------------------------------------------------------------------
	// Step 5: Complete the session (with a simultaneous name change to verify it's preserved)
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "complete-session",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Name:   dbtest.StringPointer("Final Session Name"),
				Status: dbtest.StringPointer("completed"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountsessionapp.CycleCountSession)
				if gotResp.Status != "completed" {
					return fmt.Sprintf("expected status completed, got %s", gotResp.Status)
				}
				if gotResp.Name != "Final Session Name" {
					return fmt.Sprintf("expected name 'Final Session Name', got %s", gotResp.Name)
				}
				if gotResp.CompletedDate == "" {
					return "expected completed_date to be set"
				}
				return ""
			},
		},
	}, "step5-complete")

	// -------------------------------------------------------------------------
	// Step 6: Verify inventory adjustments were created AND approved
	// -------------------------------------------------------------------------
	ctx := context.Background()
	busDomain := test.DB.BusDomain

	reasonCode := inventoryadjustmentbus.ReasonCodeCycleCount
	filter := inventoryadjustmentbus.QueryFilter{
		ReasonCode: &reasonCode,
	}

	adjs, err := busDomain.InventoryAdjustment.Query(ctx, filter, inventoryadjustmentbus.DefaultOrderBy, page.MustParse("1", "10"))
	if err != nil {
		t.Fatalf("querying inventory adjustments: %s", err)
	}

	if len(adjs) == 0 {
		t.Fatal("expected at least one inventory adjustment to be created")
	}

	for _, adj := range adjs {
		if adj.ApprovalStatus != inventoryadjustmentbus.ApprovalStatusApproved {
			t.Errorf("expected adjustment %s to be approved, got %s", adj.InventoryAdjustmentID, adj.ApprovalStatus)
		}
		if adj.ReasonCode != inventoryadjustmentbus.ReasonCodeCycleCount {
			t.Errorf("expected reason_code cycle_count, got %s", adj.ReasonCode)
		}
		if adj.QuantityChange != -5 {
			t.Errorf("expected quantity_change -5, got %d", adj.QuantityChange)
		}
	}

	// -------------------------------------------------------------------------
	// Step 7: Verify TOCTOU — completing again returns FailedPrecondition
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "already-completed",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("completed"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.FailedPrecondition, "session is already completed and cannot be transitioned"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}, "step7-already-completed")
}
