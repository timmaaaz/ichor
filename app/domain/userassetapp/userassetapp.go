package userassetapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/userassetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the user asset domain.
type App struct {
	userassetbus *userassetbus.Business
	auth         *auth.Auth
}

// NewApp constructs a user asset app API for use.
func NewApp(userassetbus *userassetbus.Business) *App {
	return &App{
		userassetbus: userassetbus,
	}
}

// NewAppWithAuth constructs a user asset app API for use with auth support.
func NewAppWithAuth(userassetbus *userassetbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:         ath,
		userassetbus: userassetbus,
	}
}

// Create adds a new user asset to the system.
func (a *App) Create(ctx context.Context, app NewUserAsset) (UserAsset, error) {
	na, err := toBusNewUserAsset(app)
	if err != nil {
		return UserAsset{}, errs.New(errs.InvalidArgument, err)
	}

	ass, err := a.userassetbus.Create(ctx, na)
	if err != nil {
		if errors.Is(err, userassetbus.ErrUniqueEntry) {
			return UserAsset{}, errs.New(errs.Aborted, userassetbus.ErrUniqueEntry)
		}
		return UserAsset{}, errs.Newf(errs.Internal, "create: user asset[%+v]: %s", ass, err)
	}

	return ToAppUserAsset(ass), err
}

// Update updates an existing user asset.
func (a *App) Update(ctx context.Context, app UpdateUserAsset, id uuid.UUID) (UserAsset, error) {
	us, err := toBusUpdateUserAsset(app)
	if err != nil {
		return UserAsset{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.userassetbus.QueryByID(ctx, id)
	if err != nil {
		return UserAsset{}, errs.New(errs.NotFound, userassetbus.ErrNotFound)
	}

	asset, err := a.userassetbus.Update(ctx, st, us)
	if err != nil {
		if errors.Is(err, userassetbus.ErrNotFound) {
			return UserAsset{}, errs.New(errs.NotFound, err)
		}
		return UserAsset{}, errs.Newf(errs.Internal, "update: user asset[%+v]: %s", asset, err)
	}

	return ToAppUserAsset(asset), nil
}

// Delete removes an existing user asset.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := a.userassetbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, userassetbus.ErrNotFound)
	}

	err = a.userassetbus.Delete(ctx, st)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: user asset[%+v]: %s", st, err)
	}

	return nil
}

// Query returns a list of user assets based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[UserAsset], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[UserAsset]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[UserAsset]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[UserAsset]{}, errs.NewFieldsError("orderby", err)
	}

	assets, err := a.userassetbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[UserAsset]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.userassetbus.Count(ctx, filter)
	if err != nil {
		return query.Result[UserAsset]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppUserAssets(assets), total, page), nil
}

// QueryByID retrieves a single user asset by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (UserAsset, error) {
	asset, err := a.userassetbus.QueryByID(ctx, id)
	if err != nil {
		return UserAsset{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppUserAsset(asset), nil
}
