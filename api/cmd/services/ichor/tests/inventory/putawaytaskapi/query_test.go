package putawaytaskapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/putawaytaskapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/v1/inventory/put-away-tasks?rows=10&page=1",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp: &query.Result[putawaytaskapp.PutAwayTask]{},
			ExpResp: &query.Result[putawaytaskapp.PutAwayTask]{
				Items:       sd.PutAwayTasks,
				Total:       len(sd.PutAwayTasks),
				Page:        1,
				RowsPerPage: 10,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[putawaytaskapp.PutAwayTask])
				expResp := exp.(*query.Result[putawaytaskapp.PutAwayTask])
				return cmp.Diff(*gotResp, *expResp)
			},
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", sd.PutAwayTasks[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &putawaytaskapp.PutAwayTask{},
			ExpResp:    &sd.PutAwayTasks[0],
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
			URL:        fmt.Sprintf("/v1/inventory/put-away-tasks/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "put-away task not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
