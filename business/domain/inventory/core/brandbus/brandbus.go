package brandbus

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
	ErrNotFound              = errors.New("brand not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("brand entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, brand Brand) error
	Update(ctx context.Context, brand Brand) error
	Delete(ctx context.Context, brand Brand) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Brand, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, brandID uuid.UUID) (Brand, error)
}

// Business manages the set of APIs for brand access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a brand business API for use.
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

// Create inserts a new brand into the database.
func (b *Business) Create(ctx context.Context, na NewBrand) (Brand, error) {
	ctx, span := otel.AddSpan(ctx, "business.brandbus.create")
	defer span.End()

	now := time.Now()

	brand := Brand{
		BrandID:        uuid.New(),
		Name:           na.Name,
		ManufacturerID: na.ManufacturerID,
		ContactInfo:    na.ContactInfo,
		CreatedDate:    now,
		UpdatedDate:    now,
	}

	if err := b.storer.Create(ctx, brand); err != nil {
		return Brand{}, fmt.Errorf("create: %w", err)
	}

	return brand, nil
}

// Update replaces an brand document in the database.
func (b *Business) Update(ctx context.Context, brand Brand, ub UpdateBrand) (Brand, error) {
	ctx, span := otel.AddSpan(ctx, "business.brandbus.update")
	defer span.End()

	if ub.ContactInfo != nil {
		brand.ContactInfo = *ub.ContactInfo
	}

	if ub.ManufacturerID != nil {
		brand.ManufacturerID = *ub.ManufacturerID
	}

	if ub.Name != nil {
		brand.Name = *ub.Name
	}

	brand.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, brand); err != nil {
		return Brand{}, fmt.Errorf("update: %w", err)
	}

	return brand, nil
}

// Delete removes the specified brand.
func (b *Business) Delete(ctx context.Context, ass Brand) error {
	ctx, span := otel.AddSpan(ctx, "business.brandbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ass); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of brands from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Brand, error) {
	ctx, span := otel.AddSpan(ctx, "business.brandbus.Query")
	defer span.End()

	brands, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return brands, nil
}

// Count returns the total number of brands.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.brandbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the brand by the specified ID.
func (b *Business) QueryByID(ctx context.Context, brandID uuid.UUID) (Brand, error) {
	ctx, span := otel.AddSpan(ctx, "business.brandbus.querybyid")
	defer span.End()

	brand, err := b.storer.QueryByID(ctx, brandID)
	if err != nil {
		return Brand{}, fmt.Errorf("query: brandID[%s]: %w", brandID, err)
	}

	return brand, nil
}
