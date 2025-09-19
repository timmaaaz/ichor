package tableaccess_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/core/tableaccessapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	// make copy of sd.tableaccess into exp, sort, and get first 10
	exp := make([]tableaccessapp.TableAccess, len(sd.TableAccesses))
	copy(exp, sd.TableAccesses)

	// sort
	sort.Slice(exp, func(i, j int) bool {
		return exp[i].ID < exp[j].ID
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/table-access?page=1&rows=10&orderBy=id,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[tableaccessapp.TableAccess]{},
			ExpResp: &query.Result[tableaccessapp.TableAccess]{
				Page:        1,
				RowsPerPage: 10,
				Total:       624,
				Items:       exp[:10],
			},
			CmpFunc: func(got any, exp any) string {

				// compare items
				gotItems := got.(*query.Result[tableaccessapp.TableAccess]).Items
				expItems := exp.(*query.Result[tableaccessapp.TableAccess]).Items

				return cmp.Diff(expItems, gotItems)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/table-access/" + sd.TableAccesses[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &tableaccessapp.TableAccess{},
			ExpResp:    &sd.TableAccesses[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func query401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/core/table-access?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &query.Result[tableaccessapp.TableAccess]{},
			ExpResp:    &query.Result[tableaccessapp.TableAccess]{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(exp, got)
			},
		},
	}
	return table
}
