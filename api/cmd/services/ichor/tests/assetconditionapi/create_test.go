package assetconditionapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assetconditionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assetcondition",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &assetconditionapp.NewAssetCondition{
				Name: "TestAssetCondition",
			},
			GotResp: &assetconditionapp.AssetCondition{},
			ExpResp: &assetconditionapp.AssetCondition{
				Name: "TestAssetCondition",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*assetconditionapp.AssetCondition)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*assetconditionapp.AssetCondition)
				expResp.ID = gotResp.ID

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {

	table := []apitest.Table{
		{
			Name:       "missing name",
			URL:        "/v1/assetcondition",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &assetconditionapp.NewAssetCondition{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
