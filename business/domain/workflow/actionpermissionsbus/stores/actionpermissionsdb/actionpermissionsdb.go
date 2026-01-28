// Package actionpermissionsdb provides database operations for workflow action permissions.
package actionpermissionsdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for action permissions database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (actionpermissionsbus.Storer, error) {
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

// Create adds a new action permission to the system.
func (s *Store) Create(ctx context.Context, ap actionpermissionsbus.ActionPermission) error {
	const q = `
	INSERT INTO workflow.action_permissions (
		id, role_id, action_type, is_allowed, constraints, created_at, updated_at
	) VALUES (
		:id, :role_id, :action_type, :is_allowed, :constraints, :created_at, :updated_at
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBActionPermission(ap)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return actionpermissionsbus.ErrUnique
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing action permission in the system.
func (s *Store) Update(ctx context.Context, ap actionpermissionsbus.ActionPermission) error {
	const q = `
	UPDATE
		workflow.action_permissions
	SET
		is_allowed = :is_allowed,
		constraints = :constraints,
		updated_at = :updated_at
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBActionPermission(ap)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an action permission from the system.
func (s *Store) Delete(ctx context.Context, ap actionpermissionsbus.ActionPermission) error {
	data := struct {
		ID string `db:"id"`
	}{
		ID: ap.ID.String(),
	}

	const q = `
	DELETE FROM
		workflow.action_permissions
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of action permissions from the system.
func (s *Store) Query(ctx context.Context, filter actionpermissionsbus.QueryFilter, orderBy order.By, pg page.Page) ([]actionpermissionsbus.ActionPermission, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
	SELECT
		id, role_id, action_type, is_allowed, constraints, created_at, updated_at
	FROM
		workflow.action_permissions`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbPerms []dbActionPermission
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbPerms); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusActionPermissions(dbPerms), nil
}

// Count returns the total number of action permissions in the DB.
func (s *Store) Count(ctx context.Context, filter actionpermissionsbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		workflow.action_permissions`

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

// QueryByID retrieves a single action permission by its ID.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (actionpermissionsbus.ActionPermission, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT
		id, role_id, action_type, is_allowed, constraints, created_at, updated_at
	FROM
		workflow.action_permissions
	WHERE
		id = :id`

	var dbPerm dbActionPermission
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbPerm); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return actionpermissionsbus.ActionPermission{}, fmt.Errorf("db: %w", actionpermissionsbus.ErrNotFound)
		}
		return actionpermissionsbus.ActionPermission{}, fmt.Errorf("db: %w", err)
	}

	return toBusActionPermission(dbPerm), nil
}

// QueryByRoleAndAction retrieves a permission by role ID and action type.
func (s *Store) QueryByRoleAndAction(ctx context.Context, roleID uuid.UUID, actionType string) (actionpermissionsbus.ActionPermission, error) {
	data := struct {
		RoleID     string `db:"role_id"`
		ActionType string `db:"action_type"`
	}{
		RoleID:     roleID.String(),
		ActionType: actionType,
	}

	const q = `
	SELECT
		id, role_id, action_type, is_allowed, constraints, created_at, updated_at
	FROM
		workflow.action_permissions
	WHERE
		role_id = :role_id AND action_type = :action_type`

	var dbPerm dbActionPermission
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbPerm); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return actionpermissionsbus.ActionPermission{}, fmt.Errorf("db: %w", actionpermissionsbus.ErrNotFound)
		}
		return actionpermissionsbus.ActionPermission{}, fmt.Errorf("db: %w", err)
	}

	return toBusActionPermission(dbPerm), nil
}

// QueryByRoleIDs retrieves permissions for multiple roles for a specific action type.
// This is used to check if any of a user's roles grants permission for an action.
func (s *Store) QueryByRoleIDs(ctx context.Context, roleIDs []uuid.UUID, actionType string) ([]actionpermissionsbus.ActionPermission, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	// Convert role IDs to strings for the IN clause
	roleIDStrings := make([]string, len(roleIDs))
	for i, id := range roleIDs {
		roleIDStrings[i] = id.String()
	}

	data := map[string]any{
		"role_ids":    roleIDStrings,
		"action_type": actionType,
	}

	const q = `
	SELECT
		id, role_id, action_type, is_allowed, constraints, created_at, updated_at
	FROM
		workflow.action_permissions
	WHERE
		role_id IN (:role_ids) AND action_type = :action_type`

	var dbPerms []dbActionPermission
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &dbPerms); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusActionPermissions(dbPerms), nil
}
