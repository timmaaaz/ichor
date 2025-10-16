// Package formbus provides business logic for form configuration management.
package formbus

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
	ErrNotFound    = errors.New("form not found")
	ErrUniqueEntry = errors.New("form entry is not unique")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, form Form) error
	Update(ctx context.Context, form Form) error
	Delete(ctx context.Context, form Form) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Form, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, formID uuid.UUID) (Form, error)
	QueryByName(ctx context.Context, name string) (Form, error)
}

// Business manages the set of APIs for form access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a form business API for use.
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

// Create inserts a new form into the database.
func (b *Business) Create(ctx context.Context, nf NewForm) (Form, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.create")
	defer span.End()

	form := Form{
		ID:   uuid.New(),
		Name: nf.Name,
	}

	if err := b.storer.Create(ctx, form); err != nil {
		return Form{}, fmt.Errorf("create: %w", err)
	}

	return form, nil
}

// Update replaces a form document in the database.
func (b *Business) Update(ctx context.Context, form Form, uf UpdateForm) (Form, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.update")
	defer span.End()

	if uf.Name != nil {
		form.Name = *uf.Name
	}

	if err := b.storer.Update(ctx, form); err != nil {
		return Form{}, fmt.Errorf("update: %w", err)
	}

	return form, nil
}

// Delete removes the specified form.
func (b *Business) Delete(ctx context.Context, form Form) error {
	ctx, span := otel.AddSpan(ctx, "business.formbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, form); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of forms from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Form, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.query")
	defer span.End()

	forms, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return forms, nil
}

// Count returns the total number of forms.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the form by the specified ID.
func (b *Business) QueryByID(ctx context.Context, formID uuid.UUID) (Form, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.querybyid")
	defer span.End()

	form, err := b.storer.QueryByID(ctx, formID)
	if err != nil {
		return Form{}, fmt.Errorf("query: formID[%s]: %w", formID, err)
	}

	return form, nil
}

// QueryByName finds the form by its unique name.
func (b *Business) QueryByName(ctx context.Context, name string) (Form, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.querybyname")
	defer span.End()

	form, err := b.storer.QueryByName(ctx, name)
	if err != nil {
		return Form{}, fmt.Errorf("query: name[%s]: %w", name, err)
	}

	return form, nil
}