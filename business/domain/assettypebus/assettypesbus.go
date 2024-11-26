package assettypebus

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
	ErrNotFound              = errors.New("asset type not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("asset type entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, assetType AssetType) error
	Update(ctx context.Context, assetType AssetType) error
	Delete(ctx context.Context, assetType AssetType) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]AssetType, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, assetTypeID uuid.UUID) (AssetType, error)
}

type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a new asset type business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls
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

func (b *Business) Create(ctx context.Context, nat NewAssetType) (AssetType, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettypebus.Create")
	defer span.End()

	as := AssetType{
		ID:   uuid.New(),
		Name: nat.Name,
	}

	if err := b.storer.Create(ctx, as); err != nil {
		return AssetType{}, fmt.Errorf("store create: %w", err)
	}

	return as, nil
}

// Update modifies information about an asset type.
func (b *Business) Update(ctx context.Context, at AssetType, uat UpdateAssetType) (AssetType, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettypebus.Update")
	defer span.End()

	if uat.Name != nil {
		at.Name = *uat.Name
	}

	if err := b.storer.Update(ctx, at); err != nil {
		return AssetType{}, fmt.Errorf("store update: %w", err)
	}

	return at, nil
}

// Delete removes an asset type from the system.
func (b *Business) Delete(ctx context.Context, at AssetType) error {
	ctx, span := otel.AddSpan(ctx, "business.assettypebus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, at); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	return nil
}

// Query returns a list of asset types
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]AssetType, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettypebus.Query")
	defer span.End()

	assetTypes, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return assetTypes, nil
}

// Count returns the total number of asset types
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettypebus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the asset type by the specified ID.
func (b *Business) QueryByID(ctx context.Context, assetTypeID uuid.UUID) (AssetType, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettypebus.QueryByID")
	defer span.End()

	as, err := b.storer.QueryByID(ctx, assetTypeID)
	if err != nil {
		return AssetType{}, fmt.Errorf("query: assetTypeID[%s]: %w", assetTypeID, err)
	}

	return as, nil
}
