package purchaseorderstatusbus

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
	ErrNotFound = errors.New("purchase order status not found")
	ErrUnique   = errors.New("not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, pos PurchaseOrderStatus) error
	Update(ctx context.Context, pos PurchaseOrderStatus) error
	Delete(ctx context.Context, pos PurchaseOrderStatus) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PurchaseOrderStatus, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, posID uuid.UUID) (PurchaseOrderStatus, error)
	QueryByIDs(ctx context.Context, posIDs []uuid.UUID) ([]PurchaseOrderStatus, error)
	QueryAll(ctx context.Context) ([]PurchaseOrderStatus, error)
}

// Business manages the set of APIs for purchase order status access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
}

// NewBusiness constructs a purchase order status business API for use.
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

// Create adds a new purchase order status to the system.
func (b *Business) Create(ctx context.Context, npos NewPurchaseOrderStatus) (PurchaseOrderStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchase_order_statusbus.create")
	defer span.End()

	pos := PurchaseOrderStatus{
		ID:          uuid.New(),
		Name:        npos.Name,
		Description: npos.Description,
		SortOrder:   npos.SortOrder,
	}

	if err := b.storer.Create(ctx, pos); err != nil {
		return PurchaseOrderStatus{}, fmt.Errorf("create: %w", err)
	}

	return pos, nil
}

// Update modifies a purchase order status in the system.
func (b *Business) Update(ctx context.Context, pos PurchaseOrderStatus, upos UpdatePurchaseOrderStatus) (PurchaseOrderStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchase_order_statusbus.update")
	defer span.End()

	if upos.Name != nil {
		pos.Name = *upos.Name
	}
	if upos.Description != nil {
		pos.Description = *upos.Description
	}
	if upos.SortOrder != nil {
		pos.SortOrder = *upos.SortOrder
	}

	if err := b.storer.Update(ctx, pos); err != nil {
		return PurchaseOrderStatus{}, fmt.Errorf("update: %w", err)
	}

	return pos, nil
}

// Delete removes a purchase order status from the system.
func (b *Business) Delete(ctx context.Context, pos PurchaseOrderStatus) error {
	ctx, span := otel.AddSpan(ctx, "business.purchase_order_statusbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, pos); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of purchase order statuses from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PurchaseOrderStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchase_order_statusbus.query")
	defer span.End()

	statuses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return statuses, nil
}

// Count returns the total number of purchase order statuses.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchase_order_statusbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the purchase order status by the specified ID.
func (b *Business) QueryByID(ctx context.Context, posID uuid.UUID) (PurchaseOrderStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchase_order_statusbus.querybyid")
	defer span.End()

	pos, err := b.storer.QueryByID(ctx, posID)
	if err != nil {
		return PurchaseOrderStatus{}, fmt.Errorf("querybyid: %w", err)
	}

	return pos, nil
}

// QueryByIDs finds the purchase order statuses by the specified IDs.
func (b *Business) QueryByIDs(ctx context.Context, posIDs []uuid.UUID) ([]PurchaseOrderStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchase_order_statusbus.querybyids")
	defer span.End()

	statuses, err := b.storer.QueryByIDs(ctx, posIDs)
	if err != nil {
		return nil, fmt.Errorf("querybyids: %w", err)
	}

	return statuses, nil
}

// QueryAll retrieves all purchase order statuses from the system.
func (b *Business) QueryAll(ctx context.Context) ([]PurchaseOrderStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.purchase_order_statusbus.queryall")
	defer span.End()

	statuses, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("queryall: %w", err)
	}

	return statuses, nil
}