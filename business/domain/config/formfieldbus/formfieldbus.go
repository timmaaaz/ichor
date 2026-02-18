// Package formfieldbus provides business logic for form field configuration management.
package formfieldbus

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
	ErrNotFound             = errors.New("form field not found")
	ErrUniqueEntry          = errors.New("form field entry is not unique")
	ErrForeignKeyViolation  = errors.New("foreign key violation")
	ErrNonexistentTableName = errors.New("table does not exist in the specified schema")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, field FormField) error
	Update(ctx context.Context, field FormField) error
	Delete(ctx context.Context, field FormField) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]FormField, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, fieldID uuid.UUID) (FormField, error)
	QueryByFormID(ctx context.Context, formID uuid.UUID) ([]FormField, error)
}

// Business manages the set of APIs for form field access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a form field business API for use.
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

// Create inserts a new form field into the database.
func (b *Business) Create(ctx context.Context, nff NewFormField) (FormField, error) {
	ctx, span := otel.AddSpan(ctx, "business.formfieldbus.create")
	defer span.End()

	field := FormField{
		ID:           uuid.New(),
		FormID:       nff.FormID,
		EntityID:     nff.EntityID,
		EntitySchema: nff.EntitySchema,
		EntityTable:  nff.EntityTable,
		Name:         nff.Name,
		Label:        nff.Label,
		FieldType:    nff.FieldType,
		FieldOrder:   nff.FieldOrder,
		Required:     nff.Required,
		Config:       nff.Config,
	}

	if err := b.storer.Create(ctx, field); err != nil {
		return FormField{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(field)); err != nil {
		b.log.Error(ctx, "formfieldbus: delegate call failed", "action", ActionCreated, "err", err)
	}

	return field, nil
}

// Update replaces a form field document in the database.
func (b *Business) Update(ctx context.Context, field FormField, uff UpdateFormField) (FormField, error) {
	ctx, span := otel.AddSpan(ctx, "business.formfieldbus.update")
	defer span.End()

	before := field

	if uff.FormID != nil {
		field.FormID = *uff.FormID
	}

	if uff.EntityID != nil {
		field.EntityID = *uff.EntityID
	}

	if uff.Name != nil {
		field.Name = *uff.Name
	}

	if uff.Label != nil {
		field.Label = *uff.Label
	}

	if uff.FieldType != nil {
		field.FieldType = *uff.FieldType
	}

	if uff.FieldOrder != nil {
		field.FieldOrder = *uff.FieldOrder
	}

	if uff.Required != nil {
		field.Required = *uff.Required
	}

	if uff.Config != nil {
		field.Config = *uff.Config
	}

	if err := b.storer.Update(ctx, field); err != nil {
		return FormField{}, fmt.Errorf("update: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(before, field)); err != nil {
		b.log.Error(ctx, "formfieldbus: delegate call failed", "action", ActionUpdated, "err", err)
	}

	return field, nil
}

// Delete removes the specified form field.
func (b *Business) Delete(ctx context.Context, field FormField) error {
	ctx, span := otel.AddSpan(ctx, "business.formfieldbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, field); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(field)); err != nil {
		b.log.Error(ctx, "formfieldbus: delegate call failed", "action", ActionDeleted, "err", err)
	}

	return nil
}

// Query retrieves a list of form fields from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]FormField, error) {
	ctx, span := otel.AddSpan(ctx, "business.formfieldbus.query")
	defer span.End()

	fields, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return fields, nil
}

// Count returns the total number of form fields.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.formfieldbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the form field by the specified ID.
func (b *Business) QueryByID(ctx context.Context, fieldID uuid.UUID) (FormField, error) {
	ctx, span := otel.AddSpan(ctx, "business.formfieldbus.querybyid")
	defer span.End()

	field, err := b.storer.QueryByID(ctx, fieldID)
	if err != nil {
		return FormField{}, fmt.Errorf("query: fieldID[%s]: %w", fieldID, err)
	}

	return field, nil
}

// QueryByFormID retrieves all fields for a specific form.
func (b *Business) QueryByFormID(ctx context.Context, formID uuid.UUID) ([]FormField, error) {
	ctx, span := otel.AddSpan(ctx, "business.formfieldbus.querybyformid")
	defer span.End()

	fields, err := b.storer.QueryByFormID(ctx, formID)
	if err != nil {
		return nil, fmt.Errorf("query: formID[%s]: %w", formID, err)
	}

	return fields, nil
}
