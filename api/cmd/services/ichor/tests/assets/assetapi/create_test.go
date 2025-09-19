package asset_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/assetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assets/assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &assetapp.NewAsset{
				AssetConditionID: sd.Assets[0].AssetConditionID,
				ValidAssetID:     sd.Assets[0].ValidAssetID,
				LastMaintenance:  sd.Assets[0].LastMaintenance,
				SerialNumber:     sd.Assets[0].SerialNumber,
			},
			GotResp: &assetapp.Asset{},
			ExpResp: &assetapp.Asset{
				AssetConditionID: sd.Assets[0].AssetConditionID,
				ValidAssetID:     sd.Assets[0].ValidAssetID,
				LastMaintenance:  sd.Assets[0].LastMaintenance,
				SerialNumber:     sd.Assets[0].SerialNumber,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*assetapp.Asset)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*assetapp.Asset)
				expResp.ID = gotResp.ID

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing serial number",
			URL:        "/v1/assets/assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &assetapp.NewAsset{
				ValidAssetID:     sd.Assets[0].ValidAssetID,
				AssetConditionID: sd.Assets[0].AssetConditionID,
				LastMaintenance:  sd.Assets[0].LastMaintenance,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"serial_number\",\"error\":\"serial_number is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing valid asset id",
			URL:        "/v1/assets/assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &assetapp.NewAsset{
				AssetConditionID: sd.Assets[0].AssetConditionID,
				LastMaintenance:  sd.Assets[0].LastMaintenance,
				SerialNumber:     sd.Assets[0].SerialNumber,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"valid_asset_id\",\"error\":\"valid_asset_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing condition id",
			URL:        "/v1/assets/assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &assetapp.NewAsset{
				ValidAssetID:    sd.Assets[0].ValidAssetID,
				LastMaintenance: sd.Assets[0].LastMaintenance,
				SerialNumber:    sd.Assets[0].SerialNumber,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"asset_condition_id\",\"error\":\"asset_condition_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing last maintenance time",
			URL:        "/v1/assets/assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &assetapp.NewAsset{
				ValidAssetID:     sd.Assets[0].ValidAssetID,
				AssetConditionID: sd.Assets[0].AssetConditionID,
				SerialNumber:     sd.Assets[0].SerialNumber,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"last_maintenance\",\"error\":\"last_maintenance is a required field\"}]"),
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
			URL:        "/v1/assets/assets",
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
			URL:        "/v1/assets/assets",
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
			URL:        "/v1/assets/assets",
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
			URL:        "/v1/assets/assets",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: assets"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
