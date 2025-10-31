package formapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/formapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	formapp *formapp.App
}

func newAPI(formapp *formapp.App) *api {
	return &api{
		formapp: formapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app formapp.NewForm
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	form, err := api.formapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return form
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app formapp.UpdateForm
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	formID := web.Param(r, "form_id")
	parsed, err := uuid.Parse(formID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	form, err := api.formapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return form
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	formID := web.Param(r, "form_id")

	parsed, err := uuid.Parse(formID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.formapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	forms, err := api.formapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return forms
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	formID := web.Param(r, "form_id")

	parsed, err := uuid.Parse(formID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	form, err := api.formapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return form
}