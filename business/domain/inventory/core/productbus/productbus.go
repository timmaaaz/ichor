package productbus

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
	"github.com/timmaaaz/ichor/foundation/convert"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("product not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("product entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, product Product) error
	Update(ctx context.Context, product Product) error
	Delete(ctx context.Context, product Product) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Product, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, productID uuid.UUID) (Product, error)
}

// Business manages the set of APIs for product access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a product business API for use.
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

// Create inserts a new product into the database.
func (b *Business) Create(ctx context.Context, np NewProduct) (Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.create")
	defer span.End()

	now := time.Now()

	product := Product{
		ProductID:            uuid.New(),
		Name:                 np.Name,
		Description:          np.Description,
		SKU:                  np.SKU,
		BrandID:              np.BrandID,
		ProductCategoryID:    np.ProductCategoryID,
		ModelNumber:          np.ModelNumber,
		UpcCode:              np.UpcCode,
		Status:               np.Status,
		IsActive:             np.IsActive,
		IsPerishable:         np.IsPerishable,
		HandlingInstructions: np.HandlingInstructions,
		UnitsPerCase:         np.UnitsPerCase,
		CreatedDate:          now,
		UpdatedDate:          now,
	}

	err := b.storer.Create(ctx, product)
	if err != nil {
		return Product{}, fmt.Errorf("create: %w", err)
	}

	return product, nil
}

// Update replaces an product document in the database.
func (b *Business) Update(ctx context.Context, product Product, ub UpdateProduct) (Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.update")
	defer span.End()

	product.UpdatedDate = time.Now()

	if err := convert.PopulateSameTypes(ub, &product); err != nil {
		return Product{}, fmt.Errorf("populate product struct: %w", err)
	}

	if err := b.storer.Update(ctx, product); err != nil {
		return Product{}, fmt.Errorf("update: %w", err)
	}

	return product, nil
}

// Delete removes the specified product.
func (b *Business) Delete(ctx context.Context, product Product) error {
	ctx, span := otel.AddSpan(ctx, "business.productbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, product); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of products from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.Query")
	defer span.End()

	products, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return products, nil
}

// Count returns the total number of products.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the product by the specified ID.
func (b *Business) QueryByID(ctx context.Context, productID uuid.UUID) (Product, error) {
	ctx, span := otel.AddSpan(ctx, "business.productbus.querybyid")
	defer span.End()

	product, err := b.storer.QueryByID(ctx, productID)
	if err != nil {
		return Product{}, fmt.Errorf("query: productID[%s]: %w", productID, err)
	}

	return product, nil
}
