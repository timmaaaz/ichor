package purchaseorderlineitemstatusapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the purchase order line item status domain.
type App struct {
	purchaseorderlineitemstatusbus *purchaseorderlineitemstatusbus.Business
	auth                           *auth.Auth
}

// NewApp constructs a purchase order line item status app API for use.
func NewApp(purchaseorderlineitemstatusbus *purchaseorderlineitemstatusbus.Business) *App {
	return &App{
		purchaseorderlineitemstatusbus: purchaseorderlineitemstatusbus,
	}
}

// NewAppWithAuth constructs a purchase order line item status app API for use with auth support.
func NewAppWithAuth(purchaseorderlineitemstatusbus *purchaseorderlineitemstatusbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:                           ath,
		purchaseorderlineitemstatusbus: purchaseorderlineitemstatusbus,
	}
}

// Create adds a new purchase order line item status to the system.
func (a *App) Create(ctx context.Context, app NewPurchaseOrderLineItemStatus) (PurchaseOrderLineItemStatus, error) {
	nb, err := toBusNewPurchaseOrderLineItemStatus(app)
	if err != nil {
		return PurchaseOrderLineItemStatus{}, errs.New(errs.InvalidArgument, err)
	}

	polis, err := a.purchaseorderlineitemstatusbus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, purchaseorderlineitemstatusbus.ErrUnique) {
			return PurchaseOrderLineItemStatus{}, errs.New(errs.AlreadyExists, err)
		}
		return PurchaseOrderLineItemStatus{}, errs.Newf(errs.Internal, "create: purchaseorderlineitemstatus[%+v]: %s", polis, err)
	}

	return ToAppPurchaseOrderLineItemStatus(polis), nil
}

// Update updates an existing purchase order line item status.
func (a *App) Update(ctx context.Context, app UpdatePurchaseOrderLineItemStatus, id uuid.UUID) (PurchaseOrderLineItemStatus, error) {
	upolis, err := toBusUpdatePurchaseOrderLineItemStatus(app)
	if err != nil {
		return PurchaseOrderLineItemStatus{}, errs.New(errs.InvalidArgument, err)
	}

	polis, err := a.purchaseorderlineitemstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderlineitemstatusbus.ErrNotFound) {
			return PurchaseOrderLineItemStatus{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderLineItemStatus{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	updatedPolis, err := a.purchaseorderlineitemstatusbus.Update(ctx, polis, upolis)
	if err != nil {
		if errors.Is(err, purchaseorderlineitemstatusbus.ErrNotFound) {
			return PurchaseOrderLineItemStatus{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderLineItemStatus{}, errs.Newf(errs.Internal, "update: purchaseorderlineitemstatus[%+v]: %s", updatedPolis, err)
	}

	return ToAppPurchaseOrderLineItemStatus(updatedPolis), nil
}

// Delete removes an existing purchase order line item status.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	polis, err := a.purchaseorderlineitemstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderlineitemstatusbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if err := a.purchaseorderlineitemstatusbus.Delete(ctx, polis); err != nil {
		return errs.Newf(errs.Internal, "delete: purchaseorderlineitemstatus[%+v]: %s", polis, err)
	}

	return nil
}

// Query returns a list of purchase order line item statuses based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[PurchaseOrderLineItemStatus], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[PurchaseOrderLineItemStatus]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[PurchaseOrderLineItemStatus]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[PurchaseOrderLineItemStatus]{}, errs.NewFieldsError("orderby", err)
	}

	statuses, err := a.purchaseorderlineitemstatusbus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[PurchaseOrderLineItemStatus]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.purchaseorderlineitemstatusbus.Count(ctx, filter)
	if err != nil {
		return query.Result[PurchaseOrderLineItemStatus]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppPurchaseOrderLineItemStatuses(statuses), total, pg), nil
}

// QueryByID retrieves a single purchase order line item status by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (PurchaseOrderLineItemStatus, error) {
	polis, err := a.purchaseorderlineitemstatusbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderlineitemstatusbus.ErrNotFound) {
			return PurchaseOrderLineItemStatus{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderLineItemStatus{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppPurchaseOrderLineItemStatus(polis), nil
}

// QueryByIDs retrieves purchase order line item statuses by their ids.
func (a *App) QueryByIDs(ctx context.Context, ids []string) (PurchaseOrderLineItemStatuses, error) {
	uuids, err := toBusIDs(ids)
	if err != nil {
		return nil, errs.NewFieldsError("ids", err)
	}

	statuses, err := a.purchaseorderlineitemstatusbus.QueryByIDs(ctx, uuids)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "querybyids: %s", err)
	}

	return PurchaseOrderLineItemStatuses(ToAppPurchaseOrderLineItemStatuses(statuses)), nil
}

// QueryAll retrieves all purchase order line item statuses from the system.
func (a *App) QueryAll(ctx context.Context) (PurchaseOrderLineItemStatuses, error) {
	statuses, err := a.purchaseorderlineitemstatusbus.QueryAll(ctx)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "queryall: %s", err)
	}

	return PurchaseOrderLineItemStatuses(ToAppPurchaseOrderLineItemStatuses(statuses)), nil
}
