package orderfulfillmentstatusbus

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
	ErrNotFound              = errors.New("order fulfillment status not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("order fulfillment status entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, newStatus OrderFulfillmentStatus) error
	Update(ctx context.Context, status OrderFulfillmentStatus) error
	Delete(ctx context.Context, status OrderFulfillmentStatus) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]OrderFulfillmentStatus, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, statusID uuid.UUID) (OrderFulfillmentStatus, error)
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

func (b *Business) Create(ctx context.Context, newStatus NewOrderFulfillmentStatus) (OrderFulfillmentStatus, error) {

	ctx, span := otel.AddSpan(ctx, "business.orderfulfillmentstatusbus.create")
	defer span.End()

	status := OrderFulfillmentStatus{
		ID:             uuid.New(),
		Name:           newStatus.Name,
		Description:    newStatus.Description,
		PrimaryColor:   newStatus.PrimaryColor,
		SecondaryColor: newStatus.SecondaryColor,
		Icon:           newStatus.Icon,
	}

	if err := b.storer.Create(ctx, status); err != nil {
		return OrderFulfillmentStatus{}, err
	}
	return status, nil
}

func (b *Business) Update(ctx context.Context, status OrderFulfillmentStatus, uStatus UpdateOrderFulfillmentStatus) (OrderFulfillmentStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.orderfulfillmentstatusbus.update")
	defer span.End()

	if uStatus.Name != nil {
		status.Name = *uStatus.Name
	}
	if uStatus.Description != nil {
		status.Description = *uStatus.Description
	}
	if uStatus.PrimaryColor != nil {
		status.PrimaryColor = *uStatus.PrimaryColor
	}
	if uStatus.SecondaryColor != nil {
		status.SecondaryColor = *uStatus.SecondaryColor
	}
	if uStatus.Icon != nil {
		status.Icon = *uStatus.Icon
	}

	if err := b.storer.Update(ctx, status); err != nil {
		return OrderFulfillmentStatus{}, fmt.Errorf("update: %w", err)
	}

	return status, nil
}
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]OrderFulfillmentStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.orderfulfillmentstatusbus.query")
	defer span.End()

	statuses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return statuses, nil
}

func (b *Business) Delete(ctx context.Context, status OrderFulfillmentStatus) error {
	ctx, span := otel.AddSpan(ctx, "business.orderfulfillmentstatusbus.delete")
	defer span.End()

	return b.storer.Delete(ctx, status)
}

// Count returns the total number of order fulfillment statuses.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.orderfulfillmentstatusbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the order fulfillment status by the specified ID.
func (b *Business) QueryByID(ctx context.Context, statusID uuid.UUID) (OrderFulfillmentStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.orderfulfillmentstatusbus.querybyid")
	defer span.End()

	result, err := b.storer.QueryByID(ctx, statusID)
	if err != nil {
		return OrderFulfillmentStatus{}, fmt.Errorf("queryByID: statusID: %w", err)
	}

	return result, nil
}
