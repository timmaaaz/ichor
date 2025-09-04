package tableaccessapp

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the tableaccess domain.
type App struct {
	tableaccessbus *tableaccessbus.Business
	auth           *auth.Auth
}

// NewApp constructs a tableaccess app API for use.
func NewApp(tableaccessbus *tableaccessbus.Business) *App {
	return &App{
		tableaccessbus: tableaccessbus,
	}
}

// NewAppWithAuth constructs a tableaccess app API for use with auth support.
func NewAppWithAuth(tableaccessbus *tableaccessbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:           ath,
		tableaccessbus: tableaccessbus,
	}
}

// Create adds a new tableaccess to the system.
func (a *App) Create(ctx context.Context, app NewTableAccess) (TableAccess, error) {
	nta, err := toBusNewTableAccess(app)
	if err != nil {
		return TableAccess{}, errs.New(errs.InvalidArgument, err)
	}

	ta, err := a.tableaccessbus.Create(ctx, nta)
	if err != nil {
		return TableAccess{}, errs.New(errs.Internal, err)
	}

	return ToAppTableAccess(ta), err
}

// Update updates an existing tableaccess.
func (a *App) Update(ctx context.Context, app UpdateTableAccess, id uuid.UUID) (TableAccess, error) {
	uta, err := toBusUpdateTableAccess(app)
	if err != nil {
		return TableAccess{}, errs.New(errs.InvalidArgument, err)
	}

	ta, err := a.tableaccessbus.QueryByID(ctx, id)
	if err != nil {
		return TableAccess{}, errs.New(errs.NotFound, tableaccessbus.ErrNotFound)
	}

	updated, err := a.tableaccessbus.Update(ctx, ta, uta)
	if err != nil {
		if err == tableaccessbus.ErrNotFound {
			return TableAccess{}, errs.New(errs.NotFound, tableaccessbus.ErrNotFound)
		}
		return TableAccess{}, errs.Newf(errs.Internal, "update: tableaccess[%+v]: %s", updated, err)
	}

	return ToAppTableAccess(updated), err
}

// Delete removes a tableaccess from the system.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	ta, err := a.tableaccessbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, tableaccessbus.ErrNotFound)
	}

	if err := a.tableaccessbus.Delete(ctx, ta); err != nil {
		return errs.Newf(errs.Internal, "delete: tableaccess[%s]: %s", id, err)
	}

	return nil
}

// Query retrieves a list of tableaccesses from the system.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[TableAccess], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[TableAccess]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[TableAccess]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[TableAccess]{}, errs.NewFieldsError("orderby", err)
	}

	tableaccesses, err := a.tableaccessbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[TableAccess]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.tableaccessbus.Count(ctx, filter)
	if err != nil {
		return query.Result[TableAccess]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppTableAccesses(tableaccesses), total, page), nil
}

// QueryByID retrieves a single tableaccess by its ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (TableAccess, error) {
	ta, err := a.tableaccessbus.QueryByID(ctx, id)
	if err != nil {
		return TableAccess{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppTableAccess(ta), err
}

// QueryAll retrieves all tableaccesses from the system.
func (a *App) QueryAll(ctx context.Context) ([]TableAccess, error) {
	ta, err := a.tableaccessbus.QueryAll(ctx)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "queryall: %s", err)
	}

	return ToAppTableAccesses(ta), nil
}
