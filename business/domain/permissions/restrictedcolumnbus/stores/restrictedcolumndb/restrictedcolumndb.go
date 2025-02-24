package restrictedcolumndb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/permissions/restrictedcolumnbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for restricted column database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (restrictedcolumnbus.Storer, error) {
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

func (s *Store) Exists(ctx context.Context, rc restrictedcolumnbus.RestrictedColumn) error {
	const checkExists = `
   	SELECT EXISTS (
    SELECT 
		1 
    FROM 
		information_schema.columns 
    WHERE 
		table_schema = 'public' 
        AND 
			table_name = :table_name 
        AND 
			column_name = :column_name
	)`

	tmp := struct {
		Exists bool
	}{}
	data := map[string]any{
		"table_name":  rc.TableName,
		"column_name": rc.ColumnName,
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, checkExists, data, &tmp); err != nil {
		return fmt.Errorf("checking table/column existence: %w", err)
	}

	if !tmp.Exists {
		return restrictedcolumnbus.ErrColumnNotExists
	}

	return nil
}

// Create adds a new restricted column to the system
func (s *Store) Create(ctx context.Context, rc restrictedcolumnbus.RestrictedColumn) error {
	const q = `
	INSERT INTO restricted_columns (
		restricted_column_id, table_name, column_name
	) VALUES (
		:restricted_column_id, :table_name, :column_name
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRestrictedColumn(rc)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", restrictedcolumnbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a restricted column from the system
func (s *Store) Delete(ctx context.Context, rc restrictedcolumnbus.RestrictedColumn) error {
	const q = `
	DELETE FROM 
		restricted_columns
	WHERE
		restricted_column_id = :restricted_column_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRestrictedColumn(rc)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of restricted columns from the system
func (s *Store) Query(ctx context.Context, filter restrictedcolumnbus.QueryFilter, orderBy order.By, page page.Page) ([]restrictedcolumnbus.RestrictedColumn, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		restricted_column_id, table_name, column_name
	FROM
		restricted_columns`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbRestrictedColumns []restrictedColumn
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRestrictedColumns); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusRestrictedColumns(dbRestrictedColumns), nil
}

// Count returns the number of restricted columns in the system
func (s *Store) Count(ctx context.Context, filter restrictedcolumnbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(*)
	FROM
		restricted_columns`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count int
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerysingle: %w", err)
	}

	return count, nil
}

// QueryByID retrieves a restricted column from the system by its ID
func (s *Store) QueryByID(ctx context.Context, rcID uuid.UUID) (restrictedcolumnbus.RestrictedColumn, error) {
	data := struct {
		ID string `db:"restricted_column_id"`
	}{
		ID: rcID.String(),
	}

	const q = `
	SELECT
		restricted_column_id, table_name, column_name
	FROM
		restricted_columns
	WHERE
		restricted_column_id = :restricted_column_id
	`

	var dbRC restrictedColumn
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRC); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return restrictedcolumnbus.RestrictedColumn{}, fmt.Errorf("db: %w", rolebus.ErrNotFound)
		}
		return restrictedcolumnbus.RestrictedColumn{}, fmt.Errorf("db: %w", err)
	}

	return toBusRestrictedColumn(dbRC), nil
}
