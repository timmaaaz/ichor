// Package labelapi exposes HTTP routes for the label subsystem: catalog
// print, transaction-label render+print, and a type-filtered list query.
package labelapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/labels/labelapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	labelapp *labelapp.App
}

func newAPI(labelapp *labelapp.App) *api {
	return &api{labelapp: labelapp}
}

// print handles POST /v1/labels/print — catalog-label print by ID.
func (a *api) print(ctx context.Context, r *http.Request) web.Encoder {
	var req labelapp.PrintRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := a.labelapp.Print(ctx, req); err != nil {
		return errs.NewError(err)
	}
	return nil
}

// renderPrint handles POST /v1/labels/render-print — transaction-label
// print from an in-memory payload; no catalog row is created.
func (a *api) renderPrint(ctx context.Context, r *http.Request) web.Encoder {
	var req labelapp.RenderPrintRequest
	if err := web.Decode(r, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := a.labelapp.RenderPrint(ctx, req); err != nil {
		return errs.NewError(err)
	}
	return nil
}

// query handles GET /v1/labels — type-filtered catalog list for admin UI.
func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	labels, err := a.labelapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}
	return labels
}
