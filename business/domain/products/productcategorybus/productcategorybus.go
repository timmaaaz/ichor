package productcategorybus

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
	ErrNotFound              = errors.New("product category not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("product category entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, pc ProductCategory) error
	Update(ctx context.Context, pc ProductCategory) error
	Delete(ctx context.Context, pc ProductCategory) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ProductCategory, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, pcID uuid.UUID) (ProductCategory, error)
}

// Business manages the set of APIs for product category access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a product category business API for use.
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

// Create inserts a new pc into the database.
func (b *Business) Create(ctx context.Context, npc NewProductCategory) (ProductCategory, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcategory.create")
	defer span.End()

	now := time.Now()

	pc := ProductCategory{
		ProductCategoryID: uuid.New(),
		Name:              npc.Name,
		Description:       npc.Description,
		CreatedDate:       now,
		UpdatedDate:       now,
	}

	if err := b.storer.Create(ctx, pc); err != nil {
		return ProductCategory{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(pc)); err != nil {
		b.log.Error(ctx, "productcategorybus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return pc, nil
}

// Update replaces an pc document in the database.
func (b *Business) Update(ctx context.Context, pc ProductCategory, upc UpdateProductCategory) (ProductCategory, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcategory.update")
	defer span.End()

	before := pc

	if upc.Description != nil {
		pc.Description = *upc.Description
	}

	if upc.Name != nil {
		pc.Name = *upc.Name
	}

	pc.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, pc); err != nil {
		return ProductCategory{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, pc)); err != nil {
		b.log.Error(ctx, "productcategorybus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return pc, nil
}

// Delete removes the specified pc.
func (b *Business) Delete(ctx context.Context, ass ProductCategory) error {
	ctx, span := otel.AddSpan(ctx, "business.productcategory.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ass); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(ass)); err != nil {
		b.log.Error(ctx, "productcategorybus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of pcs from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ProductCategory, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcategory.Query")
	defer span.End()

	pcs, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return pcs, nil
}

// Count returns the total number of pcs.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcategory.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the pc by the specified ID.
func (b *Business) QueryByID(ctx context.Context, pcID uuid.UUID) (ProductCategory, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcategory.querybyid")
	defer span.End()

	pc, err := b.storer.QueryByID(ctx, pcID)
	if err != nil {
		return ProductCategory{}, fmt.Errorf("query: product category ID[%s]: %w", pcID, err)
	}

	return pc, nil
}
