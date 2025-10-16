package formapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the form domain.
type App struct {
	formbus *formbus.Business
	auth    *auth.Auth
}

// NewApp constructs a form app API for use.
func NewApp(formbus *formbus.Business) *App {
	return &App{
		formbus: formbus,
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