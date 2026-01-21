package currencyapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/currencyapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	currencyapp *currencyapp.App
}

func newAPI(currencyapp *currencyapp.App) *api {
	return &api{
		currencyapp: currencyapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app currencyapp.NewCurrency
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	currency, err := api.currencyapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return currency
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app currencyapp.UpdateCurrency
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	currencyID := web.Param(r, "currency_id")
	parsed, err := uuid.Parse(currencyID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	currency, err := api.currencyapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return currency
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	currencyID := web.Param(r, "currency_id")
	parsed, err := uuid.Parse(currencyID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.currencyapp.Delete(ctx, parsed)
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

	currency, err := api.currencyapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return currency
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	currencyID := web.Param(r, "currency_id")
	parsed, err := uuid.Parse(currencyID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	currency, err := api.currencyapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return currency
}

func (api *api) queryAll(ctx context.Context, r *http.Request) web.Encoder {
	currencies, err := api.currencyapp.QueryAll(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return currencies
}
