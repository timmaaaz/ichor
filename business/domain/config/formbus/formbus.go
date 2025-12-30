// Package formbus provides business logic for form configuration management.
package formbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
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
	QueryAll(ctx context.Context) ([]Form, error)
}

// Business manages the set of APIs for form access.
type Business struct {
	log          *logger.Logger
	storer       Storer
	delegate     *delegate.Delegate
	formFieldBus *formfieldbus.Business
}

// NewBusiness constructs a form business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer, formFieldBus *formfieldbus.Business) *Business {
	return &Business{
		log:          log,
		delegate:     delegate,
		storer:       storer,
		formFieldBus: formFieldBus,
	}
}

// NewWithTx constructs a new Business value replacing the Storer
// value with a Storer value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	formFieldBus, err := b.formFieldBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:          b.log,
		delegate:     b.delegate,
		storer:       storer,
		formFieldBus: formFieldBus,
	}, nil
}

// Create inserts a new form into the database.
func (b *Business) Create(ctx context.Context, nf NewForm) (Form, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.create")
	defer span.End()

	form := Form{
		ID:                uuid.New(),
		Name:              nf.Name,
		IsReferenceData:   nf.IsReferenceData,
		AllowInlineCreate: nf.AllowInlineCreate,
	}

	if err := b.storer.Create(ctx, form); err != nil {
		return Form{}, fmt.Errorf("create: %w", err)
	}

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionCreatedData(form)); err != nil {
		b.log.Error(ctx, "formbus: delegate call failed", "action", ActionCreated, "err", err)
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

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionUpdatedData(form)); err != nil {
		b.log.Error(ctx, "formbus: delegate call failed", "action", ActionUpdated, "err", err)
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

	// Fire delegate event for workflow automation
	if err := b.delegate.Call(ctx, ActionDeletedData(form)); err != nil {
		b.log.Error(ctx, "formbus: delegate call failed", "action", ActionDeleted, "err", err)
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

// QueryAll retrieves all forms from the system.
func (b *Business) QueryAll(ctx context.Context) ([]Form, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.queryall")
	defer span.End()

	forms, err := b.storer.QueryAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("queryall: %w", err)
	}

	return forms, nil
}

// ExportByIDs exports forms and their fields by IDs.
func (b *Business) ExportByIDs(ctx context.Context, formIDs []uuid.UUID) ([]FormWithFields, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.exportbyids")
	defer span.End()

	var results []FormWithFields

	for _, formID := range formIDs {
		form, err := b.storer.QueryByID(ctx, formID)
		if err != nil {
			return nil, fmt.Errorf("query form %s: %w", formID, err)
		}

		fields, err := b.formFieldBus.QueryByFormID(ctx, formID)
		if err != nil {
			return nil, fmt.Errorf("query fields for form %s: %w", formID, err)
		}

		results = append(results, FormWithFields{
			Form:   form,
			Fields: fields,
		})
	}

	return results, nil
}

// ImportForms imports forms with conflict resolution.
func (b *Business) ImportForms(ctx context.Context, packages []FormWithFields, mode string) (ImportStats, error) {
	ctx, span := otel.AddSpan(ctx, "business.formbus.importforms")
	defer span.End()

	stats := ImportStats{}

	for _, pkg := range packages {
		// Check if form exists by name
		existing, err := b.storer.QueryByName(ctx, pkg.Form.Name)
		existsAlready := err == nil

		switch mode {
		case "skip":
			if existsAlready {
				stats.SkippedCount++
				continue
			}
			// Create new
			if err := b.createFormWithFields(ctx, pkg); err != nil {
				return stats, err
			}
			stats.ImportedCount++

		case "replace":
			if existsAlready {
				// Delete existing and create new
				if err := b.Delete(ctx, existing); err != nil {
					return stats, fmt.Errorf("delete existing: %w", err)
				}
				stats.UpdatedCount++
			}
			if err := b.createFormWithFields(ctx, pkg); err != nil {
				return stats, err
			}
			if !existsAlready {
				stats.ImportedCount++
			}

		case "merge":
			if existsAlready {
				// Update existing form, merge fields
				if err := b.updateFormWithFields(ctx, existing.ID, pkg); err != nil {
					return stats, fmt.Errorf("update form: %w", err)
				}
				stats.UpdatedCount++
			} else {
				// Create new
				if err := b.createFormWithFields(ctx, pkg); err != nil {
					return stats, err
				}
				stats.ImportedCount++
			}
		}
	}

	return stats, nil
}

func (b *Business) createFormWithFields(ctx context.Context, pkg FormWithFields) error {
	// Create form
	newForm := NewForm{
		Name:              pkg.Form.Name,
		IsReferenceData:   pkg.Form.IsReferenceData,
		AllowInlineCreate: pkg.Form.AllowInlineCreate,
	}

	form, err := b.Create(ctx, newForm)
	if err != nil {
		return fmt.Errorf("create form: %w", err)
	}

	// Create fields
	for _, field := range pkg.Fields {
		newField := formfieldbus.NewFormField{
			FormID:       form.ID,
			EntityID:     field.EntityID,
			EntitySchema: field.EntitySchema,
			EntityTable:  field.EntityTable,
			Name:         field.Name,
			Label:        field.Label,
			FieldType:    field.FieldType,
			FieldOrder:   field.FieldOrder,
			Required:     field.Required,
			Config:       field.Config,
		}
		if _, err := b.formFieldBus.Create(ctx, newField); err != nil {
			return fmt.Errorf("create field %s: %w", field.Name, err)
		}
	}

	return nil
}

func (b *Business) updateFormWithFields(ctx context.Context, formID uuid.UUID, pkg FormWithFields) error {
	// First get the current form
	form, err := b.storer.QueryByID(ctx, formID)
	if err != nil {
		return fmt.Errorf("query form: %w", err)
	}

	// Update form
	updateForm := UpdateForm{
		Name:              &pkg.Form.Name,
		IsReferenceData:   &pkg.Form.IsReferenceData,
		AllowInlineCreate: &pkg.Form.AllowInlineCreate,
	}

	if _, err := b.Update(ctx, form, updateForm); err != nil {
		return fmt.Errorf("update form: %w", err)
	}

	// Delete existing fields and recreate (simple approach)
	existingFields, err := b.formFieldBus.QueryByFormID(ctx, formID)
	if err != nil {
		return fmt.Errorf("query existing fields: %w", err)
	}

	for _, field := range existingFields {
		if err := b.formFieldBus.Delete(ctx, field); err != nil {
			return fmt.Errorf("delete field %s: %w", field.ID, err)
		}
	}

	// Create new fields
	for _, field := range pkg.Fields {
		newField := formfieldbus.NewFormField{
			FormID:       formID,
			EntityID:     field.EntityID,
			EntitySchema: field.EntitySchema,
			EntityTable:  field.EntityTable,
			Name:         field.Name,
			Label:        field.Label,
			FieldType:    field.FieldType,
			FieldOrder:   field.FieldOrder,
			Required:     field.Required,
			Config:       field.Config,
		}
		if _, err := b.formFieldBus.Create(ctx, newField); err != nil {
			return fmt.Errorf("create field %s: %w", field.Name, err)
		}
	}

	return nil
}