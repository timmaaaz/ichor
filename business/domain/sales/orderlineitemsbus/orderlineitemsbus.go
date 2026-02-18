package orderlineitemsbus

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
	ErrNotFound              = errors.New("order line item not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("order line item entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, newStatus OrderLineItem) error
	Update(ctx context.Context, status OrderLineItem) error
	Delete(ctx context.Context, status OrderLineItem) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]OrderLineItem, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, statusID uuid.UUID) (OrderLineItem, error)
}

// Business manages the set of APIs for brand access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a brand business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// NewWithTx constructs a new Business value replacing the Storer
// value with a Storer value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
	}, nil
}

func (b *Business) Create(ctx context.Context, newStatus NewOrderLineItem) (OrderLineItem, error) {

	ctx, span := otel.AddSpan(ctx, "business.orderlineitemsbus.create")
	defer span.End()

	now := time.Now().UTC()
	if newStatus.CreatedDate != nil {
		now = *newStatus.CreatedDate // Use provided date for seeding
	}

	// Default discount_type to 'flat' if not specified
	discountType := newStatus.DiscountType
	if discountType == "" {
		discountType = "flat"
	}

	status := OrderLineItem{
		ID:                            uuid.New(),
		OrderID:                       newStatus.OrderID,
		ProductID:                     newStatus.ProductID,
		Description:                   newStatus.Description,
		Quantity:                      newStatus.Quantity,
		UnitPrice:                     newStatus.UnitPrice,
		Discount:                      newStatus.Discount,
		DiscountType:                  discountType,
		LineTotal:                     newStatus.LineTotal,
		LineItemFulfillmentStatusesID: newStatus.LineItemFulfillmentStatusesID,
		CreatedBy:                     newStatus.CreatedBy,
		UpdatedBy:                     newStatus.CreatedBy,
		CreatedDate:                   now,
		UpdatedDate:                   now,
	}

	if err := b.storer.Create(ctx, status); err != nil {
		return OrderLineItem{}, err
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(status)); err != nil {
		b.log.Error(ctx, "orderlineitemsbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return status, nil
}

func (b *Business) Update(ctx context.Context, status OrderLineItem, uStatus UpdateOrderLineItem) (OrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.orderlineitemsbus.update")
	defer span.End()

	before := status

	if uStatus.OrderID != nil {
		status.OrderID = *uStatus.OrderID
	}
	if uStatus.ProductID != nil {
		status.ProductID = *uStatus.ProductID
	}
	if uStatus.Description != nil {
		status.Description = *uStatus.Description
	}
	if uStatus.Quantity != nil {
		status.Quantity = *uStatus.Quantity
	}
	if uStatus.UnitPrice != nil {
		status.UnitPrice = *uStatus.UnitPrice
	}
	if uStatus.Discount != nil {
		status.Discount = *uStatus.Discount
	}
	if uStatus.DiscountType != nil {
		status.DiscountType = *uStatus.DiscountType
	}
	if uStatus.LineTotal != nil {
		status.LineTotal = *uStatus.LineTotal
	}
	if uStatus.LineItemFulfillmentStatusesID != nil {
		status.LineItemFulfillmentStatusesID = *uStatus.LineItemFulfillmentStatusesID
	}
	if uStatus.UpdatedBy != nil {
		status.UpdatedBy = *uStatus.UpdatedBy
	}

	status.UpdatedDate = time.Now().UTC()

	if err := b.storer.Update(ctx, status); err != nil {
		return OrderLineItem{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, status)); err != nil {
		b.log.Error(ctx, "orderlineitemsbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return status, nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]OrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.orderlineitemsbus.query")
	defer span.End()

	statuses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return statuses, nil
}

func (b *Business) Delete(ctx context.Context, status OrderLineItem) error {
	ctx, span := otel.AddSpan(ctx, "business.orderlineitemsbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, status); err != nil {
		return err
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(status)); err != nil {
		b.log.Error(ctx, "orderlineitemsbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Count returns the total number of order line itemes.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.orderlineitemsbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the order line item by the specified ID.
func (b *Business) QueryByID(ctx context.Context, statusID uuid.UUID) (OrderLineItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.orderlineitemsbus.querybyid")
	defer span.End()

	result, err := b.storer.QueryByID(ctx, statusID)
	if err != nil {
		return OrderLineItem{}, fmt.Errorf("queryByID: statusID: %w", err)
	}

	return result, nil
}
