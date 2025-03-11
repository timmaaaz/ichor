package validasset_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assets/validassets?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[validassetapp.ValidAsset]{},
			ExpResp: &query.Result[validassetapp.ValidAsset]{
				Page:        1,
				RowsPerPage: 10,
				Total:       20,
				Items:       sd.ValidAssets[:10],
			},
			CmpFunc: func(got any, exp any) string {

				// modify exp result to remove protected fields
				expResp, ok := exp.(*query.Result[validassetapp.ValidAsset])
				if !ok {
					return "error occurred"
				}
				gotResp, ok := got.(*query.Result[validassetapp.ValidAsset])
				if !ok {
					return "error occurred"
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assets/validassets/" + sd.ValidAssets[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &validassetapp.ValidAsset{},
			ExpResp:    &sd.ValidAssets[0],
			CmpFunc: func(got any, exp any) string {

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
