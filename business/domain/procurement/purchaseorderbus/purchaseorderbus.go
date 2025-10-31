package purchaseorderbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("purchase order not found")
	ErrUnique   = errors.New("not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, po PurchaseOrder) error
	Update(ctx context.Context, po PurchaseOrder) error
	Delete(ctx context.Context, po PurchaseOrder) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PurchaseOrder, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, poID uuid.UUID) (PurchaseOrder, error)
	QueryByIDs(ctx context.Context, poIDs []uuid.UUID) ([]PurchaseOrder, error)
}

// Business manages the set of APIs for purchase order access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
}

// NewBusiness constructs a purchase order business API for use.
func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:    log,
		del:    del,
		storer: storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
		del:    b.del,
	}

	return &bus, nil
}

// Create adds a new purchase order to the system.
func (b *Business) Create(ctx context.Context, npo NewPurchaseOrder) (PurchaseOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderbus.create")
	defer span.End()

	now := time.Now().UTC()

	po := PurchaseOrder{
		ID:                       uuid.New(),
		OrderNumber:              npo.OrderNumber,
		SupplierID:               npo.SupplierID,
		PurchaseOrderStatusID:    npo.PurchaseOrderStatusID,
		DeliveryWarehouseID:      npo.DeliveryWarehouseID,
		DeliveryLocationID:       npo.DeliveryLocationID,
		DeliveryStreetID:         npo.DeliveryStreetID,
		OrderDate:                npo.OrderDate,
		ExpectedDeliveryDate:     npo.ExpectedDeliveryDate,
		Subtotal:                 npo.Subtotal,
		TaxAmount:                npo.TaxAmount,
		ShippingCost:             npo.ShippingCost,
		TotalAmount:              npo.TotalAmount,
		Currency:                 npo.Currency,
		RequestedBy:              npo.RequestedBy,
		Notes:                    npo.Notes,
		SupplierReferenceNumber:  npo.SupplierReferenceNumber,
		CreatedBy:                npo.CreatedBy,
		UpdatedBy:                npo.CreatedBy,
		CreatedDate:              now,
		UpdatedDate:              now,
	}

	if err := b.storer.Create(ctx, po); err != nil {
		return PurchaseOrder{}, fmt.Errorf("create: %w", err)
	}

	return po, nil
}

// Update modifies a purchase order in the system.
func (b *Business) Update(ctx context.Context, po PurchaseOrder, upo UpdatePurchaseOrder) (PurchaseOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderbus.update")
	defer span.End()

	err := convert.PopulateSameTypes(upo, &po)
	if err != nil {
		return PurchaseOrder{}, fmt.Errorf("populate same types: %w", err)
	}

	po.UpdatedDate = time.Now().UTC()

	if err := b.storer.Update(ctx, po); err != nil {
		return PurchaseOrder{}, fmt.Errorf("update: %w", err)
	}

	return po, nil
}

// Delete removes a purchase order from the system.
func (b *Business) Delete(ctx context.Context, po PurchaseOrder) error {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, po); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of purchase orders from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PurchaseOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderbus.query")
	defer span.End()

	orders, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return orders, nil
}

// Count returns the total number of purchase orders.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the purchase order by the specified ID.
func (b *Business) QueryByID(ctx context.Context, poID uuid.UUID) (PurchaseOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderbus.querybyid")
	defer span.End()

	po, err := b.storer.QueryByID(ctx, poID)
	if err != nil {
		return PurchaseOrder{}, fmt.Errorf("querybyid: %w", err)
	}

	return po, nil
}

// QueryByIDs finds the purchase orders by the specified IDs.
func (b *Business) QueryByIDs(ctx context.Context, poIDs []uuid.UUID) ([]PurchaseOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderbus.querybyids")
	defer span.End()

	orders, err := b.storer.QueryByIDs(ctx, poIDs)
	if err != nil {
		return nil, fmt.Errorf("querybyids: %w", err)
	}

	return orders, nil
}

// Approve approves a purchase order.
func (b *Business) Approve(ctx context.Context, po PurchaseOrder, approvedBy uuid.UUID) (PurchaseOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderbus.approve")
	defer span.End()

	now := time.Now().UTC()
	po.ApprovedBy = approvedBy
	po.ApprovedDate = now
	po.UpdatedBy = approvedBy
	po.UpdatedDate = now

	if err := b.storer.Update(ctx, po); err != nil {
		return PurchaseOrder{}, fmt.Errorf("approve: %w", err)
	}

	return po, nil
}