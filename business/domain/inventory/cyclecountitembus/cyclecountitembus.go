package cyclecountitembus

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
	ErrNotFound            = errors.New("cycle count item not found")
	ErrUniqueEntry         = errors.New("cycle count item entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, item CycleCountItem) error
	Update(ctx context.Context, item CycleCountItem) error
	Delete(ctx context.Context, item CycleCountItem) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CycleCountItem, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, itemID uuid.UUID) (CycleCountItem, error)
}

// Business manages the set of APIs for cycle count item access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a cycle count item business API for use.
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
		storer:   storer,
		delegate: b.delegate,
	}, nil
}

// Create adds a new cycle count item to the system.
func (b *Business) Create(ctx context.Context, ncci NewCycleCountItem) (CycleCountItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.create")
	defer span.End()

	now := time.Now()

	item := CycleCountItem{
		ID:             uuid.New(),
		SessionID:      ncci.SessionID,
		ProductID:      ncci.ProductID,
		LocationID:     ncci.LocationID,
		SystemQuantity: ncci.SystemQuantity,
		Status:         Statuses.Pending,
		CreatedDate:    now,
		UpdatedDate:    now,
	}

	if err := b.storer.Create(ctx, item); err != nil {
		return CycleCountItem{}, fmt.Errorf("create: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionCreatedData(item)); err != nil {
		b.log.Error(ctx, "cyclecountitembus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return item, nil
}

// Update modifies an existing cycle count item in the system.
// When CountedQuantity is provided, Variance is automatically computed.
func (b *Business) Update(ctx context.Context, item CycleCountItem, ucci UpdateCycleCountItem) (CycleCountItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.update")
	defer span.End()

	before := item

	if ucci.CountedQuantity != nil {
		item.CountedQuantity = ucci.CountedQuantity
		variance := *ucci.CountedQuantity - item.SystemQuantity
		item.Variance = &variance
	}
	if ucci.Status != nil {
		item.Status = *ucci.Status
	}
	if ucci.CountedBy != nil {
		item.CountedBy = *ucci.CountedBy
	}
	if ucci.CountedDate != nil {
		item.CountedDate = *ucci.CountedDate
	}

	item.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, item); err != nil {
		return CycleCountItem{}, fmt.Errorf("update: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, item)); err != nil {
		b.log.Error(ctx, "cyclecountitembus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return item, nil
}

// Delete removes a cycle count item from the system.
func (b *Business) Delete(ctx context.Context, item CycleCountItem) error {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, item); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionDeletedData(item)); err != nil {
		b.log.Error(ctx, "cyclecountitembus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of cycle count items from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CycleCountItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.query")
	defer span.End()

	items, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return items, nil
}

// Count returns the total number of cycle count items matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a single cycle count item by its ID.
func (b *Business) QueryByID(ctx context.Context, itemID uuid.UUID) (CycleCountItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.querybyid")
	defer span.End()

	item, err := b.storer.QueryByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return CycleCountItem{}, err
		}
		return CycleCountItem{}, fmt.Errorf("queryByID: itemID[%s]: %w", itemID, err)
	}

	return item, nil
}
