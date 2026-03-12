package productuombus

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Storer defines the required persistence interface for ProductUOMs.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, uom ProductUOM) error
	Update(ctx context.Context, uom ProductUOM) error
	Delete(ctx context.Context, uom ProductUOM) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ProductUOM, error)
	QueryByID(ctx context.Context, uomID uuid.UUID) (ProductUOM, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages the set of APIs for product UOM access.
type Business struct {
	log      *logger.Logger
	delegate *delegate.Delegate
	storer   Storer
}

// NewBusiness constructs a Business for product UOM API access.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// Create adds a new product UOM to the system.
func (b *Business) Create(ctx context.Context, npu NewProductUOM) (ProductUOM, error) {
	now := time.Now()

	uom := ProductUOM{
		ID:               uuid.New(),
		ProductID:        npu.ProductID,
		Name:             npu.Name,
		Abbreviation:     npu.Abbreviation,
		ConversionFactor: npu.ConversionFactor,
		IsBase:           npu.IsBase,
		IsApproximate:    npu.IsApproximate,
		Notes:            npu.Notes,
		CreatedDate:      now,
		UpdatedDate:      now,
	}

	if err := b.storer.Create(ctx, uom); err != nil {
		return ProductUOM{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(uom)); err != nil {
		b.log.Error(ctx, "productuombus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return uom, nil
}

// Update modifies information about an existing product UOM.
func (b *Business) Update(ctx context.Context, uom ProductUOM, upu UpdateProductUOM) (ProductUOM, error) {
	before := uom

	if upu.Name != nil {
		uom.Name = *upu.Name
	}
	if upu.Abbreviation != nil {
		uom.Abbreviation = *upu.Abbreviation
	}
	if upu.ConversionFactor != nil {
		uom.ConversionFactor = *upu.ConversionFactor
	}
	if upu.IsBase != nil {
		uom.IsBase = *upu.IsBase
	}
	if upu.IsApproximate != nil {
		uom.IsApproximate = *upu.IsApproximate
	}
	if upu.Notes != nil {
		uom.Notes = *upu.Notes
	}

	uom.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, uom); err != nil {
		return ProductUOM{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, uom)); err != nil {
		b.log.Error(ctx, "productuombus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return uom, nil
}

// Delete removes a product UOM from the system.
func (b *Business) Delete(ctx context.Context, uom ProductUOM) error {
	if err := b.storer.Delete(ctx, uom); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(uom)); err != nil {
		b.log.Error(ctx, "productuombus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of product UOMs from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ProductUOM, error) {
	uoms, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return uoms, nil
}

// QueryByID finds the product UOM by the specified ID.
func (b *Business) QueryByID(ctx context.Context, uomID uuid.UUID) (ProductUOM, error) {
	uom, err := b.storer.QueryByID(ctx, uomID)
	if err != nil {
		return ProductUOM{}, fmt.Errorf("querybyid: %w", err)
	}
	return uom, nil
}

// Count returns the total number of product UOMs.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return b.storer.Count(ctx, filter)
}
