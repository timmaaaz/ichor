package inspectiondb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for inspection database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (inspectionbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

func (s *Store) Create(ctx context.Context, inspection inspectionbus.Inspection) error {
	const q = `
    INSERT INTO inventory.quality_inspections (
        id, product_id, inspector_id, lot_id, inspection_date, 
		next_inspection_date, status, notes,  created_date, updated_date
    ) VALUES (
        :id, :product_id, :inspector_id, :lot_id, :inspection_date, 
		:next_inspection_date, :status, :notes, :created_date, :updated_date
    )
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInspection(inspection)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inspectionbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inspectionbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Update(ctx context.Context, inspection inspectionbus.Inspection) error {
	const q = `
    UPDATE
        inventory.quality_inspections
    SET
        id = :id,
        product_id = :product_id,
        inspector_id = :inspector_id,
        lot_id = :lot_id,
        inspection_date = :inspection_date,
        next_inspection_date = :next_inspection_date,
        status = :status,
        notes = :notes,
        updated_date = :updated_date
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInspection(inspection)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inspectionbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inspectionbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, inspection inspectionbus.Inspection) error {
	const q = `
    DELETE FROM
        inventory.quality_inspections
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInspection(inspection)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter inspectionbus.QueryFilter, orderBy order.By, page page.Page) ([]inspectionbus.Inspection, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
	    id, product_id, inspector_id, lot_id, inspection_date, 
        next_inspection_date, status, notes,  created_date, updated_date
	FROM
		inventory.quality_inspections
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var inspections []inspection

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &inspections); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusInspections(inspections), nil
}

func (s *Store) Count(ctx context.Context, filter inspectionbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        inventory.quality_inspections
    `

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryint: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, inspectionID uuid.UUID) (inspectionbus.Inspection, error) {
	data := struct {
		InspectionID uuid.UUID `db:"id"`
	}{
		InspectionID: inspectionID,
	}

	const q = `
    SELECT
        id, product_id, inspector_id, lot_id, inspection_date, 
        next_inspection_date, status, notes,  created_date, updated_date
    FROM
        inventory.quality_inspections
    WHERE
        id = :id
    `

	var inspection inspection

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &inspection); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return inspectionbus.Inspection{}, inspectionbus.ErrNotFound
		}
		return inspectionbus.Inspection{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusInspection(inspection), nil
}
