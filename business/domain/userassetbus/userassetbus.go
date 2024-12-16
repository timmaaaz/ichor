package userassetbus

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
	ErrNotFound              = errors.New("user-asset not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("user-asset entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, asset UserAsset) error
	Update(ctx context.Context, asset UserAsset) error
	Delete(ctx context.Context, asset UserAsset) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserAsset, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, assetID uuid.UUID) (UserAsset, error)
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
func (b *Business) Create(ctx context.Context, nua NewUserAsset) (UserAsset, error) {
	ctx, span := otel.AddSpan(ctx, "business.userassetbus.create")
	defer span.End()

	asset := UserAsset{
		ID:                  uuid.New(),
		UserID:              nua.UserID,
		AssetID:             nua.AssetID,
		ConditionID:         nua.ConditionID,
		ApprovedBy:          nua.ApprovedBy,
		ApprovalStatusID:    nua.ApprovalStatusID,
		FulfillmentStatusID: nua.FulfillmentStatusID,
		DateReceived:        nua.DateReceived,
		LastMaintenance:     nua.LastMaintenance,
	}

	if err := b.storer.Create(ctx, asset); err != nil {
		return UserAsset{}, fmt.Errorf("create: %w", err)
	}

	return asset, nil
}

// Update replaces an asset document in the database.
func (b *Business) Update(ctx context.Context, ass UserAsset, uua UpdateUserAsset) (UserAsset, error) {
	ctx, span := otel.AddSpan(ctx, "business.userassetbus.update")
	defer span.End()

	if uua.ApprovalStatusID != nil {
		ass.ApprovalStatusID = *uua.ApprovalStatusID
	}

	if uua.FulfillmentStatusID != nil {
		ass.FulfillmentStatusID = *uua.FulfillmentStatusID
	}

	if uua.DateReceived != nil {
		ass.DateReceived = *uua.DateReceived
	}

	if uua.LastMaintenance != nil {
		ass.LastMaintenance = *uua.LastMaintenance
	}

	if uua.ApprovedBy != nil {
		ass.ApprovedBy = *uua.ApprovedBy
	}

	if uua.AssetID != nil {
		ass.AssetID = *uua.AssetID
	}

	if uua.UserID != nil {
		ass.UserID = *uua.UserID
	}

	if uua.ConditionID != nil {
		ass.ConditionID = *uua.ConditionID
	}

	if err := b.storer.Update(ctx, ass); err != nil {
		return UserAsset{}, fmt.Errorf("update: %w", err)
	}

	return ass, nil
}

// Delete removes the specified asset.
func (b *Business) Delete(ctx context.Context, ass UserAsset) error {
	ctx, span := otel.AddSpan(ctx, "business.userassetbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ass); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of assets from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]UserAsset, error) {
	ctx, span := otel.AddSpan(ctx, "business.userassetbus.Query")
	defer span.End()

	userAssets, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return userAssets, nil
}

// Count returns the total number of assets.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.userassetbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the asset by the specified ID.
func (b *Business) QueryByID(ctx context.Context, assetID uuid.UUID) (UserAsset, error) {
	ctx, span := otel.AddSpan(ctx, "business.userassetbus.querybyid")
	defer span.End()

	asset, err := b.storer.QueryByID(ctx, assetID)
	if err != nil {
		return UserAsset{}, fmt.Errorf("query: assetID[%s]: %w", assetID, err)
	}

	return asset, nil
}
