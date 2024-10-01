package cityapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the city domain.
type App struct {
	cityBus *citybus.Business
	auth    *auth.Auth
}

// NewApp constructs a city app API for use.
func NewApp(cityBus *citybus.Business) *App {
	return &App{
		cityBus: cityBus,
	}
}

// NewAppWithAuth constructs a city app API for use with auth support.
func NewAppWithAuth(cityBus *citybus.Business, ath *auth.Auth) *App {
	return &App{
		auth:    ath,
		cityBus: cityBus,
	}
}

// Create adds a new city to the system.
func (a *App) Create(ctx context.Context, app NewCity) (City, error) {
	nc, err := toBusNewCity(app)
	if err != nil {
		return City{}, errs.New(errs.InvalidArgument, err)
	}

	city, err := a.cityBus.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, citybus.ErrUniqueEntry) {
			return City{}, errs.New(errs.Aborted, citybus.ErrUniqueEntry)
		}
		return City{}, errs.Newf(errs.Internal, "create: city[%+v]: %s", city, err)
	}

	return ToAppCity(city), nil
}

// Update updates an existing city.
func (a *App) Update(ctx context.Context, app UpdateCity, id uuid.UUID) (City, error) {
	uc, err := toBusUpdateCity(app)
	if err != nil {
		return City{}, errs.New(errs.InvalidArgument, err)
	}

	city, err := a.cityBus.QueryByID(ctx, id)
	if err != nil {
		return City{}, errs.New(errs.NotFound, citybus.ErrNotFound)
	}

	updated, err := a.cityBus.Update(ctx, city, uc)
	if err != nil {
		if errors.Is(err, citybus.ErrNotFound) {
			return City{}, errs.New(errs.NotFound, err)
		}
		return City{}, errs.Newf(errs.Internal, "update: city[%+v]: %s", city, err)
	}

	return ToAppCity(updated), nil
}

// Delete removes an existing city.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	city, err := a.cityBus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, citybus.ErrNotFound)
	}

	err = a.cityBus.Delete(ctx, city)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: city[%+v]: %s", city, err)
	}

	return nil
}

// Query retrieves a list of cities based on the filter, order, and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[City], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[City]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[City]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[City]{}, errs.NewFieldsError("orderby", err)
	}

	cities, err := a.cityBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[City]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.cityBus.Count(ctx, filter)
	if err != nil {
		return query.Result[City]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppCities(cities), total, page), nil
}

// QueryByID retrieves the city by the specified ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (City, error) {
	city, err := a.cityBus.QueryByID(ctx, id)
	if err != nil {
		return City{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppCity(city), nil
}
