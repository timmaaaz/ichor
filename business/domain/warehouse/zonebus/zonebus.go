package zonebus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("zone not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("zone entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, zone Zone) error
	Update(ctx context.Context, zone Zone) error
	Delete(ctx context.Context, zone Zone) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Zone, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, zoneID uuid.UUID) (Zone, error)
}

// Business manages the set of APIs for zone access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a zone business API for use.
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

// Create adds a new zone to the system.
func (b *Business) Create(ctx context.Context, nz NewZone) (Zone, error) {
	ctx, span := otel.AddSpan(ctx, "business.zone.create")
	defer span.End()

	now := time.Now()

	zone := Zone{
		ID:          uuid.New(),
		WarehouseID: nz.WarehouseID,
		Name:        nz.Name,
		Description: nz.Description,
		IsActive:    true,
		DateCreated: now,
		DateUpdated: now,
		CreatedBy:   nz.CreatedBy,
		UpdatedBy:   nz.CreatedBy,
	}

	if err := b.storer.Create(ctx, zone); err != nil {
		return Zone{}, fmt.Errorf("create: %w", err)
	}
	return zone, nil
}

// Update modifies a zone in the system.
func (b *Business) Update(ctx context.Context, bus Zone, uz UpdateZone) (Zone, error) {
	ctx, span := otel.AddSpan(ctx, "business.zone.update")
	defer span.End()

	err := convert.PopulateSameTypes(uz, &bus)
	if err != nil {
		return Zone{}, fmt.Errorf("populate zone from update zone: %w", err)
	}

	if err := b.storer.Update(ctx, bus); err != nil {
		return Zone{}, fmt.Errorf("update: %w", err)
	}
	return bus, nil
}

// Delete removes a zone from the system by its ID.
func (b *Business) Delete(ctx context.Context, bus Zone) error {
	ctx, span := otel.AddSpan(ctx, "business.zone.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, bus); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Count returns the number of zones in the system.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.zone.Count")
	defer span.End()
	return b.storer.Count(ctx, filter)
}

// Query retrieves a list of zones from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Zone, error) {
	ctx, span := otel.AddSpan(ctx, "business.zone.query")
	defer span.End()

	zones, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return zones, nil
}

// QueryByID retrieves a zone from the system by its ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Zone, error) {
	ctx, span := otel.AddSpan(ctx, "business.zone.querybyid")
	defer span.End()

	z, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Zone{}, fmt.Errorf("query by id: %w", err)
	}

	return z, nil
}
