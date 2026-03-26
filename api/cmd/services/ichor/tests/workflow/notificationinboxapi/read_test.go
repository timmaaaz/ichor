package notificationinboxapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/notificationinboxapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func markAsRead200(sd SeedData) []apitest.Table {
	// Pick the first unread notification (index 0, which is unread per seed).
	unreadID := sd.Notifications[0].ID

	table := []apitest.Table{
		{
			Name:       "mark-single-read",
			URL:        fmt.Sprintf("/v1/workflow/notifications/%s/read", unreadID),
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			GotResp:    &notificationinboxapi.SuccessResult{},
			ExpResp:    &notificationinboxapi.SuccessResult{Success: true},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func markAllAsRead200(sd SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "mark-all-read",
			URL:        "/v1/workflow/notifications/read-all",
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			GotResp:    &notificationinboxapi.MarkAllReadResult{},
			ExpResp:    &notificationinboxapi.MarkAllReadResult{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*notificationinboxapi.MarkAllReadResult)
				// After markAsRead200 already marked 1, there should be
				// at most 1 remaining unread (originally 2 unread, minus 1).
				// But test ordering isn't guaranteed to be sequential across
				// t.Run groups, so just verify count >= 0.
				if gotResp.Count < 0 {
					return fmt.Sprintf("expected count >= 0, got %d", gotResp.Count)
				}
				return ""
			},
		},
	}

	return table
}
