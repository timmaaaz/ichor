package streetbus

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
	ErrNotFound              = errors.New("street not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("street entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, street Street) error
	Update(ctx context.Context, street Street) error
	Delete(ctx context.Context, street Street) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Street, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, streetID uuid.UUID) (Street, error)
}

// Business manages the set of APIs for street access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a street business API for use.
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

// Create adds a new street to the system.
func (b *Business) Create(ctx context.Context, ns NewStreet) (Street, error) {
	ctx, span := otel.AddSpan(ctx, "business.streetbus.Create")
	defer span.End()

	str := Street{
		ID:         uuid.New(),
		CityID:     ns.CityID,
		Line1:      ns.Line1,
		Line2:      ns.Line2,
		PostalCode: ns.PostalCode,
	}

	if err := b.storer.Create(ctx, str); err != nil {
		return Street{}, fmt.Errorf("store create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(str)); err != nil {
		b.log.Error(ctx, "streetbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return str, nil
}

// Update modifies data about a street.
func (b *Business) Update(ctx context.Context, str Street, us UpdateStreet) (Street, error) {
	ctx, span := otel.AddSpan(ctx, "business.streetbus.Update")
	defer span.End()

	if us.CityID != nil {
		str.CityID = *us.CityID
	}

	if us.Line1 != nil {
		str.Line1 = *us.Line1
	}

	if us.Line2 != nil {
		str.Line2 = *us.Line2
	}

	if us.PostalCode != nil {
		str.PostalCode = *us.PostalCode
	}

	if err := b.storer.Update(ctx, str); err != nil {
		return Street{}, fmt.Errorf("store update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(str)); err != nil {
		b.log.Error(ctx, "streetbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return str, nil
}

// Delete removes a street from the system.
func (b *Business) Delete(ctx context.Context, str Street) error {
	ctx, span := otel.AddSpan(ctx, "business.streetbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, str); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(str)); err != nil {
		b.log.Error(ctx, "streetbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of streets from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Street, error) {
	ctx, span := otel.AddSpan(ctx, "business.streetbus.Query")
	defer span.End()

	strs, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return strs, nil
}

// Count returns the total number of cities.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.streetbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the street by the specified ID.
func (b *Business) QueryByID(ctx context.Context, streetID uuid.UUID) (Street, error) {
	ctx, span := otel.AddSpan(ctx, "business.streetbus.QueryByID")
	defer span.End()

	str, err := b.storer.QueryByID(ctx, streetID)
	if err != nil {
		return Street{}, fmt.Errorf("query: streetID[%s]: %w", streetID, err)
	}

	return str, nil
}
