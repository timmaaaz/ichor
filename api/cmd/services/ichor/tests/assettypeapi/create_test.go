package assettype_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assettypeapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assettype",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &assettypeapp.NewAssetType{
				Name: "TestAssetType",
			},
			GotResp: &assettypeapp.AssetType{},
			ExpResp: &assettypeapp.AssetType{
				Name: "TestAssetType",
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*assettypeapp.AssetType)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*assettypeapp.AssetType)
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
			URL:        "/v1/assettype",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &assettypeapp.NewAssetType{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "Validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
