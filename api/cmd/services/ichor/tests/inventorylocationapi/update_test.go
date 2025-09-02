package inventorylocationapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/warehouse/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/warehouses/inventory-locations/%s", sd.InventoryLocations[0].LocationID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &inventorylocationapp.UpdateInventoryLocation{
				WarehouseID:        &sd.Warehouses[0].ID,
				ZoneID:             &sd.Zones[0].ZoneID,
				Aisle:              dbtest.StringPointer("UpdateAisle"),
				Rack:               dbtest.StringPointer("UpdateRack"),
				Shelf:              dbtest.StringPointer("UpdateShelf"),
				Bin:                dbtest.StringPointer("UpdateBin"),
				IsPickLocation:     dbtest.StringPointer("True"),
				IsReserveLocation:  dbtest.StringPointer("false"),
				MaxCapacity:        dbtest.StringPointer("5"),
				CurrentUtilization: dbtest.StringPointer("0.57"),
			},
			GotResp: &inventorylocationapp.InventoryLocation{},
			ExpResp: &inventorylocationapp.InventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "UpdateAisle",
				Rack:               "UpdateRack",
				Shelf:              "UpdateShelf",
				Bin:                "UpdateBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "false",
				MaxCapacity:        "5",
				CurrentUtilization: "0.57",
				LocationID:         sd.InventoryLocations[0].LocationID,
				CreatedDate:        sd.InventoryLocations[0].CreatedDate,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inventorylocationapp.InventoryLocation)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inventorylocationapp.InventoryLocation)
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "malformed-warehouse-id",
			URL:        fmt.Sprintf("/v1/warehouses/inventory-locations/%s", sd.InventoryLocations[0].LocationID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.UpdateInventoryLocation{
				WarehouseID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"warehouse_id","error":"warehouse_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-zone-id",
			URL:        fmt.Sprintf("/v1/warehouses/inventory-locations/%s", sd.InventoryLocations[0].LocationID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.UpdateInventoryLocation{
				ZoneID: dbtest.StringPointer("not-a-uuid"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"zone_id","error":"zone_id must be at least 36 characters in length"}]`),
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
			URL:        fmt.Sprintf("/v1/warehouses/inventory-locations/%s", sd.InventoryLocations[0].LocationID),
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
			URL:        fmt.Sprintf("/v1/warehouses/inventory-locations/%s", sd.InventoryLocations[0].LocationID),
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
			URL:        fmt.Sprintf("/v1/warehouses/inventory-locations/%s", sd.InventoryLocations[0].LocationID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: inventory_locations"),
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
			Name:       "zone-dne",
			URL:        fmt.Sprintf("/v1/warehouses/inventory-locations/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &inventorylocationapp.UpdateInventoryLocation{
				WarehouseID: &sd.InventoryLocations[0].WarehouseID,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "namedexeccontext: inventoryLocation not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "contact-id-dne-as-fk",
			URL:        fmt.Sprintf("/v1/warehouses/inventory-locations/%s", sd.InventoryLocations[0].LocationID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusConflict,
			Input: &inventorylocationapp.UpdateInventoryLocation{
				WarehouseID: dbtest.StringPointer(uuid.NewString()),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "update: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
