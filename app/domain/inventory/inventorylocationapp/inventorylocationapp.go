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
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	inventorylocationbus *inventorylocationbus.Business
	auth                 *auth.Auth
}

func NewApp(inventorylocationbus *inventorylocationbus.Business) *App {
	return &App{
		inventorylocationbus: inventorylocationbus,
	}
}

func NewAppWithAuth(inventoryLocationBus *inventorylocationbus.Business, auth *auth.Auth) *App {
	return &App{
		inventorylocationbus: inventoryLocationBus,
		auth:                 auth,
	}
}

func (a *App) Create(ctx context.Context, app NewInventoryLocation) (InventoryLocation, error) {
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
