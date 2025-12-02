package supplierbus

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
	ErrNotFound              = errors.New("supplier not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("supplier entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, supplier Supplier) error
	Update(ctx context.Context, supplier Supplier) error
	Delete(ctx context.Context, supplier Supplier) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Supplier, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, supplierID uuid.UUID) (Supplier, error)
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

// Create creates a new supplier in the database
func (b *Business) Create(ctx context.Context, newSupplier NewSupplier) (Supplier, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierbus.create")
	defer span.End()

	now := time.Now()

	supplier := Supplier{
		SupplierID:     uuid.New(),
		Name:           newSupplier.Name,
		ContactInfosID: newSupplier.ContactInfosID,
		IsActive:       newSupplier.IsActive,
		PaymentTerms:   newSupplier.PaymentTerms,
		LeadTimeDays:   newSupplier.LeadTimeDays,
		Rating:         newSupplier.Rating,
		CreatedDate:    now,
		UpdatedDate:    now,
	}

	if err := b.storer.Create(ctx, supplier); err != nil {
		return Supplier{}, fmt.Errorf("create: %w", err)
	}

	return supplier, nil
}

// Update modifies a supplier in the database
func (b *Business) Update(ctx context.Context, supplier Supplier, us UpdateSupplier) (Supplier, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierbus.update")
	defer span.End()

	if us.ContactInfosID != nil {
		supplier.ContactInfosID = *us.ContactInfosID
	}
	if us.Name != nil {
		supplier.Name = *us.Name
	}
	if us.PaymentTerms != nil {
		supplier.PaymentTerms = *us.PaymentTerms
	}
	if us.LeadTimeDays != nil {
		supplier.LeadTimeDays = *us.LeadTimeDays
	}
	if us.Rating != nil {
		supplier.Rating = *us.Rating
	}
	if us.IsActive != nil {
		supplier.IsActive = *us.IsActive
	}

	supplier.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, supplier); err != nil {
		return Supplier{}, fmt.Errorf("update: %w", err)
	}

	return supplier, nil

}

// Delete removes the specified brand.
func (b *Business) Delete(ctx context.Context, supplier Supplier) error {
	ctx, span := otel.AddSpan(ctx, "business.supplierbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, supplier); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of product costs from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Supplier, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierbus.Query")
	defer span.End()

	pcs, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return pcs, nil
}

// Count returns the total number of product costs.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the product cost by the specified ID.
func (b *Business) QueryByID(ctx context.Context, supplierID uuid.UUID) (Supplier, error) {
	ctx, span := otel.AddSpan(ctx, "business.supplierbus.querybyid")
	defer span.End()

	supplier, err := b.storer.QueryByID(ctx, supplierID)
	if err != nil {
		return Supplier{}, fmt.Errorf("query: supplierID[%s]: %w", supplierID, err)
	}

	return supplier, nil
}
