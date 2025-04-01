package inventoryitembus

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
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, inventoryItemID uuid.UUID) (InventoryItem, error)
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
		ItemID:                uuid.New(),
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

	return inventoryItem, nil
}

// Update modifies an inventoryItem in the system.
func (b *Business) Update(ctx context.Context, ip InventoryItem, up UpdateInventoryItem) (InventoryItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.update")
	defer span.End()

	err := convert.PopulateSameTypes(up, &ip)
	if err != nil {
		return InventoryItem{}, fmt.Errorf("convert: populate same types: %w", err)
	}

	ip.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, ip); err != nil {
		return InventoryItem{}, fmt.Errorf("update: %w", err)
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

	return nil
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

// Count returns the total number of inventoryItems in the system.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves an inventoryItem by its ID.
func (b *Business) QueryByID(ctx context.Context, inventoryItemID uuid.UUID) (InventoryItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryitembus.querybyid")
	defer span.End()

	inventoryItem, err := b.storer.QueryByID(ctx, inventoryItemID)
	if err != nil {
		return InventoryItem{}, fmt.Errorf("query by id: %w", err)
	}

	return inventoryItem, nil
}
