package ordersapp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

type App struct {
	ordersbus *ordersbus.Business
	auth      *auth.Auth
}

func NewApp(ordersbus *ordersbus.Business) *App {
	return &App{
		ordersbus: ordersbus,
	}
}

func NewAppWithAuth(ordersbus *ordersbus.Business, auth *auth.Auth) *App {
	return &App{
		ordersbus: ordersbus,
		auth:      auth,
	}
}

// NewWithTx returns a copy of App whose bus(es) run on the given transaction, so callers
// (e.g. formdataapp.UpsertFormData via formdataregistry.TxBind) can enroll this app's writes
// in a larger atomic unit of work.
func (a *App) NewWithTx(tx sqldb.CommitRollbacker) (*App, error) {
	busTx, err := a.ordersbus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}
	nb := *a
	nb.ordersbus = busTx
	return &nb, nil
}

func (a *App) Create(ctx context.Context, app NewOrder) (Order, error) {
	nt, err := toBusNewOrder(app)
	if err != nil {
		return Order{}, err
	}

	status, err := a.ordersbus.Create(ctx, nt)
	if err != nil {
		if errors.Is(err, ordersbus.ErrUniqueEntry) {
			return Order{}, errs.New(errs.AlreadyExists, err)
		}
		return Order{}, err
	}

	return ToAppOrder(status), nil
}

func (a *App) Update(ctx context.Context, app UpdateOrder, id uuid.UUID) (Order, error) {
	ui, err := toBusUpdateOrder(app)
	if err != nil {
		return Order{}, errs.New(errs.InvalidArgument, err)
	}

	u, err := a.ordersbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return Order{}, errs.New(errs.NotFound, err)
		}
		return Order{}, err
	}

	status, err := a.ordersbus.Update(ctx, u, ui)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return Order{}, errs.New(errs.NotFound, err)
		}
		return Order{}, err
	}

	return ToAppOrder(status), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	t, err := a.ordersbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return errs.New(errs.NotFound, ordersbus.ErrNotFound)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.ordersbus.Delete(ctx, t)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Order], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Order]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Order]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Order]{}, errs.NewFieldsError("orderBy", err)
	}

	items, err := a.ordersbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Order]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.ordersbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Order]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppOrders(items), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Order, error) {
	status, err := a.ordersbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return Order{}, errs.New(errs.NotFound, err)
		}
		return Order{}, err
	}

	return ToAppOrder(status), nil
}

// =============================================================================
// Order container bindings (Phase 0g.B7)
// =============================================================================

// BindContainer creates a new active binding between an order and a container
// label. Pre-checks the order's existence (matches App.Update at line 52)
// so unknown orders surface as 404. Maps the bus-layer EXCLUDE-constraint
// violation (one_active_binding_per_container — raw pq error with substring
// "exclusion") to errs.Aborted → 409 Conflict per app/sdk/errs/codes.go:179.
func (a *App) BindContainer(ctx context.Context, orderID uuid.UUID, app NewOrderContainerBinding) (OrderContainerBinding, error) {
	if _, err := a.ordersbus.QueryByID(ctx, orderID); err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return OrderContainerBinding{}, errs.New(errs.NotFound, err)
		}
		return OrderContainerBinding{}, err
	}

	nb, err := toBusNewOrderContainerBinding(app, orderID)
	if err != nil {
		return OrderContainerBinding{}, err
	}

	result, err := a.ordersbus.BindContainer(ctx, nb)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "exclusion") {
			return OrderContainerBinding{}, errs.New(errs.Aborted, err)
		}
		return OrderContainerBinding{}, errs.Newf(errs.Internal, "bindContainer: %s", err)
	}
	return ToAppOrderContainerBinding(result), nil
}

// UnbindContainer marks an active binding as released. The bus layer is
// idempotent: unbinding an already-unbound binding is a silent no-op, but
// an unknown bindingID surfaces ErrBindingNotFound → 404.
func (a *App) UnbindContainer(ctx context.Context, bindingID uuid.UUID) error {
	if err := a.ordersbus.UnbindContainer(ctx, bindingID); err != nil {
		if errors.Is(err, ordersbus.ErrBindingNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "unbindContainer: %s", err)
	}
	return nil
}

// QueryActiveBindingsByOrder returns the active (unbound_at IS NULL) bindings
// for the given order, ordered by bound_at ASC. Returns an empty slice for
// orders with no active bindings — NOT a 404. The order's existence is not
// pre-checked because an unknown orderID legitimately maps to "no bindings"
// from a read perspective.
func (a *App) QueryActiveBindingsByOrder(ctx context.Context, orderID uuid.UUID) (OrderContainerBindings, error) {
	bindings, err := a.ordersbus.QueryActiveBindingsByOrder(ctx, orderID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "queryActiveBindingsByOrder: %s", err)
	}
	return OrderContainerBindings(ToAppOrderContainerBindings(bindings)), nil
}
