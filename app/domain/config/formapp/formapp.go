package formapp

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the form domain.
type App struct {
	formbus      *formbus.Business
	formfieldbus *formfieldbus.Business
	auth         *auth.Auth
}

// NewApp constructs a form app API for use.
func NewApp(formbus *formbus.Business) *App {
	return &App{
		formbus: formbus,
	}
}

// NewAppWithFormFields constructs a form app API with form field business support.
func NewAppWithFormFields(formbus *formbus.Business, formfieldbus *formfieldbus.Business) *App {
	return &App{
		formbus:      formbus,
		formfieldbus: formfieldbus,
	}
}

// NewAppWithAuth constructs a form app API for use with auth support.
func NewAppWithAuth(formbus *formbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:    ath,
		formbus: formbus,
	}
}

// Create adds a new form to the system.
func (a *App) Create(ctx context.Context, app NewForm) (Form, error) {
	nb := toBusNewForm(app)

	form, err := a.formbus.Create(ctx, nb)
	if err != nil {
		if errors.Is(err, formbus.ErrUniqueEntry) {
			return Form{}, errs.New(errs.AlreadyExists, formbus.ErrUniqueEntry)
		}
		return Form{}, errs.Newf(errs.Internal, "create: form[%+v]: %s", form, err)
	}

	return ToAppForm(form), nil
}

// Update updates an existing form.
func (a *App) Update(ctx context.Context, app UpdateForm, id uuid.UUID) (Form, error) {
	uf := toBusUpdateForm(app)

	form, err := a.formbus.QueryByID(ctx, id)
	if err != nil {
		return Form{}, errs.New(errs.NotFound, formbus.ErrNotFound)
	}

	form, err = a.formbus.Update(ctx, form, uf)
	if err != nil {
		if errors.Is(err, formbus.ErrUniqueEntry) {
			return Form{}, errs.New(errs.AlreadyExists, formbus.ErrUniqueEntry)
		}
		if errors.Is(err, formbus.ErrNotFound) {
			return Form{}, errs.New(errs.NotFound, err)
		}
		return Form{}, errs.Newf(errs.Internal, "update: form[%+v]: %s", form, err)
	}

	return ToAppForm(form), nil
}

// Delete removes an existing form.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	form, err := a.formbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, formbus.ErrNotFound)
	}

	if err := a.formbus.Delete(ctx, form); err != nil {
		return errs.Newf(errs.Internal, "delete: form[%+v]: %s", form, err)
	}

	return nil
}

// Query returns a list of forms based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Form], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Form]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Form]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Form]{}, errs.NewFieldsError("orderby", err)
	}

	forms, err := a.formbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Form]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.formbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Form]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppForms(forms), total, page), nil
}

// QueryByID retrieves a single form by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Form, error) {
	form, err := a.formbus.QueryByID(ctx, id)
	if err != nil {
		return Form{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppForm(form), nil
}

// QueryFullByID retrieves a single form by its id along with all its fields.
func (a *App) QueryFullByID(ctx context.Context, id uuid.UUID) (FormFull, error) {
	if a.formfieldbus == nil {
		return FormFull{}, errs.Newf(errs.Internal, "formfieldbus not configured")
	}

	form, err := a.formbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, formbus.ErrNotFound) {
			return FormFull{}, errs.New(errs.NotFound, formbus.ErrNotFound)
		}
		return FormFull{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	busFields, err := a.formfieldbus.QueryByFormID(ctx, form.ID)
	if err != nil {
		return FormFull{}, errs.Newf(errs.Internal, "querybyformid: %s", err)
	}

	fields := formfieldapp.ToAppFormFieldSlice(busFields)

	return ToAppFormFull(form, fields), nil
}

// QueryFullByName retrieves a single form by its name along with all its fields.
func (a *App) QueryFullByName(ctx context.Context, name string) (FormFull, error) {
	if a.formfieldbus == nil {
		return FormFull{}, errs.Newf(errs.Internal, "formfieldbus not configured")
	}

	form, err := a.formbus.QueryByName(ctx, name)
	if err != nil {
		if errors.Is(err, formbus.ErrNotFound) {
			return FormFull{}, errs.New(errs.NotFound, formbus.ErrNotFound)
		}
		return FormFull{}, errs.Newf(errs.Internal, "querybyname: %s", err)
	}

	busFields, err := a.formfieldbus.QueryByFormID(ctx, form.ID)
	if err != nil {
		return FormFull{}, errs.Newf(errs.Internal, "querybyformid: %s", err)
	}

	fields := formfieldapp.ToAppFormFieldSlice(busFields)

	return ToAppFormFull(form, fields), nil
}

// QueryAll retrieves all forms from the system.
func (a *App) QueryAll(ctx context.Context) (Forms, error) {
	forms, err := a.formbus.QueryAll(ctx)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "queryall: %s", err)
	}

	return Forms(ToAppForms(forms)), nil
}

// ExportByIDs exports forms by IDs as a JSON package.
func (a *App) ExportByIDs(ctx context.Context, formIDs []string) (ExportPackage, error) {
	// Convert string IDs to UUIDs
	uuids := make([]uuid.UUID, len(formIDs))
	for i, id := range formIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return ExportPackage{}, errs.Newf(errs.InvalidArgument, "invalid form ID %s: %s", id, err)
		}
		uuids[i] = uid
	}

	// Export from business layer
	results, err := a.formbus.ExportByIDs(ctx, uuids)
	if err != nil {
		return ExportPackage{}, errs.Newf(errs.Internal, "export: %s", err)
	}

	// Convert to app models
	var packages []FormPackage
	for _, result := range results {
		packages = append(packages, FormPackage{
			Form:   ToAppForm(result.Form),
			Fields: formfieldapp.ToAppFormFieldSlice(result.Fields),
		})
	}

	return ExportPackage{
		Version:    "1.0",
		Type:       "forms",
		ExportedAt: timeNow().Format(timeRFC3339),
		Count:      len(packages),
		Data:       packages,
	}, nil
}

// ImportForms imports forms from a JSON package.
func (a *App) ImportForms(ctx context.Context, pkg ImportPackage) (ImportResult, error) {
	// Validate package
	if err := pkg.Validate(); err != nil {
		return ImportResult{}, err
	}

	// Convert app models to business models
	var busPackages []formbus.FormWithFields
	for i, formPkg := range pkg.Data {
		busPkg, err := ToBusFormWithFields(formPkg)
		if err != nil {
			return ImportResult{
				Errors: []string{err.Error()},
			}, errs.Newf(errs.InvalidArgument, "convert form %d: %s", i, err)
		}
		busPackages = append(busPackages, busPkg)
	}

	// Import via business layer
	stats, err := a.formbus.ImportForms(ctx, busPackages, pkg.Mode)
	if err != nil {
		return ImportResult{
			Errors: []string{err.Error()},
		}, errs.Newf(errs.Internal, "import: %s", err)
	}

	return ImportResult{
		ImportedCount: stats.ImportedCount,
		SkippedCount:  stats.SkippedCount,
		UpdatedCount:  stats.UpdatedCount,
	}, nil
}

// Variables for testing
var (
	timeNow      = func() interface{ Format(string) string } { return timeNowImpl{} }
	timeRFC3339  = "2006-01-02T15:04:05Z07:00"
)

type timeNowImpl struct{}

func (t timeNowImpl) Format(layout string) string {
	return time.Now().Format(layout)
}