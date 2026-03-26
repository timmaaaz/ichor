package cyclecountsessionapp

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
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer APIs for cycle count session access.
type App struct {
	cycleCountSessionBus *cyclecountsessionbus.Business
	cycleCountItemBus    *cyclecountitembus.Business
	invAdjustmentBus     *inventoryadjustmentbus.Business
	db                   *sqlx.DB
}

// NewApp constructs a cycle count session app.
func NewApp(
	sessionBus *cyclecountsessionbus.Business,
	itemBus *cyclecountitembus.Business,
	adjBus *inventoryadjustmentbus.Business,
	db *sqlx.DB,
) *App {
	return &App{
		cycleCountSessionBus: sessionBus,
		cycleCountItemBus:    itemBus,
		invAdjustmentBus:     adjBus,
		db:                   db,
	}
}

// Create adds a new cycle count session to the system.
func (a *App) Create(ctx context.Context, app NewCycleCountSession) (CycleCountSession, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return CycleCountSession{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	session, err := a.cycleCountSessionBus.Create(ctx, toBusNewCycleCountSession(app, userID))
	if err != nil {
		if errors.Is(err, cyclecountsessionbus.ErrForeignKeyViolation) {
			return CycleCountSession{}, errs.New(errs.Aborted, err)
		}
		return CycleCountSession{}, fmt.Errorf("create: %w", err)
	}

	return ToAppCycleCountSession(session), nil
}

// Update modifies an existing cycle count session.
//
// Status transitions have special handling:
//   - → completed: atomic write (session update + inventory adjustments for all variance-approved items)
//   - Terminal states (completed, cancelled): no further transitions allowed
func (a *App) Update(ctx context.Context, sessionID uuid.UUID, app UpdateCycleCountSession) (CycleCountSession, error) {
	ucs, err := toBusUpdateCycleCountSession(app)
	if err != nil {
		return CycleCountSession{}, errs.New(errs.InvalidArgument, err)
	}

	session, err := a.cycleCountSessionBus.QueryByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, cyclecountsessionbus.ErrNotFound) {
			return CycleCountSession{}, errs.New(errs.NotFound, err)
		}
		return CycleCountSession{}, fmt.Errorf("querybyid: %w", err)
	}

	// Guard against transitioning out of terminal states.
	if ucs.Status != nil {
		if session.Status == cyclecountsessionbus.Statuses.Completed ||
			session.Status == cyclecountsessionbus.Statuses.Cancelled {
			return CycleCountSession{}, errs.Newf(errs.FailedPrecondition,
				"session is already %s and cannot be transitioned", session.Status)
		}

		// Handle status-driven side effects.
		if *ucs.Status == cyclecountsessionbus.Statuses.Completed {
			if session.Status != cyclecountsessionbus.Statuses.InProgress {
				return CycleCountSession{}, errs.Newf(errs.FailedPrecondition,
					"session must be in_progress to complete, current status: %s", session.Status)
			}
			return a.complete(ctx, session)
		}
	}

	// Plain update.
	updated, err := a.cycleCountSessionBus.Update(ctx, session, ucs)
	if err != nil {
		if errors.Is(err, cyclecountsessionbus.ErrForeignKeyViolation) {
			return CycleCountSession{}, errs.New(errs.Aborted, err)
		}
		return CycleCountSession{}, fmt.Errorf("update: %w", err)
	}

	return ToAppCycleCountSession(updated), nil
}

// complete handles the atomic write when a session is completed:
//  1. Update session status to Completed with completed_date = now
//  2. Create inventory adjustments for all variance_approved items where variance != 0
//
// All writes are wrapped in a single DB transaction.
func (a *App) complete(ctx context.Context, session cyclecountsessionbus.CycleCountSession) (CycleCountSession, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return CycleCountSession{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	now := time.Now()

	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return CycleCountSession{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Update session status to Completed inside the transaction.
	sessionBusTx, err := a.cycleCountSessionBus.NewWithTx(tx)
	if err != nil {
		return CycleCountSession{}, fmt.Errorf("new session tx: %w", err)
	}

	completedDate := now
	updated, err := sessionBusTx.Update(ctx, session, cyclecountsessionbus.UpdateCycleCountSession{
		Status:        &cyclecountsessionbus.Statuses.Completed,
		CompletedDate: &completedDate,
	})
	if err != nil {
		return CycleCountSession{}, fmt.Errorf("update session: %w", err)
	}

	// 2. Query all variance_approved items for this session.
	itemBusTx, err := a.cycleCountItemBus.NewWithTx(tx)
	if err != nil {
		return CycleCountSession{}, fmt.Errorf("new item tx: %w", err)
	}

	varianceApproved := cyclecountitembus.Statuses.VarianceApproved
	items, err := itemBusTx.Query(ctx, cyclecountitembus.QueryFilter{
		SessionID: &session.ID,
		Status:    &varianceApproved,
	}, order.NewBy("id", order.ASC), page.MustParse("1", "1000"))
	if err != nil {
		return CycleCountSession{}, fmt.Errorf("query items: %w", err)
	}

	// 3. Create inventory adjustments for items with non-zero variance.
	adjBusTx, err := a.invAdjustmentBus.NewWithTx(tx)
	if err != nil {
		return CycleCountSession{}, fmt.Errorf("new adjustment tx: %w", err)
	}

	for _, item := range items {
		if item.Variance == nil || *item.Variance == 0 {
			continue
		}

		countedQty := 0
		if item.CountedQuantity != nil {
			countedQty = *item.CountedQuantity
		}

		notes := fmt.Sprintf("Cycle count session %s: system_qty=%d, counted_qty=%d",
			session.Name, item.SystemQuantity, countedQty)

		_, err = adjBusTx.Create(ctx, inventoryadjustmentbus.NewInventoryAdjustment{
			ProductID:      item.ProductID,
			LocationID:     item.LocationID,
			AdjustedBy:     userID,
			ApprovedBy:     &userID,
			ApprovalStatus: inventoryadjustmentbus.ApprovalStatusApproved,
			QuantityChange: *item.Variance,
			ReasonCode:     inventoryadjustmentbus.ReasonCodeCycleCount,
			Notes:          notes,
			AdjustmentDate: now,
		})
		if err != nil {
			return CycleCountSession{}, fmt.Errorf("create adjustment for item %s: %w", item.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return CycleCountSession{}, fmt.Errorf("commit transaction: %w", err)
	}

	return ToAppCycleCountSession(updated), nil
}

// Delete removes a cycle count session from the system.
func (a *App) Delete(ctx context.Context, sessionID uuid.UUID) error {
	session, err := a.cycleCountSessionBus.QueryByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, cyclecountsessionbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [querybyid]: %w", err)
	}

	if err := a.cycleCountSessionBus.Delete(ctx, session); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of cycle count sessions based on query parameters.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[CycleCountSession], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.NewFieldsError("orderBy", err)
	}

	sessions, err := a.cycleCountSessionBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.cycleCountSessionBus.Count(ctx, filter)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppCycleCountSessions(sessions), total, pg), nil
}

// QueryByID retrieves a single cycle count session by ID.
func (a *App) QueryByID(ctx context.Context, sessionID uuid.UUID) (CycleCountSession, error) {
	session, err := a.cycleCountSessionBus.QueryByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, cyclecountsessionbus.ErrNotFound) {
			return CycleCountSession{}, errs.New(errs.NotFound, err)
		}
		return CycleCountSession{}, fmt.Errorf("querybyid: %w", err)
	}

	return ToAppCycleCountSession(session), nil
}
