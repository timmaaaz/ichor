package inventorylocationbus

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
	ErrNotFound              = errors.New("inventoryLocation not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("inventoryLocation entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, inventoryLocation InventoryLocation) error
	Update(ctx context.Context, inventoryLocation InventoryLocation) error
	Delete(ctx context.Context, inventoryLocation InventoryLocation) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryLocation, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, inventoryLocationID uuid.UUID) (InventoryLocation, error)
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

func (b *Business) Create(ctx context.Context, newInvLocation NewInventoryLocation) (InventoryLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.create")
	defer span.End()

	now := time.Now()

	invLocation := InventoryLocation{
		LocationID:         uuid.New(),
		WarehouseID:        newInvLocation.WarehouseID,
		ZoneID:             newInvLocation.ZoneID,
		Aisle:              newInvLocation.Aisle,
		Bin:                newInvLocation.Bin,
		Rack:               newInvLocation.Rack,
		Shelf:              newInvLocation.Shelf,
		IsPickLocation:     newInvLocation.IsPickLocation,
		IsReserveLocation:  newInvLocation.IsReserveLocation,
		MaxCapacity:        newInvLocation.MaxCapacity,
		CurrentUtilization: newInvLocation.CurrentUtilization,
		UpdatedDate:        now,
		CreatedDate:        now,
	}

	err := b.storer.Create(ctx, invLocation)
	if err != nil {
		return InventoryLocation{}, fmt.Errorf("create: %w", err)
	}

	return invLocation, nil

}

// Update modifies an inventory location in the system.
func (b *Business) Update(ctx context.Context, invLocation InventoryLocation, u UpdateInventoryLocation) (InventoryLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.update")
	defer span.End()

	now := time.Now()

	err := convert.PopulateSameTypes(u, &invLocation)
	if err != nil {
		return InventoryLocation{}, fmt.Errorf("update: %w", err)
	}

	if u.CurrentUtilization != nil {
		invLocation.CurrentUtilization = *u.CurrentUtilization
	}

	invLocation.UpdatedDate = now

	err = b.storer.Update(ctx, invLocation)
	if err != nil {
		return InventoryLocation{}, fmt.Errorf("update: %w", err)
	}

	return invLocation, nil
}

// Delete removes an inventory location from the system.
func (b *Business) Delete(ctx context.Context, invLocation InventoryLocation) error {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.delete")
	defer span.End()

	err := b.storer.Delete(ctx, invLocation)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves inventory locations based on the given query filter.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.query")
	defer span.End()

	inventoryLocations, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return inventoryLocations, nil
}

// Count returns the total number of inventory locations.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return count, nil
}

// QueryByID retrieves an inventory location by its ID.
func (b *Business) QueryByID(ctx context.Context, inventoryLocationID uuid.UUID) (InventoryLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.querybyid")
	defer span.End()

	invLocation, err := b.storer.QueryByID(ctx, inventoryLocationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return InventoryLocation{}, err
		}
		return InventoryLocation{}, fmt.Errorf("queryByID [inventoryLocation]: %w", err)
	}

	return invLocation, nil
}
