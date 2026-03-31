package cyclecountsessionapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/inventory/cycle-count-sessions?rows=10&page=1",
			Token:      "&nbsp;",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        "/v1/inventory/cycle-count-sessions?rows=10&page=1",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-read-permission",
			URL:        "/v1/inventory/cycle-count-sessions?rows=10&page=1",
			Token:      sd.Users[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusForbidden,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.PermissionDenied, "user does not have permission READ for table: inventory.cycle_count_sessions"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryByID401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-read-permission",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusForbidden,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.PermissionDenied, "user does not have permission READ for table: inventory.cycle_count_sessions"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func query200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/v1/inventory/cycle-count-sessions?rows=10&page=1",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &query.Result[cyclecountsessionapp.CycleCountSession]{},
			ExpResp: &query.Result[cyclecountsessionapp.CycleCountSession]{
				Items:       sd.CycleCountSessions,
				Total:       len(sd.CycleCountSessions),
				Page:        1,
				RowsPerPage: 10,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &cyclecountsessionapp.CycleCountSession{},
			ExpResp:    &sd.CycleCountSessions[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryByID404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "cycle count session not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
