package metricsapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/quality/metricsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the supplier domain.
type App struct {
	metricsbus *metricsbus.Business
	auth       *auth.Auth
}

// NewApp constructs a supplier app API for use.
func NewApp(metricsbus *metricsbus.Business) *App {
	return &App{
		metricsbus: metricsbus,
	}
}

// NewAppWithAuth constructs a supplier app API for use with auth support.
func NewAppWithAuth(metricsbus *metricsbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:       ath,
		metricsbus: metricsbus,
	}
}

// Create adds a new quality metric to the system.
func (a *App) Create(ctx context.Context, app NewMetric) (Metric, error) {
	nb, err := toBusNewMetric(app)
	if err != nil {
		return Metric{}, errs.New(errs.InvalidArgument, err)
	}

	qualityMetric, err := a.metricsbus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, metricsbus.ErrUniqueEntry) {
			return Metric{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, metricsbus.ErrForeignKeyViolation) {
			return Metric{}, errs.New(errs.Aborted, err)
		}
		return Metric{}, err
	}

	return ToAppMetric(qualityMetric), nil
}

// Update updates an existing quality metric.
func (a *App) Update(ctx context.Context, app UpdateMetric, id uuid.UUID) (Metric, error) {
	up, err := toBusUpdateMetric(app)
	if err != nil {
		return Metric{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.metricsbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, metricsbus.ErrNotFound) {
			return Metric{}, errs.New(errs.NotFound, err)
		}
		return Metric{}, err
	}

	qualityMetric, err := a.metricsbus.Update(ctx, st, up)
	if err != nil {
		if errors.Is(err, metricsbus.ErrForeignKeyViolation) {
			return Metric{}, errs.New(errs.Aborted, err)
		}
		if errors.Is(err, metricsbus.ErrNotFound) {
			return Metric{}, errs.New(errs.NotFound, err)
		}
		return Metric{}, fmt.Errorf("update: metric[%+v]: %w", qualityMetric, err)
	}

	return ToAppMetric(qualityMetric), nil
}

// Delete removes an existing quality metric.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := a.metricsbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, metricsbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete: queryByID failed: %v", err)
	}

	if err := a.metricsbus.Delete(ctx, st); err != nil {
		return fmt.Errorf("delete: metric[%+v]: %w", st, err)
	}
	return nil
}

// Query return a list of metrics based on filters provided
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Metric], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Metric]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Metric]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Metric]{}, errs.NewFieldsError("orderBy", err)
	}

	results, err := a.metricsbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Metric]{}, errs.Newf(errs.Internal, "query %v", err)
	}

	total, err := a.metricsbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Metric]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppMetrics(results), total, page), nil
}

// QueryByID retrieves a single quality metric by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Metric, error) {
	st, err := a.metricsbus.QueryByID(ctx, id)
	if err != nil {
		return Metric{}, errs.Newf(errs.Internal, "queryByID: %v", err)
	}

	return ToAppMetric(st), nil
}
