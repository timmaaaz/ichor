package putawaytaskapi_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/putawaytaskapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "claim-pending-to-in-progress",
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", sd.PutAwayTasks[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &putawaytaskapp.UpdatePutAwayTask{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &putawaytaskapp.PutAwayTask{},
			ExpResp: &putawaytaskapp.PutAwayTask{
				ID:         sd.PutAwayTasks[0].ID,
				ProductID:  sd.PutAwayTasks[0].ProductID,
				LocationID: sd.PutAwayTasks[0].LocationID,
				Quantity:   sd.PutAwayTasks[0].Quantity,
				Status:     "in_progress",
				CreatedBy:  sd.PutAwayTasks[0].CreatedBy,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*putawaytaskapp.PutAwayTask)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*putawaytaskapp.PutAwayTask)
				expResp.AssignedTo = gotResp.AssignedTo
				expResp.AssignedAt = gotResp.AssignedAt
				expResp.ReferenceNumber = gotResp.ReferenceNumber
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

// TestUpdate200Complete verifies that completing a put-away task atomically creates
// an inventory transaction and upserts the inventory item quantity.
func TestUpdate200Complete(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_PutAwayTask_Complete")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	ctx := context.Background()
	task := sd.PutAwayTasks[2]

	// Step 1: Claim the task (pending → in_progress).
	test.Run(t, []apitest.Table{
		{
			Name:       "claim",
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &putawaytaskapp.UpdatePutAwayTask{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &putawaytaskapp.PutAwayTask{},
			ExpResp: &putawaytaskapp.PutAwayTask{},
			CmpFunc: func(got, exp any) string { return "" },
		},
	}, "claim")

	// Step 2: Complete the task (in_progress → completed).
	test.Run(t, []apitest.Table{
		{
			Name:       "complete",
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &putawaytaskapp.UpdatePutAwayTask{
				Status: dbtest.StringPointer("completed"),
			},
			GotResp: &putawaytaskapp.PutAwayTask{},
			ExpResp: &putawaytaskapp.PutAwayTask{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*putawaytaskapp.PutAwayTask)
				if !ok {
					return "error occurred"
				}
				if gotResp.Status != "completed" {
					return fmt.Sprintf("expected status=completed, got %s", gotResp.Status)
				}
				if gotResp.CompletedBy == "" {
					return "expected completed_by to be set"
				}
				if gotResp.CompletedAt == "" {
					return "expected completed_at to be set"
				}
				return ""
			},
		},
	}, "complete")

	// Step 3: Verify a PUT_AWAY inventory transaction was created.
	productID, err := uuid.Parse(task.ProductID)
	if err != nil {
		t.Fatalf("parse product_id: %v", err)
	}
	locationID, err := uuid.Parse(task.LocationID)
	if err != nil {
		t.Fatalf("parse location_id: %v", err)
	}

	txType := "PUT_AWAY"
	txns, err := test.DB.BusDomain.InventoryTransaction.Query(ctx,
		inventorytransactionbus.QueryFilter{
			ProductID:       &productID,
			LocationID:      &locationID,
			TransactionType: &txType,
		},
		inventorytransactionbus.DefaultOrderBy,
		page.MustParse("1", "10"),
	)
	if err != nil {
		t.Fatalf("query inventory transactions: %v", err)
	}
	if len(txns) == 0 {
		t.Error("expected a PUT_AWAY inventory transaction to be created, got none")
	}

	// Step 4: Verify inventory item quantity was upserted.
	items, err := test.DB.BusDomain.InventoryItem.Query(ctx,
		inventoryitembus.QueryFilter{
			ProductID:  &productID,
			LocationID: &locationID,
		},
		inventoryitembus.DefaultOrderBy,
		page.MustParse("1", "10"),
	)
	if err != nil {
		t.Fatalf("query inventory items: %v", err)
	}
	if len(items) == 0 {
		t.Error("expected an inventory item to be upserted, got none")
	}
}

// TestUpdate400TerminalState verifies that transitioning out of a terminal state returns 400.
func TestUpdate400TerminalState(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_PutAwayTask_TerminalState")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	task := sd.PutAwayTasks[3]

	// Cancel the task first.
	test.Run(t, []apitest.Table{
		{
			Name:       "cancel",
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input:      &putawaytaskapp.UpdatePutAwayTask{Status: dbtest.StringPointer("cancelled")},
			GotResp:    &putawaytaskapp.PutAwayTask{},
			ExpResp:    &putawaytaskapp.PutAwayTask{},
			CmpFunc:    func(got, exp any) string { return "" },
		},
	}, "cancel")

	// Attempt to transition out of cancelled — expect 400.
	test.Run(t, []apitest.Table{
		{
			Name:       "transition-from-cancelled",
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input:      &putawaytaskapp.UpdatePutAwayTask{Status: dbtest.StringPointer("in_progress")},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.FailedPrecondition, "task is already cancelled and cannot be transitioned"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}, "transition-from-cancelled")
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-status-value",
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", sd.PutAwayTasks[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &putawaytaskapp.UpdatePutAwayTask{
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
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", sd.PutAwayTasks[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &putawaytaskapp.UpdatePutAwayTask{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", sd.PutAwayTasks[0].ID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &putawaytaskapp.UpdatePutAwayTask{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-update-permission",
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", sd.PutAwayTasks[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusForbidden,
			Input:      &putawaytaskapp.UpdatePutAwayTask{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.PermissionDenied, "user does not have permission UPDATE for table: inventory.put_away_tasks"),
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
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input:      &putawaytaskapp.UpdatePutAwayTask{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "put-away task not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
