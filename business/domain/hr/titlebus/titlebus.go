package titlebus

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
	ErrNotFound              = errors.New("title not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry           = errors.New("title entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, title Title) error
	Update(ctx context.Context, title Title) error
	Delete(ctx context.Context, title Title) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Title, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, titleID uuid.UUID) (Title, error)
}

type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a new title business API for use.
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

// Create creates a new title to the system.
func (b *Business) Create(ctx context.Context, nt NewTitle) (Title, error) {
	ctx, span := otel.AddSpan(ctx, "business.titlebus.Create")
	defer span.End()

	fs := Title{
		ID:          uuid.New(),
		Name:        nt.Name,
		Description: nt.Description,
	}

	if err := b.storer.Create(ctx, fs); err != nil {
		return Title{}, fmt.Errorf("store create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(fs)); err != nil {
		b.log.Error(ctx, "titlebus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return fs, nil
}

// Update modifies information about an title.
func (b *Business) Update(ctx context.Context, fs Title, ut UpdateTitle) (Title, error) {
	ctx, span := otel.AddSpan(ctx, "business.titlebus.Update")
	defer span.End()

	if ut.Description != nil {
		fs.Description = *ut.Description
	}

	if ut.Name != nil {
		fs.Name = *ut.Name
	}

	if err := b.storer.Update(ctx, fs); err != nil {
		return Title{}, fmt.Errorf("store update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(fs)); err != nil {
		b.log.Error(ctx, "titlebus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return fs, nil
}

// Delete removes an title from the system.
func (b *Business) Delete(ctx context.Context, fs Title) error {
	ctx, span := otel.AddSpan(ctx, "business.titlebus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, fs); err != nil {
		return fmt.Errorf("store delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(fs)); err != nil {
		b.log.Error(ctx, "titlebus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query returns a list of titles
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Title, error) {
	ctx, span := otel.AddSpan(ctx, "business.titlebus.Query")
	defer span.End()

	titles, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return titles, nil
}

// Count returns the total number of titles
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.titlebus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the title by the specified ID.
func (b *Business) QueryByID(ctx context.Context, titleStatusID uuid.UUID) (Title, error) {
	ctx, span := otel.AddSpan(ctx, "business.titlebus.QueryByID")
	defer span.End()

	ts, err := b.storer.QueryByID(ctx, titleStatusID)
	if err != nil {
		return Title{}, fmt.Errorf("query: titleStatusID[%s]: %w", titleStatusID, err)
	}

	return ts, nil
}
