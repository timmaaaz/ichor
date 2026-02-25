package lotlocationapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	lotlocationbus *lotlocationbus.Business
	auth           *auth.Auth
}

func NewApp(lotlocationbus *lotlocationbus.Business) *App {
	return &App{
		lotlocationbus: lotlocationbus,
	}
}

func NewAppWithAuth(lotlocationbus *lotlocationbus.Business, auth *auth.Auth) *App {
	return &App{
		lotlocationbus: lotlocationbus,
		auth:           auth,
	}
}

func (a *App) Create(ctx context.Context, app NewLotLocation) (LotLocation, error) {
	nll, err := toBusNewLotLocation(app)
	if err != nil {
		return LotLocation{}, errs.New(errs.InvalidArgument, err)
	}

	ll, err := a.lotlocationbus.Create(ctx, nll)
	if err != nil {
		if errors.Is(err, lotlocationbus.ErrUniqueEntry) {
			return LotLocation{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, lotlocationbus.ErrForeignKeyViolation) {
			return LotLocation{}, errs.New(errs.Aborted, err)
		}
		return LotLocation{}, err
	}

	return ToAppLotLocation(ll), nil
}

func (a *App) Update(ctx context.Context, app UpdateLotLocation, id uuid.UUID) (LotLocation, error) {
	ull, err := toBusUpdateLotLocation(app)
	if err != nil {
		return LotLocation{}, errs.New(errs.InvalidArgument, err)
	}

	ll, err := a.lotlocationbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lotlocationbus.ErrNotFound) {
			return LotLocation{}, errs.New(errs.NotFound, lotlocationbus.ErrNotFound)
		}
		return LotLocation{}, err
	}

	lotLocation, err := a.lotlocationbus.Update(ctx, ll, ull)
	if err != nil {
		if errors.Is(err, lotlocationbus.ErrUniqueEntry) {
			return LotLocation{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, lotlocationbus.ErrForeignKeyViolation) {
			return LotLocation{}, errs.New(errs.Aborted, err)
		}
		return LotLocation{}, err
	}

	return ToAppLotLocation(lotLocation), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	ll, err := a.lotlocationbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lotlocationbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	if err := a.lotlocationbus.Delete(ctx, ll); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[LotLocation], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[LotLocation]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[LotLocation]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[LotLocation]{}, errs.NewFieldsError("orderby", err)
	}

	lls, err := a.lotlocationbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[LotLocation]{}, errs.Newf(errs.Internal, "query %v", err)
	}

	total, err := a.lotlocationbus.Count(ctx, filter)
	if err != nil {
		return query.Result[LotLocation]{}, errs.Newf(errs.Internal, "count %v", err)
	}

	return query.NewResult(ToAppLotLocations(lls), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (LotLocation, error) {
	ll, err := a.lotlocationbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lotlocationbus.ErrNotFound) {
			return LotLocation{}, errs.New(errs.NotFound, lotlocationbus.ErrNotFound)
		}
		return LotLocation{}, err
	}

	return ToAppLotLocation(ll), nil
}
