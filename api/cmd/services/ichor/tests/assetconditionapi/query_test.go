package assetcondition_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/assetconditionapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assets/assetconditions?page=1&rows=10",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[assetconditionapp.AssetCondition]{},
			ExpResp: &query.Result[assetconditionapp.AssetCondition]{
				Page:        1,
				RowsPerPage: 10,
				Total:       15,
				Items:       sd.AssetConditions,
			},
			CmpFunc: func(got any, exp any) string {
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
			URL:        "/v1/assets/assetconditions/" + sd.AssetConditions[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &assetconditionapp.AssetCondition{},
			ExpResp:    &sd.AssetConditions[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
