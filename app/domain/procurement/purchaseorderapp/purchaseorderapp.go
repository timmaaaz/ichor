package purchaseorderapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the purchase order domain.
type App struct {
	purchaseorderbus *purchaseorderbus.Business
	auth             *auth.Auth
}

// NewApp constructs a purchase order app API for use.
func NewApp(purchaseorderbus *purchaseorderbus.Business) *App {
	return &App{
		purchaseorderbus: purchaseorderbus,
	}
}

// NewAppWithAuth constructs a purchase order app API for use with auth support.
func NewAppWithAuth(purchaseorderbus *purchaseorderbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:             ath,
		purchaseorderbus: purchaseorderbus,
	}
}

// Create adds a new purchase order to the system.
func (a *App) Create(ctx context.Context, app NewPurchaseOrder) (PurchaseOrder, error) {
	nb, err := toBusNewPurchaseOrder(app)
	if err != nil {
		return PurchaseOrder{}, errs.New(errs.InvalidArgument, err)
	}

	po, err := a.purchaseorderbus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrUnique) {
			return PurchaseOrder{}, errs.New(errs.AlreadyExists, err)
		}
		return PurchaseOrder{}, errs.Newf(errs.Internal, "create: purchaseorder[%+v]: %s", po, err)
	}

	return ToAppPurchaseOrder(po), nil
}

// Update updates an existing purchase order.
func (a *App) Update(ctx context.Context, app UpdatePurchaseOrder, id uuid.UUID) (PurchaseOrder, error) {
	upo, err := toBusUpdatePurchaseOrder(app)
	if err != nil {
		return PurchaseOrder{}, errs.New(errs.InvalidArgument, err)
	}

	po, err := a.purchaseorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrNotFound) {
			return PurchaseOrder{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrder{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	updatedPO, err := a.purchaseorderbus.Update(ctx, po, upo)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrNotFound) {
			return PurchaseOrder{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrder{}, errs.Newf(errs.Internal, "update: purchaseorder[%+v]: %s", updatedPO, err)
	}

	return ToAppPurchaseOrder(updatedPO), nil
}

// Delete removes an existing purchase order.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	po, err := a.purchaseorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if err := a.purchaseorderbus.Delete(ctx, po); err != nil {
		return errs.Newf(errs.Internal, "delete: purchaseorder[%+v]: %s", po, err)
	}

	return nil
}

// Query returns a list of purchase orders based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[PurchaseOrder], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[PurchaseOrder]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[PurchaseOrder]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[PurchaseOrder]{}, errs.NewFieldsError("orderby", err)
	}

	orders, err := a.purchaseorderbus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[PurchaseOrder]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.purchaseorderbus.Count(ctx, filter)
	if err != nil {
		return query.Result[PurchaseOrder]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppPurchaseOrders(orders), total, pg), nil
}

// QueryByID retrieves a single purchase order by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	po, err := a.purchaseorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrNotFound) {
			return PurchaseOrder{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrder{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppPurchaseOrder(po), nil
}

// QueryByIDs retrieves purchase orders by their ids.
func (a *App) QueryByIDs(ctx context.Context, ids []string) (PurchaseOrders, error) {
	uuids, err := query.ParseIDs(ids)
	if err != nil {
		return nil, errs.NewFieldsError("ids", err)
	}

	orders, err := a.purchaseorderbus.QueryByIDs(ctx, uuids)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "querybyids: %s", err)
	}

	return PurchaseOrders(ToAppPurchaseOrders(orders)), nil
}

// Approve approves a purchase order.
func (a *App) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) (PurchaseOrder, error) {
	po, err := a.purchaseorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrNotFound) {
			return PurchaseOrder{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrder{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	approvedPO, err := a.purchaseorderbus.Approve(ctx, po, approvedBy)
	if err != nil {
		return PurchaseOrder{}, errs.Newf(errs.Internal, "approve: purchaseorder[%+v]: %s", approvedPO, err)
	}

	return ToAppPurchaseOrder(approvedPO), nil
}
