package assetconditionapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/assets/assetconditionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the asset condition domain.
type App struct {
	assetConditionBus *assetconditionbus.Business
	auth              *auth.Auth
}

// NewApp constructs a asset condition app API for use.
func NewApp(assetConditionBus *assetconditionbus.Business) *App {
	return &App{
		assetConditionBus: assetConditionBus,
	}
}

// NewAppWithAuth constructs a asset condition app API for use with auth support.
func NewAppWithAuth(assetConditionBus *assetconditionbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:              ath,
		assetConditionBus: assetConditionBus,
	}
}

// Create adds a new asset condition to the system.
func (a *App) Create(ctx context.Context, app NewAssetCondition) (AssetCondition, error) {
	assetCondition, err := a.assetConditionBus.Create(ctx, ToBusNewAssetCondition(app))
	if err != nil {
		if errors.Is(err, assetconditionbus.ErrUniqueEntry) {
			return AssetCondition{}, errs.New(errs.Aborted, assetconditionbus.ErrUniqueEntry)
		}
		return AssetCondition{}, errs.Newf(errs.Internal, "create: asset condition[%+v]: %s", assetCondition, err)
	}

	return ToAppAssetCondition(assetCondition), nil
}

// Update updates an existing asset condition.
func (a *App) Update(ctx context.Context, app UpdateAssetCondition, id uuid.UUID) (AssetCondition, error) {
	uat := ToBusUpdateAssetCondition(app)

	at, err := a.assetConditionBus.QueryByID(ctx, id)
	if err != nil {
		return AssetCondition{}, errs.Newf(errs.NotFound, "update: asset condition[%s]: %s", id, err)
	}

	assetCondition, err := a.assetConditionBus.Update(ctx, at, uat)
	if err != nil {
		if errors.Is(err, assetconditionbus.ErrNotFound) {
			return AssetCondition{}, errs.New(errs.NotFound, err)
		}
		return AssetCondition{}, errs.Newf(errs.Internal, "update: asset condition[%+v]: %s", assetCondition, err)
	}

	return ToAppAssetCondition(assetCondition), nil
}

// Delete removes an existing asset condition.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	at, err := a.assetConditionBus.QueryByID(ctx, id)
	if err != nil {
		return errs.Newf(errs.NotFound, "delete: asset condition[%s]: %s", id, err)
	}

	if err := a.assetConditionBus.Delete(ctx, at); err != nil {
		return errs.Newf(errs.Internal, "delete: asset condition[%+v]: %s", at, err)
	}

	return nil
}

// Query returns a list of asset conditions.
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

	ats, err := a.assetConditionBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[AssetCondition]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.assetConditionBus.Count(ctx, filter)
	if err != nil {
		return query.Result[AssetCondition]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppAssetConditions(ats), total, page), nil
}

// QueryByID returns a single asset condition based on the id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (AssetCondition, error) {
	at, err := a.assetConditionBus.QueryByID(ctx, id)
	if err != nil {
		return AssetCondition{}, errs.Newf(errs.NotFound, "query: asset condition[%s]: %s", id, err)
	}

	return ToAppAssetCondition(at), nil
}
