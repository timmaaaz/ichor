package rolepageapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the role page domain.
type App struct {
	rolepagebus *rolepagebus.Business
	auth        *auth.Auth
}

// NewApp constructs a role page app API for use.
func NewApp(rolepagebus *rolepagebus.Business) *App {
	return &App{
		rolepagebus: rolepagebus,
	}
}

// NewAppWithAuth constructs a role page app API for use with auth support.
func NewAppWithAuth(rolepagebus *rolepagebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:        ath,
		rolepagebus: rolepagebus,
	}
}

// Create adds a new role page mapping to the system.
func (a *App) Create(ctx context.Context, app NewRolePage) (RolePage, error) {
	nrp, err := toBusNewRolePage(app)
	if err != nil {
		return RolePage{}, errs.New(errs.InvalidArgument, err)
	}

	rp, err := a.rolepagebus.Create(ctx, nrp)
	if err != nil {
		if errors.Is(err, rolepagebus.ErrUnique) {
			return RolePage{}, errs.New(errs.Aborted, rolepagebus.ErrUnique)
		}
		return RolePage{}, errs.Newf(errs.Internal, "create: rolepage[%+v]: %s", rp, err)
	}

	return ToAppRolePage(rp), err
}

// Update updates an existing role page mapping.
func (a *App) Update(ctx context.Context, app UpdateRolePage, id uuid.UUID) (RolePage, error) {
	urp, err := toBusUpdateRolePage(app)
	if err != nil {
		return RolePage{}, errs.New(errs.InvalidArgument, err)
	}

	rp, err := a.rolepagebus.QueryByID(ctx, id)
	if err != nil {
		return RolePage{}, errs.New(errs.NotFound, rolepagebus.ErrNotFound)
	}

	updated, err := a.rolepagebus.Update(ctx, rp, urp)
	if err != nil {
		if errors.Is(err, rolepagebus.ErrNotFound) {
			return RolePage{}, errs.New(errs.NotFound, rolepagebus.ErrNotFound)
		}
		return RolePage{}, errs.Newf(errs.Internal, "update: rolepage[%+v]: %s", updated, err)
	}

	return ToAppRolePage(updated), err
}

// Delete removes an existing role page mapping.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	rp, err := a.rolepagebus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, rolepagebus.ErrNotFound)
	}

	if err := a.rolepagebus.Delete(ctx, rp); err != nil {
		return errs.Newf(errs.Internal, "delete: rolepage[%+v]: %s", rp, err)
	}

	return nil
}

// Query retrieves a list of role page mappings from the system.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[RolePage], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[RolePage]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[RolePage]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[RolePage]{}, errs.NewFieldsError("orderby", err)
	}

	rolePages, err := a.rolepagebus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[RolePage]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.rolepagebus.Count(ctx, filter)
	if err != nil {
		return query.Result[RolePage]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppRolePages(rolePages), total, page), nil
}

// QueryByID finds the role page mapping by the specified ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (RolePage, error) {
	rp, err := a.rolepagebus.QueryByID(ctx, id)
	if err != nil {
		return RolePage{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppRolePage(rp), nil
}
