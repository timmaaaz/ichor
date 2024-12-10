package reportsto_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/reportstoapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/reportsto",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &reportstoapp.NewReportsTo{
				BossID:     sd.ReportsTo[0].BossID,
				ReporterID: sd.ReportsTo[1].ReporterID,
			},
			GotResp: &reportstoapp.ReportsTo{},
			ExpResp: &reportstoapp.ReportsTo{
				BossID:     sd.ReportsTo[0].BossID,
				ReporterID: sd.ReportsTo[1].ReporterID,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*reportstoapp.ReportsTo)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*reportstoapp.ReportsTo)
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
			Name:       "missing reporter_id",
			URL:        "/v1/reportsto",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &reportstoapp.NewReportsTo{
				BossID: uuid.New().String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"reporter_id\",\"error\":\"reporter_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
		{
			Name:       "missing boss_id",
			URL:        "/v1/reportsto",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &reportstoapp.NewReportsTo{
				ReporterID: uuid.New().String(),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"boss_id\",\"error\":\"boss_id is a required field\"}]"),
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				return cmp.Diff(exp, gotResp)
			},
		},
	}
	return table
}

func create401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "emptytoken",
			URL:        "/v1/reportsto",
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
			Name:       "badtoken",
			URL:        "/v1/reportsto",
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
			Name:       "badsig",
			URL:        "/v1/reportsto",
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
			Name:       "wronguser",
			URL:        "/v1/reportsto",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
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
