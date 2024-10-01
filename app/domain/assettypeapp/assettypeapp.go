package assettypeapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/assettypebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the asset type domain.
type App struct {
	assetTypeBus *assettypebus.Business
	auth         *auth.Auth
}

// NewApp constructs a asset type app API for use.
func NewApp(assetTypeBus *assettypebus.Business) *App {
	return &App{
		assetTypeBus: assetTypeBus,
	}
}

// NewAppWithAuth constructs a asset type app API for use with auth support.
func NewAppWithAuth(assetTypeBus *assettypebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:         ath,
		assetTypeBus: assetTypeBus,
	}
}

// Create adds a new asset type to the system.
func (a *App) Create(ctx context.Context, app NewAssetType) (AssetType, error) {
	assetType, err := a.assetTypeBus.Create(ctx, ToBusNewAssetType(app))
	if err != nil {
		if errors.Is(err, assettypebus.ErrUniqueEntry) {
			return AssetType{}, errs.New(errs.Aborted, assettypebus.ErrUniqueEntry)
		}
		return AssetType{}, errs.Newf(errs.Internal, "create: asset type[%+v]: %s", assetType, err)
	}

	return ToAppAssetType(assetType), nil
}

// Update updates an existing asset type.
func (a *App) Update(ctx context.Context, app UpdateAssetType, id uuid.UUID) (AssetType, error) {
	uat := ToBusUpdateAssetType(app)

	at, err := a.assetTypeBus.QueryByID(ctx, id)
	if err != nil {
		return AssetType{}, errs.Newf(errs.NotFound, "update: asset type[%s]: %s", id, err)
	}

	assetType, err := a.assetTypeBus.Update(ctx, at, uat)
	if err != nil {
		if errors.Is(err, assettypebus.ErrNotFound) {
			return AssetType{}, errs.New(errs.NotFound, err)
		}
		return AssetType{}, errs.Newf(errs.Internal, "update: asset type[%+v]: %s", assetType, err)
	}

	return ToAppAssetType(assetType), nil
}

// Delete removes an existing asset type.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	at, err := a.assetTypeBus.QueryByID(ctx, id)
	if err != nil {
		return errs.Newf(errs.NotFound, "delete: asset type[%s]: %s", id, err)
	}

	if err := a.assetTypeBus.Delete(ctx, at); err != nil {
		return errs.Newf(errs.Internal, "delete: asset type[%+v]: %s", at, err)
	}

	return nil
}

// Query returns a list of asset types.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[AssetType], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[AssetType]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[AssetType]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[AssetType]{}, errs.NewFieldsError("orderby", err)
	}

	ats, err := a.assetTypeBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[AssetType]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.assetTypeBus.Count(ctx, filter)
	if err != nil {
		return query.Result[AssetType]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppAssetTypes(ats), total, page), nil
}

// QueryByID returns a single asset type based on the id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (AssetType, error) {
	at, err := a.assetTypeBus.QueryByID(ctx, id)
	if err != nil {
		return AssetType{}, errs.Newf(errs.NotFound, "query: asset type[%s]: %s", id, err)
	}

	return ToAppAssetType(at), nil
}
