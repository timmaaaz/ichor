package putawaytaskapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer APIs for put-away task access.
type App struct {
	putAwayTaskBus    *putawaytaskbus.Business
	invTransactionBus *inventorytransactionbus.Business
	invItemBus        *inventoryitembus.Business
	db                *sqlx.DB
}

// NewApp constructs a put-away task app.
func NewApp(
	putAwayTaskBus *putawaytaskbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
) *App {
	return &App{
		putAwayTaskBus:    putAwayTaskBus,
		invTransactionBus: invTransactionBus,
		invItemBus:        invItemBus,
		db:                db,
	}
}

// Create adds a new put-away task to the system.
func (a *App) Create(ctx context.Context, app NewPutAwayTask) (PutAwayTask, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return PutAwayTask{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	npt, err := toBusNewPutAwayTask(app, userID)
	if err != nil {
		return PutAwayTask{}, errs.New(errs.InvalidArgument, err)
	}

	task, err := a.putAwayTaskBus.Create(ctx, npt)
	if err != nil {
		if errors.Is(err, putawaytaskbus.ErrForeignKeyViolation) {
			return PutAwayTask{}, errs.New(errs.Aborted, err)
		}
		return PutAwayTask{}, fmt.Errorf("create: %w", err)
	}

	return ToAppPutAwayTask(task), nil
}

// Update modifies an existing put-away task.
//
// Status transitions have special handling:
//   - → in_progress: auto-sets assigned_to from the authenticated user + assigned_at = now
//   - → completed:   atomic 3-way write (task update + PUT_AWAY transaction + inventory upsert)
//   - → cancelled:   plain update, no side effects
func (a *App) Update(ctx context.Context, taskID uuid.UUID, app UpdatePutAwayTask) (PutAwayTask, error) {
	upt, err := toBusUpdatePutAwayTask(app)
	if err != nil {
		return PutAwayTask{}, errs.New(errs.InvalidArgument, err)
	}

	task, err := a.putAwayTaskBus.QueryByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, putawaytaskbus.ErrNotFound) {
			return PutAwayTask{}, errs.New(errs.NotFound, err)
		}
		return PutAwayTask{}, fmt.Errorf("querybyid: %w", err)
	}

	// Guard against transitioning out of terminal states.
	if upt.Status != nil {
		if task.Status == putawaytaskbus.Statuses.Completed || task.Status == putawaytaskbus.Statuses.Cancelled {
			return PutAwayTask{}, errs.Newf(errs.FailedPrecondition,
				"task is already %s and cannot be transitioned", task.Status)
		}
	}

	// Handle status-driven side effects.
	if upt.Status != nil {
		newStatus := *upt.Status

		switch newStatus {
		case putawaytaskbus.Statuses.InProgress:
			// Auto-assign to the authenticated user.
			userID, err := mid.GetUserID(ctx)
			if err != nil {
				return PutAwayTask{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
			}
			now := time.Now()
			upt.AssignedTo = &userID
			upt.AssignedAt = &now

		case putawaytaskbus.Statuses.Completed:
			return a.complete(ctx, task, upt)
		}
	}

	// Plain update (pending field edits, cancel, claim).
	updated, err := a.putAwayTaskBus.Update(ctx, task, upt)
	if err != nil {
		if errors.Is(err, putawaytaskbus.ErrForeignKeyViolation) {
			return PutAwayTask{}, errs.New(errs.Aborted, err)
		}
		return PutAwayTask{}, fmt.Errorf("update: %w", err)
	}

	return ToAppPutAwayTask(updated), nil
}

// complete handles the atomic 3-way write when a task is marked completed:
//  1. Update put-away task (status=completed, completed_by, completed_at)
//  2. Create PUT_AWAY inventory transaction (ledger entry)
//  3. Upsert inventory_item quantity at destination location
//
// All three writes are wrapped in a single DB transaction.
func (a *App) complete(ctx context.Context, task putawaytaskbus.PutAwayTask, upt putawaytaskbus.UpdatePutAwayTask) (PutAwayTask, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return PutAwayTask{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	now := time.Now()
	upt.CompletedBy = &userID
	upt.CompletedAt = &now

	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return PutAwayTask{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Update put-away task status inside the transaction.
	patBusTx, err := a.putAwayTaskBus.NewWithTx(tx)
	if err != nil {
		return PutAwayTask{}, fmt.Errorf("new putawaytask tx: %w", err)
	}

	updated, err := patBusTx.Update(ctx, task, upt)
	if err != nil {
		if errors.Is(err, putawaytaskbus.ErrForeignKeyViolation) {
			return PutAwayTask{}, errs.New(errs.Aborted, err)
		}
		return PutAwayTask{}, fmt.Errorf("update task: %w", err)
	}

	// 2. Create PUT_AWAY inventory transaction (ledger record).
	txBusTx, err := a.invTransactionBus.NewWithTx(tx)
	if err != nil {
		return PutAwayTask{}, fmt.Errorf("new invtransaction tx: %w", err)
	}

	_, err = txBusTx.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       task.ProductID,
		LocationID:      task.LocationID,
		UserID:          userID,
		Quantity:        task.Quantity,
		TransactionType: "PUT_AWAY",
		ReferenceNumber: task.ReferenceNumber,
		TransactionDate: now,
	})
	if err != nil {
		if errors.Is(err, inventorytransactionbus.ErrForeignKeyViolation) {
			return PutAwayTask{}, errs.New(errs.Aborted, err)
		}
		return PutAwayTask{}, fmt.Errorf("create inventory transaction: %w", err)
	}

	// 3. Upsert inventory_item quantity at the destination location.
	itemBusTx, err := a.invItemBus.NewWithTx(tx)
	if err != nil {
		return PutAwayTask{}, fmt.Errorf("new invitem tx: %w", err)
	}

	if err := itemBusTx.UpsertQuantity(ctx, task.ProductID, task.LocationID, task.Quantity); err != nil {
		return PutAwayTask{}, fmt.Errorf("upsert inventory quantity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return PutAwayTask{}, fmt.Errorf("commit transaction: %w", err)
	}

	return ToAppPutAwayTask(updated), nil
}

// Delete removes a put-away task from the system.
func (a *App) Delete(ctx context.Context, taskID uuid.UUID) error {
	task, err := a.putAwayTaskBus.QueryByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, putawaytaskbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	if err := a.putAwayTaskBus.Delete(ctx, task); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of put-away tasks based on query parameters.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[PutAwayTask], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[PutAwayTask]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[PutAwayTask]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[PutAwayTask]{}, errs.NewFieldsError("orderBy", err)
	}

	tasks, err := a.putAwayTaskBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[PutAwayTask]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.putAwayTaskBus.Count(ctx, filter)
	if err != nil {
		return query.Result[PutAwayTask]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppPutAwayTasks(tasks), total, pg), nil
}

// QueryByID retrieves a single put-away task by ID.
func (a *App) QueryByID(ctx context.Context, taskID uuid.UUID) (PutAwayTask, error) {
	task, err := a.putAwayTaskBus.QueryByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, putawaytaskbus.ErrNotFound) {
			return PutAwayTask{}, errs.New(errs.NotFound, err)
		}
		return PutAwayTask{}, fmt.Errorf("querybyid: %w", err)
	}

	return ToAppPutAwayTask(task), nil
}
