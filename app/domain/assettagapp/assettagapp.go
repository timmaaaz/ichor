package assettagapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/assettagbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the asset domain.
type App struct {
	assetTagBus *assettagbus.Business
	auth        *auth.Auth
}

// NewApp constructs a asset tag app API for use.
func NewApp(assetTagBus *assettagbus.Business) *App {
	return &App{
		assetTagBus: assetTagBus,
	}
}

// NewAppWithAuth constructs a asset tag app API for use with auth support.
func NewAppWithAuth(assetTagBus *assettagbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:        ath,
		assetTagBus: assetTagBus,
	}
}

// Create adds a new asset to the system.
func (a *App) Create(ctx context.Context, app NewAssetTag) (AssetTag, error) {
	na, err := toBusNewAssetTag(app)
	if err != nil {
		return AssetTag{}, errs.New(errs.InvalidArgument, err)
	}

	ass, err := a.assetTagBus.Create(ctx, na)
	if err != nil {
		if errors.Is(err, assettagbus.ErrUniqueEntry) {
			return AssetTag{}, errs.New(errs.Aborted, assettagbus.ErrUniqueEntry)
		}
		return AssetTag{}, errs.Newf(errs.Internal, "create: ass[%+v]: %s", ass, err)
	}

	return ToAppAssetTag(ass), err
}

// Update updates an existing asset.
func (a *App) Update(ctx context.Context, app UpdateAssetTag, id uuid.UUID) (AssetTag, error) {
	us, err := toBusUpdateAssetTag(app)
	if err != nil {
		return AssetTag{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.assetTagBus.QueryByID(ctx, id)
	if err != nil {
		return AssetTag{}, errs.New(errs.NotFound, assettagbus.ErrNotFound)
	}

	asset, err := a.assetTagBus.Update(ctx, st, us)
	if err != nil {
		if errors.Is(err, assettagbus.ErrNotFound) {
			return AssetTag{}, errs.New(errs.NotFound, err)
		}
		return AssetTag{}, errs.Newf(errs.Internal, "update: asset[%+v]: %s", asset, err)
	}

	return ToAppAssetTag(asset), nil
}

// Delete removes an existing asset.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := a.assetTagBus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, assettagbus.ErrNotFound)
	}

	err = a.assetTagBus.Delete(ctx, st)
	if err != nil {
		return errs.Newf(errs.Internal, "delete: asset[%+v]: %s", st, err)
	}

	return nil
}

// Query returns a list of assets based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[AssetTag], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[AssetTag]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[AssetTag]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[AssetTag]{}, errs.NewFieldsError("orderby", err)
	}

	assets, err := a.assetTagBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[AssetTag]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.assetTagBus.Count(ctx, filter)
	if err != nil {
		return query.Result[AssetTag]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppAssetTags(assets), total, page), nil
}

// QueryByID retrieves a single asset by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (AssetTag, error) {
	asset, err := a.assetTagBus.QueryByID(ctx, id)
	if err != nil {
		return AssetTag{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppAssetTag(asset), nil
}
