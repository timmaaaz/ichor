package paymenttermbus

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
	ErrNotFound              = errors.New("payment term not found")
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrUniqueEntry           = errors.New("payment term entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, paymentTerm PaymentTerm) error
	Update(ctx context.Context, paymentTerm PaymentTerm) error
	Delete(ctx context.Context, paymentTerm PaymentTerm) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PaymentTerm, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, paymentTermID uuid.UUID) (PaymentTerm, error)
}

// Business manages the set of APIs for payment term access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a payment term business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
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
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
	}

	return &bus, nil
}

// Create adds a new payment term to the system.
func (b *Business) Create(ctx context.Context, npt NewPaymentTerm) (PaymentTerm, error) {
	ctx, span := otel.AddSpan(ctx, "business.paymenttermbus.Create")
	defer span.End()

	pt := PaymentTerm{
		ID:          uuid.New(),
		Name:        npt.Name,
		Description: npt.Description,
	}

	if err := b.storer.Create(ctx, pt); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return PaymentTerm{}, fmt.Errorf("create: %w", ErrUniqueEntry)
		}
		return PaymentTerm{}, err
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(pt)); err != nil {
		b.log.Error(ctx, "paymenttermbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return pt, nil
}

// Update updates an existing payment term.
func (b *Business) Update(ctx context.Context, pt PaymentTerm, upt UpdatePaymentTerm) (PaymentTerm, error) {
	ctx, span := otel.AddSpan(ctx, "business.paymenttermbus.Update")
	defer span.End()

	if upt.Name != nil {
		pt.Name = *upt.Name
	}

	if upt.Description != nil {
		pt.Description = *upt.Description
	}

	if err := b.storer.Update(ctx, pt); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return PaymentTerm{}, fmt.Errorf("update: %w", ErrUniqueEntry)
		}
		return PaymentTerm{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(pt)); err != nil {
		b.log.Error(ctx, "paymenttermbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return pt, nil
}

// Delete removes a payment term from the system.
func (b *Business) Delete(ctx context.Context, pt PaymentTerm) error {
	ctx, span := otel.AddSpan(ctx, "business.paymenttermbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, pt); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(pt)); err != nil {
		b.log.Error(ctx, "paymenttermbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of existing payment terms from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PaymentTerm, error) {
	ctx, span := otel.AddSpan(ctx, "business.paymenttermbus.Query")
	defer span.End()

	paymentTerms, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return paymentTerms, nil
}

// Count returns the total number of payment terms.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.paymenttermbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the payment term by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (PaymentTerm, error) {
	ctx, span := otel.AddSpan(ctx, "business.paymenttermbus.QueryByID")
	defer span.End()

	paymentTerm, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return PaymentTerm{}, fmt.Errorf("query: paymentTermID[%s]: %w", id, err)
	}

	return paymentTerm, nil
}
