package orderfulfillmentstatusapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/order/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	orderfulfillmentstatusbus *orderfulfillmentstatusbus.Business
	auth                      *auth.Auth
}

func NewApp(orderfulfillmentstatusbus *orderfulfillmentstatusbus.Business) *App {
	return &App{
		orderfulfillmentstatusbus: orderfulfillmentstatusbus,
	}
}

func NewAppWithAuth(orderfulfillmentstatusbus *orderfulfillmentstatusbus.Business, auth *auth.Auth) *App {
	return &App{
		orderfulfillmentstatusbus: orderfulfillmentstatusbus,
		auth:                      auth,
	}
}

func (a *App) Create(ctx context.Context, app NewOrderFulfillmentStatus) (OrderFulfillmentStatus, error) {
	nt, err := toBusNewOrderFulfillmentStatus(app)
	if err != nil {
		return OrderFulfillmentStatus{}, err
	}

	status, err := a.orderfulfillmentstatusbus.Create(ctx, nt)
	if err != nil {
		if errors.Is(err, orderfulfillmentstatusbus.ErrUniqueEntry) {
			return OrderFulfillmentStatus{}, errs.New(errs.AlreadyExists, err)
		}
		return OrderFulfillmentStatus{}, err
	}

	return ToAppOrderFulfillmentStatus(status), nil
}

func (a *App) Update(ctx context.Context, app UpdateOrderFulfillmentStatus, id uuid.UUID) (OrderFulfillmentStatus, error) {
	ui, err := toBusUpdateOrderFulfillmentStatus(app)
	if err != nil {
		return OrderFulfillmentStatus{}, errs.New(errs.InvalidArgument, err)
	}

	u, err := a.orderfulfillmentstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, orderfulfillmentstatusbus.ErrNotFound) {
			return OrderFulfillmentStatus{}, errs.New(errs.NotFound, err)
		}
		return OrderFulfillmentStatus{}, err
	}

	status, err := a.orderfulfillmentstatusbus.Update(ctx, u, ui)
	if err != nil {
		if errors.Is(err, orderfulfillmentstatusbus.ErrNotFound) {
			return OrderFulfillmentStatus{}, errs.New(errs.NotFound, err)
		}
		return OrderFulfillmentStatus{}, err
	}

	return ToAppOrderFulfillmentStatus(status), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	t, err := a.orderfulfillmentstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, orderfulfillmentstatusbus.ErrNotFound) {
			return errs.New(errs.NotFound, orderfulfillmentstatusbus.ErrNotFound)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.orderfulfillmentstatusbus.Delete(ctx, t)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[OrderFulfillmentStatus], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[OrderFulfillmentStatus]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[OrderFulfillmentStatus]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[OrderFulfillmentStatus]{}, errs.NewFieldsError("orderBy", err)
	}

	items, err := a.orderfulfillmentstatusbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[OrderFulfillmentStatus]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.orderfulfillmentstatusbus.Count(ctx, filter)
	if err != nil {
		return query.Result[OrderFulfillmentStatus]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppOrderFulfillmentStatuses(items), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (OrderFulfillmentStatus, error) {
	status, err := a.orderfulfillmentstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, orderfulfillmentstatusbus.ErrNotFound) {
			return OrderFulfillmentStatus{}, errs.New(errs.NotFound, err)
		}
		return OrderFulfillmentStatus{}, err
	}

	return ToAppOrderFulfillmentStatus(status), nil
}
