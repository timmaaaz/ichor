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

type App struct {
	assettypebus *assettypebus.Business
	auth         *auth.Auth
}

// NewApp constructs an asset type app API for use.
func NewApp(assetTypeBus *assettypebus.Business) *App {
	return &App{
		assettypebus: assetTypeBus,
	}
}

// NewAppWithAuth constructs an asset type app API for use with auth support.
func NewAppWithAuth(assetTypeBus *assettypebus.Business, ath *auth.Auth) *App {
	return &App{
		auth:         ath,
		assettypebus: assetTypeBus,
	}
}

// Create adds a new asset type to the system
func (a *App) Create(ctx context.Context, app NewAssetType) (AssetType, error) {
	nas, err := toBusNewAssetType(app)
	if err != nil {
		return AssetType{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.assettypebus.Create(ctx, nas)
	if err != nil {
		return AssetType{}, err
	}

	return ToAppAssetType(as), nil
}

// Update updates an existing asset type
func (a *App) Update(ctx context.Context, app UpdateAssetType, id uuid.UUID) (AssetType, error) {
	uas, err := toBusUpdateAssetType(app)
	if err != nil {
		return AssetType{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.assettypebus.QueryByID(ctx, id)
	if err != nil {
		return AssetType{}, errs.New(errs.NotFound, assettypebus.ErrNotFound)
	}

	updated, err := a.assettypebus.Update(ctx, as, uas)
	if err != nil {
		if errors.Is(err, assettypebus.ErrNotFound) {
			return AssetType{}, errs.New(errs.NotFound, err)
		}
		return AssetType{}, errs.Newf(errs.Internal, "update: assetType[%+v]: %s", updated, err)
	}

	return ToAppAssetType(updated), nil
}

// Delete removes an existing asset type
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	as, err := a.assettypebus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, assettypebus.ErrNotFound)
	}

	err = a.assettypebus.Delete(ctx, as)
	if err != nil {
		return errs.Newf(errs.Internal, "delete asset type[%+v]: %s", as, err)
	}

	return nil
}

// Query returns a list of asset types based on the filter, order and page
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

	as, err := a.assettypebus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[AssetType]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.assettypebus.Count(ctx, filter)
	if err != nil {
		return query.Result[AssetType]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppAssetTypes(as), total, page), nil
}

// QueryByID retrieves the asset type by ID
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (AssetType, error) {
	as, err := a.assettypebus.QueryByID(ctx, id)
	if err != nil {
		return AssetType{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppAssetType(as), nil
}
