package fulfillmentstatusbus

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
	ErrNotFound              = errors.New("fulfillment status not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("fulfillment status entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, fulfillmentStatus FulfillmentStatus) error
	Update(ctx context.Context, fulfillmentStatus FulfillmentStatus) error
	Delete(ctx context.Context, fulfillmentStatus FulfillmentStatus) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]FulfillmentStatus, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, fulfillmentStatusID uuid.UUID) (FulfillmentStatus, error)
}

type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a new fulfillment status business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
//
// This function seems like it could be implemented only once, but same with a lot of other pieces of this
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

// Create creates a new fulfillment status to the system.
func (b *Business) Create(ctx context.Context, nfs NewFulfillmentStatus) (FulfillmentStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.fulfillmentstatusbus.Create")
	defer span.End()

	fs := FulfillmentStatus{
		ID:             uuid.New(),
		Name:           nfs.Name,
		IconID:         nfs.IconID,
		PrimaryColor:   nfs.PrimaryColor,
		SecondaryColor: nfs.SecondaryColor,
		Icon:           nfs.Icon,
	}

	if err := b.storer.Create(ctx, fs); err != nil {
		return FulfillmentStatus{}, fmt.Errorf("store create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(fs)); err != nil {
		b.log.Error(ctx, "fulfillmentstatusbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return fs, nil
}

// Update modifies information about an fulfillment status.
func (b *Business) Update(ctx context.Context, fs FulfillmentStatus, ufs UpdateFulfillmentStatus) (FulfillmentStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.fulfillmentstatusbus.Update")
	defer span.End()

	before := fs

	if ufs.IconID != nil {
		fs.IconID = *ufs.IconID
	}

	if ufs.Name != nil {
		fs.Name = *ufs.Name
	}

	if ufs.PrimaryColor != nil {
		fs.PrimaryColor = *ufs.PrimaryColor
	}

	if ufs.SecondaryColor != nil {
		fs.SecondaryColor = *ufs.SecondaryColor
	}

	if ufs.Icon != nil {
		fs.Icon = *ufs.Icon
	}

	if err := b.storer.Update(ctx, fs); err != nil {
		return FulfillmentStatus{}, fmt.Errorf("store update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, fs)); err != nil {
		b.log.Error(ctx, "fulfillmentstatusbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return fs, nil
}

// Delete removes an fulfillment status from the system.
func (b *Business) Delete(ctx context.Context, fs FulfillmentStatus) error {
	ctx, span := otel.AddSpan(ctx, "business.fulfillmentstatusbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, fs); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(fs)); err != nil {
		b.log.Error(ctx, "fulfillmentstatusbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query returns a list of fulfillment statuses
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]FulfillmentStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.fulfillmentstatusbus.Query")
	defer span.End()

	fulfillmentStatuses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return fulfillmentStatuses, nil
}

// Count returns the total number of fulfillment statuses
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.fulfillmentstatusbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the fulfillment status by the specified ID.
func (b *Business) QueryByID(ctx context.Context, fulfillmentStatusID uuid.UUID) (FulfillmentStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.fulfillmentstatusbus.QueryByID")
	defer span.End()

	as, err := b.storer.QueryByID(ctx, fulfillmentStatusID)
	if err != nil {
		return FulfillmentStatus{}, fmt.Errorf("query: fulfillmentStatusID[%s]: %w", fulfillmentStatusID, err)
	}

	return as, nil
}
