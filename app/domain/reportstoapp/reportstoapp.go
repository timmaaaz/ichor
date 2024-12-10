package reportstoapp

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/reportstobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the reports to domain.
type App struct {
	reportsToBus *reportstobus.Business
	auth         *auth.Auth
}

// NewApp constructs a reports to app API for use.
func NewApp(reportsToBus *reportstobus.Business) *App {
	return &App{
		reportsToBus: reportsToBus,
	}
}

// NewAppWithAuth constructs a reports to app API for use with auth support.
func NewAppWithAuth(reportsToBus *reportstobus.Business, ath *auth.Auth) *App {
	return &App{
		auth:         ath,
		reportsToBus: reportsToBus,
	}
}

// Create adds a new reports to entry to the system.
func (a *App) Create(ctx context.Context, app NewReportsTo) (ReportsTo, error) {
	reportsTo, err := a.reportsToBus.Create(ctx, ToBusNewReportsTo(app))
	if err != nil {
		if errors.Is(err, reportstobus.ErrUniqueEntry) {
			return ReportsTo{}, errs.New(errs.Aborted, reportstobus.ErrUniqueEntry)
		}
		return ReportsTo{}, errs.Newf(errs.Internal, "create: reports to[%+v]: %s", reportsTo, err)
	}

	return ToAppReportsTo(reportsTo), nil
}

// Update updates an existing reports to.
func (a *App) Update(ctx context.Context, app UpdateReportsTo, id uuid.UUID) (ReportsTo, error) {
	urt := ToBusUpdateReportsTo(app)

	rt, err := a.reportsToBus.QueryByID(ctx, id)
	if err != nil {
		return ReportsTo{}, errs.Newf(errs.NotFound, "update: reports to[%s]: %s", id, err)
	}

	reportsTo, err := a.reportsToBus.Update(ctx, rt, urt)
	if err != nil {
		if errors.Is(err, reportstobus.ErrNotFound) {
			return ReportsTo{}, errs.New(errs.NotFound, err)
		}
		return ReportsTo{}, errs.Newf(errs.Internal, "update: reports to[%+v]: %s", reportsTo, err)
	}

	return ToAppReportsTo(reportsTo), nil
}

// Delete removes an existing reports to.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	at, err := a.reportsToBus.QueryByID(ctx, id)
	if err != nil {
		return errs.Newf(errs.NotFound, "delete: reports to[%s]: %s", id, err)
	}

	if err := a.reportsToBus.Delete(ctx, at); err != nil {
		return errs.Newf(errs.Internal, "delete: reports to[%+v]: %s", at, err)
	}

	return nil
}

// Query returns a list of reports tos.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[ReportsTo], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[ReportsTo]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[ReportsTo]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[ReportsTo]{}, errs.NewFieldsError("orderby", err)
	}

	rts, err := a.reportsToBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[ReportsTo]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.reportsToBus.Count(ctx, filter)
	if err != nil {
		return query.Result[ReportsTo]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppReportsTos(rts), total, page), nil
}

// QueryByID returns a single reports to based on the id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (ReportsTo, error) {
	rt, err := a.reportsToBus.QueryByID(ctx, id)
	if err != nil {
		return ReportsTo{}, errs.Newf(errs.NotFound, "query: reports to[%s]: %s", id, err)
	}

	return ToAppReportsTo(rt), nil
}
