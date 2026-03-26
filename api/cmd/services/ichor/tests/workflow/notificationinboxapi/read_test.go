package notificationinboxapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/notificationinboxapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func markAsRead200(sd SeedData) []apitest.Table {
	// Pick the first unread notification (index 0, which is unread per seed).
	unreadID := sd.Notifications[0].ID

	table := []apitest.Table{
		{
			Name:       "mark-single-read",
			URL:        fmt.Sprintf("/v1/workflow/notifications/%s/read", unreadID),
			Token:      sd.Users[0].Token,
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
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			GotResp:    &notificationinboxapi.MarkAllReadResult{},
			ExpResp:    &notificationinboxapi.MarkAllReadResult{Count: 1},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func markAsRead404(sd SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "non-existent-notification",
			URL:        fmt.Sprintf("/v1/workflow/notifications/%s/read", uuid.NewString()),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPost,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "notification not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func markAsReadOwnership404(sd SeedData) []apitest.Table {
	// sd.Users[1] tries to mark sd.Users[0]'s notification as read.
	// The app returns NotFound to avoid leaking existence.
	targetID := sd.Notifications[0].ID

	return []apitest.Table{
		{
			Name:       "wrong-owner",
			URL:        fmt.Sprintf("/v1/workflow/notifications/%s/read", targetID),
			Token:      sd.Users[1].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPost,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "notification not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
