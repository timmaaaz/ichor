// Package vproductapi maintains the web based api for product view access.
package vproductapi

import (
	"context"
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/app/domain/vproductapp"
	"bitbucket.org/superiortechnologies/ichor/app/sdk/errs"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
)

type api struct {
	vproductApp *vproductapp.App
}

func newAPI(vproductApp *vproductapp.App) *api {
	return &api{
		vproductApp: vproductApp,
	}
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	prd, err := api.vproductApp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return prd
}
