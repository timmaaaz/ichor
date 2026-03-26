package notificationinboxapi_test

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/notificationapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

type SeedData struct {
	apitest.SeedData
	Notifications []notificationapp.Notification
	UnreadCount   int
	TotalCount    int
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// Primary user: owns the notifications under test.
	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(busDomain.User, ath, usrs[0].Email.Address),
	}

	// Secondary user: has no notifications; used for cross-user isolation tests.
	usrs2, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return SeedData{}, fmt.Errorf("seeding secondary user: %w", err)
	}
	tu2 := apitest.User{
		User:  usrs2[0],
		Token: apitest.Token(busDomain.User, ath, usrs2[0].Email.Address),
	}

	// Create 4 notifications for the primary user: 2 unread, 2 read.
	const total = 4
	nots, err := notificationbus.TestSeedNotifications(ctx, total, usrs[0].ID, busDomain.Notification)
	if err != nil {
		return SeedData{}, fmt.Errorf("seeding notifications: %w", err)
	}

	appNots := make([]notificationapp.Notification, len(nots))
	for i, n := range nots {
		appNots[i] = notificationapp.ToAppNotification(n)
	}

	return SeedData{
		SeedData: apitest.SeedData{
			Users: []apitest.User{tu1, tu2},
		},
		Notifications: appNots,
		UnreadCount:   total / 2,
		TotalCount:    total,
	}, nil
}
