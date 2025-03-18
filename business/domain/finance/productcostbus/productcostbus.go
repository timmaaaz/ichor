package productcostbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("productCost not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("productCost entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, productCost ProductCost) error
	Update(ctx context.Context, productCost ProductCost) error
	Delete(ctx context.Context, productCost ProductCost) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ProductCost, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, productCostID uuid.UUID) (ProductCost, error)
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
func (b *Business) Create(ctx context.Context, npc NewProductCost) (ProductCost, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.create")
	defer span.End()

	now := time.Now()

	pc := ProductCost{
		CostID:            uuid.New(),
		ProductID:         npc.ProductID,
		PurchaseCost:      npc.PurchaseCost,
		SellingPrice:      npc.SellingPrice,
		Currency:          npc.Currency,
		MSRP:              npc.MSRP,
		MarkupPercentage:  npc.MarkupPercentage,
		LandedCost:        npc.LandedCost,
		CarryingCost:      npc.CarryingCost,
		ABCClassification: npc.ABCClassification,
		DepreciationValue: npc.DepreciationValue,
		InsuranceValue:    npc.InsuranceValue,
		EffectiveDate:     npc.EffectiveDate,
		CreatedDate:       now,
		UpdatedDate:       now,
	}

	if err := b.storer.Create(ctx, pc); err != nil {
		return ProductCost{}, fmt.Errorf("create: %w", err)
	}

	return pc, nil
}

// Update replaces an brand document in the database.
func (b *Business) Update(ctx context.Context, pc ProductCost, upc UpdateProductCost) (ProductCost, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.update")
	defer span.End()

	err := convert.PopulateSameTypes(upc, &pc)
	if err != nil {
		return ProductCost{}, fmt.Errorf("populate product cost from update product cost: %w", err)
	}

	if upc.SellingPrice != nil {
		pc.SellingPrice = *upc.SellingPrice
	}

	if upc.PurchaseCost != nil {
		pc.PurchaseCost = *upc.PurchaseCost
	}

	if upc.MSRP != nil {
		pc.MSRP = *upc.MSRP
	}

	if upc.CarryingCost != nil {
		pc.CarryingCost = *upc.CarryingCost
	}

	if upc.InsuranceValue != nil {
		pc.InsuranceValue = *upc.InsuranceValue
	}

	if upc.LandedCost != nil {
		pc.LandedCost = *upc.LandedCost
	}

	pc.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, pc); err != nil {
		return ProductCost{}, fmt.Errorf("update: %w", err)
	}

	return pc, nil
}

// Delete removes the specified brand.
func (b *Business) Delete(ctx context.Context, pc ProductCost) error {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, pc); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of product costs from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ProductCost, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.Query")
	defer span.End()

	pcs, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return pcs, nil
}

// Count returns the total number of product costs.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the product cost by the specified ID.
func (b *Business) QueryByID(ctx context.Context, productCostID uuid.UUID) (ProductCost, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.querybyid")
	defer span.End()

	pc, err := b.storer.QueryByID(ctx, productCostID)
	if err != nil {
		return ProductCost{}, fmt.Errorf("query: productCostID[%s]: %w", productCostID, err)
	}

	return pc, nil
}
