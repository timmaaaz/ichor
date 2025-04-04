package inventorytransactionapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/movement/inventorytransactionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	inventorytransactionapp *inventorytransactionapp.App
}

func newAPI(inventorytransactionapp *inventorytransactionapp.App) *api {
	return &api{
		inventorytransactionapp: inventorytransactionapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app inventorytransactionapp.NewInventoryTransaction
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	i, err := api.inventorytransactionapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return i
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app inventorytransactionapp.UpdateInventoryTransaction
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	transactionID := web.Param(r, "transaction_id")
	parsed, err := uuid.Parse(transactionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	i, err := api.inventorytransactionapp.Update(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return i
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	transactionID := web.Param(r, "transaction_id")
	parsed, err := uuid.Parse(transactionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.inventorytransactionapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	transactions, err := api.inventorytransactionapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return transactions
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	transactionID := web.Param(r, "transaction_id")
	parsed, err := uuid.Parse(transactionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	transaction, err := api.inventorytransactionapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return transaction
}
