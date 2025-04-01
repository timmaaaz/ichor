package inventoryitemapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the product domain.
type App struct {
	inventoryitembus *inventoryitembus.Business
	auth             *auth.Auth
}

// NewApp constructs a product app API for use.
func NewApp(inventoryitembus *inventoryitembus.Business) *App {
	return &App{
		inventoryitembus: inventoryitembus,
	}
}

// NewAppWithAuth constructs a product app API for use with auth support.
func NewAppWithAuth(inventoryitembus *inventoryitembus.Business, ath *auth.Auth) *App {
	return &App{
		auth:             ath,
		inventoryitembus: inventoryitembus,
	}
}

// Create creates a new inventory item.
func (a *App) Create(ctx context.Context, app NewInventoryItem) (InventoryItem, error) {
	newItem, err := toBusNewInventoryItem(app)
	if err != nil {
		return InventoryItem{}, errs.New(errs.InvalidArgument, err)
	}

	item, err := a.inventoryitembus.Create(ctx, newItem)
	if err != nil {
		if errors.Is(err, inventoryitembus.ErrUniqueEntry) {
			return InventoryItem{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, inventoryitembus.ErrForeignKeyViolation) {
			return InventoryItem{}, errs.New(errs.Aborted, err)
		}
		return InventoryItem{}, errs.Newf(errs.Internal, "create: item[%+v]: %s", item, err)
	}

	return ToAppInventoryItem(item), nil
}

// Update updates an existing inventory item.
func (a *App) Update(ctx context.Context, app UpdateInventoryItem, id uuid.UUID) (InventoryItem, error) {
	uii, err := toBusUpdateInventoryItem(app)
	if err != nil {
		return InventoryItem{}, errs.New(errs.InvalidArgument, err)
	}

	item, err := a.inventoryitembus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryitembus.ErrNotFound) {
			return InventoryItem{}, errs.New(errs.NotFound, err)
		}
		return InventoryItem{}, errs.New(errs.Internal, err)
	}

	inventoryItem, err := a.inventoryitembus.Update(ctx, item, uii)
	if err != nil {
		if errors.Is(err, inventoryitembus.ErrForeignKeyViolation) {
			return InventoryItem{}, errs.New(errs.Aborted, err)
		}
		if errors.Is(err, inventoryitembus.ErrUniqueEntry) {
			return InventoryItem{}, errs.New(errs.AlreadyExists, err)
		}
		return InventoryItem{}, errs.Newf(errs.Internal, "update: item[%+v]: %s", inventoryItem, err)
	}

	return ToAppInventoryItem(inventoryItem), nil

}

// Delete removes an existing inventory item.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	item, err := a.inventoryitembus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryitembus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.New(errs.Internal, err)
	}

	err = a.inventoryitembus.Delete(ctx, item)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: item[%s]: %s", id, err)
	}

	return nil
}

// Query returns a list of inventory items based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[InventoryItem], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[InventoryItem]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[InventoryItem]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[InventoryItem]{}, errs.NewFieldsError("orderBy", err)
	}

	items, err := a.inventoryitembus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[InventoryItem]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.inventoryitembus.Count(ctx, filter)
	if err != nil {
		return query.Result[InventoryItem]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppInventoryItems(items), total, page), nil
}

// QueryByID retrieves the inventory item by ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (InventoryItem, error) {
	item, err := a.inventoryitembus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryitembus.ErrNotFound) {
			return InventoryItem{}, errs.New(errs.NotFound, err)
		}
		return InventoryItem{}, errs.New(errs.Internal, err)
	}

	return ToAppInventoryItem(item), nil
}
