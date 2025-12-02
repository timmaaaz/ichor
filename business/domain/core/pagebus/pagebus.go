package pagebus

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
	ErrNotFound              = errors.New("page not found")
	ErrUnique                = errors.New("not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, page Page) error
	Update(ctx context.Context, page Page) error
	Delete(ctx context.Context, page Page) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Page, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, pageID uuid.UUID) (Page, error)
	QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Page, error)
}

// Business manages the set of APIs for page access.
type Business struct {
	log    *logger.Logger
	storer Storer
	del    *delegate.Delegate
}

// NewBusiness constructs a page business API for use.
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

// Create adds a new page to the system.
func (b *Business) Create(ctx context.Context, np NewPage) (Page, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagebus.create")
	defer span.End()

	page := Page{
		ID:         uuid.New(),
		Path:       np.Path,
		Name:       np.Name,
		Module:     np.Module,
		Icon:       np.Icon,
		SortOrder:  np.SortOrder,
		IsActive:   np.IsActive,
		ShowInMenu: np.ShowInMenu,
	}

	if err := b.storer.Create(ctx, page); err != nil {
		return Page{}, fmt.Errorf("creating page: %w", err)
	}

	return page, nil
}

// Update modifies a page in the system.
func (b *Business) Update(ctx context.Context, page Page, up UpdatePage) (Page, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagebus.update")
	defer span.End()

	if up.Path != nil {
		page.Path = *up.Path
	}
	if up.Name != nil {
		page.Name = *up.Name
	}
	if up.Module != nil {
		page.Module = *up.Module
	}
	if up.Icon != nil {
		page.Icon = *up.Icon
	}
	if up.SortOrder != nil {
		page.SortOrder = *up.SortOrder
	}
	if up.IsActive != nil {
		page.IsActive = *up.IsActive
	}
	if up.ShowInMenu != nil {
		page.ShowInMenu = *up.ShowInMenu
	}

	if err := b.storer.Update(ctx, page); err != nil {
		return Page{}, fmt.Errorf("updating page: %w", err)
	}

	// Inform subscribers of the update
	if err := b.del.Call(ctx, ActionUpdatedData(page)); err != nil {
		return Page{}, fmt.Errorf("calling delegate: %w", err)
	}

	return page, nil
}

// Delete removes a page from the system.
func (b *Business) Delete(ctx context.Context, page Page) error {
	ctx, span := otel.AddSpan(ctx, "business.pagebus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, page); err != nil {
		return fmt.Errorf("deleting page: %w", err)
	}

	// Inform subscribers of the deletion
	if err := b.del.Call(ctx, ActionDeletedData(page.ID)); err != nil {
		return fmt.Errorf("calling delegate: %w", err)
	}

	return nil
}

// Query retrieves a list of pages from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Page, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagebus.query")
	defer span.End()

	pages, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("querying pages: %w", err)
	}

	return pages, nil
}

// Count returns the total number of pages.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagebus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the page by the specified ID.
func (b *Business) QueryByID(ctx context.Context, pageID uuid.UUID) (Page, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagebus.querybyid")
	defer span.End()

	page, err := b.storer.QueryByID(ctx, pageID)
	if err != nil {
		return Page{}, fmt.Errorf("querying page: pageID[%s]: %w", pageID, err)
	}

	return page, nil
}

// QueryByUserID retrieves all pages accessible to a specific user based on their roles.
func (b *Business) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Page, error) {
	ctx, span := otel.AddSpan(ctx, "business.pagebus.querybyuserid")
	defer span.End()

	pages, err := b.storer.QueryByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("querying pages by user: userID[%s]: %w", userID, err)
	}

	return pages, nil
}
