package purchaseorderlineitemstatusbus

import (
	"context"
	"errors"
	"fmt"

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
	ErrNotFound = errors.New("purchase order line item status not found")
	ErrUnique   = errors.New("not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, polis PurchaseOrderLineItemStatus) error
	Update(ctx context.Context, polis PurchaseOrderLineItemStatus) error
	Delete(ctx context.Context, polis PurchaseOrderLineItemStatus) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PurchaseOrderLineItemStatus, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, polisID uuid.UUID) (PurchaseOrderLineItemStatus, error)
	QueryByIDs(ctx context.Context, polisIDs []uuid.UUID) ([]PurchaseOrderLineItemStatus, error)
	QueryAll(ctx context.Context) ([]PurchaseOrderLineItemStatus, error)
}

// Business manages the set of APIs for purchase order line item status access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
}

// NewBusiness constructs a purchase order line item status business API for use.
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

// Create adds a new purchase order line item status to the system.
func (b *Business) Create(ctx context.Context, npolis NewPurchaseOrderLineItemStatus) (PurchaseOrderLineItemStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitemstatusbus.create")
	defer span.End()

	polis := PurchaseOrderLineItemStatus{
		ID:          uuid.New(),
		Name:        npolis.Name,
		Description: npolis.Description,
		SortOrder:   npolis.SortOrder,
	}

	if err := b.storer.Create(ctx, polis); err != nil {
		return PurchaseOrderLineItemStatus{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionCreatedData(polis)); err != nil {
		b.log.Error(ctx, "purchaseorderlineitemstatusbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return polis, nil
}

// Update modifies a purchase order line item status in the system.
func (b *Business) Update(ctx context.Context, polis PurchaseOrderLineItemStatus, upolis UpdatePurchaseOrderLineItemStatus) (PurchaseOrderLineItemStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitemstatusbus.update")
	defer span.End()

	if upolis.Name != nil {
		polis.Name = *upolis.Name
	}
	if upolis.Description != nil {
		polis.Description = *upolis.Description
	}
	if upolis.SortOrder != nil {
		polis.SortOrder = *upolis.SortOrder
	}

	if err := b.storer.Update(ctx, polis); err != nil {
		return PurchaseOrderLineItemStatus{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionUpdatedData(polis)); err != nil {
		b.log.Error(ctx, "purchaseorderlineitemstatusbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return polis, nil
}

// Delete removes a purchase order line item status from the system.
func (b *Business) Delete(ctx context.Context, polis PurchaseOrderLineItemStatus) error {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitemstatusbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, polis); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionDeletedData(polis)); err != nil {
		b.log.Error(ctx, "purchaseorderlineitemstatusbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of purchase order line item statuses from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PurchaseOrderLineItemStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitemstatusbus.query")
	defer span.End()

	statuses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return statuses, nil
}

// Count returns the total number of purchase order line item statuses.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitemstatusbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the purchase order line item status by the specified ID.
func (b *Business) QueryByID(ctx context.Context, polisID uuid.UUID) (PurchaseOrderLineItemStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitemstatusbus.querybyid")
	defer span.End()

	polis, err := b.storer.QueryByID(ctx, polisID)
	if err != nil {
		return PurchaseOrderLineItemStatus{}, fmt.Errorf("querybyid: %w", err)
	}

	return polis, nil
}

// QueryByIDs finds the purchase order line item statuses by the specified IDs.
func (b *Business) QueryByIDs(ctx context.Context, polisIDs []uuid.UUID) ([]PurchaseOrderLineItemStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitemstatusbus.querybyids")
	defer span.End()

	statuses, err := b.storer.QueryByIDs(ctx, polisIDs)
	if err != nil {
		return nil, fmt.Errorf("querybyids: %w", err)
	}

	return statuses, nil
}

// QueryAll retrieves all purchase order line item statuses from the system.
func (b *Business) QueryAll(ctx context.Context) ([]PurchaseOrderLineItemStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchaseorderlineitemstatusbus.queryall")
	defer span.End()

	statuses, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("queryall: %w", err)
	}

	return statuses, nil
}