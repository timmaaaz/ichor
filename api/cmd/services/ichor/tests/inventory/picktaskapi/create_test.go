package picktaskapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/picktaskapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/pick-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &picktaskapp.NewPickTask{
				SalesOrderID:         sd.PickTasks[0].SalesOrderID,
				SalesOrderLineItemID: sd.PickTasks[0].SalesOrderLineItemID,
				ProductID:            sd.Products[0].ProductID,
				LocationID:           sd.InventoryLocations[0].LocationID,
				QuantityToPick:       "5",
			},
			GotResp: &picktaskapp.PickTask{},
			ExpResp: &picktaskapp.PickTask{
				SalesOrderID:         sd.PickTasks[0].SalesOrderID,
				SalesOrderLineItemID: sd.PickTasks[0].SalesOrderLineItemID,
				ProductID:            sd.Products[0].ProductID,
				LocationID:           sd.InventoryLocations[0].LocationID,
				QuantityToPick:       "5",
				QuantityPicked:       "0",
				Status:               "pending",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*picktaskapp.PickTask)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*picktaskapp.PickTask)
				expResp.ID = gotResp.ID
				expResp.CreatedBy = gotResp.CreatedBy
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-sales-order-id",
			URL:        "/v1/inventory/pick-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &picktaskapp.NewPickTask{
				SalesOrderLineItemID: sd.PickTasks[0].SalesOrderLineItemID,
				ProductID:            sd.Products[0].ProductID,
				LocationID:           sd.InventoryLocations[0].LocationID,
				QuantityToPick:       "5",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"sales_order_id","error":"sales_order_id is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-product-id",
			URL:        "/v1/inventory/pick-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &picktaskapp.NewPickTask{
				SalesOrderID:         sd.PickTasks[0].SalesOrderID,
				SalesOrderLineItemID: sd.PickTasks[0].SalesOrderLineItemID,
				LocationID:           sd.InventoryLocations[0].LocationID,
				QuantityToPick:       "5",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-quantity-to-pick",
			URL:        "/v1/inventory/pick-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &picktaskapp.NewPickTask{
				SalesOrderID:         sd.PickTasks[0].SalesOrderID,
				SalesOrderLineItemID: sd.PickTasks[0].SalesOrderLineItemID,
				ProductID:            sd.Products[0].ProductID,
				LocationID:           sd.InventoryLocations[0].LocationID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"quantity_to_pick","error":"quantity_to_pick is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-product-fk",
			URL:        "/v1/inventory/pick-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &picktaskapp.NewPickTask{
				SalesOrderID:         sd.PickTasks[0].SalesOrderID,
				SalesOrderLineItemID: sd.PickTasks[0].SalesOrderLineItemID,
				ProductID:            uuid.NewString(),
				LocationID:           sd.InventoryLocations[0].LocationID,
				QuantityToPick:       "5",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "invalid-location-fk",
			URL:        "/v1/inventory/pick-tasks",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &picktaskapp.NewPickTask{
				SalesOrderID:         sd.PickTasks[0].SalesOrderID,
				SalesOrderLineItemID: sd.PickTasks[0].SalesOrderLineItemID,
				ProductID:            sd.Products[0].ProductID,
				LocationID:           uuid.NewString(),
				QuantityToPick:       "5",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/inventory/pick-tasks",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        "/v1/inventory/pick-tasks",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-create-permission",
			URL:        "/v1/inventory/pick-tasks",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusForbidden,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.PermissionDenied, "user does not have permission CREATE for table: inventory.pick_tasks"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
