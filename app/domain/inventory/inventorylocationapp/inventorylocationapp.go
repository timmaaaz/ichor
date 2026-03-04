package inventorylocationapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	inventorylocationbus *inventorylocationbus.Business
	zonebus              *zonebus.Business
	auth                 *auth.Auth
}

func NewApp(inventorylocationbus *inventorylocationbus.Business, zonebus *zonebus.Business) *App {
	return &App{
		inventorylocationbus: inventorylocationbus,
		zonebus:              zonebus,
	}
}

func NewAppWithAuth(inventoryLocationBus *inventorylocationbus.Business, zonebus *zonebus.Business, auth *auth.Auth) *App {
	return &App{
		inventorylocationbus: inventoryLocationBus,
		zonebus:              zonebus,
		auth:                 auth,
	}
}

func (a *App) Create(ctx context.Context, app NewInventoryLocation) (InventoryLocation, error) {
	// If warehouse_id is not provided, derive it from the selected zone.
	if app.WarehouseID == "" {
		zoneID, err := uuid.Parse(app.ZoneID)
		if err != nil {
			return InventoryLocation{}, errs.Newf(errs.InvalidArgument, "parse zoneID: %s", err)
		}
		zone, err := a.zonebus.QueryByID(ctx, zoneID)
		if err != nil {
			if errors.Is(err, zonebus.ErrNotFound) {
				return InventoryLocation{}, errs.New(errs.NotFound, err)
			}
			return InventoryLocation{}, fmt.Errorf("derive warehouse_id from zone: %w", err)
		}
		app.WarehouseID = zone.WarehouseID.String()
	}

	nl, err := toBusNewInventoryLocation(app)
	if err != nil {
		return InventoryLocation{}, errs.New(errs.InvalidArgument, err)
	}

	il, err := a.inventorylocationbus.Create(ctx, nl)
	if err != nil {
		if errors.Is(err, inventorylocationbus.ErrForeignKeyViolation) {
			return InventoryLocation{}, errs.New(errs.Aborted, err)
		}
		if errors.Is(err, inventorylocationbus.ErrUniqueEntry) {
			return InventoryLocation{}, errs.New(errs.AlreadyExists, err)
		}
		return InventoryLocation{}, fmt.Errorf("create: %w", err)
	}

	return ToAppInventoryLocation(il), nil
}

func (a *App) Update(ctx context.Context, app UpdateInventoryLocation, id uuid.UUID) (InventoryLocation, error) {
	ul, err := toBusUpdateInventoryLocation(app)
	if err != nil {
		return InventoryLocation{}, errs.New(errs.InvalidArgument, err)
	}

	il, err := a.inventorylocationbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventorylocationbus.ErrNotFound) {
			return InventoryLocation{}, errs.New(errs.NotFound, err)
		}
		return InventoryLocation{}, fmt.Errorf("update: %w", err)
	}

	il, err = a.inventorylocationbus.Update(ctx, il, ul)
	if err != nil {
		if errors.Is(err, inventorylocationbus.ErrForeignKeyViolation) {
			return InventoryLocation{}, errs.New(errs.Aborted, err)
		}
		if errors.Is(err, inventorylocationbus.ErrUniqueEntry) {
			return InventoryLocation{}, errs.New(errs.AlreadyExists, err)
		}
		return InventoryLocation{}, fmt.Errorf("update: %w", err)
	}

	return ToAppInventoryLocation(il), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	il, err := a.inventorylocationbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventorylocationbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete: %w", err)
	}

	err = a.inventorylocationbus.Delete(ctx, il)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[InventoryLocation], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[InventoryLocation]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[InventoryLocation]{}, errs.NewFieldsError("filter", err)
	}
	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[InventoryLocation]{}, errs.NewFieldsError("orderBy", err)
	}

	ils, err := a.inventorylocationbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[InventoryLocation]{}, errs.Newf(errs.Internal, "query %v", err)
	}

	total, err := a.inventorylocationbus.Count(ctx, filter)
	if err != nil {
		return query.Result[InventoryLocation]{}, errs.Newf(errs.Internal, "count %v", err)
	}

	return query.NewResult(ToAppInventoryLocations(ils), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (InventoryLocation, error) {
	il, err := a.inventorylocationbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventorylocationbus.ErrNotFound) {
			return InventoryLocation{}, errs.New(errs.NotFound, err)
		}
		return InventoryLocation{}, fmt.Errorf("querybyid: %w", err)
	}

	return ToAppInventoryLocation(il), nil
}
