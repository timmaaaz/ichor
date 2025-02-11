package userapprovalstatus_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/userapprovalstatusapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/userapprovalstatus?page=1&rows=2",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[userapprovalstatusapp.UserApprovalStatus]{},
			ExpResp: &query.Result[userapprovalstatusapp.UserApprovalStatus]{
				Page:        1,
				RowsPerPage: 2,
				Total:       4,
				Items:       sd.UserApprovalStatuses[:2],
			},
			CmpFunc: func(got, exp any) string {
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
			URL:        "/v1/userapprovalstatus/" + sd.UserApprovalStatuses[0].ID,
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &userapprovalstatusapp.UserApprovalStatus{},
			ExpResp:    &sd.UserApprovalStatuses[0],
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
