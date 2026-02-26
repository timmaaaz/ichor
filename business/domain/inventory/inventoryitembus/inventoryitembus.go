package inventoryitembus

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

var (
	ErrNotFound              = errors.New("inventoryItem not found")
	ErrUniqueEntry           = errors.New("inventoryItem entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, inventoryItem InventoryItem) error
	Update(ctx context.Context, inventoryItem InventoryItem) error
	Delete(ctx context.Context, inventoryItem InventoryItem) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryItem, error)
	QueryWithLocationDetails(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryItemWithLocation, error)
	QueryItemsWithProductAtLocation(ctx context.Context, locationID uuid.UUID) ([]ItemWithProduct, error)
	QueryAvailableForAllocation(ctx context.Context, productID uuid.UUID, locationID *uuid.UUID, warehouseID *uuid.UUID, strategy string, limit int) ([]InventoryItem, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, inventoryID uuid.UUID) (InventoryItem, error)
	UpsertQuantity(ctx context.Context, newID, productID, locationID uuid.UUID, quantityDelta int) error
}

type Business struct {
	storer   Storer
	delegate *delegate.Delegate
	log      *logger.Logger
}

func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		storer:   storer,
		delegate: delegate,
		log:      log,
	}
}

func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		storer:   storer,
		delegate: b.delegate,
		log:      b.log,
	}, nil
}

// Create inserts a new inventoryItem into the database.
func (b *Business) Create(ctx context.Context, nip NewInventoryItem) (InventoryItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.create")
	defer span.End()

	now := time.Now()

	inventoryItem := InventoryItem{
		ID:                    uuid.New(),
		ProductID:             nip.ProductID,
		LocationID:            nip.LocationID,
		Quantity:              nip.Quantity,
		ReservedQuantity:      nip.ReservedQuantity,
		AllocatedQuantity:     nip.AllocatedQuantity,
		AvgDailyUsage:         nip.AvgDailyUsage,
		SafetyStock:           nip.SafetyStock,
		MinimumStock:          nip.MinimumStock,
		MaximumStock:          nip.MaximumStock,
		ReorderPoint:          nip.ReorderPoint,
		EconomicOrderQuantity: nip.EconomicOrderQuantity,
		CreatedDate:           now,
		UpdatedDate:           now,
	}

	if err := b.storer.Create(ctx, inventoryItem); err != nil {
		return InventoryItem{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(inventoryItem)); err != nil {
		b.log.Error(ctx, "inventoryitembus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return inventoryItem, nil
}

// Update modifies an inventoryItem in the system.
func (b *Business) Update(ctx context.Context, ip InventoryItem, up UpdateInventoryItem) (InventoryItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.update")
	defer span.End()

	before := ip

	if up.ProductID != nil {
		ip.ProductID = *up.ProductID
	}
	if up.LocationID != nil {
		ip.LocationID = *up.LocationID
	}
	if up.Quantity != nil {
		ip.Quantity = *up.Quantity
	}
	if up.ReservedQuantity != nil {
		ip.ReservedQuantity = *up.ReservedQuantity
	}
	if up.AllocatedQuantity != nil {
		ip.AllocatedQuantity = *up.AllocatedQuantity
	}
	if up.MinimumStock != nil {
		ip.MinimumStock = *up.MinimumStock
	}
	if up.MaximumStock != nil {
		ip.MaximumStock = *up.MaximumStock
	}
	if up.ReorderPoint != nil {
		ip.ReorderPoint = *up.ReorderPoint
	}
	if up.EconomicOrderQuantity != nil {
		ip.EconomicOrderQuantity = *up.EconomicOrderQuantity
	}
	if up.SafetyStock != nil {
		ip.SafetyStock = *up.SafetyStock
	}
	if up.AvgDailyUsage != nil {
		ip.AvgDailyUsage = *up.AvgDailyUsage
	}

	ip.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, ip); err != nil {
		return InventoryItem{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, ip)); err != nil {
		b.log.Error(ctx, "inventoryitembus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return ip, nil
}

// Delete removes an inventoryItem from the system.
func (b *Business) Delete(ctx context.Context, ip InventoryItem) error {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ip); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(ip)); err != nil {
		b.log.Error(ctx, "inventoryitembus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// QueryItemsWithProductAtLocation retrieves inventory items at a location with product name/sku/tracking joined.
func (b *Business) QueryItemsWithProductAtLocation(ctx context.Context, locationID uuid.UUID) ([]ItemWithProduct, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.queryitemswithproductat")
	defer span.End()

	items, err := b.storer.QueryItemsWithProductAtLocation(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("query items with product at location: %w", err)
	}

	return items, nil
}

// QueryWithLocationDetails retrieves inventory items with location context fields joined.
func (b *Business) QueryWithLocationDetails(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryItemWithLocation, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.querywithdetails")
	defer span.End()

	items, err := b.storer.QueryWithLocationDetails(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query with location details: %w", err)
	}

	return items, nil
}

// Query retrieves inventoryItems based on the provided filter and pagination settings.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.query")
	defer span.End()

	inventoryItems, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return inventoryItems, nil
}

// UpsertQuantity creates or updates the inventory item for (productID, locationID),
// adding quantityDelta to the existing quantity atomically.
// quantityDelta must be positive; zero or negative values are rejected.
func (b *Business) UpsertQuantity(ctx context.Context, productID, locationID uuid.UUID, quantityDelta int) error {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.upsertquantity")
	defer span.End()

	if quantityDelta <= 0 {
		return fmt.Errorf("upsert quantity: quantityDelta must be positive, got %d", quantityDelta)
	}

	if err := b.storer.UpsertQuantity(ctx, uuid.New(), productID, locationID, quantityDelta); err != nil {
		return fmt.Errorf("upsert quantity: %w", err)
	}

	return nil
}

// QueryAvailableForAllocation retrieves inventory items that have available quantity for allocation.
func (b *Business) QueryAvailableForAllocation(ctx context.Context, productID uuid.UUID, locationID *uuid.UUID, warehouseID *uuid.UUID, strategy string, limit int) ([]InventoryItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.queryavailableforallocation")
	defer span.End()

	inventoryItems, err := b.storer.QueryAvailableForAllocation(ctx, productID, locationID, warehouseID, strategy, limit)
	if err != nil {
		return nil, fmt.Errorf("query available for allocation: %w", err)
	}

	return inventoryItems, nil
}

// Count returns the total number of inventoryItems in the system.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves an inventoryItem by its ID.
func (b *Business) QueryByID(ctx context.Context, inventoryID uuid.UUID) (InventoryItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.querybyid")
	defer span.End()

	inventoryItem, err := b.storer.QueryByID(ctx, inventoryID)
	if err != nil {
		return InventoryItem{}, fmt.Errorf("query by id: %w", err)
	}

	return inventoryItem, nil
}
