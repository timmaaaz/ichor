package transferorderapi_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/inventory/transferorderapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {

	now := time.Now()
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[1].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &transferorderapp.UpdateTransferOrder{
				ProductID:      &sd.Products[0].ProductID,
				FromLocationID: &sd.InventoryLocations[0].LocationID,
				ToLocationID:   &sd.InventoryLocations[2].LocationID,
				ApprovedByID:   &sd.TransferOrders[0].ApprovedByID,
				RequestedByID:  &sd.TransferOrders[0].RequestedByID,
				Quantity:       dbtest.StringPointer("20"),
				Status:         dbtest.StringPointer("Adjustment"),
				TransferDate:   dbtest.StringPointer(now.Format(timeutil.FORMAT)),
			},
			GotResp: &transferorderapp.TransferOrder{},
			ExpResp: &transferorderapp.TransferOrder{
				ProductID:      sd.Products[0].ProductID,
				FromLocationID: sd.InventoryLocations[0].LocationID,
				ToLocationID:   sd.InventoryLocations[2].LocationID,
				ApprovedByID:   sd.TransferOrders[0].ApprovedByID,
				RequestedByID:  sd.TransferOrders[0].RequestedByID,
				Quantity:       "20",
				Status:         "Adjustment",
				TransferDate:   now.Format(timeutil.FORMAT),
				CreatedDate:    sd.TransferOrders[1].CreatedDate,
				TransferID:     sd.TransferOrders[1].TransferID,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*transferorderapp.TransferOrder)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*transferorderapp.TransferOrder)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-product-id",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.UpdateTransferOrder{
				ProductID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-from-location-id",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.UpdateTransferOrder{
				FromLocationID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"from_location_id","error":"from_location_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-to-location-id",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.UpdateTransferOrder{
				ToLocationID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"to_location_id","error":"to_location_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-approved-by-id",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.UpdateTransferOrder{
				ApprovedByID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"approved_by","error":"approved_by must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-requested-by-id",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.UpdateTransferOrder{
				RequestedByID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"requested_by","error":"requested_by must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-inventory-adjustment-id",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &transferorderapp.UpdateTransferOrder{
				Quantity: dbtest.StringPointer("10"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `invalid UUID length: 10`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      "&nbsp",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "badsig",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Users[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: inventory.transfer_orders"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "transaction-dne",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &transferorderapp.UpdateTransferOrder{
				ProductID: &sd.Products[0].ProductID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "queryByID: transferOrderID: transferOrder not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "to-location-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.UpdateTransferOrder{
				ToLocationID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "from-location-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.UpdateTransferOrder{
				FromLocationID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "requested-by-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.UpdateTransferOrder{
				RequestedByID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "approved-by-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.UpdateTransferOrder{
				ApprovedByID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/transfer-orders/%s", sd.TransferOrders[0].TransferID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &transferorderapp.UpdateTransferOrder{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
