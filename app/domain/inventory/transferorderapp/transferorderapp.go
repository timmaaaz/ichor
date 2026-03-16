package transferorderapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	transferorderbus  *transferorderbus.Business
	invTransactionBus *inventorytransactionbus.Business
	invItemBus        *inventoryitembus.Business
	db                *sqlx.DB
	auth              *auth.Auth
}

func NewApp(
	transferorderbus *transferorderbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
) *App {
	return &App{
		transferorderbus:  transferorderbus,
		invTransactionBus: invTransactionBus,
		invItemBus:        invItemBus,
		db:                db,
	}
}

func NewAppWithAuth(
	transferorderbus *transferorderbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
	auth *auth.Auth,
) *App {
	return &App{
		transferorderbus:  transferorderbus,
		invTransactionBus: invTransactionBus,
		invItemBus:        invItemBus,
		db:                db,
		auth:              auth,
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

	approved, err := a.transferorderbus.Approve(ctx, to, userID, "")
	if err != nil {
		if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
			return TransferOrder{}, errs.New(errs.InvalidArgument, err)
		}
		return TransferOrder{}, fmt.Errorf("approve: %w", err)
	}

	return ToAppTransferOrder(approved), nil
}

func (a *App) Reject(ctx context.Context, id uuid.UUID) (TransferOrder, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return TransferOrder{}, errs.New(errs.Unauthenticated, err)
	}

	to, err := a.transferorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return TransferOrder{}, errs.New(errs.NotFound, err)
		}
		return TransferOrder{}, fmt.Errorf("reject [querybyid]: %w", err)
	}

	rejected, err := a.transferorderbus.Reject(ctx, to, userID, "")
	if err != nil {
		if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
			return TransferOrder{}, errs.New(errs.InvalidArgument, err)
		}
		return TransferOrder{}, fmt.Errorf("reject: %w", err)
	}

	return ToAppTransferOrder(rejected), nil
}

func (a *App) Claim(ctx context.Context, id uuid.UUID) (TransferOrder, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return TransferOrder{}, errs.New(errs.Unauthenticated, err)
	}

	to, err := a.transferorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return TransferOrder{}, errs.New(errs.NotFound, err)
		}
		return TransferOrder{}, fmt.Errorf("claim [querybyid]: %w", err)
	}

	claimed, err := a.transferorderbus.Claim(ctx, to, userID)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
			return TransferOrder{}, errs.New(errs.FailedPrecondition, err)
		}
		return TransferOrder{}, fmt.Errorf("claim: %w", err)
	}

	return ToAppTransferOrder(claimed), nil
}

// Execute atomically completes a transfer: marks the order completed, creates
// two inventory transactions (TRANSFER_OUT and TRANSFER_IN), decrements source
// location stock, and increments destination location stock.
func (a *App) Execute(ctx context.Context, id uuid.UUID) (TransferOrder, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return TransferOrder{}, errs.New(errs.Unauthenticated, err)
	}

	to, err := a.transferorderbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return TransferOrder{}, errs.New(errs.NotFound, err)
		}
		return TransferOrder{}, fmt.Errorf("execute [querybyid]: %w", err)
	}

	if to.Status != transferorderbus.StatusInTransit {
		return TransferOrder{}, errs.Newf(errs.FailedPrecondition,
			"transfer must be in_transit, got %s", to.Status)
	}

	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return TransferOrder{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Mark transfer order as completed.
	toBusTx, err := a.transferorderbus.NewWithTx(tx)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("new transferorder tx: %w", err)
	}

	completed, err := toBusTx.Execute(ctx, to, userID)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
			return TransferOrder{}, errs.New(errs.FailedPrecondition, err)
		}
		return TransferOrder{}, fmt.Errorf("execute transfer: %w", err)
	}

	// 2. Create TRANSFER_OUT inventory transaction (source location).
	txBusTx, err := a.invTransactionBus.NewWithTx(tx)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("new invtransaction tx: %w", err)
	}

	refNum := to.TransferID.String()
	now := time.Now()

	_, err = txBusTx.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       to.ProductID,
		LocationID:      to.FromLocationID,
		UserID:          userID,
		Quantity:        to.Quantity,
		TransactionType: "TRANSFER_OUT",
		ReferenceNumber: refNum,
		TransactionDate: now,
	})
	if err != nil {
		return TransferOrder{}, fmt.Errorf("create transfer_out transaction: %w", err)
	}

	// 3. Create TRANSFER_IN inventory transaction (destination location).
	_, err = txBusTx.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       to.ProductID,
		LocationID:      to.ToLocationID,
		UserID:          userID,
		Quantity:        to.Quantity,
		TransactionType: "TRANSFER_IN",
		ReferenceNumber: refNum,
		TransactionDate: now,
	})
	if err != nil {
		return TransferOrder{}, fmt.Errorf("create transfer_in transaction: %w", err)
	}

	// 4. Decrement source location stock.
	itemBusTx, err := a.invItemBus.NewWithTx(tx)
	if err != nil {
		return TransferOrder{}, fmt.Errorf("new invitem tx: %w", err)
	}

	if err := itemBusTx.DecrementQuantity(ctx, to.ProductID, to.FromLocationID, to.Quantity); err != nil {
		return TransferOrder{}, errs.Newf(errs.FailedPrecondition, "insufficient stock at source location: %s", err)
	}

	// 5. Increment destination location stock.
	if err := itemBusTx.UpsertQuantity(ctx, to.ProductID, to.ToLocationID, to.Quantity); err != nil {
		return TransferOrder{}, fmt.Errorf("upsert destination inventory: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return TransferOrder{}, fmt.Errorf("commit transaction: %w", err)
	}

	return ToAppTransferOrder(completed), nil
}
