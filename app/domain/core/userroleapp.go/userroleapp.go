package userroleapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the userrole domain.
type App struct {
	userrolebus *userrolebus.Business
	auth        *auth.Auth
}

// NewApp constructs a userrole app API for use.
func NewApp(userrolebus *userrolebus.Business) *App {
	return &App{
		userrolebus: userrolebus,
	}
}

// NewAppWithAuth constructs a userrole app API for use with auth support.
func NewAppWithAuth(userrolebus *userrolebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:        ath,
		userrolebus: userrolebus,
	}
}

// Create adds a new userrole to the system.
func (a *App) Create(ctx context.Context, app NewUserRole) (UserRole, error) {
	nur, err := toBusNewUserRole(app)
	if err != nil {
		return UserRole{}, errs.New(errs.InvalidArgument, err)
	}

	rol, err := a.userrolebus.Create(ctx, nur)
	if err != nil {
		if errors.Is(err, userrolebus.ErrUnique) {
			return UserRole{}, errs.New(errs.Aborted, userrolebus.ErrUnique)
		}
		return UserRole{}, errs.Newf(errs.Internal, "create: userrole[%+v]: %s", rol, err)
	}

	return ToAppUserRole(rol), err
}

// Delete removes a userrole from the system by its id.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	ur, err := a.userrolebus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, userrolebus.ErrNotFound)
	}

	if err := a.userrolebus.Delete(ctx, ur); err != nil {
		return errs.Newf(errs.Internal, "delete: userrole[%+v]: %s", ur, err)
	}

	return nil
}

// Query retrieves a list of roles from the system.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[UserRole], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[UserRole]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[UserRole]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[UserRole]{}, errs.NewFieldsError("orderby", err)
	}

	urs, err := a.userrolebus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[UserRole]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.userrolebus.Count(ctx, filter)
	if err != nil {
		return query.Result[UserRole]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppUserRoles(urs), total, page), nil
}

// QueryByID retrieves a single userrole from the system by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (UserRole, error) {
	ur, err := a.userrolebus.QueryByID(ctx, id)
	if err != nil {
		return UserRole{}, errs.New(errs.NotFound, userrolebus.ErrNotFound)
	}

	return ToAppUserRole(ur), err
}
