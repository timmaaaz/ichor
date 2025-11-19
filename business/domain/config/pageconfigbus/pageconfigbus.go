package pageconfigbus

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
	ErrNotFound = errors.New("page config not found")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, config PageConfig) error
	Update(ctx context.Context, config PageConfig) error
	Delete(ctx context.Context, configID uuid.UUID) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageReq page.Page) ([]PageConfig, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, configID uuid.UUID) (PageConfig, error)
	QueryByName(ctx context.Context, name string) (PageConfig, error)
	QueryByNameAndUserID(ctx context.Context, name string, userID uuid.UUID) (PageConfig, error)
}

// Business manages the set of APIs for page config access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
}

// NewBusiness constructs a page config business API for use.
func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
		del:    del,
	}
}

// NewWithTx constructs a new Business value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
		del:    b.del,
	}

	return &bus, nil
}

// Create adds a new page configuration to the system.
func (b *Business) Create(ctx context.Context, nc NewPageConfig) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Create")
	defer span.End()

	// If is_default is true, ensure user_id is zero (will be NULL in database)
	userID := nc.UserID
	if nc.IsDefault {
		userID = uuid.UUID{}
	}

	config := PageConfig{
		ID:        uuid.New(),
		Name:      nc.Name,
		UserID:    userID,
		IsDefault: nc.IsDefault,
	}

	if err := b.storer.Create(ctx, config); err != nil {
		return PageConfig{}, fmt.Errorf("create: %w", err)
	}

	return config, nil
}

// Update modifies an existing page configuration.
func (b *Business) Update(ctx context.Context, uc UpdatePageConfig, configID uuid.UUID) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Update")
	defer span.End()

	// Fetch existing config
	config, err := b.storer.QueryByID(ctx, configID)
	if err != nil {
		return PageConfig{}, fmt.Errorf("query: %w", err)
	}

	// Apply updates
	if uc.Name != nil {
		config.Name = *uc.Name
	}
	if uc.UserID != nil {
		config.UserID = *uc.UserID
	}
	if uc.IsDefault != nil {
		config.IsDefault = *uc.IsDefault
	}

	// If is_default is true, ensure user_id is zero
	if config.IsDefault {
		config.UserID = uuid.UUID{}
	}

	if err := b.storer.Update(ctx, config); err != nil {
		return PageConfig{}, fmt.Errorf("update: %w", err)
	}

	return config, nil
}

// Delete removes a page configuration from the system.
func (b *Business) Delete(ctx context.Context, configID uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, configID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of page configurations based on filters.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageReq page.Page) ([]PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Query")
	defer span.End()

	configs, err := b.storer.Query(ctx, filter, orderBy, pageReq)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return configs, nil
}

// Count returns the total number of page configurations matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.Count")
	defer span.End()

	count, err := b.storer.Count(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	return count, nil
}

// QueryByID finds a page configuration by its ID.
func (b *Business) QueryByID(ctx context.Context, configID uuid.UUID) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.QueryByID")
	defer span.End()

	config, err := b.storer.QueryByID(ctx, configID)
	if err != nil {
		return PageConfig{}, fmt.Errorf("query: %w", err)
	}

	return config, nil
}

// QueryByName retrieves the default page configuration by name.
// This returns the default page config that serves as a fallback for all users.
func (b *Business) QueryByName(ctx context.Context, name string) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.QueryByName")
	defer span.End()

	config, err := b.storer.QueryByName(ctx, name)
	if err != nil {
		return PageConfig{}, fmt.Errorf("query: %w", err)
	}

	return config, nil
}

// QueryByNameAndUserID retrieves a user-specific page configuration.
func (b *Business) QueryByNameAndUserID(ctx context.Context, name string, userID uuid.UUID) (PageConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.QueryByNameAndUserID")
	defer span.End()

	config, err := b.storer.QueryByNameAndUserID(ctx, name, userID)
	if err != nil {
		return PageConfig{}, fmt.Errorf("query: %w", err)
	}

	return config, nil
}
