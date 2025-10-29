package purchaseorderstatusapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the purchase order status domain.
type App struct {
	purchaseorderstatusbus *purchaseorderstatusbus.Business
	auth                   *auth.Auth
}

// NewApp constructs a purchase order status app API for use.
func NewApp(purchaseorderstatusbus *purchaseorderstatusbus.Business) *App {
	return &App{
		purchaseorderstatusbus: purchaseorderstatusbus,
	}
}

// NewAppWithAuth constructs a purchase order status app API for use with auth support.
func NewAppWithAuth(purchaseorderstatusbus *purchaseorderstatusbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:                   ath,
		purchaseorderstatusbus: purchaseorderstatusbus,
	}
}

// Create adds a new purchase order status to the system.
func (a *App) Create(ctx context.Context, app NewPurchaseOrderStatus) (PurchaseOrderStatus, error) {
	nb, err := toBusNewPurchaseOrderStatus(app)
	if err != nil {
		return PurchaseOrderStatus{}, errs.New(errs.InvalidArgument, err)
	}

	pos, err := a.purchaseorderstatusbus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, purchaseorderstatusbus.ErrUnique) {
			return PurchaseOrderStatus{}, errs.New(errs.AlreadyExists, err)
		}
		return PurchaseOrderStatus{}, errs.Newf(errs.Internal, "create: purchaseorderstatus[%+v]: %s", pos, err)
	}

	return ToAppPurchaseOrderStatus(pos), nil
}

// Update updates an existing purchase order status.
func (a *App) Update(ctx context.Context, app UpdatePurchaseOrderStatus, id uuid.UUID) (PurchaseOrderStatus, error) {
	upos, err := toBusUpdatePurchaseOrderStatus(app)
	if err != nil {
		return PurchaseOrderStatus{}, errs.New(errs.InvalidArgument, err)
	}

	pos, err := a.purchaseorderstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderstatusbus.ErrNotFound) {
			return PurchaseOrderStatus{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderStatus{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	updatedPos, err := a.purchaseorderstatusbus.Update(ctx, pos, upos)
	if err != nil {
		if errors.Is(err, purchaseorderstatusbus.ErrNotFound) {
			return PurchaseOrderStatus{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderStatus{}, errs.Newf(errs.Internal, "update: purchaseorderstatus[%+v]: %s", updatedPos, err)
	}

	return ToAppPurchaseOrderStatus(updatedPos), nil
}

// Delete removes an existing purchase order status.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	pos, err := a.purchaseorderstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderstatusbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if err := a.purchaseorderstatusbus.Delete(ctx, pos); err != nil {
		return errs.Newf(errs.Internal, "delete: purchaseorderstatus[%+v]: %s", pos, err)
	}

	return nil
}

// Query returns a list of purchase order statuses based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[PurchaseOrderStatus], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[PurchaseOrderStatus]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[PurchaseOrderStatus]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[PurchaseOrderStatus]{}, errs.NewFieldsError("orderby", err)
	}

	statuses, err := a.purchaseorderstatusbus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[PurchaseOrderStatus]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.purchaseorderstatusbus.Count(ctx, filter)
	if err != nil {
		return query.Result[PurchaseOrderStatus]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppPurchaseOrderStatuses(statuses), total, pg), nil
}

// QueryByID retrieves a single purchase order status by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (PurchaseOrderStatus, error) {
	pos, err := a.purchaseorderstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderstatusbus.ErrNotFound) {
			return PurchaseOrderStatus{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderStatus{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppPurchaseOrderStatus(pos), nil
}

// QueryByIDs retrieves purchase order statuses by their ids.
func (a *App) QueryByIDs(ctx context.Context, ids []string) (PurchaseOrderStatuses, error) {
	uuids, err := toBusIDs(ids)
	if err != nil {
		return nil, errs.NewFieldsError("ids", err)
	}

	statuses, err := a.purchaseorderstatusbus.QueryByIDs(ctx, uuids)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "querybyids: %s", err)
	}

	return PurchaseOrderStatuses(ToAppPurchaseOrderStatuses(statuses)), nil
}

// QueryAll retrieves all purchase order statuses from the system.
func (a *App) QueryAll(ctx context.Context) (PurchaseOrderStatuses, error) {
	statuses, err := a.purchaseorderstatusbus.QueryAll(ctx)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "queryall: %s", err)
	}

	return PurchaseOrderStatuses(ToAppPurchaseOrderStatuses(statuses)), nil
}
