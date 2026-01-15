package ordersbus

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
	ErrNotFound              = errors.New("order not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("order entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, newStatus Order) error
	Update(ctx context.Context, status Order) error
	Delete(ctx context.Context, status Order) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Order, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, statusID uuid.UUID) (Order, error)
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

func (b *Business) Create(ctx context.Context, no NewOrder) (Order, error) {

	ctx, span := otel.AddSpan(ctx, "business.ordersbus.create")
	defer span.End()

	now := time.Now().UTC()
	if no.CreatedDate != nil {
		now = *no.CreatedDate // Use provided date for seeding
	}

	order := Order{
		ID:                  uuid.New(),
		Number:              no.Number,
		CustomerID:          no.CustomerID,
		DueDate:             no.DueDate,
		FulfillmentStatusID: no.FulfillmentStatusID,
		OrderDate:           no.OrderDate,
		BillingAddressID:    no.BillingAddressID,
		ShippingAddressID:   no.ShippingAddressID,
		Subtotal:            no.Subtotal,
		TaxRate:             no.TaxRate,
		TaxAmount:           no.TaxAmount,
		ShippingCost:        no.ShippingCost,
		TotalAmount:         no.TotalAmount,
		Currency:            no.Currency,
		PaymentTermID:       no.PaymentTermID,
		Notes:               no.Notes,
		CreatedBy:           no.CreatedBy,
		UpdatedBy:           no.CreatedBy,
		CreatedDate:         now,
		UpdatedDate:         now,
	}

	if err := b.storer.Create(ctx, order); err != nil {
		return Order{}, err
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(order)); err != nil {
		b.log.Error(ctx, "ordersbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return order, nil
}

func (b *Business) Update(ctx context.Context, order Order, uo UpdateOrder) (Order, error) {
	ctx, span := otel.AddSpan(ctx, "business.ordersbus.update")
	defer span.End()

	if uo.Number != nil {
		order.Number = *uo.Number
	}
	if uo.CustomerID != nil {
		order.CustomerID = *uo.CustomerID
	}
	if uo.DueDate != nil {
		order.DueDate = *uo.DueDate
	}
	if uo.FulfillmentStatusID != nil {
		order.FulfillmentStatusID = *uo.FulfillmentStatusID
	}
	if uo.OrderDate != nil {
		order.OrderDate = *uo.OrderDate
	}
	if uo.BillingAddressID != nil {
		order.BillingAddressID = uo.BillingAddressID
	}
	if uo.ShippingAddressID != nil {
		order.ShippingAddressID = uo.ShippingAddressID
	}
	if uo.Subtotal != nil {
		order.Subtotal = *uo.Subtotal
	}
	if uo.TaxRate != nil {
		order.TaxRate = *uo.TaxRate
	}
	if uo.TaxAmount != nil {
		order.TaxAmount = *uo.TaxAmount
	}
	if uo.ShippingCost != nil {
		order.ShippingCost = *uo.ShippingCost
	}
	if uo.TotalAmount != nil {
		order.TotalAmount = *uo.TotalAmount
	}
	if uo.Currency != nil {
		order.Currency = *uo.Currency
	}
	if uo.PaymentTermID != nil {
		order.PaymentTermID = uo.PaymentTermID
	}
	if uo.Notes != nil {
		order.Notes = *uo.Notes
	}
	if uo.UpdatedBy != nil {
		order.UpdatedBy = *uo.UpdatedBy
	}

	order.UpdatedDate = time.Now().UTC()

	if err := b.storer.Update(ctx, order); err != nil {
		return Order{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(order)); err != nil {
		b.log.Error(ctx, "ordersbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return order, nil
}
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Order, error) {
	ctx, span := otel.AddSpan(ctx, "business.ordersbus.query")
	defer span.End()

	statuses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return statuses, nil
}

func (b *Business) Delete(ctx context.Context, order Order) error {
	ctx, span := otel.AddSpan(ctx, "business.ordersbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, order); err != nil {
		return err
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(order)); err != nil {
		b.log.Error(ctx, "ordersbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Count returns the total number of Orders.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.ordersbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the order by the specified ID.
func (b *Business) QueryByID(ctx context.Context, statusID uuid.UUID) (Order, error) {
	ctx, span := otel.AddSpan(ctx, "business.ordersbus.querybyid")
	defer span.End()

	result, err := b.storer.QueryByID(ctx, statusID)
	if err != nil {
		return Order{}, fmt.Errorf("queryByID: statusID: %w", err)
	}

	return result, nil
}
