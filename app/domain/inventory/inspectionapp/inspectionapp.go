package inspectionapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

type App struct {
	inspectionbus   *inspectionbus.Business
	lotTrackingsBus *lottrackingsbus.Business
	db              *sqlx.DB
	auth            *auth.Auth
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

// NewAppWithTx constructs an App with dependencies needed for transactional
// composite operations (e.g., fail + quarantine).
func NewAppWithTx(inspectionbus *inspectionbus.Business, lotTrackingsBus *lottrackingsbus.Business, db *sqlx.DB) *App {
	return &App{
		inspectionbus:   inspectionbus,
		lotTrackingsBus: lotTrackingsBus,
		db:              db,
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

// Fail atomically fails an inspection and optionally quarantines the associated lot.
// Both writes happen inside a single ReadCommitted transaction.
func (a *App) Fail(ctx context.Context, id uuid.UUID, app FailInspection) (FailInspectionResult, error) {
	// 1. Look up the inspection.
	inspection, err := a.inspectionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inspectionbus.ErrNotFound) {
			return FailInspectionResult{}, errs.New(errs.NotFound, err)
		}
		return FailInspectionResult{}, fmt.Errorf("fail [querybyid]: %w", err)
	}

	// Guard: cannot fail an already-failed inspection.
	if inspection.Status == "failed" {
		return FailInspectionResult{}, errs.Newf(errs.FailedPrecondition, "inspection is already failed")
	}

	// 2. Begin transaction.
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return FailInspectionResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 3. Update inspection status to "failed" inside the transaction.
	inspBusTx, err := a.inspectionbus.NewWithTx(tx)
	if err != nil {
		return FailInspectionResult{}, fmt.Errorf("new inspection tx: %w", err)
	}

	failedStatus := "failed"
	updateInsp := inspectionbus.UpdateInspection{
		Status: &failedStatus,
		Notes:  &app.Notes,
	}

	updated, err := inspBusTx.Update(ctx, inspection, updateInsp)
	if err != nil {
		return FailInspectionResult{}, fmt.Errorf("update inspection: %w", err)
	}

	// 4. Optionally quarantine the lot.
	lotStatus := ""
	if app.QuarantineLot {
		lot, err := a.lotTrackingsBus.QueryByID(ctx, inspection.LotID)
		if err != nil {
			if errors.Is(err, lottrackingsbus.ErrNotFound) {
				return FailInspectionResult{}, errs.Newf(errs.NotFound, "lot %s not found", inspection.LotID)
			}
			return FailInspectionResult{}, fmt.Errorf("query lot: %w", err)
		}

		ltBusTx, err := a.lotTrackingsBus.NewWithTx(tx)
		if err != nil {
			return FailInspectionResult{}, fmt.Errorf("new lottrackings tx: %w", err)
		}

		quarantined := "quarantined"
		updateLot := lottrackingsbus.UpdateLotTrackings{
			QualityStatus: &quarantined,
		}

		lot, err = ltBusTx.Update(ctx, lot, updateLot)
		if err != nil {
			return FailInspectionResult{}, fmt.Errorf("quarantine lot: %w", err)
		}

		lotStatus = lot.QualityStatus
	}

	// 5. Commit.
	if err := tx.Commit(); err != nil {
		return FailInspectionResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return FailInspectionResult{
		Inspection: ToAppInspection(updated),
		LotStatus:  lotStatus,
	}, nil
}
