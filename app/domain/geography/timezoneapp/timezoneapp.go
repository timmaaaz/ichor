package timezoneapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the timezone domain.
type App struct {
	timezoneBus *timezonebus.Business
	auth        *auth.Auth
}

// NewApp constructs a timezone app API for use.
func NewApp(timezoneBus *timezonebus.Business) *App {
	return &App{
		timezoneBus: timezoneBus,
	}
}

// NewAppWithAuth constructs a timezone app API for use with auth support.
func NewAppWithAuth(timezoneBus *timezonebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:        ath,
		timezoneBus: timezoneBus,
	}
}

// Create adds a new timezone to the system.
func (a *App) Create(ctx context.Context, app NewTimezone) (Timezone, error) {
	ntz := toBusNewTimezone(app)

	tz, err := a.timezoneBus.Create(ctx, ntz)
	if err != nil {
		if errors.Is(err, timezonebus.ErrUniqueEntry) {
			return Timezone{}, errs.New(errs.Aborted, timezonebus.ErrUniqueEntry)
		}
		return Timezone{}, errs.Newf(errs.Internal, "create: timezone[%+v]: %s", tz, err)
	}

	return ToAppTimezone(tz), nil
}

// Update updates an existing timezone.
func (a *App) Update(ctx context.Context, app UpdateTimezone, id uuid.UUID) (Timezone, error) {
	utz, err := toBusUpdateTimezone(app)
	if err != nil {
		return Timezone{}, errs.New(errs.InvalidArgument, err)
	}

	tz, err := a.timezoneBus.QueryByID(ctx, id)
	if err != nil {
		return Timezone{}, errs.New(errs.NotFound, timezonebus.ErrNotFound)
	}

	updated, err := a.timezoneBus.Update(ctx, tz, utz)
	if err != nil {
		if errors.Is(err, timezonebus.ErrNotFound) {
			return Timezone{}, errs.New(errs.NotFound, err)
		}
		return Timezone{}, errs.Newf(errs.Internal, "update: timezone[%+v]: %s", tz, err)
	}

	return ToAppTimezone(updated), nil
}

// Delete removes an existing timezone.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	tz, err := a.timezoneBus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, timezonebus.ErrNotFound)
	}

	err = a.timezoneBus.Delete(ctx, tz)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: timezone[%+v]: %s", tz, err)
	}

	return nil
}

// Query retrieves a list of timezones based on the filter, order, and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Timezone], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Timezone]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Timezone]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Timezone]{}, errs.NewFieldsError("orderby", err)
	}

	tzs, err := a.timezoneBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[Timezone]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.timezoneBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Timezone]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppTimezones(tzs), total, pg), nil
}

// QueryByID retrieves the timezone by the specified ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Timezone, error) {
	tz, err := a.timezoneBus.QueryByID(ctx, id)
	if err != nil {
		return Timezone{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppTimezone(tz), nil
}
