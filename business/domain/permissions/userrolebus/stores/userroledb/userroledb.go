package userroledb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for user database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (userrolebus.Storer, error) {
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

// Create adds a new user role to the system
func (s *Store) Create(ctx context.Context, ur userrolebus.UserRole) error {
	const q = `
	INSERT INTO user_roles (
		user_role_id, user_id, role_id
	) VALUES (
		:user_role_id, :user_id, :role_id
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserRole(ur)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", userrolebus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a user role in the system
func (s *Store) Update(ctx context.Context, ur userrolebus.UserRole) error {
	const q = `
	UPDATE user_roles
	SET
		role_id = :role_id
	WHERE
		user_role_id = :user_role_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserRole(ur)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", userbus.ErrUniqueEmail)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a role from the system
func (s *Store) Delete(ctx context.Context, ur userrolebus.UserRole) error {
	const q = `
	DELETE FROM 
		user_roles
	WHERE 
		user_role_id = :user_role_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserRole(ur)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of roles from the system
func (s *Store) Query(ctx context.Context, filter userrolebus.QueryFilter, orderBy order.By, page page.Page) ([]userrolebus.UserRole, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		user_role_id, user_id, role_id
	FROM
		user_roles`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbUserRoles []userRole
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbUserRoles); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUserRoles(dbUserRoles), nil
}

// Count returns the total number of roles in the DB.
func (s *Store) Count(ctx context.Context, filter userrolebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		user_roles`

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

// QueryByID retrieves a single role from the system by its ID.
func (s *Store) QueryByID(ctx context.Context, roleID uuid.UUID) (userrolebus.UserRole, error) {
	data := struct {
		ID string `db:"user_role_id"`
	}{
		ID: roleID.String(),
	}

	const q = `
	SELECT
		user_role_id, user_id, role_id
	FROM
		user_roles
	WHERE
		user_role_id = :user_role_id`

	var dbUserRole userRole
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUserRole); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userrolebus.UserRole{}, fmt.Errorf("db: %w", userbus.ErrNotFound)
		}
		return userrolebus.UserRole{}, fmt.Errorf("db: %w", err)
	}

	return toBusUserRole(dbUserRole), nil
}

// QueryByUserID retrieves a single role from the system by its ID.
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]userrolebus.UserRole, error) {
	data := struct {
		ID string `db:"user_id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
		user_role_id, user_id, role_id
	FROM
		user_roles
	WHERE
		user_id = :user_id`

	var dbUserRole []userRole
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbUserRole); err != nil {
		return []userrolebus.UserRole{}, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUserRoles(dbUserRole), nil
}
