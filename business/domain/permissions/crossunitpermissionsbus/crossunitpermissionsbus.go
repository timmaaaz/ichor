package crossunitpermissionsbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/convert"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("role not found")
	ErrUnique                = errors.New("org unit column access organization is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrColumnNotExists       = errors.New("column does not exist")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, cup CrossUnitPermission) error
	Update(ctx context.Context, cup CrossUnitPermission) error
	Delete(ctx context.Context, cup CrossUnitPermission) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CrossUnitPermission, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, id uuid.UUID) (CrossUnitPermission, error)
}

// Business manages the set of APIs for cross unit permission access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a cross unit permissions business API for use.
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

// Create adds a new cross unit permission to the database.
func (b *Business) Create(ctx context.Context, ncup NewCrossUnitPermission) (CrossUnitPermission, error) {
	ctx, span := otel.AddSpan(ctx, "business.crossunitpermission.create")
	defer span.End()

	cup := CrossUnitPermission{
		ID:           uuid.New(),
		SourceUnitID: ncup.SourceUnitID,
		TargetUnitID: ncup.TargetUnitID,
		CanRead:      ncup.CanRead,
		CanUpdate:    ncup.CanUpdate,
		GrantedBy:    ncup.GrantedBy,
		ValidFrom:    ncup.ValidFrom,
		ValidUntil:   ncup.ValidUntil,
		Reason:       ncup.Reason,
	}

	if err := b.storer.Create(ctx, cup); err != nil {
		return CrossUnitPermission{}, fmt.Errorf("create: %w", err)
	}

	return cup, nil
}

// Update modifies a cross unit permission in the database.
func (b *Business) Update(ctx context.Context, cup CrossUnitPermission, ucup UpdateCrossUnitPermission) (CrossUnitPermission, error) {
	ctx, span := otel.AddSpan(ctx, "business.crossunitpermission.update")
	defer span.End()

	err := convert.PopulateSameTypes(ucup, &cup)
	if err != nil {
		return CrossUnitPermission{}, fmt.Errorf("update: %w", err)
	}

	if err := b.storer.Update(ctx, cup); err != nil {
		return CrossUnitPermission{}, fmt.Errorf("update: %w", err)
	}

	return cup, nil
}

// Delete removes a cross unit permission from the database.
func (b *Business) Delete(ctx context.Context, cup CrossUnitPermission) error {
	ctx, span := otel.AddSpan(ctx, "business.crossunitpermission.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, cup); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of cross unit permissions from the database.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CrossUnitPermission, error) {
	ctx, span := otel.AddSpan(ctx, "business.crossunitpermission.query")
	defer span.End()

	cups, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return cups, nil
}

// Count returns the number of cross unit permissions that match the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.crossunitpermission.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID finds the cross unit permission by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (CrossUnitPermission, error) {
	ctx, span := otel.AddSpan(ctx, "business.crossunitpermission.querybyid")
	defer span.End()

	cup, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return CrossUnitPermission{}, fmt.Errorf("query: id[%s]: %w", id, err)
	}

	return cup, nil
}
