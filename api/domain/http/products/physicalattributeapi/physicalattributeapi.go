package physicalattributeapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/products/physicalattributeapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	physicalattributeapp *physicalattributeapp.App
}

func newAPI(physicalattributeapp *physicalattributeapp.App) *api {
	return &api{
		physicalattributeapp: physicalattributeapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app physicalattributeapp.NewPhysicalAttribute
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	brand, err := api.physicalattributeapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return brand
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app physicalattributeapp.UpdatePhysicalAttribute
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	attributeID := web.Param(r, "attribute_id")
	parsed, err := uuid.Parse(attributeID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	brand, err := api.physicalattributeapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return brand
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	attributeID := web.Param(r, "attribute_id")

	parsed, err := uuid.Parse(attributeID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.physicalattributeapp.Delete(ctx, parsed)
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

	brands, err := api.physicalattributeapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return brands
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	attributeID := web.Param(r, "attribute_id")

	parsed, err := uuid.Parse(attributeID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	brand, err := api.physicalattributeapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return brand
}
