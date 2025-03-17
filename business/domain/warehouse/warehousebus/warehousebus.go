package warehousebus

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
	ErrNotFound              = errors.New("warehouse not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("warehouse entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, warehouse Warehouse) error
	Update(ctx context.Context, warehouse Warehouse) error
	Delete(ctx context.Context, warehouse Warehouse) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Warehouse, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, warehouseID uuid.UUID) (Warehouse, error)
}

// Business manages the set of APIs for warehouse access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a warehouse business API for use.
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

// Create adds a new warehouse to the system.
func (b *Business) Create(ctx context.Context, nw NewWarehouse) (Warehouse, error) {
	ctx, span := otel.AddSpan(ctx, "business.warehouse.create")
	defer span.End()

	now := time.Now()

	warehouse := Warehouse{
		ID:          uuid.New(),
		StreetID:    nw.StreetID,
		Name:        nw.Name,
		IsActive:    true,
		CreatedBy:   nw.CreatedBy,
		UpdatedBy:   nw.CreatedBy,
		DateCreated: now,
		DateUpdated: now,
	}

	if err := b.storer.Create(ctx, warehouse); err != nil {
		return Warehouse{}, fmt.Errorf("create: %w", err)
	}

	return warehouse, nil
}

// Update modifies a warehouse in the system.
func (b *Business) Update(ctx context.Context, bus Warehouse, uw UpdateWarehouse) (Warehouse, error) {
	ctx, span := otel.AddSpan(ctx, "business.warehouse.update")
	defer span.End()

	err := convert.PopulateSameTypes(uw, &bus)
	if err != nil {
		return Warehouse{}, fmt.Errorf("populate warehouse from update warehouse: %w", err)
	}

	if err := b.storer.Update(ctx, bus); err != nil {
		return Warehouse{}, fmt.Errorf("update: %w", err)
	}

	return bus, nil
}

// Delete removes a warehouse from the system by its ID.
func (b *Business) Delete(ctx context.Context, bus Warehouse) error {
	ctx, span := otel.AddSpan(ctx, "business.warehouse.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, bus); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of warehouses from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Warehouse, error) {
	ctx, span := otel.AddSpan(ctx, "business.warehouse.query")
	defer span.End()

	warehouses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return warehouses, nil
}

// Count returns the number of warehouses in the system.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.warehouse.Count")
	defer span.End()
	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a warehouse from the system by its ID.
func (b *Business) QueryByID(ctx context.Context, warehouseID uuid.UUID) (Warehouse, error) {
	ctx, span := otel.AddSpan(ctx, "business.warehouse.querybyid")
	defer span.End()

	warehouse, err := b.storer.QueryByID(ctx, warehouseID)
	if err != nil {
		return Warehouse{}, fmt.Errorf("query by id: %w", err)
	}

	return warehouse, nil
}
