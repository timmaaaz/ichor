package contactinfoapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/contactinfoapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	contactinfoapp *contactinfoapp.App
}

func newAPI(contactinfoapp *contactinfoapp.App) *api {
	return &api{
		contactinfoapp: contactinfoapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app contactinfoapp.NewContactInfo
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	contact, err := api.contactinfoapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return contact
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app contactinfoapp.UpdateContactInfo
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	ciID := web.Param(r, "contact_info_id")
	parsed, err := uuid.Parse(ciID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	contact, err := api.contactinfoapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return contact
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	ciID := web.Param(r, "contact_info_id")

	parsed, err := uuid.Parse(ciID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.contactinfoapp.Delete(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	contacts, err := api.contactinfoapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return contacts
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	ciID := web.Param(r, "contact_info_id")

	parsed, err := uuid.Parse(ciID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	ci, err := api.contactinfoapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return ci
}
