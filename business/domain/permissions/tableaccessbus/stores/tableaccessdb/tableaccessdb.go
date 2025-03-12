package tableaccessdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for org unit database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (tableaccessbus.Storer, error) {
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

// Create adds a new table access to the system
func (s *Store) Create(ctx context.Context, ta tableaccessbus.TableAccess) error {

	// TODO: Write a test specifically for this
	// First check if the table exists in the database
	const checkTable = `
    SELECT EXISTS (
        SELECT 1 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = :table_name
    )`

	tmp := struct {
		Exists bool
	}{}
	data := map[string]any{
		"table_name": ta.TableName,
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, checkTable, data, &tmp); err != nil {
		return fmt.Errorf("namedquerystruct: %w", err)
	}

	if !tmp.Exists {
		return fmt.Errorf("table[%s]: %w", ta.TableName, tableaccessbus.ErrNonexistentTableName)
	}

	// Now we can insert
	const q = `
	INSERT INTO table_access (
		table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete
	) VALUES (
		:table_access_id, :role_id, :table_name, :can_create, :can_read, :can_update, :can_delete
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTableAccess(ta)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return tableaccessbus.ErrUnique
		}
		// TODO: Custom error for table name doesn't exist
		return err
	}

	return nil
}

// Update modifies a table access in the system
func (s *Store) Update(ctx context.Context, ta tableaccessbus.TableAccess) error {
	const q = `
	UPDATE 
		table_access
	SET 
		role_id = :role_id,
		table_name = :table_name,
		can_create = :can_create,
		can_read = :can_read,
		can_update = :can_update,
		can_delete = :can_delete
	WHERE 
		table_access_id = :table_access_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTableAccess(ta)); err != nil {
		return err
	}

	return nil
}

// Delete removes a table access from the system
func (s *Store) Delete(ctx context.Context, ta tableaccessbus.TableAccess) error {
	const q = `
	DELETE FROM 
		table_access
	WHERE 
		table_access_id = :table_access_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTableAccess(ta)); err != nil {
		return err
	}

	return nil
}

// Query retrieves a list of table accesses from the system
func (s *Store) Query(ctx context.Context, filter tableaccessbus.QueryFilter, orderBy order.By, page page.Page) ([]tableaccessbus.TableAccess, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete
	FROM
		table_access`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var tas []tableAccess
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &tas); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusTableAccesses(tas), nil
}

// Count returns the number of table accesses in the system
func (s *Store) Count(ctx context.Context, filter tableaccessbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(*)
	FROM
		table_access`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count int
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryint: %w", err)
	}

	return count, nil
}

// QueryByID retrieves a single table access from the system
func (s *Store) QueryByID(ctx context.Context, tableAccessID uuid.UUID) (tableaccessbus.TableAccess, error) {
	const q = `
	SELECT
		table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete
	FROM
		table_access
	WHERE
		table_access_id = :table_access_id
	`

	data := map[string]any{"table_access_id": tableAccessID}

	var ta tableAccess
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ta); err != nil {
		return tableaccessbus.TableAccess{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusTableAccess(ta), nil
}

// QueryByRoleIDs retrieves a list of table accesses from the system by role
func (s *Store) QueryByRoleIDs(ctx context.Context, roleIDs []uuid.UUID) ([]tableaccessbus.TableAccess, error) {
	data := map[string]any{"role_ids": roleIDs}

	const q = `
	SELECT
		table_access_id, role_id, table_name, can_create, can_read, can_update, can_delete
	FROM
		table_access
	WHERE
		role_id IN (:role_ids)
	`

	var tas []tableAccess
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &tas); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusTableAccesses(tas), nil
}
