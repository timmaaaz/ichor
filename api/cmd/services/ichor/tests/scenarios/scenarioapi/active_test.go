package scenarioapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// activeNone covers the post-seed / pre-Load state: per seed_scenarios.go
// Decision 2, the seeder does NOT set scenarios_active, so GET
// /v1/scenarios/active returns 404. A 200 case would require completing
// the Load flow, which needs fixture YAML plumbing — deferred to a
// follow-up.
func activeNone(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "nothing-loaded",
			URL:        "/v1/scenarios/active",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "scenario not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func active401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/scenarios/active",
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
			// Scenarios use auth.RuleAdminOnly → non-admin USER gets 401
			// from the rule check before the table-access middleware runs.
			Name:       "non-admin-401",
			URL:        "/v1/scenarios/active",
			Token:      sd.Users[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]: rego evaluation failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
