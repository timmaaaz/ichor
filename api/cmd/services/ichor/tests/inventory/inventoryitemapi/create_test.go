package inventoryinventoryitemapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
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
			},
			GotResp: &inventoryitemapp.InventoryItem{},
			ExpResp: &inventoryitemapp.InventoryItem{
				ProductID:             sd.Products[0].ProductID,
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
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*inventoryitemapp.InventoryItem)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*inventoryitemapp.InventoryItem)
				expResp.ID = gotResp.ID
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
			Name:       "missing-product-id",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
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
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"product_id\",\"error\":\"product_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-location-id",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				Quantity:              "10",
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"location_id\",\"error\":\"location_id is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-quantity",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"quantity\",\"error\":\"quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-reserved-quantity",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				Quantity:              "10",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"reserved_quantity\",\"error\":\"reserved_quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-allocated-quantity",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				Quantity:              "10",
				ReservedQuantity:      "50",
				MinimumStock:          "1",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"allocated_quantity\",\"error\":\"allocated_quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-minimum-stock",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				Quantity:              "10",
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"minimum_stock\",\"error\":\"minimum_stock is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-maximum-stock",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				Quantity:              "10",
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"maximum_stock\",\"error\":\"maximum_stock is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-reorder-point",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				Quantity:              "10",
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				MaximumStock:          "100",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"reorder_point\",\"error\":\"reorder_point is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-economic-order-quantity",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:         sd.Products[0].ProductID,
				LocationID:        sd.InventoryLocations[0].LocationID,
				Quantity:          "10",
				ReservedQuantity:  "50",
				AllocatedQuantity: "100",
				MinimumStock:      "1",
				MaximumStock:      "100",
				ReorderPoint:      "5",
				SafetyStock:       "40",
				AvgDailyUsage:     "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"economic_order_quantity\",\"error\":\"economic_order_quantity is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-safety-stock",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				Quantity:              "10",
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				AvgDailyUsage:         "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"safety_stock\",\"error\":\"safety_stock is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-avg-daily-usage",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            sd.InventoryLocations[0].LocationID,
				Quantity:              "10",
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"avg_daily_usage","error":"avg_daily_usage is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "malformed-handling-units-per-case",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             "not-a-uuid",
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
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"product_id","error":"product_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "malformed-handling-units-per-case",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            "not-a-uuid",
				Quantity:              "10",
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"location_id","error":"location_id must be at least 36 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "product-id-is-not-a-valid-fk",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             uuid.NewString(),
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
			},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			GotResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "location-id-is-not-valid-fk",
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &inventoryitemapp.NewInventoryItem{
				ProductID:             sd.Products[0].ProductID,
				LocationID:            uuid.NewString(),
				Quantity:              "10",
				ReservedQuantity:      "50",
				AllocatedQuantity:     "100",
				MinimumStock:          "1",
				MaximumStock:          "100",
				ReorderPoint:          "5",
				EconomicOrderQuantity: "25",
				SafetyStock:           "40",
				AvgDailyUsage:         "6",
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
			URL:        "/v1/inventory/inventory-items",
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
			URL:        "/v1/inventory/inventory-items",
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
			URL:        "/v1/inventory/inventory-items",
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
			URL:        "/v1/inventory/inventory-items",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: inventory.inventory_items"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
