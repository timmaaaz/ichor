package picktaskapp

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
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer APIs for pick task access.
type App struct {
	pickTaskBus       *picktaskbus.Business
	invTransactionBus *inventorytransactionbus.Business
	invItemBus        *inventoryitembus.Business
	db                *sqlx.DB
}

// NewApp constructs a pick task app.
func NewApp(
	pickTaskBus *picktaskbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
	db *sqlx.DB,
) *App {
	return &App{
		pickTaskBus:       pickTaskBus,
		invTransactionBus: invTransactionBus,
		invItemBus:        invItemBus,
		db:                db,
	}
}

// Create adds a new pick task to the system.
func (a *App) Create(ctx context.Context, app NewPickTask) (PickTask, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return PickTask{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	npt, err := toBusNewPickTask(app, userID)
	if err != nil {
		return PickTask{}, errs.New(errs.InvalidArgument, err)
	}

	task, err := a.pickTaskBus.Create(ctx, npt)
	if err != nil {
		if errors.Is(err, picktaskbus.ErrForeignKeyViolation) {
			return PickTask{}, errs.New(errs.Aborted, err)
		}
		return PickTask{}, fmt.Errorf("create: %w", err)
	}

	return ToAppPickTask(task), nil
}

// Update modifies an existing pick task.
//
// Status transitions have special handling:
//   - → in_progress: auto-sets assigned_to from the authenticated user + assigned_at = now
//   - → completed:   atomic 3-way write (task update + PICK transaction + inventory decrement)
//   - → short_picked: same atomic 3-way write as completed, with partial quantity
//   - → cancelled:   plain update, no side effects
func (a *App) Update(ctx context.Context, taskID uuid.UUID, app UpdatePickTask) (PickTask, error) {
	upt, err := toBusUpdatePickTask(app)
	if err != nil {
		return PickTask{}, errs.New(errs.InvalidArgument, err)
	}

	task, err := a.pickTaskBus.QueryByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, picktaskbus.ErrNotFound) {
			return PickTask{}, errs.New(errs.NotFound, err)
		}
		return PickTask{}, fmt.Errorf("querybyid: %w", err)
	}

	// Guard against transitioning out of terminal states.
	if upt.Status != nil {
		if task.Status == picktaskbus.Statuses.Completed ||
			task.Status == picktaskbus.Statuses.ShortPicked ||
			task.Status == picktaskbus.Statuses.Cancelled {
			return PickTask{}, errs.Newf(errs.FailedPrecondition,
				"task is already %s and cannot be transitioned", task.Status)
		}
	}

	// Handle status-driven side effects.
	if upt.Status != nil {
		newStatus := *upt.Status

		switch newStatus {
		case picktaskbus.Statuses.InProgress:
			// Auto-assign to the authenticated user.
			userID, err := mid.GetUserID(ctx)
			if err != nil {
				return PickTask{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
			}
			now := time.Now()
			upt.AssignedTo = &userID
			upt.AssignedAt = &now

		case picktaskbus.Statuses.Completed, picktaskbus.Statuses.ShortPicked:
			// Validate short_picked requires quantity_picked and short_pick_reason.
			if newStatus == picktaskbus.Statuses.ShortPicked {
				if upt.QuantityPicked == nil {
					return PickTask{}, errs.Newf(errs.InvalidArgument, "quantity_picked is required when status is short_picked")
				}
				if upt.ShortPickReason == nil || *upt.ShortPickReason == "" {
					return PickTask{}, errs.Newf(errs.InvalidArgument, "short_pick_reason is required when status is short_picked")
				}
			}
			return a.complete(ctx, task, upt)
		}
	}

	// Plain update (pending field edits, cancel, claim).
	updated, err := a.pickTaskBus.Update(ctx, task, upt)
	if err != nil {
		if errors.Is(err, picktaskbus.ErrForeignKeyViolation) {
			return PickTask{}, errs.New(errs.Aborted, err)
		}
		return PickTask{}, fmt.Errorf("update: %w", err)
	}

	return ToAppPickTask(updated), nil
}

// complete handles the atomic 3-way write when a task is completed or short-picked:
//  1. Update pick task (status, completed_by, completed_at, quantity_picked)
//  2. Create PICK inventory transaction (ledger entry)
//  3. Decrement inventory_item quantity at the source location
//
// All three writes are wrapped in a single DB transaction.
func (a *App) complete(ctx context.Context, task picktaskbus.PickTask, upt picktaskbus.UpdatePickTask) (PickTask, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return PickTask{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	now := time.Now()
	upt.CompletedBy = &userID
	upt.CompletedAt = &now

	// Determine actual quantity picked.
	quantityPicked := task.QuantityToPick
	if upt.QuantityPicked != nil {
		quantityPicked = *upt.QuantityPicked
	}
	upt.QuantityPicked = &quantityPicked

	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return PickTask{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Update pick task status inside the transaction.
	ptBusTx, err := a.pickTaskBus.NewWithTx(tx)
	if err != nil {
		return PickTask{}, fmt.Errorf("new picktask tx: %w", err)
	}

	updated, err := ptBusTx.Update(ctx, task, upt)
	if err != nil {
		return PickTask{}, fmt.Errorf("update task: %w", err)
	}

	// 2. Create PICK inventory transaction (ledger record).
	txBusTx, err := a.invTransactionBus.NewWithTx(tx)
	if err != nil {
		return PickTask{}, fmt.Errorf("new invtransaction tx: %w", err)
	}

	// Create inventory transaction (negative quantity = outbound pick).
	_, err = txBusTx.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       task.ProductID,
		LocationID:      task.LocationID,
		UserID:          userID,
		LotID:           task.LotID,
		Quantity:        -quantityPicked,
		TransactionType: "PICK",
		ReferenceNumber: task.SalesOrderID.String(),
		TransactionDate: now,
	})
	if err != nil {
		return PickTask{}, fmt.Errorf("create inventory transaction: %w", err)
	}

	// 3. Decrement inventory_item quantity at the source location.
	itemBusTx, err := a.invItemBus.NewWithTx(tx)
	if err != nil {
		return PickTask{}, fmt.Errorf("new invitem tx: %w", err)
	}

	if err := itemBusTx.DecrementQuantity(ctx, task.ProductID, task.LocationID, quantityPicked); err != nil {
		return PickTask{}, fmt.Errorf("decrement inventory quantity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return PickTask{}, fmt.Errorf("commit transaction: %w", err)
	}

	return ToAppPickTask(updated), nil
}

// Delete removes a pick task from the system.
func (a *App) Delete(ctx context.Context, taskID uuid.UUID) error {
	task, err := a.pickTaskBus.QueryByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, picktaskbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	if err := a.pickTaskBus.Delete(ctx, task); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of pick tasks based on query parameters.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[PickTask], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[PickTask]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[PickTask]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[PickTask]{}, errs.NewFieldsError("orderBy", err)
	}

	tasks, err := a.pickTaskBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[PickTask]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.pickTaskBus.Count(ctx, filter)
	if err != nil {
		return query.Result[PickTask]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppPickTasks(tasks), total, pg), nil
}

// QueryByID retrieves a single pick task by ID.
func (a *App) QueryByID(ctx context.Context, taskID uuid.UUID) (PickTask, error) {
	task, err := a.pickTaskBus.QueryByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, picktaskbus.ErrNotFound) {
			return PickTask{}, errs.New(errs.NotFound, err)
		}
		return PickTask{}, fmt.Errorf("querybyid: %w", err)
	}

	return ToAppPickTask(task), nil
}
