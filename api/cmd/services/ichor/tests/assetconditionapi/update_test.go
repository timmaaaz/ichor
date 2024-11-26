package assetconditionapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assetconditionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assetcondition/" + sd.AssetConditions[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &assetconditionapp.UpdateAssetCondition{
				Name: dbtest.StringPointer("UpdatedAssetCondition"),
			},
			GotResp: &assetconditionapp.AssetCondition{},
			ExpResp: &assetconditionapp.AssetCondition{Name: "UpdatedAssetCondition"},
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

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-id",
			URL:        "/v1/assetcondition/abc",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: assetconditionapp.UpdateAssetCondition{
				Name: dbtest.StringPointer(""),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"name","error":"name must be at least 3 characters in length"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
