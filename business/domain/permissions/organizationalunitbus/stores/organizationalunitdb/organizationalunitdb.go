package organizationalunitdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (organizationalunitbus.Storer, error) {
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

// Create adds a new organizational unit to the system
func (s *Store) Create(ctx context.Context, ou organizationalunitbus.OrganizationalUnit) error {
	const q = `
	INSERT INTO organizational_units (
		organizational_unit_id, parent_id, name, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active
	) VALUES (
		:organizational_unit_id, :parent_id, :name, :level, :path, :can_inherit_permissions, :can_rollup_data, :unit_type, :is_active
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrganizationalUnit(ou)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return organizationalunitbus.ErrUnique
		}
		return err
	}

	return nil
}

// Update updates an existing organizational unit in the system
func (s *Store) Update(ctx context.Context, ou organizationalunitbus.OrganizationalUnit) error {
	const q = `
	UPDATE organizational_units
	SET
		parent_id = :parent_id,
		name = :name,
		level = :level,
		path = :path,
		can_inherit_permissions = :can_inherit_permissions,
		can_rollup_data = :can_rollup_data,
		unit_type = :unit_type,
		is_active = :is_active
	WHERE
		organizational_unit_id = :organizational_unit_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrganizationalUnit(ou)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", organizationalunitbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an organizational unit from the system
func (s *Store) Delete(ctx context.Context, ou organizationalunitbus.OrganizationalUnit) error {
	const q = `
	DELETE FROM 
		organizational_units
	WHERE 
		organizational_unit_id = :organizational_unit_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBOrganizationalUnit(ou)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of organizational units from the system
func (s *Store) Query(ctx context.Context, filter organizationalunitbus.QueryFilter, orderBy order.By, page page.Page) ([]organizationalunitbus.OrganizationalUnit, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		organizational_unit_id, parent_id, name, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active
	FROM
		organizational_units`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbOrgUnits []organizationalUnit
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbOrgUnits); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusOrganizationalUnits(dbOrgUnits), nil
}

// Count returns the total number of org units in the DB.
func (s *Store) Count(ctx context.Context, filter organizationalunitbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		organizational_units`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves an organizational unit by its ID
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (organizationalunitbus.OrganizationalUnit, error) {
	const q = `
	SELECT
		organizational_unit_id, parent_id, name, level, path, can_inherit_permissions, can_rollup_data, unit_type, is_active
	FROM
		organizational_units
	WHERE
		organizational_unit_id = :organizational_unit_id
	`

	data := map[string]any{"organizational_unit_id": id}
	var dbOrgUnit organizationalUnit
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbOrgUnit); err != nil {
		return organizationalunitbus.OrganizationalUnit{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusOrganizationalUnit(dbOrgUnit), nil
}

// You'll need to add this method to your Store
func (s *Store) QueryByParentID(ctx context.Context, parentID uuid.UUID) ([]organizationalunitbus.OrganizationalUnit, error) {
	const q = `
	SELECT
		organizational_unit_id, parent_id, name, level, path,
		can_inherit_permissions, can_rollup_data, unit_type, is_active
	FROM
		organizational_units
	WHERE
		parent_id = :parent_id
	`

	data := map[string]any{"parent_id": parentID}

	var dbOUs []organizationalUnit
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbOUs); err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return toBusOrganizationalUnits(dbOUs), nil
}
