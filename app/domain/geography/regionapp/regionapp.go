package regionapp

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the region domain.
type App struct {
	regionBus *regionbus.Business
	auth      *auth.Auth
}

// NewApp constructs a region app API for use.
func NewApp(regionBus *regionbus.Business) *App {
	return &App{
		regionBus: regionBus,
	}
}

// NewAppWithAuth constructs a region app API for use with auth support.
func NewAppWithAuth(regionBus *regionbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:      ath,
		regionBus: regionBus,
	}
}

// Query retrieves a list of regions based on the filter, order, and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Region], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Region]{}, err
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Region]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Region]{}, err
	}

	regions, err := a.regionBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Region]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.regionBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Region]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppRegions(regions), total, page), nil
}

// QueryByID retrieves a single region based on the region ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Region, error) {
	region, err := a.regionBus.QueryByID(ctx, id)
	if err != nil {
		return Region{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	return ToAppRegion(region), nil
}
