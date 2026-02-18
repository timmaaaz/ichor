package assetconditionbus

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
	ErrNotFound              = errors.New("asset condition not found")
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrUniqueEntry           = errors.New("asset condition entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, assetCondition AssetCondition) error
	Update(ctx context.Context, assetCondition AssetCondition) error
	Delete(ctx context.Context, assetCondition AssetCondition) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]AssetCondition, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, assetConditionID uuid.UUID) (AssetCondition, error)
}

// Business manages the set of APIs for asset condition access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a asset condition business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
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
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
	}

	return &bus, nil
}

// Create adds a new asset condition to the system.
func (b *Business) Create(ctx context.Context, nat NewAssetCondition) (AssetCondition, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Create")
	defer span.End()

	at := AssetCondition{
		ID:          uuid.New(),
		Name:        nat.Name,
		Description: nat.Description,
	}

	if err := b.storer.Create(ctx, at); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return AssetCondition{}, fmt.Errorf("create: %w", ErrUniqueEntry)
		}
		return AssetCondition{}, err
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(at)); err != nil {
		b.log.Error(ctx, "assetconditionbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return at, nil
}

// Update updates an existing asset condition.
func (b *Business) Update(ctx context.Context, at AssetCondition, uat UpdateAssetCondition) (AssetCondition, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Update")
	defer span.End()

	before := at

	if uat.Name != nil {
		at.Name = *uat.Name
	}

	if uat.Description != nil {
		at.Description = *uat.Description
	}

	if err := b.storer.Update(ctx, at); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return AssetCondition{}, fmt.Errorf("update: %w", ErrUniqueEntry)
		}
		return AssetCondition{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, at)); err != nil {
		b.log.Error(ctx, "assetconditionbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return at, nil
}

// Delete removes an asset condition from the system.
func (b *Business) Delete(ctx context.Context, at AssetCondition) error {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, at); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(at)); err != nil {
		b.log.Error(ctx, "assetconditionbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of existing asset conditions from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]AssetCondition, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Query")
	defer span.End()

	assetConditions, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return assetConditions, nil
}

// Count returns the total number of asset conditions.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the asset condition by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (AssetCondition, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.QueryByID")
	defer span.End()

	assetCondition, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return AssetCondition{}, fmt.Errorf("query: assetConditionID[%s]: %w", id, err)
	}

	return assetCondition, nil
}
