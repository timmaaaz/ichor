package assetapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the asset domain.
type App struct {
	assetBus *assetbus.Business
	auth     *auth.Auth
}

// NewApp constructs a asset app API for use.
func NewApp(assetBus *assetbus.Business) *App {
	return &App{
		assetBus: assetBus,
	}
}

// NewAppWithAuth constructs a asset app API for use with auth support.
func NewAppWithAuth(assetBus *assetbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:     ath,
		assetBus: assetBus,
	}
}

// Create adds a new asset to the system.
func (a *App) Create(ctx context.Context, app NewAsset) (Asset, error) {
	na, err := toBusNewAsset(app)
	if err != nil {
		return Asset{}, errs.New(errs.InvalidArgument, err)
	}

	ass, err := a.assetBus.Create(ctx, na)
	if err != nil {
		if errors.Is(err, assetbus.ErrUniqueEntry) {
			return Asset{}, errs.New(errs.Aborted, assetbus.ErrUniqueEntry)
		}
		return Asset{}, errs.Newf(errs.Internal, "create: ass[%+v]: %s", ass, err)
	}

	return ToAppAsset(ass), err
}

// Update updates an existing asset.
func (a *App) Update(ctx context.Context, app UpdateAsset, id uuid.UUID) (Asset, error) {
	us, err := toBusUpdateAsset(app)
	if err != nil {
		return Asset{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.assetBus.QueryByID(ctx, id)
	if err != nil {
		return Asset{}, errs.New(errs.NotFound, assetbus.ErrNotFound)
	}

	asset, err := a.assetBus.Update(ctx, st, us)
	if err != nil {
		if errors.Is(err, assetbus.ErrNotFound) {
			return Asset{}, errs.New(errs.NotFound, err)
		}
		return Asset{}, errs.Newf(errs.Internal, "update: asset[%+v]: %s", asset, err)
	}

	return ToAppAsset(asset), nil
}

// Delete removes an existing asset.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := a.assetBus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, assetbus.ErrNotFound)
	}

	err = a.assetBus.Delete(ctx, st)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: asset[%+v]: %s", st, err)
	}

	return nil
}

// Query returns a list of assets based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Asset], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Asset]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Asset]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Asset]{}, errs.NewFieldsError("orderby", err)
	}

	assets, err := a.assetBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Asset]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.assetBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Asset]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppAssets(assets), total, page), nil
}

// QueryByID retrieves a single asset by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Asset, error) {
	asset, err := a.assetBus.QueryByID(ctx, id)
	if err != nil {
		return Asset{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppAsset(asset), nil
}
