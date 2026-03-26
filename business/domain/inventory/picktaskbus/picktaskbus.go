package picktaskbus

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
	ErrNotFound            = errors.New("pick task not found")
	ErrUniqueEntry         = errors.New("pick task entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, task PickTask) error
	Update(ctx context.Context, task PickTask) error
	Delete(ctx context.Context, task PickTask) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PickTask, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, taskID uuid.UUID) (PickTask, error)
}

// Business manages the set of APIs for pick task access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a pick task business API for use.
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

// Create adds a new pick task to the system.
func (b *Business) Create(ctx context.Context, npt NewPickTask) (PickTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.create")
	defer span.End()

	now := time.Now()

	task := PickTask{
		ID:                   uuid.New(),
		SalesOrderID:         npt.SalesOrderID,
		SalesOrderLineItemID: npt.SalesOrderLineItemID,
		ProductID:            npt.ProductID,
		LotID:                npt.LotID,
		SerialID:             npt.SerialID,
		LocationID:           npt.LocationID,
		QuantityToPick:       npt.QuantityToPick,
		QuantityPicked:       0,
		Status:               Statuses.Pending,
		CreatedBy:            npt.CreatedBy,
		CreatedDate:          now,
		UpdatedDate:          now,
	}

	if err := b.storer.Create(ctx, task); err != nil {
		return PickTask{}, fmt.Errorf("create: %w", err)
	}

	if b.delegate != nil {
		b.delegate.Call(ctx, ActionCreatedData(task))
	}

	return task, nil
}

// Update modifies an existing pick task in the system.
func (b *Business) Update(ctx context.Context, task PickTask, upt UpdatePickTask) (PickTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.update")
	defer span.End()

	before := task

	if upt.LotID != nil {
		task.LotID = upt.LotID
	}
	if upt.SerialID != nil {
		task.SerialID = upt.SerialID
	}
	if upt.LocationID != nil {
		task.LocationID = *upt.LocationID
	}
	if upt.QuantityToPick != nil {
		task.QuantityToPick = *upt.QuantityToPick
	}
	if upt.QuantityPicked != nil {
		task.QuantityPicked = *upt.QuantityPicked
	}
	if upt.Status != nil {
		task.Status = *upt.Status
	}
	if upt.AssignedTo != nil {
		task.AssignedTo = *upt.AssignedTo
	}
	if upt.AssignedAt != nil {
		task.AssignedAt = *upt.AssignedAt
	}
	if upt.CompletedBy != nil {
		task.CompletedBy = *upt.CompletedBy
	}
	if upt.CompletedAt != nil {
		task.CompletedAt = *upt.CompletedAt
	}
	if upt.ShortPickReason != nil {
		task.ShortPickReason = *upt.ShortPickReason
	}

	task.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, task); err != nil {
		return PickTask{}, fmt.Errorf("update: %w", err)
	}

	if b.delegate != nil {
		b.delegate.Call(ctx, ActionUpdatedData(before, task))
	}

	return task, nil
}

// Delete removes a pick task from the system.
func (b *Business) Delete(ctx context.Context, task PickTask) error {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, task); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if b.delegate != nil {
		b.delegate.Call(ctx, ActionDeletedData(task))
	}

	return nil
}

// Query retrieves a list of pick tasks from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PickTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.query")
	defer span.End()

	tasks, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return tasks, nil
}

// Count returns the total number of pick tasks matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a single pick task by its ID.
func (b *Business) QueryByID(ctx context.Context, taskID uuid.UUID) (PickTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.querybyid")
	defer span.End()

	task, err := b.storer.QueryByID(ctx, taskID)
	if err != nil {
		return PickTask{}, fmt.Errorf("query: %w", err)
	}

	return task, nil
}
