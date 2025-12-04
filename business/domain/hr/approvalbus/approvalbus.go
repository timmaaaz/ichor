package approvalbus

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
	Create(ctx context.Context, userApprovalStatus UserApprovalStatus) error
	Update(ctx context.Context, userApprovalStatus UserApprovalStatus) error
	Delete(ctx context.Context, userApprovalStatus UserApprovalStatus) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserApprovalStatus, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userApprovalStatusID uuid.UUID) (UserApprovalStatus, error)
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
func (b *Business) Create(ctx context.Context, nas NewUserApprovalStatus) (UserApprovalStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.userapprovalstatusbus.Create")
	defer span.End()

	as := UserApprovalStatus{
		ID:             uuid.New(),
		Name:           nas.Name,
		IconID:         nas.IconID,
		PrimaryColor:   nas.PrimaryColor,
		SecondaryColor: nas.SecondaryColor,
		Icon:           nas.Icon,
	}

	if err := b.storer.Create(ctx, as); err != nil {
		return UserApprovalStatus{}, fmt.Errorf("store create: %w", err)
	}

	return as, nil
}

// Update modifies information about an approval status.
func (b *Business) Update(ctx context.Context, as UserApprovalStatus, uas UpdateUserApprovalStatus) (UserApprovalStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.userapprovalstatusbus.Update")
	defer span.End()

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
		return UserApprovalStatus{}, fmt.Errorf("store update: %w", err)
	}

	return as, nil
}

// Delete removes an approval status from the system.
func (b *Business) Delete(ctx context.Context, as UserApprovalStatus) error {
	ctx, span := otel.AddSpan(ctx, "business.userapprovalstatusbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, as); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	return nil
}

// Query returns a list of approval statuses
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserApprovalStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.userapprovalstatusbus.Query")
	defer span.End()

	aprvlStatuses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return aprvlStatuses, nil
}

// Count returns the total number of approval statuses
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.userapprovalstatusbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the approval status by the specified ID.
func (b *Business) QueryByID(ctx context.Context, aprvlStatusID uuid.UUID) (UserApprovalStatus, error) {
	ctx, span := otel.AddSpan(ctx, "business.userapprovalstatusbus.QueryByID")
	defer span.End()

	as, err := b.storer.QueryByID(ctx, aprvlStatusID)
	if err != nil {
		return UserApprovalStatus{}, fmt.Errorf("query: aprvlStatusID[%s]: %w", aprvlStatusID, err)
	}

	return as, nil
}
