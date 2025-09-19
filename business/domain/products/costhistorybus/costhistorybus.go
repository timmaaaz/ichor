package costhistorybus

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
	ErrNotFound              = errors.New("costHistory not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("costHistory entry is not unique")
	ErrForeignKeyViolation   = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, costHistory CostHistory) error
	Update(ctx context.Context, costHistory CostHistory) error
	Delete(ctx context.Context, costHistory CostHistory) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CostHistory, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, costHistoryID uuid.UUID) (CostHistory, error)
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

// Create inserts a new cost history
func (b *Business) Create(ctx context.Context, nch NewCostHistory) (CostHistory, error) {
	ctx, span := otel.AddSpan(ctx, "business.costhistorybus.create")
	defer span.End()

	now := time.Now()

	ch := CostHistory{
		CostHistoryID: uuid.New(),
		ProductID:     nch.ProductID,
		CostType:      nch.CostType,
		Amount:        nch.Amount,
		Currency:      nch.Currency,
		EffectiveDate: nch.EffectiveDate,
		EndDate:       nch.EndDate,
		CreatedDate:   now,
		UpdatedDate:   now,
	}

	if err := b.storer.Create(ctx, ch); err != nil {
		return CostHistory{}, fmt.Errorf("create: %w", err)
	}

	return ch, nil
}

// Update replaces a cost history document
func (b *Business) Update(ctx context.Context, ch CostHistory, uch UpdateCostHistory) (CostHistory, error) {
	ctx, span := otel.AddSpan(ctx, "business.costhistorybus.update")
	defer span.End()

	err := convert.PopulateSameTypes(uch, &ch)
	if err != nil {
		return CostHistory{}, fmt.Errorf("populate cost history struct: %w", err)
	}

	if uch.Amount != nil {
		ch.Amount = *uch.Amount
	}

	ch.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, ch); err != nil {
		return CostHistory{}, fmt.Errorf("update: %w", err)
	}

	return ch, nil
}

// Delete removes a cost history from the database
func (b *Business) Delete(ctx context.Context, ch CostHistory) error {
	ctx, span := otel.AddSpan(ctx, "business.costhistorybus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, ch); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of product costs from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CostHistory, error) {
	ctx, span := otel.AddSpan(ctx, "business.costhistorybus.Query")
	defer span.End()

	chs, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return chs, nil
}

// Count returns the total number of product costs.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.costhistorybus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the product cost by the specified ID.
func (b *Business) QueryByID(ctx context.Context, costHistoryID uuid.UUID) (CostHistory, error) {
	ctx, span := otel.AddSpan(ctx, "business.costhistorybus.querybyid")
	defer span.End()

	ch, err := b.storer.QueryByID(ctx, costHistoryID)
	if err != nil {
		return CostHistory{}, fmt.Errorf("query: costHistoryID[%s]: %w", costHistoryID, err)
	}

	return ch, nil
}
