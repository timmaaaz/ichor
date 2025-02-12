// Package userapp maintains the app layer api for the user domain.
package userapp

import (
	"context"
	"errors"

	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the user domain.
type App struct {
	userBus *userbus.Business
	auth    *auth.Auth
}

// NewApp constructs a user app API for use.
func NewApp(userBus *userbus.Business) *App {
	return &App{
		userBus: userBus,
	}
}

// NewAppWithAuth constructs a user app API for use with auth support.
func NewAppWithAuth(userBus *userbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:    ath,
		userBus: userBus,
	}
}

// Create adds a new user to the system.
func (a *App) Create(ctx context.Context, app NewUser) (User, error) {
	nc, err := toBusNewUser(app)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	usr, err := a.userBus.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, userbus.ErrUniqueEmail) {
			return User{}, errs.New(errs.Aborted, userbus.ErrUniqueEmail)
		}
		return User{}, errs.Newf(errs.Internal, "create: usr[%+v]: %s", usr, err)
	}

	return toAppUser(usr), nil
}

// Update updates an existing user.
func (a *App) Update(ctx context.Context, app UpdateUser) (User, error) {
	uu, err := toBusUpdateUser(app)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	usr, err := mid.GetUser(ctx)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "user missing in context: %s", err)
	}

	updUsr, err := a.userBus.Update(ctx, usr, uu)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "update: userID[%s] uu[%+v]: %s", usr.ID, uu, err)
	}

	return toAppUser(updUsr), nil
}

// UpdateRole updates an existing user's role.
func (a *App) UpdateRole(ctx context.Context, app UpdateUserRole) (User, error) {
	uu, err := toBusUpdateUserRole(app)
	if err != nil {
		return User{}, errs.New(errs.InvalidArgument, err)
	}

	usr, err := mid.GetUser(ctx)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "user missing in context: %s", err)
	}

	updUsr, err := a.userBus.Update(ctx, usr, uu)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "updaterole: userID[%s] uu[%+v]: %s", usr.ID, uu, err)
	}

	return toAppUser(updUsr), nil
}

// Delete removes a user from the system.
func (a *App) Delete(ctx context.Context) error {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "userID missing in context: %s", err)
	}

	if err := a.userBus.Delete(ctx, usr); err != nil {
		return errs.Newf(errs.Internal, "delete: userID[%s]: %s", usr.ID, err)
	}

	return nil
}

// Query returns a list of users with paging.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[User], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[User]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[User]{}, err
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[User]{}, errs.NewFieldsError("order", err)
	}

	usrs, err := a.userBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[User]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.userBus.Count(ctx, filter)
	if err != nil {
		return query.Result[User]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppUsers(usrs), total, page), nil
}

// QueryByID returns a user by its ID.
func (a *App) QueryByID(ctx context.Context) (User, error) {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return User{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return toAppUser(usr), nil
}

func (a *App) ApproveUser(ctx context.Context) error {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "approveuser: %s", err)
	}

	usrID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "approveuser: %s", err)
	}

	// we don't need to send the user information back to the user probably
	if err := a.userBus.Approve(ctx, usr, usrID); err != nil {
		return errs.Newf(errs.Internal, "approveuser: userID[%s]: %s", usr.ID, err)
	}

	return nil
}

func (a *App) DenyUser(ctx context.Context) error {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "deny user: %s", err)
	}

	if err := a.userBus.Deny(ctx, usr); err != nil {
		return errs.Newf(errs.Internal, "deny user: userID[%s]: %s", usr.ID, err)
	}

	return nil
}

func (a *App) SetUserUnderReview(ctx context.Context) error {
	usr, err := mid.GetUser(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "SetUserPending: %s", err)
	}
	if err := a.userBus.SetUnderReview(ctx, usr); err != nil {
		return errs.Newf(errs.Internal, "SetUserPending: userID[%s]: %s", usr.ID, err)
	}

	return nil
}
