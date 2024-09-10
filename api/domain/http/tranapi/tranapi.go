// Package tranapi maintains the web based api for tran access.
package tranapi

import (
	"context"
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/app/domain/tranapp"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/errs"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
)

type api struct {
	tranApp *tranapp.App
}

func newAPI(tranApp *tranapp.App) *api {
	return &api{
		tranApp: tranApp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app tranapp.NewTran
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	prd, err := api.tranApp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return prd
}
