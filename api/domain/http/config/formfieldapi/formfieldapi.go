package formfieldapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	formfieldapp *formfieldapp.App
}

func newAPI(formfieldapp *formfieldapp.App) *api {
	return &api{
		formfieldapp: formfieldapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app formfieldapp.NewFormField
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	field, err := api.formfieldapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return field
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app formfieldapp.UpdateFormField
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	fieldID := web.Param(r, "field_id")
	parsed, err := uuid.Parse(fieldID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	field, err := api.formfieldapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return field
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	fieldID := web.Param(r, "field_id")

	parsed, err := uuid.Parse(fieldID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.formfieldapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	fields, err := api.formfieldapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return fields
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	fieldID := web.Param(r, "field_id")

	parsed, err := uuid.Parse(fieldID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	field, err := api.formfieldapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return field
}

func (api *api) queryByFormID(ctx context.Context, r *http.Request) web.Encoder {
	formIDStr := web.Param(r, "form_id")

	formID, err := uuid.Parse(formIDStr)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	fields, err := api.formfieldapp.QueryByFormID(ctx, formID)
	if err != nil {
		return errs.NewError(err)
	}

	return formfieldapp.FormFields{Fields: fields}
}
