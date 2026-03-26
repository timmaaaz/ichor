package notificationinboxapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_NotificationInboxAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_NotificationInboxAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, query401(sd), "query401")
	test.Run(t, count401(sd), "count401")
	test.Run(t, markAsRead401(sd), "markAsRead401")
	test.Run(t, markAllAsRead401(sd), "markAllAsRead401")

	test.Run(t, query200(sd), "query200")
	test.Run(t, queryIsolation200(sd), "queryIsolation200")
	test.Run(t, count200(sd), "count200")
	test.Run(t, markAsRead200(sd), "markAsRead200")
	test.Run(t, markAsRead404(sd), "markAsRead404")
	test.Run(t, markAsReadOwnership404(sd), "markAsReadOwnership404")
	test.Run(t, markAllAsRead200(sd), "markAllAsRead200")
}
