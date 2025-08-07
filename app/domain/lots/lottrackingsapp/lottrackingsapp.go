package lottrackingsapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	lottrackingsbus *lottrackingsbus.Business
	auth            *auth.Auth
}

func NewApp(lottrackingsbus *lottrackingsbus.Business) *App {
	return &App{
		lottrackingsbus: lottrackingsbus,
	}
}

func NewAppWithAuth(lottrackingsbus *lottrackingsbus.Business, auth *auth.Auth) *App {
	return &App{
		lottrackingsbus: lottrackingsbus,
		auth:            auth,
	}
}

func (a *App) Create(ctx context.Context, app NewLotTrackings) (LotTrackings, error) {
	nlt, err := toBusNewLotTrackings(app)
	if err != nil {
		return LotTrackings{}, errs.New(errs.InvalidArgument, err)
	}

	lt, err := a.lottrackingsbus.Create(ctx, nlt)
	if err != nil {
		if errors.Is(err, lottrackingsbus.ErrUniqueEntry) {
			return LotTrackings{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, lottrackingsbus.ErrForeignKeyViolation) {
			return LotTrackings{}, errs.New(errs.Aborted, err)
		}
		return LotTrackings{}, err
	}

	return ToAppLotTrackings(lt), nil
}

func (a *App) Update(ctx context.Context, app UpdateLotTrackings, id uuid.UUID) (LotTrackings, error) {
	ult, err := toBusUpdateLotTrackings(app)
	if err != nil {
		return LotTrackings{}, errs.New(errs.InvalidArgument, err)
	}

	lt, err := a.lottrackingsbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lottrackingsbus.ErrNotFound) {
			return LotTrackings{}, errs.New(errs.NotFound, lottrackingsbus.ErrNotFound)
		}
		return LotTrackings{}, err
	}

	lotTrackings, err := a.lottrackingsbus.Update(ctx, lt, ult)
	if err != nil {
		if errors.Is(err, lottrackingsbus.ErrUniqueEntry) {
			return LotTrackings{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, lottrackingsbus.ErrForeignKeyViolation) {
			return LotTrackings{}, errs.New(errs.Aborted, err)
		}
		return LotTrackings{}, err
	}

	return ToAppLotTrackings(lotTrackings), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	lt, err := a.lottrackingsbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lottrackingsbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.lottrackingsbus.Delete(ctx, lt)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[LotTrackings], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[LotTrackings]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[LotTrackings]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[LotTrackings]{}, errs.NewFieldsError("orderby", err)
	}

	lts, err := a.lottrackingsbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[LotTrackings]{}, errs.Newf(errs.Internal, "query %v", err)
	}

	total, err := a.lottrackingsbus.Count(ctx, filter)
	if err != nil {
		return query.Result[LotTrackings]{}, errs.Newf(errs.Internal, "count %v", err)
	}

	return query.NewResult(ToAppLotTrackingss(lts), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (LotTrackings, error) {
	lt, err := a.lottrackingsbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, lottrackingsbus.ErrNotFound) {
			return LotTrackings{}, errs.New(errs.NotFound, lottrackingsbus.ErrNotFound)
		}
		return LotTrackings{}, err
	}

	return ToAppLotTrackings(lt), nil
}
