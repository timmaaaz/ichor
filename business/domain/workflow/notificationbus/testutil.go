package notificationbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// TestSeedNotifications creates n notifications for the given userID with
// varying priorities and read/unread states. The first half are unread; the
// second half are marked read. Returns the created notifications in insertion
// order.
func TestSeedNotifications(ctx context.Context, n int, userID uuid.UUID, api *Business) ([]Notification, error) {
	priorities := []string{PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow}

	notifications := make([]Notification, n)

	for i := 0; i < n; i++ {
		nn := NewNotification{
			UserID:           userID,
			Title:            fmt.Sprintf("Test Notification %c", 'A'+i),
			Message:          fmt.Sprintf("Message for notification %c", 'A'+i),
			Priority:         priorities[i%len(priorities)],
			SourceEntityName: "orders",
			SourceEntityID:   uuid.New(),
			ActionURL:        "/orders/" + uuid.New().String(),
		}

		notification, err := api.Create(ctx, nn)
		if err != nil {
			return nil, fmt.Errorf("seeding notification %d: %w", i, err)
		}

		// Mark the second half as read.
		if i >= n/2 {
			if err := api.MarkAsRead(ctx, notification.ID, userID); err != nil {
				return nil, fmt.Errorf("marking notification %d as read: %w", i, err)
			}

			// Re-query to capture the updated read state.
			notification, err = api.QueryByID(ctx, notification.ID)
			if err != nil {
				return nil, fmt.Errorf("re-querying notification %d after mark-read: %w", i, err)
			}
		}

		notifications[i] = notification
	}

	return notifications, nil
}
