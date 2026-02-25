package supplierproductbus

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
	ErrNotFound              = errors.New("supplierProduct not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("supplierProduct entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, supplierProduct SupplierProduct) error
	Update(ctx context.Context, supplierProduct SupplierProduct) error
	Delete(ctx context.Context, supplierProduct SupplierProduct) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]SupplierProduct, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, supplierProductID uuid.UUID) (SupplierProduct, error)
	QueryByIDs(ctx context.Context, supplierProductIDs []uuid.UUID) ([]SupplierProduct, error)
}

// Business manages the set of APIs for cost history access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a cost history business API for use.
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

func (b *Business) Create(ctx context.Context, nsp NewSupplierProduct) (SupplierProduct, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierproductbus.create")
	defer span.End()

	now := time.Now()

	sp := SupplierProduct{
		SupplierProductID:  uuid.New(),
		SupplierID:         nsp.SupplierID,
		ProductID:          nsp.ProductID,
		SupplierPartNumber: nsp.SupplierPartNumber,
		MinOrderQuantity:   nsp.MinOrderQuantity,
		MaxOrderQuantity:   nsp.MaxOrderQuantity,
		LeadTimeDays:       nsp.LeadTimeDays,
		UnitCost:           nsp.UnitCost,
		IsPrimarySupplier:  nsp.IsPrimarySupplier,
		UpdatedDate:        now,
		CreatedDate:        now,
	}

	if err := b.storer.Create(ctx, sp); err != nil {
		return SupplierProduct{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(sp)); err != nil {
		b.log.Error(ctx, "supplierproductbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return sp, nil
}

func (b *Business) Update(ctx context.Context, sp SupplierProduct, usp UpdateSupplierProduct) (SupplierProduct, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierproductbus.update")
	defer span.End()

	before := sp

	if usp.SupplierID != nil {
		sp.SupplierID = *usp.SupplierID
	}
	if usp.ProductID != nil {
		sp.ProductID = *usp.ProductID
	}
	if usp.SupplierPartNumber != nil {
		sp.SupplierPartNumber = *usp.SupplierPartNumber
	}
	if usp.MinOrderQuantity != nil {
		sp.MinOrderQuantity = *usp.MinOrderQuantity
	}
	if usp.MaxOrderQuantity != nil {
		sp.MaxOrderQuantity = *usp.MaxOrderQuantity
	}
	if usp.LeadTimeDays != nil {
		sp.LeadTimeDays = *usp.LeadTimeDays
	}
	if usp.UnitCost != nil {
		sp.UnitCost = *usp.UnitCost
	}
	if usp.IsPrimarySupplier != nil {
		sp.IsPrimarySupplier = *usp.IsPrimarySupplier
	}

	sp.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, sp); err != nil {
		return SupplierProduct{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, sp)); err != nil {
		b.log.Error(ctx, "supplierproductbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return sp, nil
}

func (b *Business) Delete(ctx context.Context, sp SupplierProduct) error {
	ctx, span := otel.AddSpan(ctx, "business.supplierproductbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, sp); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(sp)); err != nil {
		b.log.Error(ctx, "supplierproductbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]SupplierProduct, error) {

	ctx, span := otel.AddSpan(ctx, "business.supplierproductbus.query")
	defer span.End()

	sps, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return sps, nil
}

func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierproductbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByIDs finds the supplier products by the specified IDs.
func (b *Business) QueryByIDs(ctx context.Context, supplierProductIDs []uuid.UUID) ([]SupplierProduct, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierproductbus.querybyids")
	defer span.End()

	sps, err := b.storer.QueryByIDs(ctx, supplierProductIDs)
	if err != nil {
		return nil, fmt.Errorf("querybyids: %w", err)
	}

	return sps, nil
}

func (b *Business) QueryByID(ctx context.Context, supplierProductID uuid.UUID) (SupplierProduct, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierproductbus.querybyid")
	defer span.End()

	sp, err := b.storer.QueryByID(ctx, supplierProductID)
	if err != nil {
		return SupplierProduct{}, fmt.Errorf("query: supplierProductID[%s]: %w", supplierProductID, err)
	}

	return sp, nil
}
