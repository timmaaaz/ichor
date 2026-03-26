package notificationinboxapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	User          apitest.User
	UnreadCount   int
	TotalCount    int
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return SeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	tu := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(busDomain.User, ath, usrs[0].Email.Address),
	}

	// Create 4 notifications: 2 unread, 2 read (different priorities).
	appNots := make([]notificationapp.Notification, 4)

	priorities := []string{
		notificationbus.PriorityCritical,
		notificationbus.PriorityHigh,
		notificationbus.PriorityMedium,
		notificationbus.PriorityLow,
	}

	for i := 0; i < 4; i++ {
		nn := notificationbus.NewNotification{
			UserID:           usrs[0].ID,
			Title:            fmt.Sprintf("Test Notification %c", 'A'+i),
			Message:          fmt.Sprintf("Message for notification %c", 'A'+i),
			Priority:         priorities[i],
			SourceEntityName: "orders",
			SourceEntityID:   uuid.New(),
			ActionURL:        "/orders/" + uuid.New().String(),
		}

		n, err := busDomain.Notification.Create(ctx, nn)
		if err != nil {
			return SeedData{}, fmt.Errorf("seeding notification %d: %w", i, err)
		}

		// Mark the last 2 as read.
		if i >= 2 {
			if err := busDomain.Notification.MarkAsRead(ctx, n.ID, usrs[0].ID); err != nil {
				return SeedData{}, fmt.Errorf("marking notification %d as read: %w", i, err)
			}
			// Re-query to get the updated state.
			n, err = busDomain.Notification.QueryByID(ctx, n.ID)
			if err != nil {
				return SeedData{}, fmt.Errorf("re-querying notification %d: %w", i, err)
			}
		}

		appNots[i] = notificationapp.ToAppNotification(n)
	}

	return SeedData{
		SeedData: apitest.SeedData{
			Users: []apitest.User{tu},
		},
		Notifications: appNots,
		User:          tu,
		UnreadCount:   2,
		TotalCount:    4,
	}, nil
}
