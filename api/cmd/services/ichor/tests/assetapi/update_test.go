package asset_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/assets/%s", sd.Assets[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &assetapp.UpdateAsset{
				TypeID:      dbtest.StringPointer(sd.AssetTypes[0].ID),
				ConditionID: dbtest.StringPointer(sd.AssetConditions[0].ID),
				Name:        dbtest.StringPointer("Updated Asset"),
				IsEnabled:   dbtest.BoolPointer(false),
				UpdatedBy:   dbtest.StringPointer(sd.Admins[0].ID.String()),
			},
			GotResp: &assetapp.Asset{},
			ExpResp: &assetapp.Asset{
				Name:                "Updated Asset",
				TypeID:              sd.AssetTypes[0].ID,
				ConditionID:         sd.AssetConditions[0].ID,
				IsEnabled:           false,
				UpdatedBy:           sd.Admins[0].ID.String(),
				EstPrice:            sd.Assets[0].EstPrice,
				MaintenanceInterval: sd.Assets[0].MaintenanceInterval,
				LifeExpectancy:      sd.Assets[0].LifeExpectancy,
				ModelNumber:         sd.Assets[0].ModelNumber,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*assetapp.Asset)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*assetapp.Asset)
				expResp.ID = gotResp.ID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated
				expResp.CreatedBy = gotResp.CreatedBy

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing type",
			URL:        fmt.Sprintf("/v1/assets/%s", sd.Assets[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &assetapp.UpdateAsset{
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
			URL:        fmt.Sprintf("/v1/assets/%s", sd.Assets[0].ID),
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
			URL:        fmt.Sprintf("/v1/assets/%s", sd.Assets[0].ID),
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
			URL:        fmt.Sprintf("/v1/assets/%s", sd.Assets[0].ID),
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
