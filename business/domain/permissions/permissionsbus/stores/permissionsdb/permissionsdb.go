package permissionsdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (permissionsbus.Storer, error) {
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

func (s *Store) QueryUserPermissions(ctx context.Context, userID uuid.UUID) (permissionsbus.UserPermissions, error) {
	data := struct {
		ID string `db:"user_id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
	    u.user_id,
	    u.username,
	    ta.table_name,
	    bool_or(ta.can_create) AS can_create,
	    bool_or(ta.can_read) AS can_read,
	    bool_or(ta.can_update) AS can_update,
	    bool_or(ta.can_delete) AS can_delete,
	    array_agg(DISTINCT r.name) AS roles
	FROM
	    users u
	JOIN
	    user_roles ur ON u.user_id = ur.user_id
	JOIN
	    roles r ON ur.role_id = r.role_id
	JOIN
	    table_access ta ON r.role_id = ta.role_id
	WHERE
	    u.user_id = :user_id
	GROUP BY
	    u.user_id, u.username, ta.table_name
	ORDER BY
	    ta.table_name;
	`

	var userPermissions []userPermissionsRow

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &userPermissions); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return permissionsbus.UserPermissions{}, fmt.Errorf("db: %w", rolebus.ErrNotFound)
		}
		return permissionsbus.UserPermissions{}, fmt.Errorf("db: %w", err)
	}

	return convertRowsToUserPermissions(userPermissions), nil
}
