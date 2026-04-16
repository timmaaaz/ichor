package picktaskapi_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/picktaskapp"
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
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", sd.PickTasks[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &picktaskapp.UpdatePickTask{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &picktaskapp.PickTask{},
			ExpResp: &picktaskapp.PickTask{
				ID:                   sd.PickTasks[0].ID,
				TaskNumber:           sd.PickTasks[0].TaskNumber,
				SalesOrderID:         sd.PickTasks[0].SalesOrderID,
				SalesOrderLineItemID: sd.PickTasks[0].SalesOrderLineItemID,
				ProductID:            sd.PickTasks[0].ProductID,
				LocationID:           sd.PickTasks[0].LocationID,
				QuantityToPick:       sd.PickTasks[0].QuantityToPick,
				QuantityPicked:       "0",
				Status:               "in_progress",
				CreatedBy:            sd.PickTasks[0].CreatedBy,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*picktaskapp.PickTask)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*picktaskapp.PickTask)
				expResp.AssignedTo = gotResp.AssignedTo
				expResp.AssignedAt = gotResp.AssignedAt
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

// TestUpdate200Complete verifies that completing a pick task atomically creates
// a PICK inventory transaction (negative quantity) and decrements the inventory item.
func TestUpdate200Complete(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_PickTask_Complete")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	ctx := context.Background()
	task := sd.PickTasks[2]

	// Step 1: Claim the task (pending → in_progress).
	test.Run(t, []apitest.Table{
		{
			Name:       "claim",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &picktaskapp.UpdatePickTask{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &picktaskapp.PickTask{},
			ExpResp: &picktaskapp.PickTask{},
			CmpFunc: func(got, exp any) string { return "" },
		},
	}, "claim")

	// Step 2: Complete the task (in_progress → completed).
	test.Run(t, []apitest.Table{
		{
			Name:       "complete",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &picktaskapp.UpdatePickTask{
				Status: dbtest.StringPointer("completed"),
			},
			GotResp: &picktaskapp.PickTask{},
			ExpResp: &picktaskapp.PickTask{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*picktaskapp.PickTask)
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
				if gotResp.QuantityPicked != gotResp.QuantityToPick {
					return fmt.Sprintf("expected quantity_picked=%s (full pick), got %s", gotResp.QuantityToPick, gotResp.QuantityPicked)
				}
				return ""
			},
		},
	}, "complete")

	// Step 3: Verify a PICK inventory transaction was created.
	productID, err := uuid.Parse(task.ProductID)
	if err != nil {
		t.Fatalf("parse product_id: %v", err)
	}
	locationID, err := uuid.Parse(task.LocationID)
	if err != nil {
		t.Fatalf("parse location_id: %v", err)
	}

	txType := "PICK"
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
		t.Error("expected a PICK inventory transaction to be created, got none")
	}
	if len(txns) > 0 && txns[0].Quantity >= 0 {
		t.Errorf("expected negative quantity for PICK transaction (outbound), got %d", txns[0].Quantity)
	}

	// Step 4: Verify inventory item quantity was decremented.
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
		t.Error("expected an inventory item to exist at the pick location, got none")
	}
}

// TestUpdate200ShortPicked verifies that short-picking a task requires
// quantity_picked and short_pick_reason.
func TestUpdate200ShortPicked(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_PickTask_ShortPicked")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	task := sd.PickTasks[1]

	// Step 1: Claim the task.
	test.Run(t, []apitest.Table{
		{
			Name:       "claim",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &picktaskapp.UpdatePickTask{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &picktaskapp.PickTask{},
			ExpResp: &picktaskapp.PickTask{},
			CmpFunc: func(got, exp any) string { return "" },
		},
	}, "claim")

	// Step 2: Short pick with quantity and reason.
	test.Run(t, []apitest.Table{
		{
			Name:       "short-pick",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &picktaskapp.UpdatePickTask{
				Status:          dbtest.StringPointer("short_picked"),
				QuantityPicked:  dbtest.StringPointer("2"),
				ShortPickReason: dbtest.StringPointer("insufficient stock on shelf"),
			},
			GotResp: &picktaskapp.PickTask{},
			ExpResp: &picktaskapp.PickTask{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*picktaskapp.PickTask)
				if !ok {
					return "error occurred"
				}
				if gotResp.Status != "short_picked" {
					return fmt.Sprintf("expected status=short_picked, got %s", gotResp.Status)
				}
				if gotResp.QuantityPicked != "2" {
					return fmt.Sprintf("expected quantity_picked=2, got %s", gotResp.QuantityPicked)
				}
				if gotResp.ShortPickReason != "insufficient stock on shelf" {
					return fmt.Sprintf("expected short_pick_reason set, got %q", gotResp.ShortPickReason)
				}
				return ""
			},
		},
	}, "short-pick")
}

// TestUpdate400ShortPickedMissingFields verifies that short_picked requires
// both quantity_picked and short_pick_reason.
func TestUpdate400ShortPickedMissingFields(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_PickTask_ShortPickedValidation")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	task := sd.PickTasks[0]

	// Claim first.
	test.Run(t, []apitest.Table{
		{
			Name:       "claim",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input:      &picktaskapp.UpdatePickTask{Status: dbtest.StringPointer("in_progress")},
			GotResp:    &picktaskapp.PickTask{},
			ExpResp:    &picktaskapp.PickTask{},
			CmpFunc:    func(got, exp any) string { return "" },
		},
	}, "claim")

	// Attempt short_picked without quantity_picked.
	test.Run(t, []apitest.Table{
		{
			Name:       "missing-quantity-picked",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &picktaskapp.UpdatePickTask{
				Status:          dbtest.StringPointer("short_picked"),
				ShortPickReason: dbtest.StringPointer("test reason"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "quantity_picked is required when status is short_picked"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}, "missing-quantity-picked")

	// Attempt short_picked without short_pick_reason.
	test.Run(t, []apitest.Table{
		{
			Name:       "missing-short-pick-reason",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &picktaskapp.UpdatePickTask{
				Status:         dbtest.StringPointer("short_picked"),
				QuantityPicked: dbtest.StringPointer("2"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "short_pick_reason is required when status is short_picked"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}, "missing-short-pick-reason")
}

// TestUpdate400TerminalState verifies that transitioning out of a terminal state returns 400.
func TestUpdate400TerminalState(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_PickTask_TerminalState")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	task := sd.PickTasks[3]

	// Cancel the task first.
	test.Run(t, []apitest.Table{
		{
			Name:       "cancel",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input:      &picktaskapp.UpdatePickTask{Status: dbtest.StringPointer("cancelled")},
			GotResp:    &picktaskapp.PickTask{},
			ExpResp:    &picktaskapp.PickTask{},
			CmpFunc:    func(got, exp any) string { return "" },
		},
	}, "cancel")

	// Attempt to transition out of cancelled — expect 400.
	test.Run(t, []apitest.Table{
		{
			Name:       "transition-from-cancelled",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", task.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input:      &picktaskapp.UpdatePickTask{Status: dbtest.StringPointer("in_progress")},
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
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", sd.PickTasks[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &picktaskapp.UpdatePickTask{
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
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", sd.PickTasks[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &picktaskapp.UpdatePickTask{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", sd.PickTasks[0].ID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input:      &picktaskapp.UpdatePickTask{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-update-permission",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", sd.PickTasks[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusForbidden,
			Input:      &picktaskapp.UpdatePickTask{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.PermissionDenied, "user does not have permission UPDATE for table: inventory.pick_tasks"),
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
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input:      &picktaskapp.UpdatePickTask{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "pick task not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
