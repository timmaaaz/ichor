package notificationinboxapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/notificationinboxapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/notificationapp"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func query200(sd SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic-query",
			URL:        "/v1/workflow/notifications?page=1&rows=10",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[notificationapp.Notification]{},
			ExpResp: &query.Result[notificationapp.Notification]{
				Items:       sd.Notifications,
				Total:       sd.TotalCount,
				Page:        1,
				RowsPerPage: 10,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[notificationapp.Notification])
				expResp := exp.(*query.Result[notificationapp.Notification])
				return cmp.Diff(
					*gotResp,
					*expResp,
					cmpopts.SortSlices(func(a, b notificationapp.Notification) bool {
						return a.ID < b.ID
					}),
				)
			},
		},
		{
			Name:       "filter-unread",
			URL:        "/v1/workflow/notifications?page=1&rows=10&is_read=false",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[notificationapp.Notification]{},
			ExpResp:    &query.Result[notificationapp.Notification]{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[notificationapp.Notification])
				if len(gotResp.Items) != sd.UnreadCount {
					return cmp.Diff(len(gotResp.Items), sd.UnreadCount)
				}
				for _, n := range gotResp.Items {
					if n.IsRead {
						return cmp.Diff(false, n.IsRead)
					}
				}
				return ""
			},
		},
	}

	return table
}

func count200(sd SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unread-count",
			URL:        "/v1/workflow/notifications/count?is_read=false",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &notificationinboxapi.UnreadCount{},
			ExpResp:    &notificationinboxapi.UnreadCount{Count: sd.UnreadCount},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "total-count",
			URL:        "/v1/workflow/notifications/count",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &notificationinboxapi.UnreadCount{},
			ExpResp:    &notificationinboxapi.UnreadCount{Count: sd.TotalCount},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryIsolation200(sd SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "second-user-sees-no-notifications",
			URL:        "/v1/workflow/notifications?page=1&rows=10",
			Token:      sd.Users[1].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[notificationapp.Notification]{},
			ExpResp: &query.Result[notificationapp.Notification]{
				Items:       []notificationapp.Notification{},
				Total:       0,
				Page:        1,
				RowsPerPage: 10,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[notificationapp.Notification])
				expResp := exp.(*query.Result[notificationapp.Notification])
				return cmp.Diff(*gotResp, *expResp)
			},
		},
	}
}
