package titleapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the title domain.
type App struct {
	titlebus *titlebus.Business
	auth     *auth.Auth
}

// NewApp constructs a title app API for use.
func NewApp(titlebus *titlebus.Business) *App {
	return &App{
		titlebus: titlebus,
	}
}

// NewAppWithAuth constructs a title app API for use with auth support.
func NewAppWithAuth(titlebus *titlebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:     ath,
		titlebus: titlebus,
	}
}

// Create adds a new title to the system
func (a *App) Create(ctx context.Context, app NewTitle) (Title, error) {
	nt, err := toBusNewTitle(app)
	if err != nil {
		return Title{}, err
	}

	fs, err := a.titlebus.Create(ctx, nt)
	if err != nil {
		return Title{}, err
	}

	return ToAppTitle(fs), nil
}

// Update updates an existing title
func (a *App) Update(ctx context.Context, app UpdateTitle, id uuid.UUID) (Title, error) {
	ut, err := toBusUpdateTitle(app)
	if err != nil {
		return Title{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.titlebus.QueryByID(ctx, id)
	if err != nil {
		return Title{}, errs.New(errs.NotFound, titlebus.ErrNotFound)
	}

	updated, err := a.titlebus.Update(ctx, as, ut)
	if err != nil {
		if errors.Is(err, titlebus.ErrNotFound) {
			return Title{}, errs.New(errs.NotFound, err)
		}
		return Title{}, errs.Newf(errs.Internal, "update: approvalStatus[%+v]: %s", updated, err)
	}

	return ToAppTitle(updated), nil
}

// Delete removes an existing title
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	t, err := a.titlebus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, titlebus.ErrNotFound)
	}

	err = a.titlebus.Delete(ctx, t)
	if err != nil {
		return errs.Newf(errs.Internal, "delete title[%+v]: %s", t, err)
	}

	return nil
}

// Query returns a list of title based on the filter, order and page
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Title], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Title]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Title]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Title]{}, errs.NewFieldsError("orderby", err)
	}

	t, err := a.titlebus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Title]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.titlebus.Count(ctx, filter)
	if err != nil {
		return query.Result[Title]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppTitles(t), total, page), nil
}

// QueryByID retrieves the title by ID
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Title, error) {
	t, err := a.titlebus.QueryByID(ctx, id)
	if err != nil {
		return Title{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppTitle(t), nil
}
