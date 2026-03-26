package notificationinboxapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
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
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[notificationapp.Notification]{},
			ExpResp:    &query.Result[notificationapp.Notification]{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[notificationapp.Notification])

				if len(gotResp.Items) != sd.TotalCount {
					return fmt.Sprintf("expected %d notifications, got %d", sd.TotalCount, len(gotResp.Items))
				}

				if gotResp.Total != sd.TotalCount {
					return fmt.Sprintf("expected total %d, got %d", sd.TotalCount, gotResp.Total)
				}

				if gotResp.Page != 1 {
					return fmt.Sprintf("expected page 1, got %d", gotResp.Page)
				}

				if gotResp.RowsPerPage != 10 {
					return fmt.Sprintf("expected rows_per_page 10, got %d", gotResp.RowsPerPage)
				}

				// Verify all seeded IDs are present (set comparison, order may vary).
				expIDs := make(map[string]bool)
				for _, n := range sd.Notifications {
					expIDs[n.ID] = true
				}
				for _, n := range gotResp.Items {
					if !expIDs[n.ID] {
						return fmt.Sprintf("unexpected notification ID: %s", n.ID)
					}
					delete(expIDs, n.ID)
				}
				if len(expIDs) > 0 {
					return fmt.Sprintf("missing notification IDs: %v", expIDs)
				}

				return ""
			},
		},
		{
			Name:       "filter-unread",
			URL:        "/v1/workflow/notifications?page=1&rows=10&is_read=false",
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[notificationapp.Notification]{},
			ExpResp:    &query.Result[notificationapp.Notification]{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*query.Result[notificationapp.Notification])
				if len(gotResp.Items) != sd.UnreadCount {
					return fmt.Sprintf("expected %d unread notifications, got %d", sd.UnreadCount, len(gotResp.Items))
				}
				for _, n := range gotResp.Items {
					if n.IsRead {
						return fmt.Sprintf("expected all unread, got read notification %s", n.ID)
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
			Token:      sd.User.Token,
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
			Token:      sd.User.Token,
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
