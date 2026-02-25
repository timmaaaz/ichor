package lotlocationapi_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/lotlocationapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/lot-locations?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &query.Result[lotlocationapp.LotLocation]{},
			ExpResp: &query.Result[lotlocationapp.LotLocation]{
				Page:        1,
				RowsPerPage: 10,
				Total:       15,
				Items:       sd.LotLocations[:10],
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/lot-locations/" + sd.LotLocations[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: 200,
			Method:     "GET",
			GotResp:    &lotlocationapp.LotLocation{},
			ExpResp:    &sd.LotLocations[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
