package cyclecountsessionbus

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
	ErrNotFound              = errors.New("cycle count session not found")
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrUniqueEntry           = errors.New("cycle count session entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, session CycleCountSession) error
	Update(ctx context.Context, session CycleCountSession) error
	Delete(ctx context.Context, session CycleCountSession) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CycleCountSession, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, sessionID uuid.UUID) (CycleCountSession, error)
}

// Business manages the set of APIs for cycle count session access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a cycle count session business API for use.
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
		storer:   storer,
		delegate: b.delegate,
	}, nil
}

// Create adds a new cycle count session to the system.
func (b *Business) Create(ctx context.Context, nccs NewCycleCountSession) (CycleCountSession, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.create")
	defer span.End()

	now := time.Now()

	session := CycleCountSession{
		ID:          uuid.New(),
		Name:        nccs.Name,
		Status:      Statuses.Draft,
		CreatedBy:   nccs.CreatedBy,
		CreatedDate: now,
		UpdatedDate: now,
	}

	if err := b.storer.Create(ctx, session); err != nil {
		return CycleCountSession{}, fmt.Errorf("create: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionCreatedData(session)); err != nil {
		b.log.Error(ctx, "cyclecountsessionbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return session, nil
}

// Update modifies an existing cycle count session in the system.
func (b *Business) Update(ctx context.Context, session CycleCountSession, uccs UpdateCycleCountSession) (CycleCountSession, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.update")
	defer span.End()

	before := session

	if uccs.Name != nil {
		session.Name = *uccs.Name
	}
	if uccs.Status != nil {
		session.Status = *uccs.Status
	}
	if uccs.CompletedDate != nil {
		session.CompletedDate = *uccs.CompletedDate
	}

	session.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, session); err != nil {
		return CycleCountSession{}, fmt.Errorf("update: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, session)); err != nil {
		b.log.Error(ctx, "cyclecountsessionbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return session, nil
}

// Delete removes a cycle count session from the system.
func (b *Business) Delete(ctx context.Context, session CycleCountSession) error {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, session); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionDeletedData(session)); err != nil {
		b.log.Error(ctx, "cyclecountsessionbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of cycle count sessions from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CycleCountSession, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.query")
	defer span.End()

	sessions, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return sessions, nil
}

// Count returns the total number of cycle count sessions matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a single cycle count session by its ID.
func (b *Business) QueryByID(ctx context.Context, sessionID uuid.UUID) (CycleCountSession, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.querybyid")
	defer span.End()

	session, err := b.storer.QueryByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return CycleCountSession{}, err
		}
		return CycleCountSession{}, fmt.Errorf("queryByID: sessionID[%s]: %w", sessionID, err)
	}

	return session, nil
}
