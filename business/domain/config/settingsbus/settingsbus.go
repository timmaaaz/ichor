package settingsbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound            = errors.New("setting not found")
	ErrUniqueEntry         = errors.New("setting entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, setting Setting) error
	Update(ctx context.Context, setting Setting) error
	Delete(ctx context.Context, setting Setting) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Setting, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByKey(ctx context.Context, key string) (Setting, error)
}

// Business manages the set of APIs for settings access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a settings business API for use.
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

// Create creates a new setting.
func (b *Business) Create(ctx context.Context, ns NewSetting) (Setting, error) {
	ctx, span := otel.AddSpan(ctx, "business.settingsbus.create")
	defer span.End()

	now := time.Now()

	s := Setting{
		Key:         ns.Key,
		Value:       ns.Value,
		Description: ns.Description,
		CreatedDate: now,
		UpdatedDate: now,
	}

	if err := b.storer.Create(ctx, s); err != nil {
		return Setting{}, fmt.Errorf("create: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionCreatedData(s)); err != nil {
		b.log.Error(ctx, "settingsbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return s, nil
}

// Update updates an existing setting.
func (b *Business) Update(ctx context.Context, s Setting, u UpdateSetting) (Setting, error) {
	ctx, span := otel.AddSpan(ctx, "business.settingsbus.update")
	defer span.End()

	before := s

	if u.Value != nil {
		s.Value = u.Value
	}
	if u.Description != nil {
		s.Description = *u.Description
	}

	s.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, s); err != nil {
		return Setting{}, fmt.Errorf("update: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, s)); err != nil {
		b.log.Error(ctx, "settingsbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return s, nil
}

// Delete deletes an existing setting.
func (b *Business) Delete(ctx context.Context, s Setting) error {
	ctx, span := otel.AddSpan(ctx, "business.settingsbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, s); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionDeletedData(s)); err != nil {
		b.log.Error(ctx, "settingsbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves settings based on the provided filter, order, and page.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Setting, error) {
	ctx, span := otel.AddSpan(ctx, "business.settingsbus.query")
	defer span.End()

	settings, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return settings, nil
}

// Count returns the total number of settings that match the provided filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.settingsbus.count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByKey retrieves a setting by its key.
func (b *Business) QueryByKey(ctx context.Context, key string) (Setting, error) {
	ctx, span := otel.AddSpan(ctx, "business.settingsbus.querybykey")
	defer span.End()

	s, err := b.storer.QueryByKey(ctx, key)
	if err != nil {
		return Setting{}, fmt.Errorf("query by key: %w", err)
	}

	return s, nil
}
