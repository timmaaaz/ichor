package countryapp

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/geography/countrybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the country domain.
type App struct {
	countryBus *countrybus.Business
	auth       *auth.Auth
}

// NewApp constructs a country app API for use.
func NewApp(countryBus *countrybus.Business) *App {
	return &App{
		countryBus: countryBus,
	}
}

// NewAppWithAuth constructs a country app API for use with auth support.
func NewAppWithAuth(countryBus *countrybus.Business, ath *auth.Auth) *App {
	return &App{
		auth:       ath,
		countryBus: countryBus,
	}
}

// Query retrieves a list of countries based on the filter, order, and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Country], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Country]{}, err
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Country]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Country]{}, err
	}

	countries, err := a.countryBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Country]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.countryBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Country]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppCountries(countries), total, page), nil
}

// QueryByID retrieves the country by the specified ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Country, error) {
	// Not using middleware, may change later.
	countryID, err := a.countryBus.QueryByID(ctx, id)
	if err != nil {
		return Country{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return toAppCountry(countryID), nil
}
