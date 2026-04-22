package scenarioapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// load204 verifies POST /v1/scenarios/{id}/load returns 204 on two shapes:
//   - scenarios[1]: no fixtures — ApplyFixtures is a no-op, active pointer flips
//   - scenarios[2]: one resolvable fixture — ApplyFixtures inserts it
//
// The sub-tests run sequentially (apitest.Table is executed in slice order)
// so the first Load's scenario_id becomes the currentActive before the
// second Load. DeleteScopedRows then removes scenarios[1]'s footprint
// (empty anyway) and ApplyFixtures applies scenarios[2]'s one row.
func load204(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "load-empty-scenario",
			URL:        fmt.Sprintf("/v1/scenarios/%s/load", sd.Scenarios[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNoContent,
			GotResp:    nil,
			ExpResp:    nil,
			CmpFunc:    func(_, _ any) string { return "" },
		},
		{
			Name:       "load-scenario-with-fixture",
			URL:        fmt.Sprintf("/v1/scenarios/%s/load", sd.Scenarios[2].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNoContent,
			GotResp:    nil,
			ExpResp:    nil,
			CmpFunc:    func(_, _ any) string { return "" },
		},
	}
}

func load404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unknown-scenario",
			URL:        fmt.Sprintf("/v1/scenarios/%s/load", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "scenario not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func load401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "non-admin-401",
			URL:        fmt.Sprintf("/v1/scenarios/%s/load", sd.Scenarios[1].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
