package scenarioapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/scenarios/scenarioapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func fixtures200(sd apitest.SeedData) []apitest.Table {
	scenarioWithFixtures := sd.Scenarios[0]
	scenarioWithoutFixtures := sd.Scenarios[1]

	return []apitest.Table{
		{
			Name:       "with-fixtures",
			URL:        fmt.Sprintf("/v1/scenarios/%s/fixtures", scenarioWithFixtures.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &scenarioapp.ScenarioFixtures{},
			ExpResp:    nil,
			CmpFunc: func(got, _ any) string {
				r := got.(*scenarioapp.ScenarioFixtures)
				if r.ID != scenarioWithFixtures.ID.String() {
					return fmt.Sprintf("got ID %s, want %s", r.ID, scenarioWithFixtures.ID)
				}
				if r.Name != scenarioWithFixtures.Name {
					return fmt.Sprintf("got Name %q, want %q", r.Name, scenarioWithFixtures.Name)
				}
				po := r.FixturesByTable["procurement.purchase_orders"]
				if len(po) != 2 {
					return fmt.Sprintf("procurement.purchase_orders: got %d rows, want 2", len(po))
				}
				ii := r.FixturesByTable["inventory.inventory_items"]
				if len(ii) != 1 {
					return fmt.Sprintf("inventory.inventory_items: got %d rows, want 1", len(ii))
				}
				return ""
			},
		},
		{
			Name:       "no-fixtures",
			URL:        fmt.Sprintf("/v1/scenarios/%s/fixtures", scenarioWithoutFixtures.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &scenarioapp.ScenarioFixtures{},
			ExpResp:    nil,
			CmpFunc: func(got, _ any) string {
				r := got.(*scenarioapp.ScenarioFixtures)
				if r.ID != scenarioWithoutFixtures.ID.String() {
					return fmt.Sprintf("got ID %s, want %s", r.ID, scenarioWithoutFixtures.ID)
				}
				if len(r.FixturesByTable) != 0 {
					return fmt.Sprintf("got %d tables, want 0", len(r.FixturesByTable))
				}
				return ""
			},
		},
	}
}

func fixtures404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unknown-scenario",
			URL:        fmt.Sprintf("/v1/scenarios/%s/fixtures", uuid.NewString()),
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

func fixtures401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "non-admin-401",
			URL:        fmt.Sprintf("/v1/scenarios/%s/fixtures", sd.Scenarios[0].ID),
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
