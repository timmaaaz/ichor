package assetconditionapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the asset condition domain.
type App struct {
	assetconditionbus *assetconditionbus.Business
	auth              *auth.Auth
}

// NewApp constructs an asset condition app API for use.
func NewApp(assetConditionBus *assetconditionbus.Business) *App {
	return &App{
		assetconditionbus: assetConditionBus,
	}
}

// NewAppWithAuth constructs an asset condition app API for use with auth support.
func NewAppWithAuth(assetConditionBus *assetconditionbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:              ath,
		assetconditionbus: assetConditionBus,
	}
}

// Create adds a new asset condition to the system
func (a *App) Create(ctx context.Context, app NewAssetCondition) (AssetCondition, error) {
	nas, err := toBusNewAssetCondition(app)
	if err != nil {
		return AssetCondition{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.assetconditionbus.Create(ctx, nas)
	if err != nil {
		return AssetCondition{}, err
	}

	return ToAppAssetCondition(as), nil
}

// Update updates an existing asset condition
func (a *App) Update(ctx context.Context, app UpdateAssetCondition, id uuid.UUID) (AssetCondition, error) {
	uas, err := toBusUpdateAssetCondition(app)
	if err != nil {
		return AssetCondition{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.assetconditionbus.QueryByID(ctx, id)
	if err != nil {
		return AssetCondition{}, errs.New(errs.NotFound, assetconditionbus.ErrNotFound)
	}

	updated, err := a.assetconditionbus.Update(ctx, as, uas)
	if err != nil {
		if errors.Is(err, assetconditionbus.ErrNotFound) {
			return AssetCondition{}, errs.New(errs.NotFound, err)
		}
		return AssetCondition{}, errs.Newf(errs.Internal, "update: approvalStatus[%+v]: %s", updated, err)
	}

	return ToAppAssetCondition(updated), nil
}

// Delete removes an existing asset condition
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	as, err := a.assetconditionbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, assetconditionbus.ErrNotFound)
	}

	err = a.assetconditionbus.Delete(ctx, as)
	if err != nil {
		return errs.Newf(errs.Internal, "delete asset condition[%+v]: %s", as, err)
	}

	return nil
}

// Query returns a list of asset conditions based on the filter, order and page
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[AssetCondition], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[AssetCondition]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[AssetCondition]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[AssetCondition]{}, errs.NewFieldsError("orderby", err)
	}

	as, err := a.assetconditionbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[AssetCondition]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.assetconditionbus.Count(ctx, filter)
	if err != nil {
		return query.Result[AssetCondition]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppAssetConditions(as), total, page), nil
}

// QueryByID retrieves the asset condition by ID
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (AssetCondition, error) {
	as, err := a.assetconditionbus.QueryByID(ctx, id)
	if err != nil {
		return AssetCondition{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppAssetCondition(as), nil
}
