package pagecontentapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pagecontentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func queryByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-content/" + sd.PageContents[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &pagecontentapp.PageContent{},
			ExpResp: &pagecontentapp.PageContent{
				ID:           sd.PageContents[0].ID,
				PageConfigID: sd.PageContents[0].PageConfigID,
				ContentType:  sd.PageContents[0].ContentType,
				Label:        sd.PageContents[0].Label,
				OrderIndex:   sd.PageContents[0].OrderIndex,
				IsVisible:    sd.PageContents[0].IsVisible,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*pagecontentapp.PageContent)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*pagecontentapp.PageContent)
				expResp.TableConfigID = gotResp.TableConfigID
				expResp.FormID = gotResp.FormID
				expResp.ChartConfigID = gotResp.ChartConfigID
				expResp.ParentID = gotResp.ParentID
				expResp.Layout = gotResp.Layout
				expResp.IsDefault = gotResp.IsDefault
				expResp.Children = gotResp.Children

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func queryByID400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-id",
			URL:        "/v1/config/page-content/invalid-uuid",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusBadRequest,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "invalid UUID length: 12"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryByPageConfigID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-configs/content/" + sd.PageContents[0].PageConfigID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &pagecontentapp.PageContents{},
			ExpResp:    &pagecontentapp.PageContents{},
			CmpFunc: func(got, exp any) string {
				return ""
			},
		},
	}
}

func queryWithChildren200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-configs/content/children/" + sd.PageContents[0].PageConfigID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &pagecontentapp.PageContents{},
			ExpResp:    &pagecontentapp.PageContents{},
			CmpFunc: func(got, exp any) string {
				return ""
			},
		},
	}
}
