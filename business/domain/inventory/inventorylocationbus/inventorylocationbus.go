package inventorylocationbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
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
	QueryByIDs(ctx context.Context, inventoryLocationIDs []uuid.UUID) ([]InventoryLocation, error)
}

// Business manages the set of APIs for brand access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
	outbox   *outbox.Writer
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
// WithOutbox returns a copy of the Business wired to the cascade outbox Writer.
// Inert until the Writer is injected at the F2 cutover (nil Writer -> Emit no-ops).
func (b *Business) WithOutbox(w *outbox.Writer) *Business {
	nb := *b
	nb.outbox = w
	return &nb
}

func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	nb := *b
	nb.storer = storer
	return &nb, nil
}

func (b *Business) Create(ctx context.Context, newInvLocation NewInventoryLocation) (InventoryLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.create")
	defer span.End()

	return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) (InventoryLocation, error) {
			now := time.Now()

			invLocation := InventoryLocation{
				LocationID:         uuid.New(),
				WarehouseID:        newInvLocation.WarehouseID,
				ZoneID:             newInvLocation.ZoneID,
				Aisle:              newInvLocation.Aisle,
				Bin:                newInvLocation.Bin,
				Rack:               newInvLocation.Rack,
				Shelf:              newInvLocation.Shelf,
				LocationCode:       newInvLocation.LocationCode,
				IsPickLocation:     newInvLocation.IsPickLocation,
				IsReserveLocation:  newInvLocation.IsReserveLocation,
				MaxCapacity:        newInvLocation.MaxCapacity,
				CurrentUtilization: newInvLocation.CurrentUtilization,
				UpdatedDate:        now,
				CreatedDate:        now,
			}

			if err := b.storer.Create(ctx, invLocation); err != nil {
				return InventoryLocation{}, fmt.Errorf("create: %w", err)
			}

			// Fire delegate event for workflow automation
			evtData := ActionCreatedData(invLocation)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return InventoryLocation{}, fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.delegate.Call(ctx, ActionCreatedData(invLocation)); err != nil {
				b.log.Error(ctx, "inventorylocationbus: delegate call failed", "action", ActionCreated, "err", err)
			}

			return invLocation, nil
		})
}

// Update modifies an inventory location in the system.
func (b *Business) Update(ctx context.Context, invLocation InventoryLocation, u UpdateInventoryLocation) (InventoryLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.update")
	defer span.End()

	return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) (InventoryLocation, error) {
			before := invLocation

			now := time.Now()

			if u.WarehouseID != nil {
				invLocation.WarehouseID = *u.WarehouseID
			}
			if u.ZoneID != nil {
				invLocation.ZoneID = *u.ZoneID
			}
			if u.Aisle != nil {
				invLocation.Aisle = *u.Aisle
			}
			if u.Rack != nil {
				invLocation.Rack = *u.Rack
			}
			if u.Shelf != nil {
				invLocation.Shelf = *u.Shelf
			}
			if u.Bin != nil {
				invLocation.Bin = *u.Bin
			}
			// INVARIANT: LocationCode is a scan-lookup key used by scanapp.Scan() to
			// resolve physical destinations on the warehouse floor. If physical
			// coordinates change (Aisle/Rack/Shelf/Bin) but LocationCode is not
			// explicitly updated or cleared, the existing code becomes stale — pointing
			// scanners to wrong coordinates, corrupting put-away and picking workflows.
			//
			// Rather than silently clearing LocationCode (destroys user data) or
			// silently preserving it (creates stale scan key), we reject the update
			// and force the caller to provide an explicit LocationCode value.
			//
			// History: commit 865580d8 added auto-clearing, commit 9defe77b removed it.
			// This validation replaces both approaches with an explicit rejection.
			coordinatesChanging := u.Aisle != nil || u.Rack != nil || u.Shelf != nil || u.Bin != nil
			if coordinatesChanging && invLocation.LocationCode != nil && u.LocationCode == nil {
				return InventoryLocation{}, fmt.Errorf("update: location_code must be provided or explicitly cleared when physical coordinates (aisle/rack/shelf/bin) change")
			}

			if u.LocationCode != nil {
				invLocation.LocationCode = u.LocationCode
			} else if u.Aisle != nil || u.Rack != nil || u.Shelf != nil || u.Bin != nil {
				invLocation.LocationCode = nil
			}
			if u.IsPickLocation != nil {
				invLocation.IsPickLocation = *u.IsPickLocation
			}
			if u.IsReserveLocation != nil {
				invLocation.IsReserveLocation = *u.IsReserveLocation
			}
			if u.MaxCapacity != nil {
				invLocation.MaxCapacity = *u.MaxCapacity
			}
			if u.CurrentUtilization != nil {
				invLocation.CurrentUtilization = *u.CurrentUtilization
			}

			invLocation.UpdatedDate = now

			if err := b.storer.Update(ctx, invLocation); err != nil {
				return InventoryLocation{}, fmt.Errorf("update: %w", err)
			}

			// Fire delegate event for workflow automation
			evtData := ActionUpdatedData(before, invLocation)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return InventoryLocation{}, fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.delegate.Call(ctx, ActionUpdatedData(before, invLocation)); err != nil {
				b.log.Error(ctx, "inventorylocationbus: delegate call failed", "action", ActionUpdated, "err", err)
			}

			return invLocation, nil
		})
}

// Delete removes an inventory location from the system.
func (b *Business) Delete(ctx context.Context, invLocation InventoryLocation) error {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.delete")
	defer span.End()

	return outbox.WriteAtomicVoid(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) error {
			if err := b.storer.Delete(ctx, invLocation); err != nil {
				return fmt.Errorf("delete: %w", err)
			}

			// Fire delegate event for workflow automation
			evtData := ActionDeletedData(invLocation)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.delegate.Call(ctx, ActionDeletedData(invLocation)); err != nil {
				b.log.Error(ctx, "inventorylocationbus: delegate call failed", "action", ActionDeleted, "err", err)
			}

			return nil
		})
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

// QueryByIDs retrieves the inventory locations for the specified IDs in a
// single query. Missing IDs are absent from the result rather than an error;
// callers that require every ID to resolve must check the returned length.
func (b *Business) QueryByIDs(ctx context.Context, inventoryLocationIDs []uuid.UUID) ([]InventoryLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorylocationbus.querybyids")
	defer span.End()

	invLocations, err := b.storer.QueryByIDs(ctx, inventoryLocationIDs)
	if err != nil {
		return nil, fmt.Errorf("querybyids: %w", err)
	}

	return invLocations, nil
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
