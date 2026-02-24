package putawaytaskbus

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
	ErrNotFound            = errors.New("put-away task not found")
	ErrUniqueEntry         = errors.New("put-away task entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, task PutAwayTask) error
	Update(ctx context.Context, task PutAwayTask) error
	Delete(ctx context.Context, task PutAwayTask) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PutAwayTask, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, taskID uuid.UUID) (PutAwayTask, error)
}

// Business manages the set of APIs for put-away task access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a put-away task business API for use.
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

// Create creates a new put-away task with status pending.
func (b *Business) Create(ctx context.Context, npt NewPutAwayTask) (PutAwayTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.putawaytaskbus.create")
	defer span.End()

	now := time.Now()

	task := PutAwayTask{
		ID:              uuid.New(),
		ProductID:       npt.ProductID,
		LocationID:      npt.LocationID,
		Quantity:        npt.Quantity,
		ReferenceNumber: npt.ReferenceNumber,
		Status:          Statuses.Pending,
		CreatedBy:       npt.CreatedBy,
		CreatedDate:     now,
		UpdatedDate:     now,
	}

	if err := b.storer.Create(ctx, task); err != nil {
		return PutAwayTask{}, fmt.Errorf("create: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionCreatedData(task)); err != nil {
		b.log.Error(ctx, "putawaytaskbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return task, nil
}

// Update modifies an existing put-away task.
func (b *Business) Update(ctx context.Context, pat PutAwayTask, upt UpdatePutAwayTask) (PutAwayTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.putawaytaskbus.update")
	defer span.End()

	before := pat

	if upt.ProductID != nil {
		pat.ProductID = *upt.ProductID
	}
	if upt.LocationID != nil {
		pat.LocationID = *upt.LocationID
	}
	if upt.Quantity != nil {
		pat.Quantity = *upt.Quantity
	}
	if upt.ReferenceNumber != nil {
		pat.ReferenceNumber = *upt.ReferenceNumber
	}
	if upt.Status != nil {
		pat.Status = *upt.Status
	}
	if upt.AssignedTo != nil {
		pat.AssignedTo = *upt.AssignedTo
	}
	if upt.AssignedAt != nil {
		pat.AssignedAt = *upt.AssignedAt
	}
	if upt.CompletedBy != nil {
		pat.CompletedBy = *upt.CompletedBy
	}
	if upt.CompletedAt != nil {
		pat.CompletedAt = *upt.CompletedAt
	}

	pat.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, pat); err != nil {
		return PutAwayTask{}, fmt.Errorf("update: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, pat)); err != nil {
		b.log.Error(ctx, "putawaytaskbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return pat, nil
}

// Delete removes a put-away task from the system.
func (b *Business) Delete(ctx context.Context, pat PutAwayTask) error {
	ctx, span := otel.AddSpan(ctx, "business.putawaytaskbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, pat); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionDeletedData(pat)); err != nil {
		b.log.Error(ctx, "putawaytaskbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of put-away tasks based on the given filter, order, and page.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PutAwayTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.putawaytaskbus.query")
	defer span.End()

	tasks, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return tasks, nil
}

// Count returns the total number of put-away tasks matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.putawaytaskbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds a put-away task by its ID.
func (b *Business) QueryByID(ctx context.Context, taskID uuid.UUID) (PutAwayTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.putawaytaskbus.querybyid")
	defer span.End()

	task, err := b.storer.QueryByID(ctx, taskID)
	if err != nil {
		return PutAwayTask{}, fmt.Errorf("queryByID: taskID[%s]: %w", taskID, err)
	}

	return task, nil
}
