package notificationinboxapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/workflow/notificationapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	notificationApp *notificationapp.App
}

func newAPI(cfg Config) *api {
	return &api{
		notificationApp: notificationapp.NewApp(cfg.NotificationBus),
	}
}

func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := notificationapp.QueryParams{
		Page:             r.URL.Query().Get("page"),
		Rows:             r.URL.Query().Get("rows"),
		OrderBy:          r.URL.Query().Get("orderBy"),
		ID:               r.URL.Query().Get("id"),
		IsRead:           r.URL.Query().Get("is_read"),
		Priority:         r.URL.Query().Get("priority"),
		SourceEntityName: r.URL.Query().Get("source_entity_name"),
		SourceEntityID:   r.URL.Query().Get("source_entity_id"),
	}

	result, err := a.notificationApp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

func (a *api) count(ctx context.Context, r *http.Request) web.Encoder {
	isRead := r.URL.Query().Get("is_read")

	count, err := a.notificationApp.Count(ctx, isRead)
	if err != nil {
		return errs.NewError(err)
	}

	return UnreadCount{Count: count}
}

func (a *api) markAsRead(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "notification_id")

	if err := a.notificationApp.MarkAsRead(ctx, id); err != nil {
		return errs.NewError(err)
	}

	return SuccessResult{Success: true}
}

func (a *api) markAllAsRead(ctx context.Context, r *http.Request) web.Encoder {
	count, err := a.notificationApp.MarkAllAsRead(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return MarkAllReadResult{Count: count}
}
