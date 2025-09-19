package customersapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the  customer domain.
type App struct {
	customersbus *customersbus.Business
	auth         *auth.Auth
}

// NewApp constructs a  customer app API for use.
func NewApp(customersbus *customersbus.Business) *App {
	return &App{
		customersbus: customersbus,
	}
}

// NewAppWithAuth constructs a  customer app API for use with auth support.
func NewAppWithAuth(customersbus *customersbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:         ath,
		customersbus: customersbus,
	}
}

// Create adds a new  customer to the system.
func (a *App) Create(ctx context.Context, app NewCustomers) (Customers, error) {
	na, err := toBusNewCustomers(app)
	if err != nil {
		return Customers{}, errs.New(errs.InvalidArgument, err)
	}

	ass, err := a.customersbus.Create(ctx, na)
	if err != nil {
		if errors.Is(err, customersbus.ErrUniqueEntry) {
			return Customers{}, errs.New(errs.Aborted, customersbus.ErrUniqueEntry)
		}
		return Customers{}, errs.Newf(errs.Internal, "create:  customer[%+v]: %s", ass, err)
	}

	return ToAppCustomer(ass), err
}

// Update updates an existing  customer.
func (a *App) Update(ctx context.Context, app UpdateCustomers, id uuid.UUID) (Customers, error) {
	us, err := toBusUpdateCustomers(app)
	if err != nil {
		return Customers{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.customersbus.QueryByID(ctx, id)
	if err != nil {
		return Customers{}, errs.New(errs.NotFound, customersbus.ErrNotFound)
	}

	customers, err := a.customersbus.Update(ctx, st, us)
	if err != nil {
		if errors.Is(err, customersbus.ErrNotFound) {
			return Customers{}, errs.New(errs.NotFound, err)
		}
		return Customers{}, errs.Newf(errs.Internal, "update:  customer[%+v]: %s", customers, err)
	}

	return ToAppCustomer(customers), nil
}

// Delete removes an existing  customer.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := a.customersbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, customersbus.ErrNotFound)
	}

	err = a.customersbus.Delete(ctx, st)
	if err != nil {
		return errs.Newf(errs.Internal, "delete:  customer[%+v]: %s", st, err)
	}

	return nil
}

// Query returns a list of  customers based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Customers], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Customers]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Customers]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Customers]{}, errs.NewFieldsError("orderby", err)
	}

	customerss, err := a.customersbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Customers]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.customersbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Customers]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppCustomers(customerss), total, page), nil
}

// QueryByID retrieves a single customer by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Customers, error) {
	customers, err := a.customersbus.QueryByID(ctx, id)
	if err != nil {
		return Customers{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppCustomer(customers), nil
}
