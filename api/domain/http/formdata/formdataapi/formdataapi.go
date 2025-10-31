package formdataapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/formdata/formdataapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	formdataapp *formdataapp.App
}

func newAPI(formdataapp *formdataapp.App) *api {
	return &api{
		formdataapp: formdataapp,
	}
}

// upsert handles multi-entity transactional create/update operations.
func (api *api) upsert(ctx context.Context, r *http.Request) web.Encoder {
	var req formdataapp.FormDataRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	formID := web.Param(r, "form_id")
	parsed, err := uuid.Parse(formID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.formdataapp.UpsertFormData(ctx, parsed, req)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

// validate validates that a form has all required fields for the specified operations.
func (api *api) validate(ctx context.Context, r *http.Request) web.Encoder {
	var req formdataapp.FormValidationRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	formID := web.Param(r, "form_id")
	parsed, err := uuid.Parse(formID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.formdataapp.ValidateForm(ctx, parsed, req)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
