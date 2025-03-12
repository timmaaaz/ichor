package validasset_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"

	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/assets/validassets/%s", sd.ValidAssets[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &validassetapp.UpdateValidAsset{
				TypeID:    dbtest.StringPointer(sd.AssetTypes[0].ID),
				Name:      dbtest.StringPointer("Updated Asset"),
				IsEnabled: dbtest.BoolPointer(false),
				UpdatedBy: dbtest.StringPointer(sd.Admins[0].ID.String()),
			},
			GotResp: &validassetapp.ValidAsset{},
			ExpResp: &validassetapp.ValidAsset{
				Name:                "Updated Asset",
				TypeID:              sd.AssetTypes[0].ID,
				IsEnabled:           false,
				UpdatedBy:           sd.Admins[0].ID.String(),
				EstPrice:            sd.ValidAssets[0].EstPrice,
				MaintenanceInterval: sd.ValidAssets[0].MaintenanceInterval,
				LifeExpectancy:      sd.ValidAssets[0].LifeExpectancy,
				ModelNumber:         sd.ValidAssets[0].ModelNumber,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*validassetapp.ValidAsset)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*validassetapp.ValidAsset)
				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated
				expResp.CreatedBy = gotResp.CreatedBy

				// NOTES: This is a protected field and will be returned if you
				// created it, but it will not be returned from queries if you
				// do not have permission to see it and therefore wasn't
				expResp.EstPrice = gotResp.EstPrice

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing type",
			URL:        fmt.Sprintf("/v1/assets/validassets/%s", sd.ValidAssets[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &validassetapp.UpdateValidAsset{
				TypeID: dbtest.StringPointer("asdf"),
			},
			GotResp: &struct{}{},
			ExpResp: &struct{}{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/assets/validassets/%s", sd.ValidAssets[0].ID),
			Token:      "&nbsp",
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
			URL:        fmt.Sprintf("/v1/assets/validassets/%s", sd.ValidAssets[0].ID),
			Token:      sd.Users[0].Token + "A",
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
			URL:        fmt.Sprintf("/v1/assets/validassets/%s", sd.ValidAssets[0].ID),
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
