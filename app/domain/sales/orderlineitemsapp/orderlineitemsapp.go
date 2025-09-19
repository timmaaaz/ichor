package orderlineitemsapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	orderlineitemsbus *orderlineitemsbus.Business
	auth              *auth.Auth
}

func NewApp(orderlineitemsbus *orderlineitemsbus.Business) *App {
	return &App{
		orderlineitemsbus: orderlineitemsbus,
	}
}

func NewAppWithAuth(orderlineitemsbus *orderlineitemsbus.Business, auth *auth.Auth) *App {
	return &App{
		orderlineitemsbus: orderlineitemsbus,
		auth:              auth,
	}
}

func (a *App) Create(ctx context.Context, app NewOrderLineItem) (OrderLineItem, error) {
	nt, err := toBusNewOrderLineItem(app)
	if err != nil {
		return OrderLineItem{}, err
	}

	status, err := a.orderlineitemsbus.Create(ctx, nt)
	if err != nil {
		if errors.Is(err, orderlineitemsbus.ErrUniqueEntry) {
			return OrderLineItem{}, errs.New(errs.AlreadyExists, err)
		}
		return OrderLineItem{}, err
	}

	return ToAppOrderLineItem(status), nil
}

func (a *App) Update(ctx context.Context, app UpdateOrderLineItem, id uuid.UUID) (OrderLineItem, error) {
	ui, err := toBusUpdateOrderLineItem(app)
	if err != nil {
		return OrderLineItem{}, errs.New(errs.InvalidArgument, err)
	}

	u, err := a.orderlineitemsbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, orderlineitemsbus.ErrNotFound) {
			return OrderLineItem{}, errs.New(errs.NotFound, err)
		}
		return OrderLineItem{}, err
	}

	status, err := a.orderlineitemsbus.Update(ctx, u, ui)
	if err != nil {
		if errors.Is(err, orderlineitemsbus.ErrNotFound) {
			return OrderLineItem{}, errs.New(errs.NotFound, err)
		}
		return OrderLineItem{}, err
	}

	return ToAppOrderLineItem(status), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	t, err := a.orderlineitemsbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, orderlineitemsbus.ErrNotFound) {
			return errs.New(errs.NotFound, orderlineitemsbus.ErrNotFound)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.orderlineitemsbus.Delete(ctx, t)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[OrderLineItem], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[OrderLineItem]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[OrderLineItem]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[OrderLineItem]{}, errs.NewFieldsError("orderBy", err)
	}

	items, err := a.orderlineitemsbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[OrderLineItem]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.orderlineitemsbus.Count(ctx, filter)
	if err != nil {
		return query.Result[OrderLineItem]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppOrderLineItems(items), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (OrderLineItem, error) {
	status, err := a.orderlineitemsbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, orderlineitemsbus.ErrNotFound) {
			return OrderLineItem{}, errs.New(errs.NotFound, err)
		}
		return OrderLineItem{}, err
	}

	return ToAppOrderLineItem(status), nil
}
