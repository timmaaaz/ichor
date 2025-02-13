package validasset_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/validassetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assets/validassets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &validassetapp.NewValidAsset{
				TypeID:    sd.AssetTypes[0].ID,
				Name:      "New Valid Asset",
				IsEnabled: true,
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &validassetapp.ValidAsset{},
			ExpResp: &validassetapp.ValidAsset{
				Name:      "New Valid Asset",
				TypeID:    sd.AssetTypes[0].ID,
				IsEnabled: true,
				CreatedBy: sd.Admins[0].ID.String(),
				UpdatedBy: sd.Admins[0].ID.String(),
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

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing type",
			URL:        "/v1/assets/validassets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &validassetapp.NewValidAsset{
				Name:      "New Asset",
				IsEnabled: true,
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"type_id\",\"error\":\"type_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing name",
			URL:        "/v1/assets/validassets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &validassetapp.NewValidAsset{
				TypeID:    sd.AssetTypes[0].ID,
				IsEnabled: true,
				CreatedBy: sd.Admins[0].ID.String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/assets/validassets",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad token",
			URL:        "/v1/assets/validassets",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        "/v1/assets/validassets",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/assets/validassets",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
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
