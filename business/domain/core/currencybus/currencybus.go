package currencybus

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
	ErrNotFound              = errors.New("currency not found")
	ErrUnique                = errors.New("not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, currency Currency) error
	Update(ctx context.Context, currency Currency) error
	Delete(ctx context.Context, currency Currency) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Currency, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, currencyID uuid.UUID) (Currency, error)
	QueryByIDs(ctx context.Context, currencyIDs []uuid.UUID) ([]Currency, error)
	QueryAll(ctx context.Context) ([]Currency, error)
}

// Business manages the set of APIs for currency access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
}

// NewBusiness constructs a currency business API for use.
func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:    log,
		del:    del,
		storer: storer,
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
		log:    b.log,
		storer: storer,
	}

	return &bus, nil
}

// Create adds a new currency to the system.
func (b *Business) Create(ctx context.Context, nc NewCurrency) (Currency, error) {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.create")
	defer span.End()

	now := time.Now()

	currency := Currency{
		ID:            uuid.New(),
		Code:          nc.Code,
		Name:          nc.Name,
		Symbol:        nc.Symbol,
		Locale:        nc.Locale,
		DecimalPlaces: nc.DecimalPlaces,
		IsActive:      nc.IsActive,
		SortOrder:     nc.SortOrder,
		CreatedBy:     nc.CreatedBy,
		CreatedDate:   now,
		UpdatedBy:     nc.CreatedBy,
		UpdatedDate:   now,
	}

	if err := b.storer.Create(ctx, currency); err != nil {
		return Currency{}, fmt.Errorf("creating currency: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionCreatedData(currency)); err != nil {
		b.log.Error(ctx, "currencybus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return currency, nil
}

// Update modifies a currency in the system.
func (b *Business) Update(ctx context.Context, currency Currency, uc UpdateCurrency) (Currency, error) {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.update")
	defer span.End()

	before := currency

	if uc.Code != nil {
		currency.Code = *uc.Code
	}
	if uc.Name != nil {
		currency.Name = *uc.Name
	}
	if uc.Symbol != nil {
		currency.Symbol = *uc.Symbol
	}
	if uc.Locale != nil {
		currency.Locale = *uc.Locale
	}
	if uc.DecimalPlaces != nil {
		currency.DecimalPlaces = *uc.DecimalPlaces
	}
	if uc.IsActive != nil {
		currency.IsActive = *uc.IsActive
	}
	if uc.SortOrder != nil {
		currency.SortOrder = *uc.SortOrder
	}

	currency.UpdatedBy = uc.UpdatedBy
	currency.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, currency); err != nil {
		return Currency{}, fmt.Errorf("updating currency: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionUpdatedData(before, currency)); err != nil {
		b.log.Error(ctx, "currencybus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return currency, nil
}

// Delete removes a currency from the system.
func (b *Business) Delete(ctx context.Context, currency Currency) error {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, currency); err != nil {
		return fmt.Errorf("deleting currency: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.del.Call(ctx, ActionDeletedData(currency)); err != nil {
		b.log.Error(ctx, "currencybus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of currencies from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Currency, error) {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.query")
	defer span.End()

	currencies, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying currencies: %w", err)
	}

	return currencies, nil
}

// Count returns the total number of currencies.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the currency by the specified ID.
func (b *Business) QueryByID(ctx context.Context, currencyID uuid.UUID) (Currency, error) {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.querybyid")
	defer span.End()

	currency, err := b.storer.QueryByID(ctx, currencyID)
	if err != nil {
		return Currency{}, fmt.Errorf("querying currency: currencyID[%s]: %w", currencyID, err)
	}

	return currency, nil
}

// QueryByIDs finds the currencies by the specified IDs.
func (b *Business) QueryByIDs(ctx context.Context, currencyIDs []uuid.UUID) ([]Currency, error) {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.querybyids")
	defer span.End()

	currencies, err := b.storer.QueryByIDs(ctx, currencyIDs)
	if err != nil {
		return nil, fmt.Errorf("querying currencies: currencyIDs[%s]: %w", currencyIDs, err)
	}

	return currencies, nil
}

// QueryAll retrieves all currencies from the system.
func (b *Business) QueryAll(ctx context.Context) ([]Currency, error) {
	ctx, span := otel.AddSpan(ctx, "business.currencybus.queryall")
	defer span.End()

	currencies, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying all currencies: %w", err)
	}

	return currencies, nil
}
