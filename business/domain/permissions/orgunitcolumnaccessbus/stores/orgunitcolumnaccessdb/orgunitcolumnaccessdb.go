package orgunitcolumnaccessdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/permissions/orgunitcolumnaccessbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (orgunitcolumnaccessbus.Storer, error) {
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

func (s *Store) Exists(ctx context.Context, ouca orgunitcolumnaccessbus.OrgUnitColumnAccess) error {
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
		"table_name":  ouca.TableName,
		"column_name": ouca.ColumnName,
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, checkExists, data, &tmp); err != nil {
		return fmt.Errorf("checking table/column existence: %w", err)
	}

	if !tmp.Exists {
		return fmt.Errorf("exists: %w", orgunitcolumnaccessbus.ErrColumnNotExists)
	}

	return nil
}

// Create adds a new table access to the system
func (s *Store) Create(ctx context.Context, ouca orgunitcolumnaccessbus.OrgUnitColumnAccess) error {
	const q = `
	INSERT INTO org_unit_column_access (
		org_unit_column_access_id, organizational_unit_id, table_name, column_name, can_read, can_update, can_inherit_permissions, can_rollup_data
	) VALUES (
		:org_unit_column_access_id, :organizational_unit_id, :table_name, :column_name, :can_read, :can_update, :can_inherit_permissions, :can_rollup_data
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrgUnitColumnAccess(ouca)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", orgunitcolumnaccessbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update makes changes to a table access in the system
func (s *Store) Update(ctx context.Context, ouca orgunitcolumnaccessbus.OrgUnitColumnAccess) error {
	const q = `
	UPDATE 
		org_unit_column_access
	SET
		organizational_unit_id = :organizational_unit_id,
		table_name = :table_name,
		column_name = :column_name,
		can_read = :can_read,
		can_update = :can_update,
		can_inherit_permissions = :can_inherit_permissions,
		can_rollup_data = :can_rollup_data
	WHERE
		org_unit_column_access_id = :org_unit_column_access_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrgUnitColumnAccess(ouca)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", orgunitcolumnaccessbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a table access from the system
func (s *Store) Delete(ctx context.Context, ouca orgunitcolumnaccessbus.OrgUnitColumnAccess) error {
	const q = `
	DELETE FROM 
		org_unit_column_access
	WHERE 
		org_unit_column_access_id = :org_unit_column_access_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrgUnitColumnAccess(ouca)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of table access from the system
func (s *Store) Query(ctx context.Context, filter orgunitcolumnaccessbus.QueryFilter, orderBy order.By, page page.Page) ([]orgunitcolumnaccessbus.OrgUnitColumnAccess, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		org_unit_column_access_id, organizational_unit_id, table_name, column_name, can_read, can_update, can_inherit_permissions, can_rollup_data
	FROM
		org_unit_column_access`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var oucas []orgUnitColumnAccess

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &oucas); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusOrgUnitColumnAccesses(oucas), nil
}

// Count returns the number of table access that match the filter
func (s *Store) Count(ctx context.Context, filter orgunitcolumnaccessbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(*)
	FROM
		org_unit_column_access`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryscalar: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single table access from the system by its ID
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (orgunitcolumnaccessbus.OrgUnitColumnAccess, error) {
	const q = `
	SELECT
		org_unit_column_access_id, organizational_unit_id, table_name, column_name, can_read, can_update, can_inherit_permissions, can_rollup_data
	FROM
		org_unit_column_access
	WHERE
		org_unit_column_access_id = $1
	`

	var ouca orgUnitColumnAccess
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, &ouca, id); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return orgunitcolumnaccessbus.OrgUnitColumnAccess{}, fmt.Errorf("db: %w", orgunitcolumnaccessbus.ErrNotFound)
		}
		return orgunitcolumnaccessbus.OrgUnitColumnAccess{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusOrgUnitColumnAccess(ouca), nil
}
