package picktaskapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/picktaskapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/v1/inventory/pick-tasks?rows=10&page=1",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &query.Result[picktaskapp.PickTask]{},
			ExpResp: &query.Result[picktaskapp.PickTask]{
				Items:       sd.PickTasks,
				Total:       len(sd.PickTasks),
				Page:        1,
				RowsPerPage: 10,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[picktaskapp.PickTask])
				expResp := exp.(*query.Result[picktaskapp.PickTask])
				return cmp.Diff(*gotResp, *expResp)
			},
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", sd.PickTasks[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &picktaskapp.PickTask{},
			ExpResp:    &sd.PickTasks[0],
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
			URL:        fmt.Sprintf("/v1/inventory/pick-tasks/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "pick task not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
