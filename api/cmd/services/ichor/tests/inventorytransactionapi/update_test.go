package inventorytransactionapi_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/movement/inventorytransactionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {

	now := time.Now()
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &inventorytransactionapp.UpdateInventoryTransaction{
				ProductID:       &sd.Products[0].ProductID,
				LocationID:      &sd.InventoryLocations[0].LocationID,
				UserID:          dbtest.StringPointer(sd.Users[0].ID.String()),
				Quantity:        dbtest.StringPointer("10"),
				TransactionType: dbtest.StringPointer("IN"),
				ReferenceNumber: dbtest.StringPointer("UpdateReferenceNumber"),
				TransactionDate: dbtest.StringPointer(now.Format(timeutil.FORMAT)),
			},
			GotResp: &inventorytransactionapp.InventoryTransaction{},
			ExpResp: &inventorytransactionapp.InventoryTransaction{
				ProductID:              sd.Products[0].ProductID,
				LocationID:             sd.InventoryLocations[0].LocationID,
				UserID:                 sd.Users[0].ID.String(),
				Quantity:               "10",
				TransactionType:        "IN",
				ReferenceNumber:        "UpdateReferenceNumber",
				TransactionDate:        now.Format(timeutil.FORMAT),
				InventoryTransactionID: sd.InventoryTransactions[0].InventoryTransactionID,
				CreatedDate:            sd.InventoryTransactions[0].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inventorytransactionapp.InventoryTransaction)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inventorytransactionapp.InventoryTransaction)
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
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.UpdateInventoryTransaction{
				ProductID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-location-id",
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.UpdateInventoryTransaction{
				LocationID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"location_id","error":"location_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-user-id",
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.UpdateInventoryTransaction{
				UserID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"user_id","error":"user_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-inventory-location-id",
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventorytransactionapp.UpdateInventoryTransaction{
				TransactionType: dbtest.StringPointer("IN"),
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
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
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
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
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
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: inventory_transactions"),
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
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &inventorytransactionapp.UpdateInventoryTransaction{
				ProductID: &sd.Products[0].ProductID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "queryByID: inventoryTransaction not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "location-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventorytransactionapp.UpdateInventoryTransaction{
				LocationID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "user-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventorytransactionapp.UpdateInventoryTransaction{
				UserID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/movement/inventory-transactions/%s", sd.InventoryTransactions[0].InventoryTransactionID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventorytransactionapp.UpdateInventoryTransaction{
				ProductID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
