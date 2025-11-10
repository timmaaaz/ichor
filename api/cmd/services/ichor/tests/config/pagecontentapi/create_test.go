package pagecontentapi_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pagecontentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
)

func create200(sd apitest.SeedData) []apitest.Table {
	layout := json.RawMessage(`{"colSpan":{"default":12}}`)

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-content",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &pagecontentapp.NewPageContent{
				PageConfigID: sd.PageContents[0].PageConfigID,
				ContentType:  pagecontentbus.ContentTypeText,
				Label:        "Test Content",
				OrderIndex:   100,
				Layout:       layout,
				IsVisible:    true,
			},
			GotResp: &pagecontentapp.PageContent{},
			ExpResp: &pagecontentapp.PageContent{
				PageConfigID: sd.PageContents[0].PageConfigID,
				ContentType:  pagecontentbus.ContentTypeText,
				Label:        "Test Content",
				OrderIndex:   100,
				IsVisible:    true,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*pagecontentapp.PageContent)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*pagecontentapp.PageContent)
				expResp.ID = gotResp.ID
				expResp.TableConfigID = gotResp.TableConfigID
				expResp.FormID = gotResp.FormID
				expResp.ParentID = gotResp.ParentID
				expResp.Layout = gotResp.Layout
				expResp.IsDefault = gotResp.IsDefault
				expResp.Children = gotResp.Children

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-page-config-id",
			URL:        "/v1/config/page-content",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      &pagecontentapp.NewPageContent{},
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"pageConfigId\",\"error\":\"pageConfigId is a required field\"},{\"field\":\"contentType\",\"error\":\"contentType is a required field\"}]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	layout := json.RawMessage(`{"colSpan":{"default":12}}`)

	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/config/page-content",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &pagecontentapp.NewPageContent{
				PageConfigID: sd.PageContents[0].PageConfigID,
				ContentType:  pagecontentbus.ContentTypeText,
				Label:        "Test Content",
				Layout:       layout,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad token",
			URL:        "/v1/config/page-content",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &pagecontentapp.NewPageContent{
				PageConfigID: sd.PageContents[0].PageConfigID,
				ContentType:  pagecontentbus.ContentTypeText,
				Label:        "Test Content",
				Layout:       layout,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        "/v1/config/page-content",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &pagecontentapp.NewPageContent{
				PageConfigID: sd.PageContents[0].PageConfigID,
				ContentType:  pagecontentbus.ContentTypeText,
				Label:        "Test Content",
				Layout:       layout,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/config/page-content",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &pagecontentapp.NewPageContent{
				PageConfigID: sd.PageContents[0].PageConfigID,
				ContentType:  pagecontentbus.ContentTypeText,
				Label:        "Test Content",
				Layout:       layout,
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
