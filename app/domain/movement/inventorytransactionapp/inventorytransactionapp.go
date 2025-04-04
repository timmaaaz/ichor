package inventorytransactionapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	inventorytransactionbus *inventorytransactionbus.Business
	auth                    *auth.Auth
}

func NewApp(inventorytransactionbus *inventorytransactionbus.Business) *App {
	return &App{
		inventorytransactionbus: inventorytransactionbus,
	}
}

func NewAppWithAuth(inventorytransactionbus *inventorytransactionbus.Business, auth *auth.Auth) *App {
	return &App{
		inventorytransactionbus: inventorytransactionbus,
		auth:                    auth,
	}
}

func (a *App) Create(ctx context.Context, app NewInventoryTransaction) (InventoryTransaction, error) {
	nt, err := toBusNewInventoryTransaction(app)
	if err != nil {
		return InventoryTransaction{}, errs.New(errs.InvalidArgument, err)
	}

	t, err := a.inventorytransactionbus.Create(ctx, nt)
	if err != nil {
		if errors.Is(err, inventorytransactionbus.ErrUniqueEntry) {
			return InventoryTransaction{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, inventorytransactionbus.ErrForeignKeyViolation) {
			return InventoryTransaction{}, errs.New(errs.Aborted, err)
		}
		return InventoryTransaction{}, fmt.Errorf("create: %w", err)
	}

	return ToAppInventoryTransaction(t), nil
}

func (a *App) Update(ctx context.Context, id uuid.UUID, app UpdateInventoryTransaction) (InventoryTransaction, error) {
	ui, err := toBusUpdateInventoryTransaction(app)
	if err != nil {
		return InventoryTransaction{}, errs.New(errs.InvalidArgument, err)
	}

	t, err := a.inventorytransactionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventorytransactionbus.ErrNotFound) {
			return InventoryTransaction{}, errs.New(errs.NotFound, err)
		}
		return InventoryTransaction{}, err
	}

	t, err = a.inventorytransactionbus.Update(ctx, t, ui)
	if err != nil {
		if errors.Is(err, inventorytransactionbus.ErrUniqueEntry) {
			return InventoryTransaction{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, inventorytransactionbus.ErrForeignKeyViolation) {
			return InventoryTransaction{}, errs.New(errs.Aborted, err)
		}
		return InventoryTransaction{}, fmt.Errorf("update: %w", err)
	}

	return ToAppInventoryTransaction(t), nil

}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	t, err := a.inventorytransactionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventorytransactionbus.ErrNotFound) {
			return errs.New(errs.NotFound, inventorytransactionbus.ErrNotFound)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.inventorytransactionbus.Delete(ctx, t)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[InventoryTransaction], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[InventoryTransaction]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[InventoryTransaction]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[InventoryTransaction]{}, errs.NewFieldsError("orderBy", err)
	}

	items, err := a.inventorytransactionbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[InventoryTransaction]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.inventorytransactionbus.Count(ctx, filter)
	if err != nil {
		return query.Result[InventoryTransaction]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppInventoryTransactions(items), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (InventoryTransaction, error) {
	t, err := a.inventorytransactionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventorytransactionbus.ErrNotFound) {
			return InventoryTransaction{}, errs.New(errs.NotFound, err)
		}
		return InventoryTransaction{}, errs.Newf(errs.Internal, "queryByID: %v", err)
	}

	return ToAppInventoryTransaction(t), nil
}
