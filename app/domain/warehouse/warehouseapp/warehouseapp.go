package warehouseapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the warehouse domain.
type App struct {
	warehouseBus *warehousebus.Business
	auth         *auth.Auth
}

// NewApp constructs a warehouse app API for use.
func NewApp(warehouseBus *warehousebus.Business) *App {
	return &App{
		warehouseBus: warehouseBus,
	}
}

// NewAppWithAuth constructs a warehouse app API for use with auth support.
func NewAppWithAuth(warehouseBus *warehousebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:         ath,
		warehouseBus: warehouseBus,
	}
}

// Create adds a new warehouse to the system.
func (a *App) Create(ctx context.Context, app NewWarehouse) (Warehouse, error) {
	nw, err := toBusNewWarehouse(app)
	if err != nil {
		return Warehouse{}, err
	}

	wa, err := a.warehouseBus.Create(ctx, nw)
	if err != nil {
		if errors.Is(err, warehousebus.ErrUniqueEntry) {
			return Warehouse{}, errs.New(errs.Aborted, warehousebus.ErrUniqueEntry)
		}
		return Warehouse{}, errs.Newf(errs.Internal, "create: warehouse[%+v]: %s", wa, err)
	}

	return ToAppWarehouse(wa), nil
}

// Update updates an existing warehouse.
func (a *App) Update(ctx context.Context, app UpdateWarehouse, id uuid.UUID) (Warehouse, error) {
	uw, err := toBusUpdateWarehouse(app)
	if err != nil {
		return Warehouse{}, errs.New(errs.InvalidArgument, err)
	}

	wa, err := a.warehouseBus.QueryByID(ctx, id)
	if err != nil {
		return Warehouse{}, errs.New(errs.NotFound, warehousebus.ErrNotFound)
	}

	warehouse, err := a.warehouseBus.Update(ctx, wa, uw)
	if err != nil {
		if errors.Is(err, warehousebus.ErrNotFound) {
			return Warehouse{}, errs.New(errs.NotFound, err)
		}
		return Warehouse{}, errs.Newf(errs.Internal, "update: warehouse[%+v]: %s", warehouse, err)
	}

	return ToAppWarehouse(warehouse), nil
}

// Delete removes an existing warehouse.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	wa, err := a.warehouseBus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, warehousebus.ErrNotFound)
	}

	err = a.warehouseBus.Delete(ctx, wa)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// Query returns a list of warehouses based on the filter, order, and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Warehouse], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Warehouse]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Warehouse]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Warehouse]{}, errs.NewFieldsError("orderby", err)
	}

	warehouses, err := a.warehouseBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Warehouse]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.warehouseBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Warehouse]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppWarehouses(warehouses), total, page), nil
}

// QueryByID returns a single warehouse based on the id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Warehouse, error) {
	warehouse, err := a.warehouseBus.QueryByID(ctx, id)
	if err != nil {
		return Warehouse{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}
	return ToAppWarehouse(warehouse), nil
}
