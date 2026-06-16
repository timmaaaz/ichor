package transferorderbus

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
	ErrNotFound              = errors.New("transferOrder not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("transferOrder entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
	ErrInvalidTransferStatus = errors.New("transfer order is not in a state that allows this transition")
)

// Transfer order status values.
const (
	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusInTransit = "in_transit"
	StatusCompleted = "completed"
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, transferOrder TransferOrder) error
	Update(ctx context.Context, transferOrder TransferOrder) error
	// UpdateWithStatusGuard performs the same column update as Update, but only when
	// the row's current status equals expectedStatus. It returns the number of rows
	// affected (0 means the guard did not match — e.g. a concurrent transition won).
	UpdateWithStatusGuard(ctx context.Context, transferOrder TransferOrder, expectedStatus string) (int64, error)
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

	return &Business{
		log:      b.log,
		delegate: b.delegate,
		outbox:   b.outbox,
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
		TransferNumber: nto.TransferNumber,
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

	if sid, ok := sqldb.GetScenarioFilter(ctx); ok {
		transferOrder.ScenarioID = &sid
	}

	if err := b.storer.Create(ctx, transferOrder); err != nil {
		return TransferOrder{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	evtData := ActionCreatedData(transferOrder)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return TransferOrder{}, fmt.Errorf("emit cascade event: %w", err)
	}
	if err := b.delegate.Call(ctx, ActionCreatedData(transferOrder)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return transferOrder, nil
}

// Update updates an existing transferOrder.
func (b *Business) Update(ctx context.Context, to TransferOrder, ut UpdateTransferOrder) (TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.update")
	defer span.End()

	before := to

	if ut.TransferNumber != nil {
		to.TransferNumber = ut.TransferNumber
	}
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
		to.ApprovedByID = ut.ApprovedByID
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
	if ut.ClaimedByID != nil {
		to.ClaimedByID = ut.ClaimedByID
	}
	if ut.ClaimedAt != nil {
		to.ClaimedAt = ut.ClaimedAt
	}
	if ut.CompletedByID != nil {
		to.CompletedByID = ut.CompletedByID
	}
	if ut.CompletedAt != nil {
		to.CompletedAt = ut.CompletedAt
	}

	to.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, to); err != nil {
		return TransferOrder{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	evtData := ActionUpdatedData(before, to)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return TransferOrder{}, fmt.Errorf("emit cascade event: %w", err)
	}
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, to)); err != nil {
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
	evtData := ActionDeletedData(to)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return fmt.Errorf("emit cascade event: %w", err)
	}
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

// Approve sets the approver and marks the transfer order as approved.
func (b *Business) Approve(ctx context.Context, to TransferOrder, approvedBy uuid.UUID, reason string) (TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.approve")
	defer span.End()

	if to.Status != StatusPending {
		return TransferOrder{}, fmt.Errorf("approve: %w: must be pending, got %s", ErrInvalidTransferStatus, to.Status)
	}

	before := to

	now := time.Now()
	to.ApprovedByID = &approvedBy
	to.Status = StatusApproved
	to.ApprovalReason = reason
	to.UpdatedDate = now

	if err := b.storer.Update(ctx, to); err != nil {
		return TransferOrder{}, fmt.Errorf("approve: %w", err)
	}

	evtData := ActionUpdatedData(before, to)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return TransferOrder{}, fmt.Errorf("emit cascade event: %w", err)
	}
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, to)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return to, nil
}

// Claim marks an approved transfer order as in_transit, recording who claimed it.
func (b *Business) Claim(ctx context.Context, to TransferOrder, claimedBy uuid.UUID) (TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.claim")
	defer span.End()

	if to.Status != StatusApproved {
		return TransferOrder{}, fmt.Errorf("claim: %w: must be approved, got %s", ErrInvalidTransferStatus, to.Status)
	}

	before := to

	now := time.Now()
	to.ClaimedByID = &claimedBy
	to.ClaimedAt = &now
	to.Status = StatusInTransit
	to.UpdatedDate = now

	// Guard on the still-approved DB state to close the read-check-write race: if a
	// concurrent claim already moved the row to in_transit, 0 rows match and we reject
	// rather than silently overwriting the winner's claimed_by.
	rows, err := b.storer.UpdateWithStatusGuard(ctx, to, StatusApproved)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("claim: %w", err)
	}
	if rows == 0 {
		return TransferOrder{}, fmt.Errorf("claim: %w: status changed concurrently", ErrInvalidTransferStatus)
	}

	evtData := ActionUpdatedData(before, to)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return TransferOrder{}, fmt.Errorf("emit cascade event: %w", err)
	}
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, to)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return to, nil
}

// Execute marks an in_transit transfer order as completed, recording who completed it.
// This is a simple status transition — the atomic stock move happens at the app layer.
func (b *Business) Execute(ctx context.Context, to TransferOrder, completedBy uuid.UUID) (TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.execute")
	defer span.End()

	if to.Status != StatusInTransit {
		return TransferOrder{}, fmt.Errorf("execute: %w: must be in_transit, got %s", ErrInvalidTransferStatus, to.Status)
	}

	before := to

	now := time.Now()
	to.CompletedByID = &completedBy
	to.CompletedAt = &now
	to.Status = StatusCompleted
	to.UpdatedDate = now

	// Guard on the still-in_transit DB state to close the read-check-write race: if a
	// concurrent execute already completed the row, 0 rows match and we reject rather
	// than silently overwriting the winner's completed_by.
	rows, err := b.storer.UpdateWithStatusGuard(ctx, to, StatusInTransit)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("execute: %w", err)
	}
	if rows == 0 {
		return TransferOrder{}, fmt.Errorf("execute: %w: status changed concurrently", ErrInvalidTransferStatus)
	}

	evtData := ActionUpdatedData(before, to)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return TransferOrder{}, fmt.Errorf("emit cascade event: %w", err)
	}
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, to)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return to, nil
}

// Reject sets the rejector and marks the transfer order as rejected.
func (b *Business) Reject(ctx context.Context, to TransferOrder, rejectedBy uuid.UUID, reason string) (TransferOrder, error) {
	ctx, span := otel.AddSpan(ctx, "business.transferorderbus.reject")
	defer span.End()

	if to.Status != StatusPending {
		return TransferOrder{}, fmt.Errorf("reject: %w: must be pending, got %s", ErrInvalidTransferStatus, to.Status)
	}

	before := to

	now := time.Now()
	to.RejectedByID = &rejectedBy
	to.Status = StatusRejected
	to.RejectionReason = reason
	to.UpdatedDate = now

	if err := b.storer.Update(ctx, to); err != nil {
		return TransferOrder{}, fmt.Errorf("reject: %w", err)
	}

	evtData := ActionUpdatedData(before, to)
	if err := b.outbox.Emit(ctx, evtData); err != nil {
		return TransferOrder{}, fmt.Errorf("emit cascade event: %w", err)
	}
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, to)); err != nil {
		b.log.Error(ctx, "transferorderbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return to, nil
}
