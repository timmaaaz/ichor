package assetbus

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
	ErrNotFound              = errors.New("asset not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("asset entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, asset Asset) error
	Update(ctx context.Context, asset Asset) error
	Delete(ctx context.Context, asset Asset) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Asset, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, assetID uuid.UUID) (Asset, error)
}

// Business manages the set of APIs for asset access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a asset business API for use.
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

// Create inserts a new asset into the database.
func (b *Business) Create(ctx context.Context, na NewAsset) (Asset, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.create")
	defer span.End()

	now := time.Now()

	asset := Asset{
		ID:                  uuid.New(),
		TypeID:              na.TypeID,
		Name:                na.Name,
		EstPrice:            na.EstPrice,
		MaintenanceInterval: na.MaintenanceInterval,
		LifeExpectancy:      na.LifeExpectancy,
		ModelNumber:         na.ModelNumber,
		IsEnabled:           na.IsEnabled,
		DateCreated:         now,
		DateUpdated:         now,
		CreatedBy:           na.CreatedBy,
		UpdatedBy:           na.CreatedBy,
	}

	if err := b.storer.Create(ctx, asset); err != nil {
		return Asset{}, fmt.Errorf("create: %w", err)
	}

	return asset, nil
}

// Update replaces an asset document in the database.
func (b *Business) Update(ctx context.Context, ass Asset, ua UpdateAsset) (Asset, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.update")
	defer span.End()

	now := time.Now()

	if ua.TypeID != nil {
		ass.TypeID = *ua.TypeID
	}

	if ua.Name != nil {
		ass.Name = *ua.Name
	}

	if ua.EstPrice != nil {
		ass.EstPrice = *ua.EstPrice
	}

	if ua.Price != nil {
		ass.Price = *ua.Price
	}

	if ua.MaintenanceInterval != nil {
		ass.MaintenanceInterval = *ua.MaintenanceInterval
	}

	if ua.LifeExpectancy != nil {
		ass.LifeExpectancy = *ua.LifeExpectancy
	}

	if ua.ModelNumber != nil {
		ass.ModelNumber = *ua.ModelNumber
	}

	if ua.IsEnabled != nil {
		ass.IsEnabled = *ua.IsEnabled
	}

	if ua.UpdatedBy != nil {
		ass.UpdatedBy = *ua.UpdatedBy
	}

	ass.DateUpdated = now

	if err := b.storer.Update(ctx, ass); err != nil {
		return Asset{}, fmt.Errorf("update: %w", err)
	}

	return ass, nil
}

// Delete removes the specified asset.
func (b *Business) Delete(ctx context.Context, ass Asset) error {
	ctx, span := otel.AddSpan(ctx, "business.assetbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ass); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of assets from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Asset, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetbus.Query")
	defer span.End()

	strs, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return strs, nil
}

// Count returns the total number of assets.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the asset by the specified ID.
func (b *Business) QueryByID(ctx context.Context, assetID uuid.UUID) (Asset, error) {
	ctx, span := otel.AddSpan(ctx, "business.assetbus.querybyid")
	defer span.End()

	asset, err := b.storer.QueryByID(ctx, assetID)
	if err != nil {
		return Asset{}, fmt.Errorf("query: assetID[%s]: %w", assetID, err)
	}

	return asset, nil
}
