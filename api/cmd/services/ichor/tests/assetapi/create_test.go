package asset_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assetapp"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assets",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &assetapp.NewAsset{
				Name:   "New Asset",
				TypeID: sd.AssetTypes[0].ID,
			},
			GotResp: &assetapp.Asset{},
			ExpResp: &assetapp.Asset{
				Name: "New Asset",
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*assetapp.Asset)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*assetapp.Asset)
				expResp.ID = gotResp.ID

				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
