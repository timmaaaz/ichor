package crossunitpermissionsdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/permissions/crossunitpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for cross unit permissions database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (crossunitpermissionsbus.Storer, error) {
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

// Create adds a new cross unit permission to the database.
func (s *Store) Create(ctx context.Context, cup crossunitpermissionsbus.CrossUnitPermission) error {
	const q = `
		INSERT INTO cross_unit_permissions
			(cross_unit_permission_id, source_unit_id, target_unit_id, can_read, can_update, granted_by, valid_from, valid_until, reason)
		VALUES
			(:cross_unit_permission_id, :source_unit_id, :target_unit_id, :can_read, :can_update, :granted_by, :valid_from, :valid_until, :reason)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCrossUnitPermission(cup)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", crossunitpermissionsbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update modifies a cross unit permission in the database.
func (s *Store) Update(ctx context.Context, cup crossunitpermissionsbus.CrossUnitPermission) error {
	const q = `
		UPDATE cross_unit_permissions
		SET
			source_unit_id = :source_unit_id,
			target_unit_id = :target_unit_id,
			can_read = :can_read,
			can_update = :can_update,
			granted_by = :granted_by,
			valid_from = :valid_from,
			valid_until = :valid_until,
			reason = :reason
		WHERE
			cross_unit_permission_id = :cross_unit_permission_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCrossUnitPermission(cup)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", crossunitpermissionsbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes a cross unit permission from the database.
func (s *Store) Delete(ctx context.Context, cup crossunitpermissionsbus.CrossUnitPermission) error {
	const q = `
	DELETE FROM 
		cross_unit_permissions
	WHERE 
		cross_unit_permission_id = :cross_unit_permission_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCrossUnitPermission(cup)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Query retrieves a list of cross unit permissions from the database.
func (s *Store) Query(ctx context.Context, filter crossunitpermissionsbus.QueryFilter, orderBy order.By, page page.Page) ([]crossunitpermissionsbus.CrossUnitPermission, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		cross_unit_permission_id, source_unit_id, target_unit_id, can_read, can_update, granted_by, valid_from, valid_until, reason
	FROM
		cross_unit_permissions`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var cups []crossUnitPermission
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &cups); err != nil {
		return []crossunitpermissionsbus.CrossUnitPermission{}, fmt.Errorf("namedqueryslice: %w", err)
	}
	return toBusCrossUnitPermissions(cups), nil
}

// Count returns the number of cross unit permissions in the database.
func (s *Store) Count(ctx context.Context, filter crossunitpermissionsbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(*) AS count
	FROM
		cross_unit_permissions`

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

// QueryByID retrieves a single cross unit permission from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (crossunitpermissionsbus.CrossUnitPermission, error) {
	const q = `
	SELECT
		cross_unit_permission_id, source_unit_id, target_unit_id, can_read, can_update, granted_by, valid_from, valid_until, reason
	FROM
		cross_unit_permissions
	WHERE
		cross_unit_permission_id = :cross_unit_permission_id`

	var cup crossunitpermissionsbus.CrossUnitPermission
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, &cup, id); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return crossunitpermissionsbus.CrossUnitPermission{}, fmt.Errorf("db: %w", crossunitpermissionsbus.ErrNotFound)
		}
		return crossunitpermissionsbus.CrossUnitPermission{}, fmt.Errorf("querystruct: %w", err)
	}
	return cup, nil
}
