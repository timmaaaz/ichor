package costhistoryapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/finance/costhistorybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	costhistorybus *costhistorybus.Business
	auth           *auth.Auth
}

// NewApp constructs a product cost app API for use.
func NewApp(costhistorybus *costhistorybus.Business) *App {
	return &App{
		costhistorybus: costhistorybus,
	}
}

// NewAppWithAuth constructs a product cost app API for use with auth support.
func NewAppWithAuth(costhistorybus *costhistorybus.Business, ath *auth.Auth) *App {
	return &App{
		auth:           ath,
		costhistorybus: costhistorybus,
	}
}

// Create adds a new product cost to the system.
func (a *App) Create(ctx context.Context, app NewCostHistory) (CostHistory, error) {
	nb, err := toBusNewCostHistory(app)
	if err != nil {
		return CostHistory{}, errs.New(errs.InvalidArgument, err)
	}

	productCost, err := a.costhistorybus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, costhistorybus.ErrUniqueEntry) {
			return CostHistory{}, errs.New(errs.AlreadyExists, costhistorybus.ErrUniqueEntry)
		}
		if errors.Is(err, costhistorybus.ErrForeignKeyViolation) {
			return CostHistory{}, errs.New(errs.Aborted, costhistorybus.ErrForeignKeyViolation)
		}
		return CostHistory{}, errs.Newf(errs.Internal, "create: product cost[%+v]: %s", productCost, err)
	}

	return ToAppCostHistory(productCost), err
}

// Update updates an existing product cost.
func (a *App) Update(ctx context.Context, app UpdateCostHistory, id uuid.UUID) (CostHistory, error) {
	upc, err := toBusUpdateCostHistory(app)
	if err != nil {
		return CostHistory{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.costhistorybus.QueryByID(ctx, id)
	if err != nil {
		return CostHistory{}, errs.New(errs.NotFound, costhistorybus.ErrNotFound)
	}

	productCost, err := a.costhistorybus.Update(ctx, st, upc)
	if err != nil {
		if errors.Is(err, costhistorybus.ErrForeignKeyViolation) {
			return CostHistory{}, errs.New(errs.Aborted, costhistorybus.ErrForeignKeyViolation)
		}
		if errors.Is(err, costhistorybus.ErrNotFound) {
			return CostHistory{}, errs.New(errs.NotFound, err)
		}
		return CostHistory{}, errs.Newf(errs.Internal, "update: productCost[%+v]: %s", productCost, err)
	}

	return ToAppCostHistory(productCost), nil
}

// Delete removes an existing productCost.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	productCost, err := a.costhistorybus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, costhistorybus.ErrNotFound)
	}

	err = a.costhistorybus.Delete(ctx, productCost)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: productCost[%+v]: %s", productCost, err)
	}

	return nil
}

// Query returns a list of productCosts based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[CostHistory], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[CostHistory]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[CostHistory]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[CostHistory]{}, errs.NewFieldsError("orderby", err)
	}

	productCosts, err := a.costhistorybus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[CostHistory]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.costhistorybus.Count(ctx, filter)
	if err != nil {
		return query.Result[CostHistory]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppCostHistories(productCosts), total, page), nil
}

// QueryByID retrieves a single productCost by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (CostHistory, error) {
	productCost, err := a.costhistorybus.QueryByID(ctx, id)
	if err != nil {
		return CostHistory{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppCostHistory(productCost), nil
}
