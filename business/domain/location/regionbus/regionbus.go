package regionbus

import (
	"context"
	"fmt"

	"bitbucket.org/superiortechnologies/ichor/business/sdk/delegate"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/page"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/sqldb"
	"bitbucket.org/superiortechnologies/ichor/foundation/logger"
	"bitbucket.org/superiortechnologies/ichor/foundation/otel"
	"github.com/google/uuid"
)

// Set error variables.
var (
	ErrNotFound = fmt.Errorf("region not found")
)

// Storer defines the database interaction methods.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Region, error)
	QueryByID(ctx context.Context, regionID uuid.UUID) (Region, error)
}

// Business manages the set of APIs for region access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a country business API for use.
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

// Query retrieves a list of existing regions from the database.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Region, error) {
	ctx, span := otel.AddSpan(ctx, "business.domain.location.regionbus.Query")
	defer span.End()

	regions, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return regions, nil
}

// Count returns the number of regions that match the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.domain.location.regionbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a single region by its ID.
func (b *Business) QueryByID(ctx context.Context, regionID uuid.UUID) (Region, error) {
	ctx, span := otel.AddSpan(ctx, "business.domain.location.regionbus.QueryByID")
	defer span.End()

	region, err := b.storer.QueryByID(ctx, regionID)
	if err != nil {
		return Region{}, fmt.Errorf("query by id: %w", err)
	}

	return region, nil
}
