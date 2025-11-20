package formapi

import (
	"context"
	"net/http"
	"net/url"

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

func (api *api) queryFullByID(ctx context.Context, r *http.Request) web.Encoder {
	formID := web.Param(r, "form_id")

	parsed, err := uuid.Parse(formID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	formFull, err := api.formapp.QueryFullByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return formFull
}

func (api *api) queryFullByName(ctx context.Context, r *http.Request) web.Encoder {
	formName := web.Param(r, "form_name")

	if formName == "" {
		return errs.New(errs.InvalidArgument, errs.Newf(errs.InvalidArgument, "form name is required"))
	}

	// Decode URL-encoded form name (e.g., "My%20Form" -> "My Form")
	decodedName, err := url.QueryUnescape(formName)
	if err != nil {
		return errs.New(errs.InvalidArgument, errs.Newf(errs.InvalidArgument, "invalid form name encoding: %s", err))
	}

	formFull, err := api.formapp.QueryFullByName(ctx, decodedName)
	if err != nil {
		return errs.NewError(err)
	}

	return formFull
}

func (api *api) queryAll(ctx context.Context, r *http.Request) web.Encoder {
	forms, err := api.formapp.QueryAll(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return forms
}