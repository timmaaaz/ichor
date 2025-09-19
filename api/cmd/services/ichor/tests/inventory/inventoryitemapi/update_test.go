package inventoryinventoryitemapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", sd.InventoryItems[1].ItemID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &inventoryitemapp.UpdateInventoryItem{
				ProductID:             &sd.Products[1].ProductID,
				LocationID:            &sd.InventoryLocations[0].LocationID,
				Quantity:              dbtest.StringPointer("10"),
				ReservedQuantity:      dbtest.StringPointer("50"),
				AllocatedQuantity:     dbtest.StringPointer("100"),
				MinimumStock:          dbtest.StringPointer("1"),
				MaximumStock:          dbtest.StringPointer("100"),
				ReorderPoint:          dbtest.StringPointer("5"),
				EconomicOrderQuantity: dbtest.StringPointer("25"),
				SafetyStock:           dbtest.StringPointer("40"),
				AvgDailyUsage:         dbtest.StringPointer("6"),
			},
			GotResp: &inventoryitemapp.InventoryItem{},
			ExpResp: &inventoryitemapp.InventoryItem{
				ItemID:                sd.InventoryItems[1].ItemID,
				ProductID:             sd.Products[1].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				Quantity:              "10",
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
				CreatedDate:           sd.InventoryItems[1].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inventoryitemapp.InventoryItem)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inventoryitemapp.InventoryItem)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-location-uuid",
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", sd.InventoryItems[0].ItemID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.UpdateInventoryItem{
				ProductID:  &sd.Products[1].ProductID,
				LocationID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"location_id","error":"location_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-product-uuid",
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", sd.InventoryItems[0].ItemID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.UpdateInventoryItem{
				ProductID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-inventory-item-uuid",
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", "not-a-uuid"),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input:      &inventoryitemapp.UpdateInventoryItem{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "invalid UUID length: 10"),
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
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", sd.InventoryItems[0].ItemID),
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
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", sd.InventoryItems[0].ItemID),
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
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", sd.InventoryItems[0].ItemID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: inventory_items"),
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
			Name:       "brand-dne",
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input:      &inventoryitemapp.UpdateInventoryItem{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "query by id: inventoryItem not found"),
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
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", sd.InventoryItems[0].ItemID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventoryitemapp.UpdateInventoryItem{
				LocationID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "product-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/inventory/core/inventory-items/%s", sd.InventoryItems[0].ItemID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventoryitemapp.UpdateInventoryItem{
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
