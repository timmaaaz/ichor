package officebus

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
	ErrNotFound              = errors.New("office not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("office entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, office Office) error
	Update(ctx context.Context, office Office) error
	Delete(ctx context.Context, office Office) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Office, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, officeID uuid.UUID) (Office, error)
}

// Business manages the set of APIs for office access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a office business API for use.
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

// Create adds a new asset type to the system.
func (b *Business) Create(ctx context.Context, no NewOffice) (Office, error) {
	ctx, span := otel.AddSpan(ctx, "business.officebus.Create")
	defer span.End()

	o := Office{
		ID:       uuid.New(),
		Name:     no.Name,
		StreetID: no.StreetID,
	}

	if err := b.storer.Create(ctx, o); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return Office{}, fmt.Errorf("create: %w", ErrUniqueEntry)
		}
		return Office{}, err
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(o)); err != nil {
		b.log.Error(ctx, "officebus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return o, nil
}

// Update updates an existing asset type.
func (b *Business) Update(ctx context.Context, o Office, uo UpdateOffice) (Office, error) {
	ctx, span := otel.AddSpan(ctx, "business.officebus.Update")
	defer span.End()

	if uo.Name != nil {
		o.Name = *uo.Name
	}

	if uo.StreetID != nil {
		o.StreetID = *uo.StreetID
	}

	if err := b.storer.Update(ctx, o); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return Office{}, fmt.Errorf("update: %w", ErrUniqueEntry)
		}
		return Office{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(o)); err != nil {
		b.log.Error(ctx, "officebus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return o, nil
}

// Delete removes an asset type from the system.
func (b *Business) Delete(ctx context.Context, at Office) error {
	ctx, span := otel.AddSpan(ctx, "business.officebus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, at); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(at)); err != nil {
		b.log.Error(ctx, "officebus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of existing asset types from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Office, error) {
	ctx, span := otel.AddSpan(ctx, "business.officebus.Query")
	defer span.End()

	offices, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return offices, nil
}

// Count returns the total number of asset types.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.officebus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the asset type by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Office, error) {
	ctx, span := otel.AddSpan(ctx, "business.officebus.QueryByID")
	defer span.End()

	office, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Office{}, fmt.Errorf("query: officeID[%s]: %w", id, err)
	}

	return office, nil
}
