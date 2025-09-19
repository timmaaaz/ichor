package zoneapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	zonebus *zonebus.Business
	auth    *auth.Auth
}

func NewApp(zonebus *zonebus.Business) *App {
	return &App{
		zonebus: zonebus,
	}
}

func NewAppWithAuth(zone *zonebus.Business, auth *auth.Auth) *App {
	return &App{
		zonebus: zone,
		auth:    auth,
	}
}

func (a *App) Create(ctx context.Context, app NewZone) (Zone, error) {
	nz, err := toBusNewZone(app)
	if err != nil {
		return Zone{}, errs.New(errs.InvalidArgument, err)
	}

	z, err := a.zonebus.Create(ctx, nz)
	if err != nil {
		if errors.Is(err, zonebus.ErrUniqueEntry) {
			return Zone{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, zonebus.ErrForeignKeyViolation) {
			return Zone{}, errs.New(errs.Aborted, err)
		}
		return Zone{}, err
	}

	return ToAppZone(z), nil
}

func (a *App) Update(ctx context.Context, app UpdateZone, id uuid.UUID) (Zone, error) {
	uz, err := toBusUpdateZone(app)
	if err != nil {
		return Zone{}, errs.New(errs.InvalidArgument, err)
	}

	z, err := a.zonebus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, zonebus.ErrNotFound) {
			return Zone{}, errs.New(errs.NotFound, err)
		}
		return Zone{}, err
	}

	updated, err := a.zonebus.Update(ctx, z, uz)
	if err != nil {
		if errors.Is(err, zonebus.ErrUniqueEntry) {
			return Zone{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, zonebus.ErrForeignKeyViolation) {
			return Zone{}, errs.New(errs.Aborted, err)
		}
		return Zone{}, err
	}

	return ToAppZone(updated), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	z, err := a.zonebus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, zonebus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.zonebus.Delete(ctx, z)
	if err != nil {
		return err
	}

	return nil
}

// Query returns a list of zones based on filters provided
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Zone], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Zone]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Zone]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Zone]{}, errs.NewFieldsError("orderBy", err)
	}

	zones, err := a.zonebus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[Zone]{}, errs.Newf(errs.Internal, "query %v", err)
	}

	total, err := a.zonebus.Count(ctx, filter)
	if err != nil {
		return query.Result[Zone]{}, errs.Newf(errs.Internal, "count %v", err)
	}

	return query.NewResult(ToAppZones(zones), total, page), nil
}

// QueryByID retrieves the zone by ID
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (Zone, error) {
	z, err := a.zonebus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, zonebus.ErrNotFound) {
			return Zone{}, errs.New(errs.NotFound, err)
		}
		return Zone{}, err
	}

	return ToAppZone(z), nil
}
