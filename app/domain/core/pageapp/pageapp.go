package pageapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the page domain.
type App struct {
	pagebus *pagebus.Business
	auth    *auth.Auth
}

// NewApp constructs a page app API for use.
func NewApp(pagebus *pagebus.Business) *App {
	return &App{
		pagebus: pagebus,
	}
}

// NewAppWithAuth constructs a page app API for use with auth support.
func NewAppWithAuth(pagebus *pagebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:    ath,
		pagebus: pagebus,
	}
}

// Create adds a new page to the system.
func (a *App) Create(ctx context.Context, app NewPage) (Page, error) {
	np, err := toBusNewPage(app)
	if err != nil {
		return Page{}, errs.New(errs.InvalidArgument, err)
	}

	pag, err := a.pagebus.Create(ctx, np)
	if err != nil {
		if errors.Is(err, pagebus.ErrUnique) {
			return Page{}, errs.New(errs.Aborted, pagebus.ErrUnique)
		}
		return Page{}, errs.Newf(errs.Internal, "create: page[%+v]: %s", pag, err)
	}

	return ToAppPage(pag), err
}

// Update updates an existing page.
func (a *App) Update(ctx context.Context, app UpdatePage, id uuid.UUID) (Page, error) {
	up, err := toBusUpdatePage(app)
	if err != nil {
		return Page{}, errs.New(errs.InvalidArgument, err)
	}

	pag, err := a.pagebus.QueryByID(ctx, id)
	if err != nil {
		return Page{}, errs.New(errs.NotFound, pagebus.ErrNotFound)
	}

	updated, err := a.pagebus.Update(ctx, pag, up)
	if err != nil {
		if errors.Is(err, pagebus.ErrNotFound) {
			return Page{}, errs.New(errs.NotFound, pagebus.ErrNotFound)
		}
		return Page{}, errs.Newf(errs.Internal, "update: page[%+v]: %s", updated, err)
	}

	return ToAppPage(updated), err
}

// Delete removes an existing page.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	pag, err := a.pagebus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, pagebus.ErrNotFound)
	}

	if err := a.pagebus.Delete(ctx, pag); err != nil {
		return errs.Newf(errs.Internal, "delete: page[%+v]: %s", pag, err)
	}

	return nil
}

// Query retrieves a list of pages from the system.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Page], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Page]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Page]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Page]{}, errs.NewFieldsError("orderby", err)
	}

	pages, err := a.pagebus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Page]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.pagebus.Count(ctx, filter)
	if err != nil {
		return query.Result[Page]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppPages(pages), total, page), nil
}

// QueryByID finds the page by the specified ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Page, error) {
	pag, err := a.pagebus.QueryByID(ctx, id)
	if err != nil {
		return Page{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppPage(pag), nil
}
