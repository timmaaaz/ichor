package inventorylocationapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &inventorylocationapp.InventoryLocation{},
			ExpResp: &inventorylocationapp.InventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "false",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inventorylocationapp.InventoryLocation)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inventorylocationapp.InventoryLocation)
				expResp.LocationID = gotResp.LocationID
				expResp.UpdatedDate = gotResp.UpdatedDate
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-warehouse-id",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"warehouse_id\",\"error\":\"warehouse_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-zone-id",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"zone_id\",\"error\":\"zone_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-aisle",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"aisle\",\"error\":\"aisle is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-rack",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"rack\",\"error\":\"rack is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-shelf",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"shelf\",\"error\":\"shelf is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-bin",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"bin\",\"error\":\"bin is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-is-pick-location",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"is_pick_location\",\"error\":\"is_pick_location is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-is-reserve-location",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"is_reserve_location\",\"error\":\"is_reserve_location is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-max-capacity",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"max_capacity\",\"error\":\"max_capacity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-current-utilization",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:       sd.Warehouses[0].ID,
				ZoneID:            sd.Zones[0].ZoneID,
				Aisle:             "NewAisle",
				Rack:              "NewRack",
				Shelf:             "NewShelf",
				Bin:               "NewBin",
				IsPickLocation:    "true",
				IsReserveLocation: "f",
				MaxCapacity:       "100",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"current_utilization\",\"error\":\"current_utilization is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-warehouse-id",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        "not-a-uuid",
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"warehouse_id\",\"error\":\"warehouse_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-warehouse-id",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             "not-a-uuid",
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"zone_id\",\"error\":\"zone_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "warehouse-id-not-valid-fk",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        uuid.New().String(),
				ZoneID:             sd.Zones[0].ZoneID,
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "zone-id-not-valid-fk",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventorylocationapp.NewInventoryLocation{
				WarehouseID:        sd.Warehouses[0].ID,
				ZoneID:             uuid.New().String(),
				Aisle:              "NewAisle",
				Rack:               "NewRack",
				Shelf:              "NewShelf",
				Bin:                "NewBin",
				IsPickLocation:     "true",
				IsReserveLocation:  "f",
				MaxCapacity:        "100",
				CurrentUtilization: "99.99",
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/inventory/inventory-locations",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad token",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/inventory/inventory-locations",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: inventory_locations"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
