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

type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a new asset condition business API
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		storer:   storer,
		delegate: delegate,
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

// Create creates a new asset condition to the system.
func (b *Business) Create(ctx context.Context, nac NewAssetCondition) (AssetCondition, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Create")
	defer span.End()

	as := AssetCondition{
		ID:   uuid.New(),
		Name: nac.Name,
	}

	if err := b.storer.Create(ctx, as); err != nil {
		return AssetCondition{}, fmt.Errorf("store create: %w", err)
	}

	return as, nil
}

// Update modifies information about an asset condition.
func (b *Business) Update(ctx context.Context, as AssetCondition, uac UpdateAssetCondition) (AssetCondition, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Update")
	defer span.End()

	if uac.Name != nil {
		as.Name = *uac.Name
	}

	if err := b.storer.Update(ctx, as); err != nil {
		return AssetCondition{}, fmt.Errorf("store update: %w", err)
	}

	return as, nil
}

// Delete removes an asset condition from the system.
func (b *Business) Delete(ctx context.Context, as AssetCondition) error {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, as); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	return nil
}

// Query returns a list of asset condition
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]AssetCondition, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Query")
	defer span.End()

	aprvlStatuses, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return aprvlStatuses, nil
}

// Count returns the total number of asset conditiones
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the asset condition by the specified ID.
func (b *Business) QueryByID(ctx context.Context, aprvlStatusID uuid.UUID) (AssetCondition, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetconditionbus.QueryByID")
	defer span.End()

	as, err := b.storer.QueryByID(ctx, aprvlStatusID)
	if err != nil {
		return AssetCondition{}, fmt.Errorf("query: aprvlStatusID[%s]: %w", aprvlStatusID, err)
	}

	return as, nil
}
