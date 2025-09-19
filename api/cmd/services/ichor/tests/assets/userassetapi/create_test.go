package userasset_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/userassetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/assets/user-assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &userassetapp.NewUserAsset{
				UserID:              sd.UserAssets[0].UserID,
				AssetID:             sd.UserAssets[0].AssetID,
				ApprovedBy:          sd.UserAssets[0].ApprovedBy,
				ApprovalStatusID:    sd.UserAssets[0].ApprovalStatusID,
				FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
				DateReceived:        sd.UserAssets[0].DateReceived,
				LastMaintenance:     sd.UserAssets[0].LastMaintenance,
			},
			GotResp: &userassetapp.UserAsset{},
			ExpResp: &userassetapp.UserAsset{
				UserID:              sd.UserAssets[0].UserID,
				AssetID:             sd.UserAssets[0].AssetID,
				ApprovedBy:          sd.UserAssets[0].ApprovedBy,
				ApprovalStatusID:    sd.UserAssets[0].ApprovalStatusID,
				FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
				DateReceived:        sd.UserAssets[0].DateReceived,
				LastMaintenance:     sd.UserAssets[0].LastMaintenance,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*userassetapp.UserAsset)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*userassetapp.UserAsset)
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
			Name:       "missing user_id",
			URL:        "/v1/assets/user-assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userassetapp.NewUserAsset{
				AssetID:             sd.UserAssets[0].AssetID,
				ApprovedBy:          sd.UserAssets[0].ApprovedBy,
				ApprovalStatusID:    sd.UserAssets[0].ApprovalStatusID,
				FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
				DateReceived:        sd.UserAssets[0].DateReceived,
				LastMaintenance:     sd.UserAssets[0].LastMaintenance,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"user_id\",\"error\":\"user_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing asset_id",
			URL:        "/v1/assets/user-assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userassetapp.NewUserAsset{
				UserID:              sd.UserAssets[0].UserID,
				ApprovedBy:          sd.UserAssets[0].ApprovedBy,
				ApprovalStatusID:    sd.UserAssets[0].ApprovalStatusID,
				FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
				DateReceived:        sd.UserAssets[0].DateReceived,
				LastMaintenance:     sd.UserAssets[0].LastMaintenance,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"asset_id\",\"error\":\"asset_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing approved by",
			URL:        "/v1/assets/user-assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userassetapp.NewUserAsset{
				UserID:              sd.UserAssets[0].UserID,
				AssetID:             sd.UserAssets[0].AssetID,
				ApprovalStatusID:    sd.UserAssets[0].ApprovalStatusID,
				FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
				DateReceived:        sd.UserAssets[0].DateReceived,
				LastMaintenance:     sd.UserAssets[0].LastMaintenance,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"approved_by\",\"error\":\"approved_by is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},

		{
			Name:       "missing approval status id",
			URL:        "/v1/assets/user-assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userassetapp.NewUserAsset{
				UserID:              sd.UserAssets[0].UserID,
				AssetID:             sd.UserAssets[0].AssetID,
				ApprovedBy:          sd.UserAssets[0].ApprovedBy,
				FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
				DateReceived:        sd.UserAssets[0].DateReceived,
				LastMaintenance:     sd.UserAssets[0].LastMaintenance,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"approval_status_id\",\"error\":\"approval_status_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing fulfillment status id",
			URL:        "/v1/assets/user-assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userassetapp.NewUserAsset{
				UserID:           sd.UserAssets[0].UserID,
				AssetID:          sd.UserAssets[0].AssetID,
				ApprovedBy:       sd.UserAssets[0].ApprovedBy,
				ApprovalStatusID: sd.UserAssets[0].ApprovalStatusID,
				DateReceived:     sd.UserAssets[0].DateReceived,
				LastMaintenance:  sd.UserAssets[0].LastMaintenance,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"fulfillment_status_id\",\"error\":\"fulfillment_status_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing date received",
			URL:        "/v1/assets/user-assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userassetapp.NewUserAsset{
				UserID:              sd.UserAssets[0].UserID,
				AssetID:             sd.UserAssets[0].AssetID,
				ApprovedBy:          sd.UserAssets[0].ApprovedBy,
				ApprovalStatusID:    sd.UserAssets[0].ApprovalStatusID,
				FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
				LastMaintenance:     sd.UserAssets[0].LastMaintenance,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"date_received\",\"error\":\"date_received is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing last maintenance",
			URL:        "/v1/assets/user-assets",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &userassetapp.NewUserAsset{
				UserID:              sd.UserAssets[0].UserID,
				AssetID:             sd.UserAssets[0].AssetID,
				ApprovedBy:          sd.UserAssets[0].ApprovedBy,
				ApprovalStatusID:    sd.UserAssets[0].ApprovalStatusID,
				FulfillmentStatusID: sd.UserAssets[0].FulfillmentStatusID,
				DateReceived:        sd.UserAssets[0].DateReceived,
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
			URL:        "/v1/assets/user-assets",
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
			URL:        "/v1/assets/user-assets",
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
			URL:        "/v1/assets/user-assets",
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
			URL:        "/v1/assets/user-assets",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: user_assets"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
