package purchaseorderlineitemapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the purchase order line item domain.
type App struct {
	purchaseorderlineitembus *purchaseorderlineitembus.Business
	auth                     *auth.Auth
}

// NewApp constructs a purchase order line item app API for use.
func NewApp(purchaseorderlineitembus *purchaseorderlineitembus.Business) *App {
	return &App{
		purchaseorderlineitembus: purchaseorderlineitembus,
	}
}

// NewAppWithAuth constructs a purchase order line item app API for use with auth support.
func NewAppWithAuth(purchaseorderlineitembus *purchaseorderlineitembus.Business, ath *auth.Auth) *App {
	return &App{
		auth:                     ath,
		purchaseorderlineitembus: purchaseorderlineitembus,
	}
}

// Create adds a new purchase order line item to the system.
func (a *App) Create(ctx context.Context, app NewPurchaseOrderLineItem) (PurchaseOrderLineItem, error) {
	nb, err := toBusNewPurchaseOrderLineItem(app)
	if err != nil {
		return PurchaseOrderLineItem{}, errs.New(errs.InvalidArgument, err)
	}

	poli, err := a.purchaseorderlineitembus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, purchaseorderlineitembus.ErrUnique) {
			return PurchaseOrderLineItem{}, errs.New(errs.AlreadyExists, err)
		}
		return PurchaseOrderLineItem{}, errs.Newf(errs.Internal, "create: purchaseorderlineitem[%+v]: %s", poli, err)
	}

	return ToAppPurchaseOrderLineItem(poli), nil
}

// Update updates an existing purchase order line item.
func (a *App) Update(ctx context.Context, app UpdatePurchaseOrderLineItem, id uuid.UUID) (PurchaseOrderLineItem, error) {
	upoli, err := toBusUpdatePurchaseOrderLineItem(app)
	if err != nil {
		return PurchaseOrderLineItem{}, errs.New(errs.InvalidArgument, err)
	}

	poli, err := a.purchaseorderlineitembus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderlineitembus.ErrNotFound) {
			return PurchaseOrderLineItem{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderLineItem{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	updatedPOLI, err := a.purchaseorderlineitembus.Update(ctx, poli, upoli)
	if err != nil {
		if errors.Is(err, purchaseorderlineitembus.ErrNotFound) {
			return PurchaseOrderLineItem{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderLineItem{}, errs.Newf(errs.Internal, "update: purchaseorderlineitem[%+v]: %s", updatedPOLI, err)
	}

	return ToAppPurchaseOrderLineItem(updatedPOLI), nil
}

// Delete removes an existing purchase order line item.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	poli, err := a.purchaseorderlineitembus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderlineitembus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if err := a.purchaseorderlineitembus.Delete(ctx, poli); err != nil {
		return errs.Newf(errs.Internal, "delete: purchaseorderlineitem[%+v]: %s", poli, err)
	}

	return nil
}

// Query returns a list of purchase order line items based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[PurchaseOrderLineItem], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[PurchaseOrderLineItem]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[PurchaseOrderLineItem]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[PurchaseOrderLineItem]{}, errs.NewFieldsError("orderby", err)
	}

	items, err := a.purchaseorderlineitembus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[PurchaseOrderLineItem]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.purchaseorderlineitembus.Count(ctx, filter)
	if err != nil {
		return query.Result[PurchaseOrderLineItem]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppPurchaseOrderLineItems(items), total, pg), nil
}

// QueryByID retrieves a single purchase order line item by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (PurchaseOrderLineItem, error) {
	poli, err := a.purchaseorderlineitembus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderlineitembus.ErrNotFound) {
			return PurchaseOrderLineItem{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderLineItem{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppPurchaseOrderLineItem(poli), nil
}

// QueryByIDs retrieves purchase order line items by their ids.
func (a *App) QueryByIDs(ctx context.Context, ids []string) (PurchaseOrderLineItems, error) {
	uuids, err := toBusIDs(ids)
	if err != nil {
		return nil, errs.NewFieldsError("ids", err)
	}

	items, err := a.purchaseorderlineitembus.QueryByIDs(ctx, uuids)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "querybyids: %s", err)
	}

	return PurchaseOrderLineItems(ToAppPurchaseOrderLineItems(items)), nil
}

// QueryByPurchaseOrderID retrieves all line items for a specific purchase order.
func (a *App) QueryByPurchaseOrderID(ctx context.Context, poID uuid.UUID) (PurchaseOrderLineItems, error) {
	items, err := a.purchaseorderlineitembus.QueryByPurchaseOrderID(ctx, poID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "querybypurchaseorderid: %s", err)
	}

	return PurchaseOrderLineItems(ToAppPurchaseOrderLineItems(items)), nil
}

// ReceiveQuantity updates the received quantity for a line item.
func (a *App) ReceiveQuantity(ctx context.Context, id uuid.UUID, quantity int, receivedBy uuid.UUID) (PurchaseOrderLineItem, error) {
	poli, err := a.purchaseorderlineitembus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderlineitembus.ErrNotFound) {
			return PurchaseOrderLineItem{}, errs.New(errs.NotFound, err)
		}
		return PurchaseOrderLineItem{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	updatedPOLI, err := a.purchaseorderlineitembus.ReceiveQuantity(ctx, poli, quantity, receivedBy)
	if err != nil {
		return PurchaseOrderLineItem{}, errs.Newf(errs.Internal, "receivequantity: purchaseorderlineitem[%+v]: %s", updatedPOLI, err)
	}

	return ToAppPurchaseOrderLineItem(updatedPOLI), nil
}
