package purchaseorderlineitembus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("purchase order line item not found")
	ErrUnique   = errors.New("not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, poli PurchaseOrderLineItem) error
	Update(ctx context.Context, poli PurchaseOrderLineItem) error
	Delete(ctx context.Context, poli PurchaseOrderLineItem) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PurchaseOrderLineItem, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, poliID uuid.UUID) (PurchaseOrderLineItem, error)
	QueryByIDs(ctx context.Context, poliIDs []uuid.UUID) ([]PurchaseOrderLineItem, error)
	QueryByPurchaseOrderID(ctx context.Context, poID uuid.UUID) ([]PurchaseOrderLineItem, error)
}

// Business manages the set of APIs for purchase order line item access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
}

// NewBusiness constructs a purchase order line item business API for use.
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

// Create adds a new purchase order line item to the system.
func (b *Business) Create(ctx context.Context, npoli NewPurchaseOrderLineItem) (PurchaseOrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.create")
	defer span.End()

	now := time.Now().UTC()

	poli := PurchaseOrderLineItem{
		ID:                   uuid.New(),
		PurchaseOrderID:      npoli.PurchaseOrderID,
		SupplierProductID:    npoli.SupplierProductID,
		QuantityOrdered:      npoli.QuantityOrdered,
		QuantityReceived:     0,
		QuantityCancelled:    0,
		UnitCost:             npoli.UnitCost,
		Discount:             npoli.Discount,
		LineTotal:            npoli.LineTotal,
		LineItemStatusID:     npoli.LineItemStatusID,
		ExpectedDeliveryDate: npoli.ExpectedDeliveryDate,
		Notes:                npoli.Notes,
		CreatedBy:            npoli.CreatedBy,
		UpdatedBy:            npoli.CreatedBy,
		CreatedDate:          now,
		UpdatedDate:          now,
	}

	if err := b.storer.Create(ctx, poli); err != nil {
		return PurchaseOrderLineItem{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionCreatedData(poli)); err != nil {
		b.log.Error(ctx, "purchaseorderlineitembus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return poli, nil
}

// Update modifies a purchase order line item in the system.
func (b *Business) Update(ctx context.Context, poli PurchaseOrderLineItem, upoli UpdatePurchaseOrderLineItem) (PurchaseOrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.update")
	defer span.End()

	if upoli.SupplierProductID != nil {
		poli.SupplierProductID = *upoli.SupplierProductID
	}
	if upoli.QuantityOrdered != nil {
		poli.QuantityOrdered = *upoli.QuantityOrdered
	}
	if upoli.QuantityReceived != nil {
		poli.QuantityReceived = *upoli.QuantityReceived
	}
	if upoli.QuantityCancelled != nil {
		poli.QuantityCancelled = *upoli.QuantityCancelled
	}
	if upoli.UnitCost != nil {
		poli.UnitCost = *upoli.UnitCost
	}
	if upoli.Discount != nil {
		poli.Discount = *upoli.Discount
	}
	if upoli.LineTotal != nil {
		poli.LineTotal = *upoli.LineTotal
	}
	if upoli.LineItemStatusID != nil {
		poli.LineItemStatusID = *upoli.LineItemStatusID
	}
	if upoli.ExpectedDeliveryDate != nil {
		poli.ExpectedDeliveryDate = *upoli.ExpectedDeliveryDate
	}
	if upoli.ActualDeliveryDate != nil {
		poli.ActualDeliveryDate = *upoli.ActualDeliveryDate
	}
	if upoli.Notes != nil {
		poli.Notes = *upoli.Notes
	}
	if upoli.UpdatedBy != nil {
		poli.UpdatedBy = *upoli.UpdatedBy
	}

	poli.UpdatedDate = time.Now().UTC()

	if err := b.storer.Update(ctx, poli); err != nil {
		return PurchaseOrderLineItem{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionUpdatedData(poli)); err != nil {
		b.log.Error(ctx, "purchaseorderlineitembus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return poli, nil
}

// Delete removes a purchase order line item from the system.
func (b *Business) Delete(ctx context.Context, poli PurchaseOrderLineItem) error {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, poli); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionDeletedData(poli)); err != nil {
		b.log.Error(ctx, "purchaseorderlineitembus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of purchase order line items from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PurchaseOrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.query")
	defer span.End()

	items, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return items, nil
}

// Count returns the total number of purchase order line items.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the purchase order line item by the specified ID.
func (b *Business) QueryByID(ctx context.Context, poliID uuid.UUID) (PurchaseOrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.querybyid")
	defer span.End()

	poli, err := b.storer.QueryByID(ctx, poliID)
	if err != nil {
		return PurchaseOrderLineItem{}, fmt.Errorf("querybyid: %w", err)
	}

	return poli, nil
}

// QueryByIDs finds the purchase order line items by the specified IDs.
func (b *Business) QueryByIDs(ctx context.Context, poliIDs []uuid.UUID) ([]PurchaseOrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.querybyids")
	defer span.End()

	items, err := b.storer.QueryByIDs(ctx, poliIDs)
	if err != nil {
		return nil, fmt.Errorf("querybyids: %w", err)
	}

	return items, nil
}

// QueryByPurchaseOrderID finds all line items for a specific purchase order.
func (b *Business) QueryByPurchaseOrderID(ctx context.Context, poID uuid.UUID) ([]PurchaseOrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.querybypurchaseorderid")
	defer span.End()

	items, err := b.storer.QueryByPurchaseOrderID(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("querybypurchaseorderid: %w", err)
	}

	return items, nil
}

// ReceiveQuantity updates the received quantity for a line item.
func (b *Business) ReceiveQuantity(ctx context.Context, poli PurchaseOrderLineItem, quantity int, receivedBy uuid.UUID) (PurchaseOrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitembus.receivequantity")
	defer span.End()

	poli.QuantityReceived += quantity
	poli.UpdatedBy = receivedBy
	poli.UpdatedDate = time.Now().UTC()

	if err := b.storer.Update(ctx, poli); err != nil {
		return PurchaseOrderLineItem{}, fmt.Errorf("receivequantity: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionUpdatedData(poli)); err != nil {
		b.log.Error(ctx, "purchaseorderlineitembus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return poli, nil
}