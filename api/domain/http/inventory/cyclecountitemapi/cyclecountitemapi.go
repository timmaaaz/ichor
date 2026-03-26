package cyclecountitemapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	cyclecountitemapp *cyclecountitemapp.App
}

func newAPI(cyclecountitemapp *cyclecountitemapp.App) *api {
	return &api{
		cyclecountitemapp: cyclecountitemapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app cyclecountitemapp.NewCycleCountItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	item, err := api.cyclecountitemapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return item
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app cyclecountitemapp.UpdateCycleCountItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	itemID := web.Param(r, "item_id")
	parsed, err := uuid.Parse(itemID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	item, err := api.cyclecountitemapp.Update(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return item
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	itemID := web.Param(r, "item_id")
	parsed, err := uuid.Parse(itemID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.cyclecountitemapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	items, err := api.cyclecountitemapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return items
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	itemID := web.Param(r, "item_id")
	parsed, err := uuid.Parse(itemID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	item, err := api.cyclecountitemapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return item
}
