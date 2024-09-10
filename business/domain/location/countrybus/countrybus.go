package countrybus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set error variables.
var (
	ErrNotFound = fmt.Errorf("country not found")
)

type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Country, error)
	QueryByID(ctx context.Context, countryID uuid.UUID) (Country, error)
}

// Business manages the set of APIs for user access.
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

// Query retrieves a list of existing countries from the database.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Country, error) {
	ctx, span := otel.AddSpan(ctx, "business.domain.location.countrybus.Query")
	defer span.End()

	countries, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return countries, nil
}

// Count returns the number of countries that match the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.domain.location.countrybus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves the country with the specified ID.
func (b *Business) QueryByID(ctx context.Context, countryID uuid.UUID) (Country, error) {
	ctx, span := otel.AddSpan(ctx, "business.domain.location.countrybus.QueryByID")
	defer span.End()

	country, err := b.storer.QueryByID(ctx, countryID)
	if err != nil {
		return Country{}, fmt.Errorf("querybyid: %w", err)
	}

	return country, nil
}
