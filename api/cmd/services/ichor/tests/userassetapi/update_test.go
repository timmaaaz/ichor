package userasset_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/assets/userassetapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/assets/userassets/%s", sd.UserAssets[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &userassetapp.UpdateUserAsset{
				UserID:              &sd.UserAssets[1].UserID,
				AssetID:             &sd.UserAssets[1].AssetID,
				ApprovedBy:          &sd.UserAssets[1].ApprovedBy,
				ApprovalStatusID:    &sd.UserAssets[1].ApprovalStatusID,
				FulfillmentStatusID: &sd.UserAssets[1].FulfillmentStatusID,
				DateReceived:        &sd.UserAssets[1].DateReceived,
				LastMaintenance:     &sd.UserAssets[1].LastMaintenance,
			},
			GotResp: &userassetapp.UserAsset{},
			ExpResp: &userassetapp.UserAsset{
				ID:                  sd.UserAssets[0].ID,
				UserID:              sd.UserAssets[1].UserID,
				AssetID:             sd.UserAssets[1].AssetID,
				ApprovedBy:          sd.UserAssets[1].ApprovedBy,
				ApprovalStatusID:    sd.UserAssets[1].ApprovalStatusID,
				FulfillmentStatusID: sd.UserAssets[1].FulfillmentStatusID,
				DateReceived:        sd.UserAssets[1].DateReceived,
				LastMaintenance:     sd.UserAssets[1].LastMaintenance,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*userassetapp.UserAsset)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*userassetapp.UserAsset)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing id for update",
			URL:        fmt.Sprintf("/v1/assets/userassets/%s", sd.UserAssets[0].ID[:6]),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input:      &userassetapp.UpdateUserAsset{},
			GotResp:    &struct{}{},
			ExpResp:    &struct{}{},
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
			URL:        fmt.Sprintf("/v1/assets/userassets/%s", sd.UserAssets[0].ID),
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
			URL:        fmt.Sprintf("/v1/assets/userassets/%s", sd.UserAssets[0].ID),
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
			URL:        fmt.Sprintf("/v1/assets/userassets/%s", sd.UserAssets[0].ID),
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
