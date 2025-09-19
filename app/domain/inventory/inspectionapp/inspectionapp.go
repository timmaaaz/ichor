package inspectionapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	inspectionbus *inspectionbus.Business
	auth          *auth.Auth
}

func NewApp(inspectionbus *inspectionbus.Business) *App {
	return &App{
		inspectionbus: inspectionbus,
	}
}

func NewAppWithAuth(inspectionbus *inspectionbus.Business, auth *auth.Auth) *App {
	return &App{
		inspectionbus: inspectionbus,
		auth:          auth,
	}
}

func (a *App) Create(ctx context.Context, app NewInspection) (Inspection, error) {
	ni, err := toBusNewInspection(app)
	if err != nil {
		return Inspection{}, errs.New(errs.InvalidArgument, err)
	}

	i, err := a.inspectionbus.Create(ctx, ni)
	if err != nil {
		if errors.Is(err, inspectionbus.ErrForeignKeyViolation) {
			return Inspection{}, errs.New(errs.Aborted, err)
		}
		if errors.Is(err, inspectionbus.ErrUniqueEntry) {
			return Inspection{}, errs.New(errs.AlreadyExists, err)
		}
		return Inspection{}, err
	}

	return ToAppInspection(i), nil
}

func (a *App) Update(ctx context.Context, app UpdateInspection, id uuid.UUID) (Inspection, error) {
	ui, err := toBusUpdateInspection(app)
	if err != nil {
		return Inspection{}, errs.New(errs.InvalidArgument, err)
	}

	i, err := a.inspectionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inspectionbus.ErrNotFound) {
			return Inspection{}, errs.New(errs.NotFound, err)
		}
		return Inspection{}, err
	}

	i, err = a.inspectionbus.Update(ctx, i, ui)
	if err != nil {
		if errors.Is(err, inspectionbus.ErrForeignKeyViolation) {
			return Inspection{}, errs.New(errs.Aborted, err)
		}
		if errors.Is(err, inspectionbus.ErrNotFound) {
			return Inspection{}, errs.New(errs.NotFound, err)
		}
		return Inspection{}, err
	}

	return ToAppInspection(i), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	i, err := a.inspectionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inspectionbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.inspectionbus.Delete(ctx, i)
	if err != nil {
		return fmt.Errorf("detlete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Inspection], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Inspection]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Inspection]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Inspection]{}, errs.NewFieldsError("order_by", err)
	}

	inspections, err := a.inspectionbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Inspection]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.inspectionbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Inspection]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppInspections(inspections), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Inspection, error) {
	i, err := a.inspectionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inspectionbus.ErrNotFound) {
			return Inspection{}, errs.New(errs.NotFound, err)
		}
		return Inspection{}, err
	}

	return ToAppInspection(i), nil
}
