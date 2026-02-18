package approvalrequestbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound        = errors.New("approval request not found")
	ErrAlreadyResolved = errors.New("approval request already resolved")
	ErrNotApprover     = errors.New("user is not an approver for this request")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, req ApprovalRequest) error
	QueryByID(ctx context.Context, id uuid.UUID) (ApprovalRequest, error)
	Resolve(ctx context.Context, id, resolvedBy uuid.UUID, status, reason string) (ApprovalRequest, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]ApprovalRequest, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	IsApprover(ctx context.Context, approvalID, userID uuid.UUID) (bool, error)
}

// Business manages approval request operations.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs an approval request business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
	}

	return &bus, nil
}

// Create adds a new approval request to the system.
func (b *Business) Create(ctx context.Context, na NewApprovalRequest) (ApprovalRequest, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalrequestbus.create")
	defer span.End()

	now := time.Now()

	req := ApprovalRequest{
		ID:              uuid.New(),
		ExecutionID:     na.ExecutionID,
		RuleID:          na.RuleID,
		ActionName:      na.ActionName,
		Approvers:       na.Approvers,
		ApprovalType:    na.ApprovalType,
		Status:          StatusPending,
		TimeoutHours:    na.TimeoutHours,
		TaskToken:       na.TaskToken,
		ApprovalMessage: na.ApprovalMessage,
		CreatedDate:     now,
	}

	if err := b.storer.Create(ctx, req); err != nil {
		return ApprovalRequest{}, fmt.Errorf("create approval request: %w", err)
	}

	return req, nil
}

// QueryByID returns a single approval request by ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (ApprovalRequest, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalrequestbus.querybyid")
	defer span.End()

	req, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return ApprovalRequest{}, fmt.Errorf("query approval request: id[%s]: %w", id, err)
	}

	return req, nil
}

// Resolve atomically transitions a pending approval request to approved/rejected.
// Returns ErrAlreadyResolved if the request is no longer pending.
func (b *Business) Resolve(ctx context.Context, id, resolvedBy uuid.UUID, status, reason string) (ApprovalRequest, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalrequestbus.resolve")
	defer span.End()

	req, err := b.storer.Resolve(ctx, id, resolvedBy, status, reason)
	if err != nil {
		return ApprovalRequest{}, fmt.Errorf("resolve approval request: %w", err)
	}

	return req, nil
}

// Query returns approval requests based on filter criteria.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]ApprovalRequest, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalrequestbus.query")
	defer span.End()

	reqs, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query approval requests: %w", err)
	}

	return reqs, nil
}

// Count returns the total count of approval requests based on filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalrequestbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// IsApprover checks if the given user is an approver for the request.
func (b *Business) IsApprover(ctx context.Context, approvalID, userID uuid.UUID) (bool, error) {
	ctx, span := otel.AddSpan(ctx, "business.approvalrequestbus.isapprover")
	defer span.End()

	return b.storer.IsApprover(ctx, approvalID, userID)
}
