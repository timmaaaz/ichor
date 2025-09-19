package ordersapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	ordersbus *ordersbus.Business
	auth      *auth.Auth
}

func NewApp(ordersbus *ordersbus.Business) *App {
	return &App{
		ordersbus: ordersbus,
	}
}

func NewAppWithAuth(ordersbus *ordersbus.Business, auth *auth.Auth) *App {
	return &App{
		ordersbus: ordersbus,
		auth:      auth,
	}
}

func (a *App) Create(ctx context.Context, app NewOrder) (Order, error) {
	nt, err := toBusNewOrder(app)
	if err != nil {
		return Order{}, err
	}

	status, err := a.ordersbus.Create(ctx, nt)
	if err != nil {
		if errors.Is(err, ordersbus.ErrUniqueEntry) {
			return Order{}, errs.New(errs.AlreadyExists, err)
		}
		return Order{}, err
	}

	return ToAppOrder(status), nil
}

func (a *App) Update(ctx context.Context, app UpdateOrder, id uuid.UUID) (Order, error) {
	ui, err := toBusUpdateOrder(app)
	if err != nil {
		return Order{}, errs.New(errs.InvalidArgument, err)
	}

	u, err := a.ordersbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return Order{}, errs.New(errs.NotFound, err)
		}
		return Order{}, err
	}

	status, err := a.ordersbus.Update(ctx, u, ui)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return Order{}, errs.New(errs.NotFound, err)
		}
		return Order{}, err
	}

	return ToAppOrder(status), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	t, err := a.ordersbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return errs.New(errs.NotFound, ordersbus.ErrNotFound)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.ordersbus.Delete(ctx, t)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Order], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Order]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Order]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Order]{}, errs.NewFieldsError("orderBy", err)
	}

	items, err := a.ordersbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Order]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.ordersbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Order]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppOrders(items), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Order, error) {
	status, err := a.ordersbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return Order{}, errs.New(errs.NotFound, err)
		}
		return Order{}, err
	}

	return ToAppOrder(status), nil
}
