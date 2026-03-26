package notificationapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer APIs for notification access.
type App struct {
	notificationBus *notificationbus.Business
}

// NewApp constructs a notification app.
func NewApp(notificationBus *notificationbus.Business) *App {
	return &App{
		notificationBus: notificationBus,
	}
}

// Query returns a list of notifications for the authenticated user.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Notification], error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.InvalidArgument, "page: %s", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.InvalidArgument, "filter: %s", err)
	}

	// Always scope to the authenticated user.
	filter.UserID = &userID

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, DefaultOrderBy)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.InvalidArgument, "order: %s", err)
	}

	notifications, err := a.notificationBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.notificationBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppNotifications(notifications), total, pg), nil
}

// Count returns the number of notifications matching the filter for the authenticated user.
func (a *App) Count(ctx context.Context, isRead string) (int, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return 0, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	filter, err := parseCountFilter(isRead, userID)
	if err != nil {
		return 0, errs.Newf(errs.InvalidArgument, "filter: %s", err)
	}

	count, err := a.notificationBus.Count(ctx, filter)
	if err != nil {
		return 0, errs.Newf(errs.Internal, "count: %s", err)
	}

	return count, nil
}

// MarkAsRead marks a single notification as read. Verifies the notification
// belongs to the authenticated user.
func (a *App) MarkAsRead(ctx context.Context, idStr string) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "parse notification id: %s", err)
	}

	// Verify ownership.
	notification, err := a.notificationBus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, notificationbus.ErrNotFound) {
			return errs.Newf(errs.NotFound, "notification not found")
		}
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	if notification.UserID != userID {
		return errs.Newf(errs.NotFound, "notification not found")
	}

	if err := a.notificationBus.MarkAsRead(ctx, id, userID); err != nil {
		return errs.Newf(errs.Internal, "mark as read: %s", err)
	}

	return nil
}

// MarkAllAsRead marks all unread notifications for the authenticated user as read.
func (a *App) MarkAllAsRead(ctx context.Context) (int, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return 0, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	count, err := a.notificationBus.MarkAllAsRead(ctx, userID)
	if err != nil {
		return 0, errs.Newf(errs.Internal, "mark all as read: %s", err)
	}

	return count, nil
}
