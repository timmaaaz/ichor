package paymenttermapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/paymenttermapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	paymenttermapp *paymenttermapp.App
}

func newAPI(paymenttermapp *paymenttermapp.App) *api {
	return &api{
		paymenttermapp: paymenttermapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app paymenttermapp.NewPaymentTerm
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	paymentTerm, err := api.paymenttermapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return paymentTerm
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app paymenttermapp.UpdatePaymentTerm
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	paymentTermID := web.Param(r, "payment_term_id")
	parsed, err := uuid.Parse(paymentTermID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	paymentTerm, err := api.paymenttermapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return paymentTerm
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	paymentTermID := web.Param(r, "payment_term_id")

	parsed, err := uuid.Parse(paymentTermID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.paymenttermapp.Delete(ctx, parsed)
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

	result, err := api.paymenttermapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	paymentTermID := web.Param(r, "payment_term_id")

	parsed, err := uuid.Parse(paymentTermID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.paymenttermapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
