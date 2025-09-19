package reportsto_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/hr/reportstoapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{Name: "basic",
			URL:        fmt.Sprintf("/v1/core/hr/reports-to/%s", sd.ReportsTo[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &reportstoapp.UpdateReportsTo{
				BossID:     dbtest.StringPointer(sd.ReportsTo[0].BossID),
				ReporterID: dbtest.StringPointer(sd.ReportsTo[0].ReporterID),
			},
			GotResp: &reportstoapp.ReportsTo{},
			ExpResp: &reportstoapp.ReportsTo{
				ID:         sd.ReportsTo[0].ID,
				BossID:     sd.ReportsTo[0].BossID,
				ReporterID: sd.ReportsTo[0].ReporterID,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*reportstoapp.ReportsTo)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*reportstoapp.ReportsTo)
				return cmp.Diff(expResp, gotResp)
			},
		},
	}
	return table
}

func update400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{Name: "invalid boss_id",
			URL:        fmt.Sprintf("/v1/core/hr/reports-to/%s", sd.ReportsTo[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &reportstoapp.UpdateReportsTo{
				BossID: dbtest.StringPointer("a"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"boss_id\",\"error\":\"boss_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp)
			},
		},
		{Name: "invalid reporter_id",
			URL:        fmt.Sprintf("/v1/core/hr/reports-to/%s", sd.ReportsTo[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &reportstoapp.UpdateReportsTo{
				ReporterID: dbtest.StringPointer("a"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"reporter_id\",\"error\":\"reporter_id must be at least 36 characters in length\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp)
			},
		},
	}
	return table
}

func update401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        fmt.Sprintf("/v1/core/hr/reports-to/%s", sd.ReportsTo[0].ID),
			Token:      "&nbsp;",
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
			URL:        fmt.Sprintf("/v1/core/hr/reports-to/%s", sd.ReportsTo[0].ID),
			Token:      sd.Admins[0].Token + "bad",
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
			URL:        fmt.Sprintf("/v1/core/hr/reports-to/%s", sd.ReportsTo[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: reports_to"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
