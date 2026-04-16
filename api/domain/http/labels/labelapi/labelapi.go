// Package labelapi exposes HTTP routes for the label subsystem: CRUD on
// the catalog, catalog print, transaction-label render+print, and a
// type-filtered list query.
package labelapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
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

// create handles POST /v1/labels — insert a new catalog row.
func (a *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app labelapp.NewLabel
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lc, err := a.labelapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}
	return lc
}

// update handles PUT /v1/labels/{label_id} — partial patch.
func (a *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app labelapp.UpdateLabel
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	id, err := uuid.Parse(web.Param(r, "label_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lc, err := a.labelapp.Update(ctx, id, app)
	if err != nil {
		return errs.NewError(err)
	}
	return lc
}

// delete handles DELETE /v1/labels/{label_id}.
func (a *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "label_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := a.labelapp.Delete(ctx, id); err != nil {
		return errs.NewError(err)
	}
	return nil
}

// queryByID handles GET /v1/labels/{label_id}.
func (a *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	id, err := uuid.Parse(web.Param(r, "label_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lc, err := a.labelapp.QueryByID(ctx, id)
	if err != nil {
		return errs.NewError(err)
	}
	return lc
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
