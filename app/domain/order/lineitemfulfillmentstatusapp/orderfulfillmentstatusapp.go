package lineitemfulfillmentstatusapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/order/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	lineitemfulfillmentstatusbus *lineitemfulfillmentstatusbus.Business
	auth                         *auth.Auth
}

func NewApp(lineitemfulfillmentstatusbus *lineitemfulfillmentstatusbus.Business) *App {
	return &App{
		lineitemfulfillmentstatusbus: lineitemfulfillmentstatusbus,
	}
}

func NewAppWithAuth(lineitemfulfillmentstatusbus *lineitemfulfillmentstatusbus.Business, auth *auth.Auth) *App {
	return &App{
		lineitemfulfillmentstatusbus: lineitemfulfillmentstatusbus,
		auth:                         auth,
	}
}

func (a *App) Create(ctx context.Context, app NewLineItemFulfillmentStatus) (LineItemFulfillmentStatus, error) {
	nt, err := toBusNewLineItemFulfillmentStatus(app)
	if err != nil {
		return LineItemFulfillmentStatus{}, err
	}

	status, err := a.lineitemfulfillmentstatusbus.Create(ctx, nt)
	if err != nil {
		if errors.Is(err, lineitemfulfillmentstatusbus.ErrUniqueEntry) {
			return LineItemFulfillmentStatus{}, errs.New(errs.AlreadyExists, err)
		}
		return LineItemFulfillmentStatus{}, err
	}

	return ToAppLineItemFulfillmentStatus(status), nil
}

func (a *App) Update(ctx context.Context, app UpdateLineItemFulfillmentStatus, id uuid.UUID) (LineItemFulfillmentStatus, error) {
	ui, err := toBusUpdateLineItemFulfillmentStatus(app)
	if err != nil {
		return LineItemFulfillmentStatus{}, errs.New(errs.InvalidArgument, err)
	}

	u, err := a.lineitemfulfillmentstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lineitemfulfillmentstatusbus.ErrNotFound) {
			return LineItemFulfillmentStatus{}, errs.New(errs.NotFound, err)
		}
		return LineItemFulfillmentStatus{}, err
	}

	status, err := a.lineitemfulfillmentstatusbus.Update(ctx, u, ui)
	if err != nil {
		if errors.Is(err, lineitemfulfillmentstatusbus.ErrNotFound) {
			return LineItemFulfillmentStatus{}, errs.New(errs.NotFound, err)
		}
		return LineItemFulfillmentStatus{}, err
	}

	return ToAppLineItemFulfillmentStatus(status), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	t, err := a.lineitemfulfillmentstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lineitemfulfillmentstatusbus.ErrNotFound) {
			return errs.New(errs.NotFound, lineitemfulfillmentstatusbus.ErrNotFound)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.lineitemfulfillmentstatusbus.Delete(ctx, t)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[LineItemFulfillmentStatus], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[LineItemFulfillmentStatus]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[LineItemFulfillmentStatus]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[LineItemFulfillmentStatus]{}, errs.NewFieldsError("orderBy", err)
	}

	items, err := a.lineitemfulfillmentstatusbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[LineItemFulfillmentStatus]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.lineitemfulfillmentstatusbus.Count(ctx, filter)
	if err != nil {
		return query.Result[LineItemFulfillmentStatus]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppLineItemFulfillmentStatuses(items), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (LineItemFulfillmentStatus, error) {
	status, err := a.lineitemfulfillmentstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lineitemfulfillmentstatusbus.ErrNotFound) {
			return LineItemFulfillmentStatus{}, errs.New(errs.NotFound, err)
		}
		return LineItemFulfillmentStatus{}, err
	}

	return ToAppLineItemFulfillmentStatus(status), nil
}
