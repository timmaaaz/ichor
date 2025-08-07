package roledb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for role database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (rolebus.Storer, error) {
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

// Create adds a new role to the system
func (s *Store) Create(ctx context.Context, r rolebus.Role) error {
	const q = `
	INSERT INTO roles (
		id, name, description
	) VALUES (
		:id, :name, :description
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRole(r)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", rolebus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a role in the system
func (s *Store) Update(ctx context.Context, r rolebus.Role) error {
	const q = `
	UPDATE 
		roles
	SET
		name = :name,
		description = :description
	WHERE 
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRole(r)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", rolebus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a role from the system
func (s *Store) Delete(ctx context.Context, r rolebus.Role) error {
	const q = `
	DELETE FROM 
		roles
	WHERE 
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBRole(r)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of roles from the system
func (s *Store) Query(ctx context.Context, filter rolebus.QueryFilter, orderBy order.By, page page.Page) ([]rolebus.Role, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, name, description
	FROM
		roles`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbRoles []role
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbRoles); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusRoles(dbRoles), nil
}

// Count returns the total number of roles in the DB.
func (s *Store) Count(ctx context.Context, filter rolebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		roles`

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
func (s *Store) QueryByID(ctx context.Context, roleID uuid.UUID) (rolebus.Role, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: roleID.String(),
	}

	const q = `
	SELECT
		id, name, description
	FROM
		roles
	WHERE
		id = :id`

	var dbRole role
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbRole); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return rolebus.Role{}, fmt.Errorf("db: %w", rolebus.ErrNotFound)
		}
		return rolebus.Role{}, fmt.Errorf("db: %w", err)
	}

	return toBusRole(dbRole), nil
}

// QueryByIDs retrieves a list of roles from the system by their IDs.
func (s *Store) QueryByIDs(ctx context.Context, roleIDs []uuid.UUID) ([]rolebus.Role, error) {
	uuidStrings := make([]string, len(roleIDs))
	for i, id := range roleIDs {
		uuidStrings[i] = id.String()
	}

	data := struct {
		RoleIDs []string `db:"role_ids"`
	}{
		RoleIDs: uuidStrings,
	}

	const q = `
	SELECT
		id, name, description
	FROM
		roles
	WHERE
		id IN (:role_ids)`

	var dbRoles []role
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &dbRoles); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusRoles(dbRoles), nil
}

// QueryAll retrieves all roles from the system.
func (s *Store) QueryAll(ctx context.Context) ([]rolebus.Role, error) {
	const q = `
	SELECT
		id, name, description
	FROM
		roles`

	var dbRoles []role
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbRoles); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusRoles(dbRoles), nil
}
