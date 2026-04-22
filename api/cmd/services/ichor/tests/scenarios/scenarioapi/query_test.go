package scenarioapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/scenarios/scenarioapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"

	"github.com/google/go-cmp/cmp"
)

func query200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all-scenarios",
			URL:        "/v1/scenarios?rows=100&page=1",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &scenarioapp.Scenarios{},
			ExpResp:    nil,
			// The seeder may add baseline scenarios in addition to the three
			// we TestSeed here, so assert inclusion rather than exact equality.
			CmpFunc: func(got, _ any) string {
				gotResp := got.(*scenarioapp.Scenarios)
				seen := make(map[string]bool, len(*gotResp))
				for _, s := range *gotResp {
					seen[s.ID] = true
				}
				for _, expected := range sd.Scenarios {
					if !seen[expected.ID.String()] {
						return fmt.Sprintf("seeded scenario %s missing from query response", expected.ID)
					}
				}
				return ""
			},
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/scenarios/%s", sd.Scenarios[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &scenarioapp.Scenario{},
			ExpResp:    nil,
			CmpFunc: func(got, _ any) string {
				gotResp := got.(*scenarioapp.Scenario)
				if gotResp.ID != sd.Scenarios[0].ID.String() {
					return fmt.Sprintf("got ID %s, want %s", gotResp.ID, sd.Scenarios[0].ID)
				}
				if gotResp.Name != sd.Scenarios[0].Name {
					return fmt.Sprintf("got Name %q, want %q", gotResp.Name, sd.Scenarios[0].Name)
				}
				return ""
			},
		},
	}
}

func queryByID404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/scenarios/%s", uuid.NewString()),
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

func query401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/scenarios?rows=10&page=1",
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
			// Scenarios use auth.RuleAdminOnly, so a non-admin USER token
			// fails the admin-rule check and returns 401 Unauthenticated
			// *before* reaching the table-access permission check — the
			// labelapi pattern that yields 403 doesn't apply here.
			Name:       "non-admin-401",
			URL:        "/v1/scenarios?rows=10&page=1",
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

func queryByID401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "non-admin-401",
			URL:        fmt.Sprintf("/v1/scenarios/%s", sd.Scenarios[0].ID),
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
