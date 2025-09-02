package lottrackingsdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingsbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (lottrackingsbus.Storer, error) {
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

func (s *Store) Create(ctx context.Context, lot lottrackingsbus.LotTrackings) error {
	const q = `
	INSERT INTO inventory.lot_trackings (
		id, supplier_product_id, lot_number, manufacture_date, expiration_date, received_date, 
		quantity, quality_status, created_date, updated_date
	) VALUES (
		:id, :supplier_product_id, :lot_number, :manufacture_date, :expiration_date, :received_date, 
        :quantity, :quality_status, :created_date, :updated_date
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLotTrackings(lot)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", lottrackingsbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", lottrackingsbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, lot lottrackingsbus.LotTrackings) error {
	const q = `
    UPDATE
        inventory.lot_trackings
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

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLotTrackings(lot)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", lottrackingsbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", lottrackingsbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil

}

func (s *Store) Delete(ctx context.Context, lot lottrackingsbus.LotTrackings) error {
	const q = `
	DELETE FROM
	    inventory.lot_trackings
	WHERE
	    id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLotTrackings(lot)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter lottrackingsbus.QueryFilter, orderBy order.By, page page.Page) ([]lottrackingsbus.LotTrackings, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
	    id, supplier_product_id, lot_number, manufacture_date, expiration_date, received_date, 
        quantity, quality_status, created_date, updated_date
	FROM 
		inventory.lot_trackings
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbLots []lotTrackings

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbLots); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusLotTrackingss(dbLots), nil
}

func (s *Store) Count(ctx context.Context, filter lottrackingsbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM 
        inventory.lot_trackings
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

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (lottrackingsbus.LotTrackings, error) {
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
        inventory.lot_trackings
    WHERE
        id = :id
    `

	var dbLot lotTrackings

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbLot); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return lottrackingsbus.LotTrackings{}, lottrackingsbus.ErrNotFound
		}
		return lottrackingsbus.LotTrackings{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	return toBusLotTrackings(dbLot), nil
}
