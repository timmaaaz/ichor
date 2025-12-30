package zonebus

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
	ErrNotFound              = errors.New("zone not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("zone entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
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

func (b *Business) Create(ctx context.Context, nz NewZone) (Zone, error) {
	ctx, span := otel.AddSpan(ctx, "business.zonesbus.create")
	defer span.End()

	now := time.Now()
	zone := Zone{
		ZoneID:      uuid.New(),
		WarehouseID: nz.WarehouseID,
		Name:        nz.Name,
		Description: nz.Description,
		CreatedDate: now,
		UpdatedDate: now,
	}

	if err := b.storer.Create(ctx, zone); err != nil {
		return Zone{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(zone)); err != nil {
		b.log.Error(ctx, "zonebus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return zone, nil
}

// Update modifies a zone in the system.

func (b *Business) Update(ctx context.Context, zone Zone, u UpdateZone) (Zone, error) {
	ctx, span := otel.AddSpan(ctx, "business.zonesbus.update")
	defer span.End()

	if u.Name != nil {
		zone.Name = *u.Name
	}
	if u.Description != nil {
		zone.Description = *u.Description
	}

	if u.WarehouseID != nil {
		zone.WarehouseID = *u.WarehouseID
	}

	zone.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, zone); err != nil {
		return Zone{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(zone)); err != nil {
		b.log.Error(ctx, "zonebus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return zone, nil
}

func (b *Business) Delete(ctx context.Context, zone Zone) error {
	ctx, span := otel.AddSpan(ctx, "business.zonebus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, zone); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(zone)); err != nil {
		b.log.Error(ctx, "zonebus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Zone, error) {
	ctx, span := otel.AddSpan(ctx, "business.zonesbus.query")
	defer span.End()

	zones, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying zones: %w", err)
	}

	return zones, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.zonesbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

func (b *Business) QueryByID(ctx context.Context, zoneID uuid.UUID) (Zone, error) {
	ctx, span := otel.AddSpan(ctx, "business.zonesbus.querybyid")
	defer span.End()

	zone, err := b.storer.QueryByID(ctx, zoneID)
	if err != nil {
		return Zone{}, fmt.Errorf("querybyID [zone]: %w", err)
	}
	return zone, nil
}
