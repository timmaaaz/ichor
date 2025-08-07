package lottrackingdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for metrics database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (lottrackingbus.Storer, error) {
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

func (s *Store) Create(ctx context.Context, lot lottrackingbus.LotTracking) error {
	const q = `
	INSERT INTO lot_tracking (
		id, supplier_product_id, lot_number, manufacture_date, expiration_date, received_date, 
		quantity, quality_status, created_date, updated_date
	) VALUES (
		:id, :supplier_product_id, :lot_number, :manufacture_date, :expiration_date, :received_date, 
        :quantity, :quality_status, :created_date, :updated_date
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLotTracking(lot)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", lottrackingbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", lottrackingbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, lot lottrackingbus.LotTracking) error {
	const q = `
    UPDATE
        lot_tracking
    SET
        id = :id,
        supplier_product_id = :supplier_product_id,
        lot_number = :lot_number,
        manufacture_date = :manufacture_date,
        expiration_date = :expiration_date,
        received_date = :received_date,
        quantity = :quantity,
        quality_status = :quality_status,
        updated_date = :updated_date
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLotTracking(lot)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", lottrackingbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", lottrackingbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil

}

func (s *Store) Delete(ctx context.Context, lot lottrackingbus.LotTracking) error {
	const q = `
	DELETE FROM
	    lot_tracking
	WHERE
	    id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLotTracking(lot)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter lottrackingbus.QueryFilter, orderBy order.By, page page.Page) ([]lottrackingbus.LotTracking, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
	    id, supplier_product_id, lot_number, manufacture_date, expiration_date, received_date, 
        quantity, quality_status, created_date, updated_date
	FROM 
		lot_tracking
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbLots []lotTracking

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbLots); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusLotTrackings(dbLots), nil
}

func (s *Store) Count(ctx context.Context, filter lottrackingbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM 
        lot_tracking
    `

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (lottrackingbus.LotTracking, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
    SELECT
        id, supplier_product_id, lot_number, manufacture_date, expiration_date, received_date, 
        quantity, quality_status, created_date, updated_date
    FROM 
        lot_tracking
    WHERE
        id = :id
    `

	var dbLot lotTracking

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbLot); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return lottrackingbus.LotTracking{}, lottrackingbus.ErrNotFound
		}
		return lottrackingbus.LotTracking{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	return toBusLotTracking(dbLot), nil
}
