package roleapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the  role domain.
type App struct {
	rolebus *rolebus.Business
	auth    *auth.Auth
}

// NewApp constructs a  role app API for use.
func NewApp(rolebus *rolebus.Business) *App {
	return &App{
		rolebus: rolebus,
	}
}

// NewAppWithAuth constructs a  role app API for use with auth support.
func NewAppWithAuth(rolebus *rolebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:    ath,
		rolebus: rolebus,
	}
}

// Create adds a new  role to the system.
func (a *App) Create(ctx context.Context, app NewRole) (Role, error) {
	nr, err := toBusNewRole(app)
	if err != nil {
		return Role{}, errs.New(errs.InvalidArgument, err)
	}

	rol, err := a.rolebus.Create(ctx, nr)
	if err != nil {
		if errors.Is(err, rolebus.ErrUnique) {
			return Role{}, errs.New(errs.Aborted, rolebus.ErrUnique)
		}
		return Role{}, errs.Newf(errs.Internal, "create: role[%+v]: %s", rol, err)
	}

	return ToAppRole(rol), err
}

// Update updates an existing  role.
func (a *App) Update(ctx context.Context, app UpdateRole, id uuid.UUID) (Role, error) {
	ur, err := toBusUpdateRole(app)
	if err != nil {
		return Role{}, errs.New(errs.InvalidArgument, err)
	}

	rol, err := a.rolebus.QueryByID(ctx, id)
	if err != nil {
		return Role{}, errs.New(errs.NotFound, rolebus.ErrNotFound)
	}

	updated, err := a.rolebus.Update(ctx, rol, ur)
	if err != nil {
		if errors.Is(err, rolebus.ErrNotFound) {
			return Role{}, errs.New(errs.NotFound, rolebus.ErrNotFound)
		}
		return Role{}, errs.Newf(errs.Internal, "update: role[%+v]: %s", updated, err)
	}

	return ToAppRole(updated), err
}

// Delete removes an existing  role.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	rol, err := a.rolebus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, rolebus.ErrNotFound)
	}

	if err := a.rolebus.Delete(ctx, rol); err != nil {
		return errs.Newf(errs.Internal, "delete: role[%+v]: %s", rol, err)
	}

	return nil
}

// Query retrieves a list of roles from the system.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Role], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Role]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Role]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Role]{}, errs.NewFieldsError("orderby", err)
	}

	roles, err := a.rolebus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Role]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.rolebus.Count(ctx, filter)
	if err != nil {
		return query.Result[Role]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppRoles(roles), total, page), nil
}

// QueryByID finds the role by the specified ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Role, error) {
	rol, err := a.rolebus.QueryByID(ctx, id)
	if err != nil {
		return Role{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppRole(rol), nil
}
