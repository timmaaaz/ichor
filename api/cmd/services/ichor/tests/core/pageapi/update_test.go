package page_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/pageapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/pages/" + sd.Pages[0].ID,
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &pageapp.UpdatePage{
				Name:      dbtest.StringPointer("Updated Page Name"),
				Path:      dbtest.StringPointer("/updated/path"),
				Module:    dbtest.StringPointer("updated-module"),
				Icon:      dbtest.StringPointer("updated-icon"),
				SortOrder: dbtest.IntPointer(5000),
				IsActive:  dbtest.BoolPointer(false),
			},
			GotResp: &pageapp.Page{},
			ExpResp: &pageapp.Page{
				ID:         sd.Pages[0].ID,
				Name:       "Updated Page Name",
				Path:       "/updated/path",
				Module:     "updated-module",
				Icon:       "updated-icon",
				SortOrder:  5000,
				IsActive:   false,
				ShowInMenu: sd.Pages[0].ShowInMenu,
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "invalid-id",
			URL:        "/v1/core/pages/invalid-uuid",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &pageapp.UpdatePage{
				Name: dbtest.StringPointer("Test"),
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				// Just check it's an error, don't compare specific messages
				return ""
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unauthorized",
			URL:        "/v1/core/pages/" + sd.Pages[0].ID,
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &pageapp.UpdatePage{
				Name: dbtest.StringPointer("Unauthorized Update"),
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				// Just check it's an error, don't compare specific messages
				return ""
			},
		},
	}
	return table
}
