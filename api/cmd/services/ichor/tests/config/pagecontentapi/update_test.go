package pagecontentapi_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/go-cmp/cmp"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pagecontentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	layout := json.RawMessage(`{"colSpan":{"default":6}}`)

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-content/" + sd.PageContents[1].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &pagecontentapp.UpdatePageContent{
				Label:      dbtest.StringPointer("Updated Label"),
				OrderIndex: dbtest.IntPointer(50),
				Layout:     &layout,
				IsVisible:  dbtest.BoolPointer(false),
			},
			GotResp: &pagecontentapp.PageContent{},
			ExpResp: &pagecontentapp.PageContent{
				ID:           sd.PageContents[1].ID,
				PageConfigID: sd.PageContents[1].PageConfigID,
				ContentType:  sd.PageContents[1].ContentType,
				Label:        "Updated Label",
				OrderIndex:   50,
				IsVisible:    false,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*pagecontentapp.PageContent)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*pagecontentapp.PageContent)
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
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-uuid",
			URL:        "/v1/config/page-content/invalid-uuid",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &pagecontentapp.UpdatePageContent{
				Label: dbtest.StringPointer("Updated Label"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "invalid UUID length: 12"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "empty token",
			URL:        "/v1/config/page-content/" + sd.PageContents[1].ID,
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pagecontentapp.UpdatePageContent{
				Label: dbtest.StringPointer("Updated Label"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad token",
			URL:        "/v1/config/page-content/" + sd.PageContents[1].ID,
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pagecontentapp.UpdatePageContent{
				Label: dbtest.StringPointer("Updated Label"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad sig",
			URL:        "/v1/config/page-content/" + sd.PageContents[1].ID,
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pagecontentapp.UpdatePageContent{
				Label: dbtest.StringPointer("Updated Label"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/config/page-content/" + sd.PageContents[1].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pagecontentapp.UpdatePageContent{
				Label: dbtest.StringPointer("Updated Label"),
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

func update404(sd apitest.SeedData) []apitest.Table {
	invalidID := "00000000-0000-0000-0000-000000000000"

	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/config/page-content/" + invalidID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &pagecontentapp.UpdatePageContent{
				Label: dbtest.StringPointer("Updated Label"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "page content not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
