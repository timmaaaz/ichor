package inspectionbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("inspection not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("inspection entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, inspection Inspection) error
	Update(ctx context.Context, inspection Inspection) error
	Delete(ctx context.Context, inspection Inspection) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Inspection, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, metricID uuid.UUID) (Inspection, error)
}

// Business manages the set of APIs for inspection access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a inspection business API for use.
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

// Create inserts a new inspection into the database.
func (b *Business) Create(ctx context.Context, ni NewInspection) (Inspection, error) {
	ctx, span := otel.AddSpan(ctx, "business.inspectionbus.create")
	defer span.End()

	now := time.Now()
	inspection := Inspection{
		InspectionID:       uuid.New(),
		ProductID:          ni.ProductID,
		InspectorID:        ni.InspectorID,
		InspectionDate:     ni.InspectionDate,
		Status:             ni.Status,
		LotID:              ni.LotID,
		Notes:              ni.Notes,
		NextInspectionDate: ni.NextInspectionDate,
		UpdatedDate:        now,
		CreatedDate:        now,
	}

	err := b.storer.Create(ctx, inspection)
	if err != nil {
		return Inspection{}, fmt.Errorf("create: %w", err)
	}

	return inspection, nil
}

// Update updates an existing inspection in the database.
func (b *Business) Update(ctx context.Context, i Inspection, ui UpdateInspection) (Inspection, error) {
	ctx, span := otel.AddSpan(ctx, "business.inspectionbus.update")
	defer span.End()

	now := time.Now()
	i.UpdatedDate = now

	err := convert.PopulateSameTypes(ui, &i)
	if err != nil {
		return Inspection{}, fmt.Errorf("convert: %w", err)
	}

	err = b.storer.Update(ctx, i)
	if err != nil {
		return Inspection{}, fmt.Errorf("update: %w", err)
	}

	return i, nil
}

// Delete removes an existing inspection from the database.
func (b *Business) Delete(ctx context.Context, i Inspection) error {
	ctx, span := otel.AddSpan(ctx, "business.inspectionbus.delete")
	defer span.End()

	err := b.storer.Delete(ctx, i)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of inspections based on the provided filter and sorting options.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Inspection, error) {
	ctx, span := otel.AddSpan(ctx, "business.inspectionbus.query")
	defer span.End()

	inspections, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return inspections, nil
}

// Count returns the total number of inspections that match the provided filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.inspectionbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID retrieves an inspection by its ID.
func (b *Business) QueryByID(ctx context.Context, inspectionID uuid.UUID) (Inspection, error) {
	ctx, span := otel.AddSpan(ctx, "business.inspectionbus.querybyid")
	defer span.End()

	inspection, err := b.storer.QueryByID(ctx, inspectionID)
	if err != nil {
		return Inspection{}, fmt.Errorf("queryByID: %w", err)
	}

	return inspection, nil
}
