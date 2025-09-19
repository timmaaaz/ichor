package inventoryadjustmentapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	inventoryadjustmentbus *inventoryadjustmentbus.Business
	auth                   *auth.Auth
}

// NewApp constructs an inventory adjustment app API for use.
func NewApp(inventoryadjustmentbus *inventoryadjustmentbus.Business) *App {
	return &App{
		inventoryadjustmentbus: inventoryadjustmentbus,
	}
}

// NewAppWithAuth constructs an inventory adjustment app API for use with auth support.
func NewAppWithAuth(inventoryadjustmentbus *inventoryadjustmentbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:                   ath,
		inventoryadjustmentbus: inventoryadjustmentbus,
	}
}

// Create creates a new inventory adjustment.
func (a *App) Create(ctx context.Context, app NewInventoryAdjustment) (InventoryAdjustment, error) {
	newAdjustment, err := toBusNewInventoryAdjustment(app)
	if err != nil {
		return InventoryAdjustment{}, errs.New(errs.InvalidArgument, err)
	}

	adjustment, err := a.inventoryadjustmentbus.Create(ctx, newAdjustment)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrUniqueEntry) {
			return InventoryAdjustment{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, inventoryadjustmentbus.ErrForeignKeyViolation) {
			return InventoryAdjustment{}, errs.New(errs.Aborted, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("create %w", err)
	}

	return ToAppInventoryAdjustment(adjustment), nil

}

func (a *App) Update(ctx context.Context, id uuid.UUID, app UpdateInventoryAdjustment) (InventoryAdjustment, error) {
	updateAdjustment, err := toBusUpdateInventoryAdjustment(app)
	if err != nil {
		return InventoryAdjustment{}, errs.New(errs.InvalidArgument, err)
	}

	adjustment, err := a.inventoryadjustmentbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return InventoryAdjustment{}, errs.New(errs.NotFound, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("update [queryByID] %w", err)
	}

	adjustment, err = a.inventoryadjustmentbus.Update(ctx, adjustment, updateAdjustment)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrUniqueEntry) {
			return InventoryAdjustment{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, inventoryadjustmentbus.ErrForeignKeyViolation) {
			return InventoryAdjustment{}, errs.New(errs.Aborted, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("update %w", err)
	}

	return ToAppInventoryAdjustment(adjustment), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	adjustment, err := a.inventoryadjustmentbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [queryByID]: %w", err)
	}

	err = a.inventoryadjustmentbus.Delete(ctx, adjustment)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[InventoryAdjustment], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.NewFieldsError("orderBy", err)
	}

	results, err := a.inventoryadjustmentbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.inventoryadjustmentbus.Count(ctx, filter)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppInventoryAdjustments(results), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (InventoryAdjustment, error) {
	adjustment, err := a.inventoryadjustmentbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return InventoryAdjustment{}, errs.New(errs.NotFound, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("querybyid: %w", err)
	}

	return ToAppInventoryAdjustment(adjustment), nil
}
