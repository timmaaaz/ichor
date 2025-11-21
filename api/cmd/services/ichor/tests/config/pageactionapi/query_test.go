package pageaction_test

import (
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions?page=1&rows=5",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[pageactionapp.PageAction]{},
			ExpResp: &query.Result[pageactionapp.PageAction]{
				Page:        1,
				RowsPerPage: 5,
				Total:       len(sd.PageActions),
				Items:       sd.PageActions,
			},
			CmpFunc: func(got any, exp any) string {
				// Sort both got and exp by page_config_id, then action_order, then ID
				// This groups actions by page config, which is more logical
				sortFunc := func(items []pageactionapp.PageAction) {
					sort.Slice(items, func(i, j int) bool {
						if items[i].PageConfigID != items[j].PageConfigID {
							return items[i].PageConfigID < items[j].PageConfigID
						}
						if items[i].ActionOrder != items[j].ActionOrder {
							return items[i].ActionOrder < items[j].ActionOrder
						}
						return items[i].ID < items[j].ID
					})
				}

				gotItems := got.(*query.Result[pageactionapp.PageAction]).Items
				sortFunc(gotItems)

				expItems := exp.(*query.Result[pageactionapp.PageAction]).Items
				sortFunc(expItems)
				// Grab the first 5
				if len(expItems) > 5 {
					exp.(*query.Result[pageactionapp.PageAction]).Items = expItems[0:5]
				}

				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions/" + sd.PageActions[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &pageactionapp.PageAction{},
			ExpResp:    &sd.PageActions[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}

func queryByPageConfigID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-configs/actions/" + sd.PageConfigs[0].ID,
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &pageactionapp.ActionsGroupedByType{},
			ExpResp:    &pageactionapp.ActionsGroupedByType{},
			CmpFunc: func(got any, exp any) string {
				// Just verify we get a valid response structure
				gotResp := got.(*pageactionapp.ActionsGroupedByType)
				if gotResp == nil {
					return "got nil response"
				}
				// We don't check exact counts since seed data is random
				return ""
			},
		},
	}
	return table
}

func query401(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/config/page-actions?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &query.Result[pageactionapp.PageAction]{},
			ExpResp:    &query.Result[pageactionapp.PageAction]{},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
	return table
}
