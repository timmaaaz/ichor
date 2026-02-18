package assettagbus

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
	ErrNotFound              = errors.New("asset tag not found")
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrUniqueEntry           = errors.New("asset tag entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, tag AssetTag) error
	Update(ctx context.Context, tag AssetTag) error
	Delete(ctx context.Context, tag AssetTag) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]AssetTag, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, tagID uuid.UUID) (AssetTag, error)
}

// Business manages the set of APIs for tag access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs an asset tag business API for use.
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

// Create adds a new asset tag in the system.
func (b *Business) Create(ctx context.Context, nat NewAssetTag) (AssetTag, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettagbus.Create")
	defer span.End()

	t := AssetTag{
		ID:           uuid.New(),
		ValidAssetID: nat.ValidAssetID,
		TagID:        nat.TagID,
	}

	if err := b.storer.Create(ctx, t); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return AssetTag{}, fmt.Errorf("create: %w", ErrUniqueEntry)
		}
		return AssetTag{}, err
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(t)); err != nil {
		b.log.Error(ctx, "assettagbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return t, nil
}

// Update updates an existing asset tag.
func (b *Business) Update(ctx context.Context, at AssetTag, uat UpdateAssetTag) (AssetTag, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettagbus.Update")
	defer span.End()

	before := at

	if uat.TagID != nil {
		at.TagID = *uat.TagID
	}

	if uat.ValidAssetID != nil {
		at.ValidAssetID = *uat.ValidAssetID
	}

	if err := b.storer.Update(ctx, at); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return AssetTag{}, fmt.Errorf("update: %w", ErrUniqueEntry)
		}
		return AssetTag{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, at)); err != nil {
		b.log.Error(ctx, "assettagbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return at, nil
}

// Delete removes an asset tag from the system.
func (b *Business) Delete(ctx context.Context, at AssetTag) error {
	ctx, span := otel.AddSpan(ctx, "business.assettagbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, at); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(at)); err != nil {
		b.log.Error(ctx, "assettagbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of existing asset tags from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]AssetTag, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettagbus.Query")
	defer span.End()

	tags, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return tags, nil
}

// Count returns the total number of tags.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettagbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the tag by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (AssetTag, error) {
	ctx, span := otel.AddSpan(ctx, "business.assettagbus.QueryByID")
	defer span.End()

	tag, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return AssetTag{}, fmt.Errorf("query: tagID[%s]: %w", id, err)
	}

	return tag, nil
}
