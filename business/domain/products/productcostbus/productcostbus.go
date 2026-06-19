package productcostbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/outbox"
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
	QueryByID(ctx context.Context, productID uuid.UUID) (ProductCost, error)
}

// Business manages the set of APIs for product cost access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
	outbox   *outbox.Writer
}

// NewBusiness constructs a product cost business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// WithOutbox returns a copy of the Business wired to the cascade outbox Writer.
// Inert until the Writer is injected at the F2 cutover (nil Writer -> Emit no-ops).
func (b *Business) WithOutbox(w *outbox.Writer) *Business {
	nb := *b
	nb.outbox = w
	return &nb
}

// NewWithTx constructs a new business value that will use the specified transaction
// in any store-related calls. It copies the receiver and overrides only the storer,
// so no field (delegate, outbox, log) is silently dropped (cf. commit 63f6b034).
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	nb := *b
	nb.storer = storer
	return &nb, nil
}

// Create inserts a new product cost into the database.
func (b *Business) Create(ctx context.Context, npc NewProductCost) (ProductCost, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.create")
	defer span.End()

	return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) (ProductCost, error) {
			now := time.Now()

			pc := ProductCost{
				ID:                uuid.New(),
				ProductID:         npc.ProductID,
				PurchaseCost:      npc.PurchaseCost,
				SellingPrice:      npc.SellingPrice,
				CurrencyID:        npc.CurrencyID,
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

			// Fire delegate event for workflow automation
			evtData := ActionCreatedData(pc)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return ProductCost{}, fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.delegate.Call(ctx, ActionCreatedData(pc)); err != nil {
				b.log.Error(ctx, "productcostbus: delegate call failed", "action", ActionCreated, "err", err)
			}

			return pc, nil
		})
}

// Update replaces an product cost document in the database.
func (b *Business) Update(ctx context.Context, pc ProductCost, upc UpdateProductCost) (ProductCost, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.update")
	defer span.End()

	return outbox.WriteAtomic(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) (ProductCost, error) {
			before := pc

			if upc.ProductID != nil {
				pc.ProductID = *upc.ProductID
			}
			if upc.PurchaseCost != nil {
				pc.PurchaseCost = *upc.PurchaseCost
			}
			if upc.SellingPrice != nil {
				pc.SellingPrice = *upc.SellingPrice
			}
			if upc.CurrencyID != nil {
				pc.CurrencyID = *upc.CurrencyID
			}
			if upc.MSRP != nil {
				pc.MSRP = *upc.MSRP
			}
			if upc.MarkupPercentage != nil {
				pc.MarkupPercentage = *upc.MarkupPercentage
			}
			if upc.LandedCost != nil {
				pc.LandedCost = *upc.LandedCost
			}
			if upc.CarryingCost != nil {
				pc.CarryingCost = *upc.CarryingCost
			}
			if upc.ABCClassification != nil {
				pc.ABCClassification = *upc.ABCClassification
			}
			if upc.DepreciationValue != nil {
				pc.DepreciationValue = *upc.DepreciationValue
			}
			if upc.InsuranceValue != nil {
				pc.InsuranceValue = *upc.InsuranceValue
			}
			if upc.EffectiveDate != nil {
				pc.EffectiveDate = *upc.EffectiveDate
			}

			pc.UpdatedDate = time.Now()

			if err := b.storer.Update(ctx, pc); err != nil {
				return ProductCost{}, fmt.Errorf("update: %w", err)
			}

			// Fire delegate event for workflow automation
			evtData := ActionUpdatedData(before, pc)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return ProductCost{}, fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.delegate.Call(ctx, ActionUpdatedData(before, pc)); err != nil {
				b.log.Error(ctx, "productcostbus: delegate call failed", "action", ActionUpdated, "err", err)
			}

			return pc, nil
		})
}

// Delete removes the specified product cost.
func (b *Business) Delete(ctx context.Context, pc ProductCost) error {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.delete")
	defer span.End()

	return outbox.WriteAtomicVoid(ctx, b.outbox, b, (*Business).NewWithTx,
		func(ctx context.Context, b *Business) error {
			if err := b.storer.Delete(ctx, pc); err != nil {
				return fmt.Errorf("delete: %w", err)
			}

			// Fire delegate event for workflow automation
			evtData := ActionDeletedData(pc)
			if err := b.outbox.Emit(ctx, evtData); err != nil {
				return fmt.Errorf("emit cascade event: %w", err)
			}
			if err := b.delegate.Call(ctx, ActionDeletedData(pc)); err != nil {
				b.log.Error(ctx, "productcostbus: delegate call failed", "action", ActionDeleted, "err", err)
			}

			return nil
		})
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
func (b *Business) QueryByID(ctx context.Context, productID uuid.UUID) (ProductCost, error) {
	ctx, span := otel.AddSpan(ctx, "business.productcostbus.querybyid")
	defer span.End()

	pc, err := b.storer.QueryByID(ctx, productID)
	if err != nil {
		return ProductCost{}, fmt.Errorf("query: productID[%s]: %w", productID, err)
	}

	return pc, nil
}
