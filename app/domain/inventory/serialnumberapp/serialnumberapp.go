package serialnumberapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	serialnumberbus *serialnumberbus.Business
	auth            *auth.Auth
}

func NewApp(serialnumberbus *serialnumberbus.Business) *App {
	return &App{
		serialnumberbus: serialnumberbus,
	}
}

func NewAppWithAuth(serialnumberbus *serialnumberbus.Business, auth *auth.Auth) *App {
	return &App{
		serialnumberbus: serialnumberbus,
		auth:            auth,
	}
}

func (a *App) Create(ctx context.Context, newSN NewSerialNumber) (SerialNumber, error) {
	nsn, err := toBusNewSerialNumber(newSN)
	if err != nil {
		return SerialNumber{}, errs.New(errs.InvalidArgument, err)
	}

	sn, err := a.serialnumberbus.Create(ctx, nsn)
	if err != nil {
		if errors.Is(err, serialnumberbus.ErrUniqueEntry) {
			return SerialNumber{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, serialnumberbus.ErrForeignKeyViolation) {
			return SerialNumber{}, errs.New(errs.Aborted, err)
		}
		return SerialNumber{}, err
	}

	return ToAppSerialNumber(sn), nil
}

func (a *App) Update(ctx context.Context, updatedSN UpdateSerialNumber, id uuid.UUID) (SerialNumber, error) {
	usnBus, err := toBusUpdateSerialNumber(updatedSN)
	if err != nil {
		return SerialNumber{}, errs.New(errs.InvalidArgument, err)
	}

	currentSN, err := a.serialnumberbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, serialnumberbus.ErrNotFound) {
			return SerialNumber{}, errs.New(errs.NotFound, err)
		}
		return SerialNumber{}, fmt.Errorf("update [querybyid]: %w", err)
	}

	sn, err := a.serialnumberbus.Update(ctx, currentSN, usnBus)
	if err != nil {
		if errors.Is(err, serialnumberbus.ErrUniqueEntry) {
			return SerialNumber{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, serialnumberbus.ErrForeignKeyViolation) {
			return SerialNumber{}, errs.New(errs.Aborted, err)
		}
		return SerialNumber{}, fmt.Errorf("update: %w", err)
	}

	return ToAppSerialNumber(sn), nil

}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	sn, err := a.serialnumberbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, serialnumberbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.serialnumberbus.Delete(ctx, sn)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[SerialNumber], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[SerialNumber]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[SerialNumber]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[SerialNumber]{}, errs.NewFieldsError("orderBy", err)
	}

	snList, err := a.serialnumberbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[SerialNumber]{}, fmt.Errorf("query: %w", err)
	}

	total, err := a.serialnumberbus.Count(ctx, filter)
	if err != nil {
		return query.Result[SerialNumber]{}, fmt.Errorf("count: %w", err)
	}

	return query.NewResult(ToAppSerialNumbers(snList), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (SerialNumber, error) {
	sn, err := a.serialnumberbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, serialnumberbus.ErrNotFound) {
			return SerialNumber{}, errs.New(errs.NotFound, err)
		}
		return SerialNumber{}, fmt.Errorf("querybyid: %w", err)
	}

	return ToAppSerialNumber(sn), nil
}
