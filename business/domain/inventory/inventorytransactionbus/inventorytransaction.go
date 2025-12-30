package inventorytransactionbus

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
	ErrNotFound              = errors.New("inventoryTransaction not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("inventoryTransaction entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, inventoryTransaction InventoryTransaction) error
	Update(ctx context.Context, inventoryTransaction InventoryTransaction) error
	Delete(ctx context.Context, inventoryTransaction InventoryTransaction) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryTransaction, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, inventoryTransactionID uuid.UUID) (InventoryTransaction, error)
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

// Create creates a new inventoryTransaction.
func (b *Business) Create(ctx context.Context, nit NewInventoryTransaction) (InventoryTransaction, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorytransactionbus.Create")
	defer span.End()

	now := time.Now()

	it := InventoryTransaction{
		InventoryTransactionID: uuid.New(),
		LocationID:             nit.LocationID,
		ProductID:              nit.ProductID,
		UserID:                 nit.UserID,
		TransactionType:        nit.TransactionType,
		Quantity:               nit.Quantity,
		ReferenceNumber:        nit.ReferenceNumber,
		TransactionDate:        nit.TransactionDate,
		CreatedDate:            now,
		UpdatedDate:            now,
	}

	err := b.storer.Create(ctx, it)
	if err != nil {
		return InventoryTransaction{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(it)); err != nil {
		b.log.Error(ctx, "inventorytransactionbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return it, nil
}

// Update updates an existing inventoryTransaction.
func (b *Business) Update(ctx context.Context, it InventoryTransaction, u UpdateInventoryTransaction) (InventoryTransaction, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorytransactionbus.update")
	defer span.End()

	if u.ProductID != nil {
		it.ProductID = *u.ProductID
	}
	if u.LocationID != nil {
		it.LocationID = *u.LocationID
	}
	if u.UserID != nil {
		it.UserID = *u.UserID
	}
	if u.Quantity != nil {
		it.Quantity = *u.Quantity
	}
	if u.TransactionType != nil {
		it.TransactionType = *u.TransactionType
	}
	if u.ReferenceNumber != nil {
		it.ReferenceNumber = *u.ReferenceNumber
	}
	if u.TransactionDate != nil {
		it.TransactionDate = *u.TransactionDate
	}

	it.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, it); err != nil {
		return InventoryTransaction{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(it)); err != nil {
		b.log.Error(ctx, "inventorytransactionbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return it, nil
}

// Delete removes a inventoryTransaction from the system.
func (b *Business) Delete(ctx context.Context, it InventoryTransaction) error {
	ctx, span := otel.AddSpan(ctx, "business.inventorytransactionbus.delete")
	defer span.End()

	err := b.storer.Delete(ctx, it)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(it)); err != nil {
		b.log.Error(ctx, "inventorytransactionbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves inventoryTransactions based on the given filter, order, and page.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]InventoryTransaction, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorytransactionbus.query")
	defer span.End()

	inventoryTransactions, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return inventoryTransactions, nil
}

// Count returns the total number of inventoryTransactions.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorytransactionbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID retrieves an inventoryTransaction by its ID.
func (b *Business) QueryByID(ctx context.Context, inventoryTransactionID uuid.UUID) (InventoryTransaction, error) {
	ctx, span := otel.AddSpan(ctx, "business.inventorytransactionbus.querybyid")
	defer span.End()

	it, err := b.storer.QueryByID(ctx, inventoryTransactionID)
	if err != nil {
		return InventoryTransaction{}, fmt.Errorf("queryByID: %w", err)
	}

	return it, nil
}
