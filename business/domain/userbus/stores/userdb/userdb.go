// Package userdb contains user related CRUD functionality.
package userdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/mail"

	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/order"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/page"
	"bitbucket.org/superiortechnologies/ichor/business/sdk/sqldb"
	"bitbucket.org/superiortechnologies/ichor/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (userbus.Storer, error) {
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

// Create inserts a new user into the database.
func (s *Store) Create(ctx context.Context, usr userbus.User) error {
	const q = `
	INSERT INTO users (
		user_id, requested_by, approved_by, title_id, office_id, work_phone_id, 
		cell_phone_id, username, first_name, last_name, email, birthday, roles, 
		system_roles, password_hash, enabled, date_hired, date_requested, 
		date_approved, date_created, date_updated
	) VALUES (
		:user_id, :requested_by, :approved_by, :title_id, :office_id, 
		:work_phone_id, :cell_phone_id, :username, :first_name, :last_name, 
		:email, :birthday, :roles, :system_roles, :password_hash, :enabled, 
		:date_hired, :date_requested, :date_approved, :date_created, 
		:date_updated
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", userbus.ErrUniqueEmail)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s *Store) Update(ctx context.Context, usr userbus.User) error {
	const q = `
	UPDATE
		users
	SET 
		requested_by = :requested_by,
		approved_by = :approved_by,
		title_id = :title_id,
		office_id = :office_id,
		work_phone_id = :work_phone_id,
		cell_phone_id = :cell_phone_id,
		username = :username,
		first_name = :first_name,
		last_name = :last_name,
		email = :email,
		birthday = :birthday,
		roles = :roles,
		system_roles = :system_roles,
		password_hash = :password_hash,
		enabled = :enabled,
		date_hired = :date_hired,
		date_requested = :date_requested,
		date_approved = :date_approved,
		date_updated = :date_updated
	WHERE
		user_id = :user_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return userbus.ErrUniqueEmail
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a user from the database.
func (s *Store) Delete(ctx context.Context, usr userbus.User) error {
	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(ctx context.Context, filter userbus.QueryFilter, orderBy order.By, page page.Page) ([]userbus.User, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		user_id, requested_by, approved_by, title_id, office_id, work_phone_id, 
		cell_phone_id, username, first_name, last_name, email, birthday, roles, 
		system_roles, password_hash, enabled, date_hired, date_requested, 
		date_approved, date_created, date_updated
	FROM
		users`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbUsrs []user
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbUsrs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUsers(dbUsrs)
}

// Count returns the total number of users in the DB.
func (s *Store) Count(ctx context.Context, filter userbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		count(1)
	FROM
		users`

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

// QueryByID gets the specified user from the database.
func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (userbus.User, error) {
	data := struct {
		ID string `db:"user_id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
        user_id, requested_by, approved_by, title_id, office_id, work_phone_id, 
		cell_phone_id, username, first_name, last_name, email, birthday, roles, 
		system_roles, password_hash, enabled, date_hired, date_requested, 
		date_approved, date_created, date_updated
	FROM
		users
	WHERE 
		user_id = :user_id`

	var dbUsr user
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userbus.User{}, fmt.Errorf("db: %w", userbus.ErrNotFound)
		}
		return userbus.User{}, fmt.Errorf("db: %w", err)
	}

	return toBusUser(dbUsr)
}

// QueryByEmail gets the specified user from the database by email.
func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (userbus.User, error) {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email.Address,
	}

	const q = `
	SELECT
        user_id, requested_by, approved_by, title_id, office_id, work_phone_id, 
		cell_phone_id, username, first_name, last_name, email, birthday, roles, 
		system_roles, password_hash, enabled, date_hired, date_requested, 
		date_approved, date_created, date_updated
	FROM
		users
	WHERE
		email = :email`

	var dbUsr user
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userbus.User{}, fmt.Errorf("db: %w", userbus.ErrNotFound)
		}
		return userbus.User{}, fmt.Errorf("db: %w", err)
	}

	return toBusUser(dbUsr)
}
