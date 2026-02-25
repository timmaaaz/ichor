package inventoryadjustmentbus

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
	ErrNotFound              = errors.New("inventoryAdjustment not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("inventoryAdjustment entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, inventoryAdjustment InventoryAdjustment) error
	Update(ctx context.Context, inventoryAdjustment InventoryAdjustment) error
	Delete(ctx context.Context, inventoryAdjustment InventoryAdjustment) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryAdjustment, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, inventoryAdjustmentID uuid.UUID) (InventoryAdjustment, error)
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

// Create creates a new inventoryAdjustment.
func (b *Business) Create(ctx context.Context, nia NewInventoryAdjustment) (InventoryAdjustment, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryadjustmentbus.create")
	defer span.End()

	now := time.Now()

	ia := InventoryAdjustment{
		InventoryAdjustmentID: uuid.New(),
		ProductID:             nia.ProductID,
		LocationID:            nia.LocationID,
		AdjustedBy:            nia.AdjustedBy,
		ApprovedBy:            nia.ApprovedBy,
		ApprovalStatus:        "pending",
		QuantityChange:        nia.QuantityChange,
		ReasonCode:            nia.ReasonCode,
		Notes:                 nia.Notes,
		AdjustmentDate:        nia.AdjustmentDate,
		UpdatedDate:           now,
		CreatedDate:           now,
	}

	err := b.storer.Create(ctx, ia)
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(ia)); err != nil {
		b.log.Error(ctx, "inventoryadjustmentbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return ia, nil
}

// Update updates an existing inventoryAdjustment.
func (b *Business) Update(ctx context.Context, ia InventoryAdjustment, u UpdateInventoryAdjustment) (InventoryAdjustment, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryadjustmentbus.update")
	defer span.End()

	before := ia

	if u.ProductID != nil {
		ia.ProductID = *u.ProductID
	}
	if u.LocationID != nil {
		ia.LocationID = *u.LocationID
	}
	if u.AdjustedBy != nil {
		ia.AdjustedBy = *u.AdjustedBy
	}
	if u.ApprovedBy != nil {
		ia.ApprovedBy = u.ApprovedBy
	}
	if u.ApprovalStatus != nil {
		ia.ApprovalStatus = *u.ApprovalStatus
	}
	if u.QuantityChange != nil {
		ia.QuantityChange = *u.QuantityChange
	}
	if u.ReasonCode != nil {
		ia.ReasonCode = *u.ReasonCode
	}
	if u.Notes != nil {
		ia.Notes = *u.Notes
	}
	if u.AdjustmentDate != nil {
		ia.AdjustmentDate = *u.AdjustmentDate
	}

	ia.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, ia); err != nil {
		return InventoryAdjustment{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, ia)); err != nil {
		b.log.Error(ctx, "inventoryadjustmentbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return ia, nil
}

// Delete deletes an existing inventoryAdjustment.
func (b *Business) Delete(ctx context.Context, ia InventoryAdjustment) error {
	ctx, span := otel.AddSpan(ctx, "business.inventoryadjustmentbus.delete")
	defer span.End()

	err := b.storer.Delete(ctx, ia)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(ia)); err != nil {
		b.log.Error(ctx, "inventoryadjustmentbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves inventoryAdjustments based on the provided filter, order, and page.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryAdjustment, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryadjustmentbus.query")
	defer span.End()

	inventoryAdjustments, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return inventoryAdjustments, nil
}

// Count returns the total number of inventoryAdjustments that match the provided filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryadjustmentbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID retrieves an inventoryAdjustment by its ID.
func (b *Business) QueryByID(ctx context.Context, inventoryAdjustmentID uuid.UUID) (InventoryAdjustment, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventoryadjustmentbus.querybyid")
	defer span.End()

	ia, err := b.storer.QueryByID(ctx, inventoryAdjustmentID)
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("query by id: %w", err)
	}

	return ia, nil
}
