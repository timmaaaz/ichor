package assettype_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/assettypeapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{Name: "basic",
			URL:        fmt.Sprintf("/v1/assets/assettypes/%s", sd.AssetTypes[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &assettypeapp.UpdateAssetType{
				Name:        dbtest.StringPointer("Updated New Asset Type"),
				Description: dbtest.StringPointer("Updated Optional Description"),
			},
			GotResp: &assettypeapp.AssetType{},
			ExpResp: &assettypeapp.AssetType{
				ID:          sd.AssetTypes[0].ID,
				Name:        "Updated New Asset Type",
				Description: "Updated Optional Description",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*assettypeapp.AssetType)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*assettypeapp.AssetType)
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{Name: "invalid name",
			URL:        fmt.Sprintf("/v1/assets/assettypes/%s", sd.AssetTypes[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &assettypeapp.UpdateAssetType{
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
			URL:        fmt.Sprintf("/v1/assets/assettypes/%s", sd.AssetTypes[0].ID),
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
			URL:        fmt.Sprintf("/v1/assets/assettypes/%s", sd.AssetTypes[0].ID),
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
			URL:        fmt.Sprintf("/v1/assets/assettypes/%s", sd.AssetTypes[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
