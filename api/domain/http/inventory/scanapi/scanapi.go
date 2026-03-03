package scanapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/scanapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	scanapp *scanapp.App
}

func newAPI(scanApp *scanapp.App) *api {
	return &api{scanapp: scanApp}
}

func (api *api) scan(ctx context.Context, r *http.Request) web.Encoder {
	barcode := r.URL.Query().Get("barcode")
	if barcode == "" {
		return errs.Newf(errs.InvalidArgument, "barcode query parameter is required")
	}

	result, err := api.scanapp.Scan(ctx, barcode)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
