package lottrackingapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	lottrackingbus *lottrackingbus.Business
	auth           *auth.Auth
}

func NewApp(lottrackingbus *lottrackingbus.Business) *App {
	return &App{
		lottrackingbus: lottrackingbus,
	}
}

func NewAppWithAuth(lottrackingbus *lottrackingbus.Business, auth *auth.Auth) *App {
	return &App{
		lottrackingbus: lottrackingbus,
		auth:           auth,
	}
}

func (a *App) Create(ctx context.Context, app NewLotTracking) (LotTracking, error) {
	nlt, err := toBusNewLotTracking(app)
	if err != nil {
		return LotTracking{}, errs.New(errs.InvalidArgument, err)
	}

	lt, err := a.lottrackingbus.Create(ctx, nlt)
	if err != nil {
		if errors.Is(err, lottrackingbus.ErrUniqueEntry) {
			return LotTracking{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, lottrackingbus.ErrForeignKeyViolation) {
			return LotTracking{}, errs.New(errs.Aborted, err)
		}
		return LotTracking{}, err
	}

	return ToAppLotTracking(lt), nil
}

func (a *App) Update(ctx context.Context, app UpdateLotTracking, id uuid.UUID) (LotTracking, error) {
	ult, err := toBusUpdateLotTracking(app)
	if err != nil {
		return LotTracking{}, errs.New(errs.InvalidArgument, err)
	}

	lt, err := a.lottrackingbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lottrackingbus.ErrNotFound) {
			return LotTracking{}, errs.New(errs.NotFound, lottrackingbus.ErrNotFound)
		}
		return LotTracking{}, err
	}

	lotTracking, err := a.lottrackingbus.Update(ctx, lt, ult)
	if err != nil {
		if errors.Is(err, lottrackingbus.ErrUniqueEntry) {
			return LotTracking{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, lottrackingbus.ErrForeignKeyViolation) {
			return LotTracking{}, errs.New(errs.Aborted, err)
		}
		return LotTracking{}, err
	}

	return ToAppLotTracking(lotTracking), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	lt, err := a.lottrackingbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lottrackingbus.ErrNotFound) {
			return errs.New(errs.NotFound, lottrackingbus.ErrNotFound)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.lottrackingbus.Delete(ctx, lt)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[LotTracking], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[LotTracking]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[LotTracking]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[LotTracking]{}, errs.NewFieldsError("orderby", err)
	}

	lts, err := a.lottrackingbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[LotTracking]{}, errs.Newf(errs.Internal, "query %v", err)
	}

	total, err := a.lottrackingbus.Count(ctx, filter)
	if err != nil {
		return query.Result[LotTracking]{}, errs.Newf(errs.Internal, "count %v", err)
	}

	return query.NewResult(ToAppLotTrackings(lts), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (LotTracking, error) {
	lt, err := a.lottrackingbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lottrackingbus.ErrNotFound) {
			return LotTracking{}, errs.New(errs.NotFound, lottrackingbus.ErrNotFound)
		}
		return LotTracking{}, err
	}

	return ToAppLotTracking(lt), nil
}
