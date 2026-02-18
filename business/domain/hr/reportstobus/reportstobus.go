package reportstobus

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
	ErrNotFound              = errors.New("reports to not found")
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrUniqueEntry           = errors.New("reports to entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, tag ReportsTo) error
	Update(ctx context.Context, tag ReportsTo) error
	Delete(ctx context.Context, tag ReportsTo) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ReportsTo, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, tagID uuid.UUID) (ReportsTo, error)
}

// Business manages the set of APIs for reports to access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs an reports to business API for use.
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

// Create adds a new asset tag in the system.
func (b *Business) Create(ctx context.Context, nrt NewReportsTo) (ReportsTo, error) {
	ctx, span := otel.AddSpan(ctx, "business.reportstobus.Create")
	defer span.End()

	rt := ReportsTo{
		ID:         uuid.New(),
		ReporterID: nrt.ReporterID,
		BossID:     nrt.BossID,
	}

	if err := b.storer.Create(ctx, rt); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return ReportsTo{}, fmt.Errorf("create: %w", ErrUniqueEntry)
		}
		return ReportsTo{}, err
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(rt)); err != nil {
		b.log.Error(ctx, "reportstobus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return rt, nil
}

// Update updates an existing asset tag.
func (b *Business) Update(ctx context.Context, rt ReportsTo, urt UpdateReportsTo) (ReportsTo, error) {
	ctx, span := otel.AddSpan(ctx, "business.reportstobus.Update")
	defer span.End()

	before := rt

	if urt.BossID != nil {
		rt.BossID = *urt.BossID
	}

	if urt.ReporterID != nil {
		rt.ReporterID = *urt.ReporterID
	}

	if err := b.storer.Update(ctx, rt); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return ReportsTo{}, fmt.Errorf("update: %w", ErrUniqueEntry)
		}
		return ReportsTo{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, rt)); err != nil {
		b.log.Error(ctx, "reportstobus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return rt, nil
}

// Delete removes an asset tag from the system.
func (b *Business) Delete(ctx context.Context, at ReportsTo) error {
	ctx, span := otel.AddSpan(ctx, "business.reportstobus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, at); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(at)); err != nil {
		b.log.Error(ctx, "reportstobus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of existing asset tags from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ReportsTo, error) {
	ctx, span := otel.AddSpan(ctx, "business.reportstobus.Query")
	defer span.End()

	reportsTo, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return reportsTo, nil
}

// Count returns the total number of tags.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.reportstobus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the tag by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (ReportsTo, error) {
	ctx, span := otel.AddSpan(ctx, "business.reportstobus.QueryByID")
	defer span.End()

	reportsTo, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return ReportsTo{}, fmt.Errorf("query: tagID[%s]: %w", id, err)
	}

	return reportsTo, nil
}
