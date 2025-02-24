package restrictedcolumnbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("restricted column not found")
	ErrUnique                = errors.New("not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, rc RestrictedColumn) error
	Delete(ctx context.Context, rc RestrictedColumn) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]RestrictedColumn, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (RestrictedColumn, error)
}

// Business manages the set of APIs for user access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a user business API for use.
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

// Create adds a new restricted column to the system
func (b *Business) Create(ctx context.Context, nrc NewRestrictedColumn) (RestrictedColumn, error) {
	ctx, span := otel.AddSpan(ctx, "business.restrictedcolumnbus.create")
	defer span.End()

	rc := RestrictedColumn{
		ID:         uuid.New(),
		TableName:  nrc.TableName,
		ColumnName: nrc.ColumnName,
	}

	if err := b.storer.Create(ctx, rc); err != nil {
		return RestrictedColumn{}, fmt.Errorf("creating restricted column: %w", err)
	}

	return rc, nil
}

// Delete removes a restricted column from the system
func (b *Business) Delete(ctx context.Context, rc RestrictedColumn) error {
	ctx, span := otel.AddSpan(ctx, "business.restrictedcolumnbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, rc); err != nil {
		return fmt.Errorf("deleting restricted column: %w", err)
	}

	return nil
}

// Query retrieves a list of restricted columns from the system
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]RestrictedColumn, error) {
	ctx, span := otel.AddSpan(ctx, "business.restrictedcolumnbus.query")
	defer span.End()

	rcs, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying restricted columns: %w", err)
	}

	return rcs, nil
}

// Count returns the number of restricted columns in the system
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.restrictedcolumnbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("counting restricted columns: %w", err)
	}

	return count, nil
}

// QueryByID retrieves a restricted column from the system by its ID
func (b *Business) QueryByID(ctx context.Context, rcID uuid.UUID) (RestrictedColumn, error) {
	ctx, span := otel.AddSpan(ctx, "business.restrictedcolumnbus.querybyid")
	defer span.End()

	rc, err := b.storer.QueryByID(ctx, rcID)
	if err != nil {
		return RestrictedColumn{}, fmt.Errorf("querying restricted column by ID: %w", err)
	}

	return rc, nil
}
