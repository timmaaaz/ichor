package inventorylocationapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/inventory-locations?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[inventorylocationapp.InventoryLocation]{},
			ExpResp: &query.Result[inventorylocationapp.InventoryLocation]{
				Page:        1,
				RowsPerPage: 10,
				Total:       25,
				Items:       sd.InventoryLocations[:10],
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByLocationCode200(sd apitest.SeedData) []apitest.Table {
	// seed_test sets location_code = "TESTLOC-001" on inventoryLocations[0].
	table := []apitest.Table{
		{
			Name:       "by-location-code",
			URL:        "/v1/inventory/inventory-locations?page=1&rows=10&location_code=TESTLOC-001",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[inventorylocationapp.InventoryLocation]{},
			ExpResp: &query.Result[inventorylocationapp.InventoryLocation]{
				Page:        1,
				RowsPerPage: 10,
				Total:       1,
				Items:       []inventorylocationapp.InventoryLocation{sd.InventoryLocations[0]},
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/inventory-locations/" + sd.InventoryLocations[0].LocationID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &inventorylocationapp.InventoryLocation{},
			ExpResp:    &sd.InventoryLocations[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
