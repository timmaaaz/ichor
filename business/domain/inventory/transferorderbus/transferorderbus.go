package transferorderbus

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
	ErrNotFound              = errors.New("transferOrder not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("transferOrder entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, transferOrder TransferOrder) error
	Update(ctx context.Context, transferOrder TransferOrder) error
	Delete(ctx context.Context, transferOrder TransferOrder) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]TransferOrder, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, transferOrderID uuid.UUID) (TransferOrder, error)
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

// Create creates a new transferOrder.
func (b *Business) Create(ctx context.Context, nto NewTransferOrder) (TransferOrder, error) {

	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.create")
	defer span.End()

	now := time.Now()

	transferOrder := TransferOrder{
		TransferID:     uuid.New(),
		ProductID:      nto.ProductID,
		FromLocationID: nto.FromLocationID,
		ToLocationID:   nto.ToLocationID,
		RequestedByID:  nto.RequestedByID,
		ApprovedByID:   nto.ApprovedByID,
		Quantity:       nto.Quantity,
		Status:         nto.Status,
		TransferDate:   nto.TransferDate,
		CreatedDate:    now,
		UpdatedDate:    now,
	}

	if err := b.storer.Create(ctx, transferOrder); err != nil {
		return TransferOrder{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(transferOrder)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return transferOrder, nil
}

// Update updates an existing transferOrder.
func (b *Business) Update(ctx context.Context, to TransferOrder, ut UpdateTransferOrder) (TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.update")
	defer span.End()

	if ut.ProductID != nil {
		to.ProductID = *ut.ProductID
	}
	if ut.FromLocationID != nil {
		to.FromLocationID = *ut.FromLocationID
	}
	if ut.ToLocationID != nil {
		to.ToLocationID = *ut.ToLocationID
	}
	if ut.RequestedByID != nil {
		to.RequestedByID = *ut.RequestedByID
	}
	if ut.ApprovedByID != nil {
		to.ApprovedByID = *ut.ApprovedByID
	}
	if ut.Quantity != nil {
		to.Quantity = *ut.Quantity
	}
	if ut.Status != nil {
		to.Status = *ut.Status
	}
	if ut.TransferDate != nil {
		to.TransferDate = *ut.TransferDate
	}

	to.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, to); err != nil {
		return TransferOrder{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(to)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return to, nil
}

// Delete removes a transferOrder from the system.
func (b *Business) Delete(ctx context.Context, to TransferOrder) error {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, to); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(to)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of transferOrders based on the given query filter, order by, and pagination.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.query")
	defer span.End()

	transferOrders, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return transferOrders, nil
}

// Count returns the total number of transferOrders.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the transferOrder by the specified ID.
func (b *Business) QueryByID(ctx context.Context, transferOrderID uuid.UUID) (TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.querybyid")
	defer span.End()

	transferOrder, err := b.storer.QueryByID(ctx, transferOrderID)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("queryByID: transferOrderID: %w", err)
	}

	return transferOrder, nil
}
