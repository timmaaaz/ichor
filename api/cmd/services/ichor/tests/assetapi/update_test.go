package asset_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"

	"github.com/timmaaaz/ichor/app/domain/assetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/assets/%s", sd.Assets[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &assetapp.UpdateAsset{
				ValidAssetID:     &sd.Assets[0].ValidAssetID,
				LastMaintenance:  &sd.Assets[0].LastMaintenance,
				SerialNumber:     &sd.Assets[0].SerialNumber,
				AssetConditionID: &sd.Assets[0].AssetConditionID,
			},
			GotResp: &assetapp.Asset{},
			ExpResp: &assetapp.Asset{
				ID:               sd.Assets[1].ID,
				ValidAssetID:     sd.Assets[0].ValidAssetID,
				LastMaintenance:  sd.Assets[0].LastMaintenance,
				SerialNumber:     sd.Assets[0].SerialNumber,
				AssetConditionID: sd.Assets[0].AssetConditionID,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*assetapp.Asset)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*assetapp.Asset)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "the valid asset iD is malformed",
			URL:        fmt.Sprintf("/v1/assets/%s", sd.Assets[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &assetapp.UpdateAsset{
				ValidAssetID: dbtest.StringPointer(sd.Assets[0].ID[:7]),
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
