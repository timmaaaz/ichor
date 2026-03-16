package inventoryadjustmentapp

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
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	inventoryadjustmentbus *inventoryadjustmentbus.Business
	invTransactionBus      *inventorytransactionbus.Business
	invItemBus             *inventoryitembus.Business
	db                     *sqlx.DB
	auth                   *auth.Auth
}

// NewApp constructs an inventory adjustment app API for use.
func NewApp(
	inventoryadjustmentbus *inventoryadjustmentbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
) *App {
	return &App{
		inventoryadjustmentbus: inventoryadjustmentbus,
		invTransactionBus:      invTransactionBus,
		invItemBus:             invItemBus,
		db:                     db,
	}
}

// NewAppWithAuth constructs an inventory adjustment app API for use with auth support.
func NewAppWithAuth(
	inventoryadjustmentbus *inventoryadjustmentbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
	ath *auth.Auth,
) *App {
	return &App{
		inventoryadjustmentbus: inventoryadjustmentbus,
		invTransactionBus:      invTransactionBus,
		invItemBus:             invItemBus,
		db:                     db,
		auth:                   ath,
	}
}

// Create creates a new inventory adjustment.
func (a *App) Create(ctx context.Context, app NewInventoryAdjustment) (InventoryAdjustment, error) {
	newAdjustment, err := toBusNewInventoryAdjustment(app)
	if err != nil {
		return InventoryAdjustment{}, errs.New(errs.InvalidArgument, err)
	}

	adjustment, err := a.inventoryadjustmentbus.Create(ctx, newAdjustment)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrUniqueEntry) {
			return InventoryAdjustment{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, inventoryadjustmentbus.ErrForeignKeyViolation) {
			return InventoryAdjustment{}, errs.New(errs.Aborted, err)
		}
		if errors.Is(err, inventoryadjustmentbus.ErrInvalidReasonCode) {
			return InventoryAdjustment{}, errs.New(errs.InvalidArgument, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("create %w", err)
	}

	return ToAppInventoryAdjustment(adjustment), nil

}

func (a *App) Update(ctx context.Context, id uuid.UUID, app UpdateInventoryAdjustment) (InventoryAdjustment, error) {
	updateAdjustment, err := toBusUpdateInventoryAdjustment(app)
	if err != nil {
		return InventoryAdjustment{}, errs.New(errs.InvalidArgument, err)
	}

	adjustment, err := a.inventoryadjustmentbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return InventoryAdjustment{}, errs.New(errs.NotFound, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("update [queryByID] %w", err)
	}

	adjustment, err = a.inventoryadjustmentbus.Update(ctx, adjustment, updateAdjustment)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrUniqueEntry) {
			return InventoryAdjustment{}, errs.New(errs.AlreadyExists, err)
		}
		if errors.Is(err, inventoryadjustmentbus.ErrForeignKeyViolation) {
			return InventoryAdjustment{}, errs.New(errs.Aborted, err)
		}
		if errors.Is(err, inventoryadjustmentbus.ErrInvalidReasonCode) {
			return InventoryAdjustment{}, errs.New(errs.InvalidArgument, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("update %w", err)
	}

	return ToAppInventoryAdjustment(adjustment), nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	adjustment, err := a.inventoryadjustmentbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [queryByID]: %w", err)
	}

	err = a.inventoryadjustmentbus.Delete(ctx, adjustment)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[InventoryAdjustment], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.NewFieldsError("orderBy", err)
	}

	results, err := a.inventoryadjustmentbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.inventoryadjustmentbus.Count(ctx, filter)
	if err != nil {
		return query.Result[InventoryAdjustment]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppInventoryAdjustments(results), total, page), nil
}

// Approve marks an adjustment as approved and atomically applies the quantity
// change to inventory. This is a 3-way atomic write:
//  1. Update adjustment status to approved
//  2. Create an ADJUSTMENT inventory transaction (ledger entry)
//  3. Adjust inventory_item quantity at the adjustment's location
func (a *App) Approve(ctx context.Context, id uuid.UUID) (InventoryAdjustment, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return InventoryAdjustment{}, errs.New(errs.Unauthenticated, err)
	}

	ia, err := a.inventoryadjustmentbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return InventoryAdjustment{}, errs.New(errs.NotFound, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("approve [querybyid]: %w", err)
	}

	if ia.ApprovalStatus != inventoryadjustmentbus.ApprovalStatusPending {
		return InventoryAdjustment{}, errs.New(errs.FailedPrecondition, inventoryadjustmentbus.ErrInvalidApprovalStatus)
	}

	// --- Begin atomic 3-way write ---

	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Update adjustment status to approved.
	adjBusTx, err := a.inventoryadjustmentbus.NewWithTx(tx)
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("new adjustment tx: %w", err)
	}

	updatedIA, err := adjBusTx.Approve(ctx, ia, userID, "")
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrInvalidApprovalStatus) {
			return InventoryAdjustment{}, errs.New(errs.FailedPrecondition, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("approve adjustment: %w", err)
	}

	// 2. Create ADJUSTMENT inventory transaction (ledger entry).
	txBusTx, err := a.invTransactionBus.NewWithTx(tx)
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("new invtransaction tx: %w", err)
	}

	_, err = txBusTx.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       ia.ProductID,
		LocationID:      ia.LocationID,
		UserID:          userID,
		Quantity:        ia.QuantityChange,
		TransactionType: "ADJUSTMENT",
		ReferenceNumber: ia.InventoryAdjustmentID.String(),
		TransactionDate: time.Now(),
	})
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("create inventory transaction: %w", err)
	}

	// 3. Adjust inventory_item quantity at the adjustment location.
	itemBusTx, err := a.invItemBus.NewWithTx(tx)
	if err != nil {
		return InventoryAdjustment{}, fmt.Errorf("new invitem tx: %w", err)
	}

	if err := itemBusTx.AdjustQuantity(ctx, ia.ProductID, ia.LocationID, ia.QuantityChange); err != nil {
		return InventoryAdjustment{}, fmt.Errorf("adjust inventory quantity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return InventoryAdjustment{}, fmt.Errorf("commit transaction: %w", err)
	}

	return ToAppInventoryAdjustment(updatedIA), nil
}

func (a *App) Reject(ctx context.Context, id uuid.UUID) (InventoryAdjustment, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return InventoryAdjustment{}, errs.New(errs.Unauthenticated, err)
	}

	ia, err := a.inventoryadjustmentbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return InventoryAdjustment{}, errs.New(errs.NotFound, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("reject [querybyid]: %w", err)
	}

	ia, err = a.inventoryadjustmentbus.Reject(ctx, ia, userID, "")
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrInvalidApprovalStatus) {
			return InventoryAdjustment{}, errs.New(errs.InvalidArgument, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("reject: %w", err)
	}

	return ToAppInventoryAdjustment(ia), nil
}

func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (InventoryAdjustment, error) {
	adjustment, err := a.inventoryadjustmentbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return InventoryAdjustment{}, errs.New(errs.NotFound, err)
		}
		return InventoryAdjustment{}, fmt.Errorf("querybyid: %w", err)
	}

	return ToAppInventoryAdjustment(adjustment), nil
}
