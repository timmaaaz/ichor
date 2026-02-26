package inventoryitemapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/inventoryitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	inventoryitemapp *inventoryitemapp.App
}

func newAPI(inventoryitemapp *inventoryitemapp.App) *api {
	return &api{
		inventoryitemapp: inventoryitemapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app inventoryitemapp.NewInventoryItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inventoryItem, err := api.inventoryitemapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return inventoryItem
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app inventoryitemapp.UpdateInventoryItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inventoryID := web.Param(r, "item_id")
	parsed, err := uuid.Parse(inventoryID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inventoryItem, err := api.inventoryitemapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return inventoryItem
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	inventoryID := web.Param(r, "item_id")
	parsed, err := uuid.Parse(inventoryID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.inventoryitemapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if qp.IncludeLocationDetails == "true" {
		inventoryItems, err := api.inventoryitemapp.QueryWithLocationDetails(ctx, qp)
		if err != nil {
			return errs.NewError(err)
		}
		return inventoryItems
	}

	inventoryItems, err := api.inventoryitemapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return inventoryItems
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	inventoryID := web.Param(r, "item_id")
	parsed, err := uuid.Parse(inventoryID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inventoryItem, err := api.inventoryitemapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return inventoryItem
}
