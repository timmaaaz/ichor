package approvalstatusbus

import (
	"context"
	"errors"
	"fmt"

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
	ErrNotFound              = errors.New("approval status not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("approval status entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, approvalStatus ApprovalStatus) error
	Update(ctx context.Context, approvalStatus ApprovalStatus) error
	Delete(ctx context.Context, approvalStatus ApprovalStatus) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ApprovalStatus, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, approvalStatusID uuid.UUID) (ApprovalStatus, error)
}

type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a new approval status business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
//
// This function seems like it could be implemented only once, but same with a lot of other pieces of this
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
	}

	return &bus, nil
}

// Create creates a new approval status to the system.
func (b *Business) Create(ctx context.Context, nas NewApprovalStatus) (ApprovalStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalstatusbus.Create")
	defer span.End()

	as := ApprovalStatus{
		ID:             uuid.New(),
		Name:           nas.Name,
		IconID:         nas.IconID,
		PrimaryColor:   nas.PrimaryColor,
		SecondaryColor: nas.SecondaryColor,
		Icon:           nas.Icon,
	}

	if err := b.storer.Create(ctx, as); err != nil {
		return ApprovalStatus{}, fmt.Errorf("store create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(as)); err != nil {
		b.log.Error(ctx, "approvalstatusbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return as, nil
}

// Update modifies information about an approval status.
func (b *Business) Update(ctx context.Context, as ApprovalStatus, uas UpdateApprovalStatus) (ApprovalStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalstatusbus.Update")
	defer span.End()

	before := as

	if uas.Name != nil {
		as.Name = *uas.Name
	}
	if uas.IconID != nil {
		as.IconID = *uas.IconID
	}
	if uas.PrimaryColor != nil {
		as.PrimaryColor = *uas.PrimaryColor
	}
	if uas.SecondaryColor != nil {
		as.SecondaryColor = *uas.SecondaryColor
	}
	if uas.Icon != nil {
		as.Icon = *uas.Icon
	}

	if err := b.storer.Update(ctx, as); err != nil {
		return ApprovalStatus{}, fmt.Errorf("store update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, as)); err != nil {
		b.log.Error(ctx, "approvalstatusbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return as, nil
}

// Delete removes an approval status from the system.
func (b *Business) Delete(ctx context.Context, as ApprovalStatus) error {
	ctx, span := otel.AddSpan(ctx, "business.approvalstatusbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, as); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(as)); err != nil {
		b.log.Error(ctx, "approvalstatusbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query returns a list of approval statuses
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ApprovalStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalstatusbus.Query")
	defer span.End()

	aprvlStatuses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return aprvlStatuses, nil
}

// Count returns the total number of approval statuses
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalstatusbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the approval status by the specified ID.
func (b *Business) QueryByID(ctx context.Context, aprvlStatusID uuid.UUID) (ApprovalStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalstatusbus.QueryByID")
	defer span.End()

	as, err := b.storer.QueryByID(ctx, aprvlStatusID)
	if err != nil {
		return ApprovalStatus{}, fmt.Errorf("query: aprvlStatusID[%s]: %w", aprvlStatusID, err)
	}

	return as, nil
}
