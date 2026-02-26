package transferorderapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	transferorderbus *transferorderbus.Business
	auth             *auth.Auth
}

func NewApp(transferorderbus *transferorderbus.Business) *App {
	return &App{
		transferorderbus: transferorderbus,
	}
}

func NewAppWithAuth(transferorderbus *transferorderbus.Business, auth *auth.Auth) *App {
	return &App{
		transferorderbus: transferorderbus,
		auth:             auth,
	}
}

func (a *App) Create(ctx context.Context, app NewTransferOrder) (TransferOrder, error) {
	nt, err := toBusNewTransferOrder(app)
	if err != nil {
		return TransferOrder{}, errs.New(errs.InvalidArgument, err)
	}

	to, err := a.transferorderbus.Create(ctx, nt)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrUniqueEntry) {
			return TransferOrder{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, transferorderbus.ErrForeignKeyViolation) {
			return TransferOrder{}, errs.New(errs.Aborted, err)
		}
		return TransferOrder{}, fmt.Errorf("create: %w", err)
	}

	return ToAppTransferOrder(to), nil
}

func (a *App) Update(ctx context.Context, id uuid.UUID, app UpdateTransferOrder) (TransferOrder, error) {
	uto, err := toBusUpdateTransferOrder(app)
	if err != nil {
		return TransferOrder{}, errs.New(errs.InvalidArgument, err)
	}

	to, err := a.transferorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return TransferOrder{}, errs.New(errs.NotFound, err)
		}
		return TransferOrder{}, fmt.Errorf("querybyid: %w", err)
	}

	to, err = a.transferorderbus.Update(ctx, to, uto)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrUniqueEntry) {
			return TransferOrder{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, transferorderbus.ErrForeignKeyViolation) {
			return TransferOrder{}, errs.New(errs.Aborted, err)
		}
		return TransferOrder{}, fmt.Errorf("update: %w", err)
	}

	return ToAppTransferOrder(to), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	to, err := a.transferorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	err = a.transferorderbus.Delete(ctx, to)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[TransferOrder], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[TransferOrder]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[TransferOrder]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[TransferOrder]{}, errs.NewFieldsError("orderBy", err)
	}

	items, err := a.transferorderbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[TransferOrder]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.transferorderbus.Count(ctx, filter)
	if err != nil {
		return query.Result[TransferOrder]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppTransferOrders(items), total, page), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (TransferOrder, error) {
	to, err := a.transferorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return TransferOrder{}, errs.New(errs.NotFound, err)
		}
		return TransferOrder{}, fmt.Errorf("querybyid: %w", err)
	}

	return ToAppTransferOrder(to), nil
}

func (a *App) Approve(ctx context.Context, id uuid.UUID) (TransferOrder, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return TransferOrder{}, errs.New(errs.Unauthenticated, err)
	}

	to, err := a.transferorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return TransferOrder{}, errs.New(errs.NotFound, err)
		}
		return TransferOrder{}, fmt.Errorf("approve [querybyid]: %w", err)
	}

	approved, err := a.transferorderbus.Approve(ctx, to, userID)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("approve: %w", err)
	}

	return ToAppTransferOrder(approved), nil
}
