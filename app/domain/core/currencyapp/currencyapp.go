package currencyapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the currency domain.
type App struct {
	currencybus *currencybus.Business
	auth        *auth.Auth
}

// NewApp constructs a currency app API for use.
func NewApp(currencybus *currencybus.Business) *App {
	return &App{
		currencybus: currencybus,
	}
}

// NewAppWithAuth constructs a currency app API for use with auth support.
func NewAppWithAuth(currencybus *currencybus.Business, ath *auth.Auth) *App {
	return &App{
		auth:        ath,
		currencybus: currencybus,
	}
}

// Create adds a new currency to the system.
func (a *App) Create(ctx context.Context, app NewCurrency) (Currency, error) {
	nc := toBusNewCurrency(app)

	currency, err := a.currencybus.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, currencybus.ErrUnique) {
			return Currency{}, errs.New(errs.Aborted, currencybus.ErrUnique)
		}
		return Currency{}, errs.Newf(errs.Internal, "create: currency[%+v]: %s", currency, err)
	}

	return ToAppCurrency(currency), nil
}

// Update updates an existing currency.
func (a *App) Update(ctx context.Context, app UpdateCurrency, id uuid.UUID) (Currency, error) {
	uc := toBusUpdateCurrency(app)

	currency, err := a.currencybus.QueryByID(ctx, id)
	if err != nil {
		return Currency{}, errs.New(errs.NotFound, currencybus.ErrNotFound)
	}

	updated, err := a.currencybus.Update(ctx, currency, uc)
	if err != nil {
		if errors.Is(err, currencybus.ErrNotFound) {
			return Currency{}, errs.New(errs.NotFound, currencybus.ErrNotFound)
		}
		if errors.Is(err, currencybus.ErrUnique) {
			return Currency{}, errs.New(errs.Aborted, currencybus.ErrUnique)
		}
		return Currency{}, errs.Newf(errs.Internal, "update: currency[%+v]: %s", updated, err)
	}

	return ToAppCurrency(updated), nil
}

// Delete removes an existing currency.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	currency, err := a.currencybus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, currencybus.ErrNotFound)
	}

	if err := a.currencybus.Delete(ctx, currency); err != nil {
		return errs.Newf(errs.Internal, "delete: currency[%+v]: %s", currency, err)
	}

	return nil
}

// Query retrieves a list of currencies from the system.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Currency], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Currency]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Currency]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Currency]{}, errs.NewFieldsError("orderby", err)
	}

	currencies, err := a.currencybus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Currency]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.currencybus.Count(ctx, filter)
	if err != nil {
		return query.Result[Currency]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppCurrencies(currencies), total, page), nil
}

// QueryByID finds the currency by the specified ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Currency, error) {
	currency, err := a.currencybus.QueryByID(ctx, id)
	if err != nil {
		return Currency{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppCurrency(currency), nil
}

// QueryAll retrieves all currencies from the system.
func (a *App) QueryAll(ctx context.Context) (Currencies, error) {
	currencies, err := a.currencybus.QueryAll(ctx)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "queryall: %s", err)
	}

	return Currencies(ToAppCurrencies(currencies)), nil
}
