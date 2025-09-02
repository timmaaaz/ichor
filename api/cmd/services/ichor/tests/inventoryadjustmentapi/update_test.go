package inventoryadjustmentapi_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/app/domain/movement/inventoryadjustmentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {

	now := time.Now()
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[1].InventoryAdjustmentID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
				ProductID:      &sd.Products[0].ProductID,
				LocationID:     &sd.InventoryLocations[0].LocationID,
				AdjustedBy:     &sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:     &sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange: dbtest.StringPointer("20"),
				ReasonCode:     dbtest.StringPointer("Adjustment"),
				Notes:          dbtest.StringPointer("Updated adjustment"),
				AdjustmentDate: dbtest.StringPointer(now.Format(timeutil.FORMAT)),
			},
			GotResp: &inventoryadjustmentapp.InventoryAdjustment{},
			ExpResp: &inventoryadjustmentapp.InventoryAdjustment{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				AdjustedBy:            sd.InventoryAdjustments[0].AdjustedBy,
				ApprovedBy:            sd.InventoryAdjustments[0].ApprovedBy,
				QuantityChange:        "20",
				ReasonCode:            "Adjustment",
				Notes:                 "Updated adjustment",
				AdjustmentDate:        now.Format(timeutil.FORMAT),
				InventoryAdjustmentID: sd.InventoryAdjustments[1].InventoryAdjustmentID,
				CreatedDate:           sd.InventoryAdjustments[1].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inventoryadjustmentapp.InventoryAdjustment)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inventoryadjustmentapp.InventoryAdjustment)
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
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
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
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
				LocationID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"location_id","error":"location_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-adjust-by-id",
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
				AdjustedBy: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"adjusted_by","error":"adjusted_by must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-approved-by-id",
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
				ApprovedBy: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"approved_by","error":"approved_by must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-inventory-adjustment-id",
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
				QuantityChange: dbtest.StringPointer("10"),
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
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
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
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
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
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: inventory_adjustments"),
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
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
				ProductID: &sd.Products[0].ProductID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "query by id: inventoryAdjustment not found"),
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
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
				LocationID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "adjust-by-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
				AdjustedBy: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "approved-by-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
				ApprovedBy: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/movement/inventory-adjustment/%s", sd.InventoryAdjustments[0].InventoryAdjustmentID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventoryadjustmentapp.UpdateInventoryAdjustment{
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
