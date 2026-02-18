package timezonebus

import (
	"context"
	"errors"
	"fmt"

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
	ErrNotFound              = errors.New("timezone not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("timezone entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, tz Timezone) error
	Update(ctx context.Context, tz Timezone) error
	Delete(ctx context.Context, tz Timezone) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Timezone, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, timezoneID uuid.UUID) (Timezone, error)
	QueryAll(ctx context.Context) ([]Timezone, error)
}

// Business manages the set of APIs for timezone access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a timezone business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
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
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
	}

	return &bus, nil
}

// Create adds a new timezone to the system.
func (b *Business) Create(ctx context.Context, ntz NewTimezone) (Timezone, error) {
	ctx, span := otel.AddSpan(ctx, "business.timezonebus.Create")
	defer span.End()

	tz := Timezone{
		ID:          uuid.New(),
		Name:        ntz.Name,
		DisplayName: ntz.DisplayName,
		UTCOffset:   ntz.UTCOffset,
		IsActive:    ntz.IsActive,
	}

	if err := b.storer.Create(ctx, tz); err != nil {
		return Timezone{}, fmt.Errorf("store create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(tz)); err != nil {
		b.log.Error(ctx, "timezonebus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return tz, nil
}

// Update modifies information about a timezone.
func (b *Business) Update(ctx context.Context, tz Timezone, utz UpdateTimezone) (Timezone, error) {
	ctx, span := otel.AddSpan(ctx, "business.timezonebus.Update")
	defer span.End()

	before := tz

	if utz.Name != nil {
		tz.Name = *utz.Name
	}

	if utz.DisplayName != nil {
		tz.DisplayName = *utz.DisplayName
	}

	if utz.UTCOffset != nil {
		tz.UTCOffset = *utz.UTCOffset
	}

	if utz.IsActive != nil {
		tz.IsActive = *utz.IsActive
	}

	if err := b.storer.Update(ctx, tz); err != nil {
		return Timezone{}, fmt.Errorf("store update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, tz)); err != nil {
		b.log.Error(ctx, "timezonebus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return tz, nil
}

// Delete removes a timezone from the system.
func (b *Business) Delete(ctx context.Context, tz Timezone) error {
	ctx, span := otel.AddSpan(ctx, "business.timezonebus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, tz); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(tz)); err != nil {
		b.log.Error(ctx, "timezonebus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of existing timezones.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Timezone, error) {
	ctx, span := otel.AddSpan(ctx, "business.timezonebus.Query")
	defer span.End()

	tzs, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return tzs, nil
}

// Count returns the total number of timezones.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.timezonebus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the timezone by the specified ID.
func (b *Business) QueryByID(ctx context.Context, timezoneID uuid.UUID) (Timezone, error) {
	ctx, span := otel.AddSpan(ctx, "business.timezonebus.QueryByID")
	defer span.End()

	tz, err := b.storer.QueryByID(ctx, timezoneID)
	if err != nil {
		return Timezone{}, fmt.Errorf("query: timezoneID[%s]: %w", timezoneID, err)
	}

	return tz, nil
}

// QueryAll retrieves all timezones from the system.
func (b *Business) QueryAll(ctx context.Context) ([]Timezone, error) {
	ctx, span := otel.AddSpan(ctx, "business.timezonebus.QueryAll")
	defer span.End()

	tzs, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("queryall: %w", err)
	}

	return tzs, nil
}
