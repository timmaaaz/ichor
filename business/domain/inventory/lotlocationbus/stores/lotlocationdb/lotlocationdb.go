package lotlocationdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/lotlocationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for lot location database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (lotlocationbus.Storer, error) {
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

func (s *Store) Create(ctx context.Context, ll lotlocationbus.LotLocation) error {
	const q = `
	INSERT INTO inventory.lot_locations (
		id, lot_id, location_id, quantity, created_date, updated_date
	) VALUES (
		:id, :lot_id, :location_id, :quantity, :created_date, :updated_date
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLotLocation(ll)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", lotlocationbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", lotlocationbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, ll lotlocationbus.LotLocation) error {
	const q = `
	UPDATE
		inventory.lot_locations
	SET
		id = :id,
		lot_id = :lot_id,
		location_id = :location_id,
		quantity = :quantity,
		updated_date = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLotLocation(ll)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", lotlocationbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", lotlocationbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, ll lotlocationbus.LotLocation) error {
	const q = `
	DELETE FROM
		inventory.lot_locations
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLotLocation(ll)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter lotlocationbus.QueryFilter, orderBy order.By, page page.Page) ([]lotlocationbus.LotLocation, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, lot_id, location_id, quantity, created_date, updated_date
	FROM
		inventory.lot_locations
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbLotLocations []lotLocation

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbLotLocations); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusLotLocations(dbLotLocations), nil
}

func (s *Store) Count(ctx context.Context, filter lotlocationbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		inventory.lot_locations
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

func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (lotlocationbus.LotLocation, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT
		id, lot_id, location_id, quantity, created_date, updated_date
	FROM
		inventory.lot_locations
	WHERE
		id = :id
	`

	var dbLL lotLocation

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbLL); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return lotlocationbus.LotLocation{}, lotlocationbus.ErrNotFound
		}
		return lotlocationbus.LotLocation{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	return toBusLotLocation(dbLL), nil
}
