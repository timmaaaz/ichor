package lotlocationbus

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
	ErrNotFound            = errors.New("lot location not found")
	ErrUniqueEntry         = errors.New("lot location entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, ll LotLocation) error
	Update(ctx context.Context, ll LotLocation) error
	Delete(ctx context.Context, ll LotLocation) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LotLocation, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, id uuid.UUID) (LotLocation, error)
}

// Business manages the set of APIs for lot location access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a lot location business API for use.
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

// Create adds a new lot location to the system.
func (b *Business) Create(ctx context.Context, nll NewLotLocation) (LotLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.lotlocationbus.create")
	defer span.End()

	now := time.Now()

	ll := LotLocation{
		ID:          uuid.New(),
		LotID:       nll.LotID,
		LocationID:  nll.LocationID,
		Quantity:    nll.Quantity,
		CreatedDate: now,
		UpdatedDate: now,
	}

	if err := b.storer.Create(ctx, ll); err != nil {
		return LotLocation{}, fmt.Errorf("create: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionCreatedData(ll)); err != nil {
		b.log.Error(ctx, "lotlocationbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return ll, nil
}

// Update modifies a lot location in the system.
func (b *Business) Update(ctx context.Context, ll LotLocation, ull UpdateLotLocation) (LotLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.lotlocationbus.update")
	defer span.End()

	before := ll

	if ull.LotID != nil {
		ll.LotID = *ull.LotID
	}
	if ull.LocationID != nil {
		ll.LocationID = *ull.LocationID
	}
	if ull.Quantity != nil {
		ll.Quantity = *ull.Quantity
	}

	ll.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, ll); err != nil {
		return LotLocation{}, fmt.Errorf("update: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, ll)); err != nil {
		b.log.Error(ctx, "lotlocationbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return ll, nil
}

// Delete removes a lot location from the system.
func (b *Business) Delete(ctx context.Context, ll LotLocation) error {
	ctx, span := otel.AddSpan(ctx, "business.lotlocationbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ll); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionDeletedData(ll)); err != nil {
		b.log.Error(ctx, "lotlocationbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of lot locations based on the provided query filter,
// order, and pagination options.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]LotLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.lotlocationbus.query")
	defer span.End()

	items, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return items, nil
}

// Count returns the total number of lot locations.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.lotlocationbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID retrieves a lot location by its unique ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (LotLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.lotlocationbus.querybyid")
	defer span.End()

	ll, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return LotLocation{}, fmt.Errorf("query by id: %w", err)
	}

	return ll, nil
}
