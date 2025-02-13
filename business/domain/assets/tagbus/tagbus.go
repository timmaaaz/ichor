package tagbus

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
	ErrNotFound              = errors.New("tag not found")
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrUniqueEntry           = errors.New("tag entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, tag Tag) error
	Update(ctx context.Context, tag Tag) error
	Delete(ctx context.Context, tag Tag) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Tag, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, tagID uuid.UUID) (Tag, error)
}

// Business manages the set of APIs for tag access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a tag business API for use.
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

// Create adds a new tag to the system.
func (b *Business) Create(ctx context.Context, nt NewTag) (Tag, error) {
	ctx, span := otel.AddSpan(ctx, "business.tagbus.Create")
	defer span.End()

	t := Tag{
		ID:          uuid.New(),
		Name:        nt.Name,
		Description: nt.Description,
	}

	if err := b.storer.Create(ctx, t); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return Tag{}, fmt.Errorf("create: %w", ErrUniqueEntry)
		}
		return Tag{}, err
	}

	return t, nil
}

// Update updates an existing tag.
func (b *Business) Update(ctx context.Context, t Tag, ut UpdateTag) (Tag, error) {
	ctx, span := otel.AddSpan(ctx, "business.tagbus.Update")
	defer span.End()

	if ut.Name != nil {
		t.Name = *ut.Name
	}

	if ut.Description != nil {
		t.Description = *ut.Description
	}

	if err := b.storer.Update(ctx, t); err != nil {
		if errors.Is(err, ErrUniqueEntry) {
			return Tag{}, fmt.Errorf("update: %w", ErrUniqueEntry)
		}
		return Tag{}, fmt.Errorf("update: %w", err)
	}

	return t, nil
}

// Delete removes an tag from the system.
func (b *Business) Delete(ctx context.Context, at Tag) error {
	ctx, span := otel.AddSpan(ctx, "business.tagbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, at); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing tags from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Tag, error) {
	ctx, span := otel.AddSpan(ctx, "business.tagbus.Query")
	defer span.End()

	tags, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return tags, nil
}

// Count returns the total number of tags.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.tagbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the tag by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Tag, error) {
	ctx, span := otel.AddSpan(ctx, "business.tagbus.QueryByID")
	defer span.End()

	tag, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Tag{}, fmt.Errorf("query: tagID[%s]: %w", id, err)
	}

	return tag, nil
}
