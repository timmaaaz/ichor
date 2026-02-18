package formfieldapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the form field domain.
type App struct {
	formfieldbus *formfieldbus.Business
	auth         *auth.Auth
}

// NewApp constructs a form field app API for use.
func NewApp(formfieldbus *formfieldbus.Business) *App {
	return &App{
		formfieldbus: formfieldbus,
	}
}

// NewAppWithAuth constructs a form field app API for use with auth support.
func NewAppWithAuth(formfieldbus *formfieldbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:         ath,
		formfieldbus: formfieldbus,
	}
}

// Create adds a new form field to the system.
func (a *App) Create(ctx context.Context, app NewFormField) (FormField, error) {
	nff, err := toBusNewFormField(app)
	if err != nil {
		return FormField{}, errs.New(errs.InvalidArgument, err)
	}

	if err := validateCopyFromField(ctx, a.formfieldbus, nff.FormID,
		nff.EntitySchema, nff.EntityTable, nff.Name, nff.FieldType, nff.Config); err != nil {
		return FormField{}, err
	}

	field, err := a.formfieldbus.Create(ctx, nff)
	if err != nil {
		if errors.Is(err, formfieldbus.ErrUniqueEntry) {
			return FormField{}, errs.New(errs.AlreadyExists, formfieldbus.ErrUniqueEntry)
		}
		if errors.Is(err, formfieldbus.ErrForeignKeyViolation) {
			return FormField{}, errs.New(errs.Aborted, formfieldbus.ErrForeignKeyViolation)
		}
		return FormField{}, errs.Newf(errs.Internal, "create: field[%+v]: %s", field, err)
	}

	return ToAppFormField(field), nil
}

// Update updates an existing form field.
func (a *App) Update(ctx context.Context, app UpdateFormField, id uuid.UUID) (FormField, error) {
	uff, err := toBusUpdateFormField(app)
	if err != nil {
		return FormField{}, errs.New(errs.InvalidArgument, err)
	}

	field, err := a.formfieldbus.QueryByID(ctx, id)
	if err != nil {
		return FormField{}, errs.New(errs.NotFound, formfieldbus.ErrNotFound)
	}

	// Validate copy_from_field if config is being updated.
	if uff.Config != nil {
		entitySchema := field.EntitySchema
		entityTable := field.EntityTable
		fieldName := field.Name
		fieldType := field.FieldType
		if uff.EntitySchema != nil {
			entitySchema = *uff.EntitySchema
		}
		if uff.EntityTable != nil {
			entityTable = *uff.EntityTable
		}
		if uff.Name != nil {
			fieldName = *uff.Name
		}
		if uff.FieldType != nil {
			fieldType = *uff.FieldType
		}

		if err := validateCopyFromField(ctx, a.formfieldbus, field.FormID,
			entitySchema, entityTable, fieldName, fieldType, *uff.Config); err != nil {
			return FormField{}, err
		}
	}

	field, err = a.formfieldbus.Update(ctx, field, uff)
	if err != nil {
		if errors.Is(err, formfieldbus.ErrUniqueEntry) {
			return FormField{}, errs.New(errs.AlreadyExists, formfieldbus.ErrUniqueEntry)
		}
		if errors.Is(err, formfieldbus.ErrForeignKeyViolation) {
			return FormField{}, errs.New(errs.Aborted, formfieldbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, formfieldbus.ErrNotFound) {
			return FormField{}, errs.New(errs.NotFound, err)
		}
		return FormField{}, errs.Newf(errs.Internal, "update: field[%+v]: %s", field, err)
	}

	return ToAppFormField(field), nil
}

// Delete removes an existing form field.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	field, err := a.formfieldbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, formfieldbus.ErrNotFound)
	}

	if err := a.formfieldbus.Delete(ctx, field); err != nil {
		return errs.Newf(errs.Internal, "delete: field[%+v]: %s", field, err)
	}

	return nil
}

// Query returns a list of form fields based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[FormField], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[FormField]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[FormField]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[FormField]{}, errs.NewFieldsError("orderby", err)
	}

	fields, err := a.formfieldbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[FormField]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.formfieldbus.Count(ctx, filter)
	if err != nil {
		return query.Result[FormField]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppFormFieldSlice(fields), total, page), nil
}

// QueryByID retrieves a single form field by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (FormField, error) {
	field, err := a.formfieldbus.QueryByID(ctx, id)
	if err != nil {
		return FormField{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppFormField(field), nil
}

// QueryByFormID retrieves all fields for a specific form.
func (a *App) QueryByFormID(ctx context.Context, formID uuid.UUID) ([]FormField, error) {
	fields, err := a.formfieldbus.QueryByFormID(ctx, formID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "querybyformid: %s", err)
	}

	return ToAppFormFieldSlice(fields), nil
}
