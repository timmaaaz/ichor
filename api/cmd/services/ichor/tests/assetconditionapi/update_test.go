package assetcondition_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/assetconditionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/assets/assetconditions/%s", sd.AssetConditions[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &assetconditionapp.UpdateAssetCondition{
				Name:        dbtest.StringPointer("Updated New Asset Type"),
				Description: dbtest.StringPointer("Updated Optional Description"),
			},
			GotResp: &assetconditionapp.AssetCondition{},
			ExpResp: &assetconditionapp.AssetCondition{
				ID:          sd.AssetConditions[0].ID,
				Name:        "Updated New Asset Type",
				Description: "Updated Optional Description",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*assetconditionapp.AssetCondition)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*assetconditionapp.AssetCondition)
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid name",
			URL:        fmt.Sprintf("/v1/assets/assetconditions/%s", sd.AssetConditions[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &assetconditionapp.UpdateAssetCondition{
				Name: dbtest.StringPointer("a"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name must be at least 3 characters in length\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp)
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/assets/assetconditions/%s", sd.AssetConditions[0].ID),
			Token:      "&nbsp;",
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
			URL:        fmt.Sprintf("/v1/assets/assetconditions/%s", sd.AssetConditions[0].ID),
			Token:      sd.Admins[0].Token + "bad",
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
			URL:        fmt.Sprintf("/v1/assets/assetconditions/%s", sd.AssetConditions[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: asset_conditions"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
