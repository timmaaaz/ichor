package citybus

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
	ErrNotFound              = errors.New("city not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("city entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, city City) error
	Update(ctx context.Context, city City) error
	Delete(ctx context.Context, city City) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]City, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, cityID uuid.UUID) (City, error)
}

// Business manages the set of APIs for city access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a city business API for use.
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

// Create adds a new city to the system.
func (b *Business) Create(ctx context.Context, nc NewCity) (City, error) {
	ctx, span := otel.AddSpan(ctx, "business.citybus.Create")
	defer span.End()

	cty := City{
		ID:       uuid.New(),
		RegionID: nc.RegionID,
		Name:     nc.Name,
	}

	if err := b.storer.Create(ctx, cty); err != nil {
		return City{}, fmt.Errorf("store create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(cty)); err != nil {
		b.log.Error(ctx, "citybus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return cty, nil
}

// Update modifies information about a city.
func (b *Business) Update(ctx context.Context, cty City, uc UpdateCity) (City, error) {
	ctx, span := otel.AddSpan(ctx, "business.citybus.Update")
	defer span.End()

	if uc.RegionID != nil {
		cty.RegionID = *uc.RegionID
	}

	if uc.Name != nil {
		cty.Name = *uc.Name
	}

	if err := b.storer.Update(ctx, cty); err != nil {
		return City{}, fmt.Errorf("store update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(cty)); err != nil {
		b.log.Error(ctx, "citybus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return cty, nil
}

// Delete removes a city from the system.
func (b *Business) Delete(ctx context.Context, cty City) error {
	ctx, span := otel.AddSpan(ctx, "business.citybus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, cty); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(cty)); err != nil {
		b.log.Error(ctx, "citybus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of existing cities.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]City, error) {
	ctx, span := otel.AddSpan(ctx, "business.citybus.Query")
	defer span.End()

	ctys, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return ctys, nil
}

// Count returns the total number of cities.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.citybus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the city by the specified ID.
func (b *Business) QueryByID(ctx context.Context, cityID uuid.UUID) (City, error) {
	ctx, span := otel.AddSpan(ctx, "business.citybus.QueryByID")
	defer span.End()

	cty, err := b.storer.QueryByID(ctx, cityID)
	if err != nil {
		return City{}, fmt.Errorf("query: cityID[%s]: %w", cityID, err)
	}

	return cty, nil
}
