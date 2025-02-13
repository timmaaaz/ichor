package tagapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/domain/assets/tagapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	tagapp *tagapp.App
}

func newAPI(tagapp *tagapp.App) *api {
	return &api{
		tagapp: tagapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app tagapp.NewTag
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tag, err := api.tagapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tag
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app tagapp.UpdateTag
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tagID := web.Param(r, "tag_id")
	parsed, err := uuid.Parse(tagID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	assetCondition, err := api.tagapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return assetCondition
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	tagID := web.Param(r, "tag_id")

	parsed, err := uuid.Parse(tagID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.tagapp.Delete(ctx, parsed)
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

	result, err := api.tagapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	tagID := web.Param(r, "tag_id")

	parsed, err := uuid.Parse(tagID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.tagapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
