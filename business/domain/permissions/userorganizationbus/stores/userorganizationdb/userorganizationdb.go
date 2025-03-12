package userorganizationdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/permissions/userorganizationbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (userorganizationbus.Storer, error) {
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
func (s *Store) Create(ctx context.Context, ua userorganizationbus.UserOrganization) error {
	const q = `
	INSERT INTO user_organizations (
		user_organization_id, user_id, organizational_unit_id, role_id, is_unit_manager, start_date, end_date, created_by
	) VALUES (
		:user_organization_id, :user_id, :organizational_unit_id, :role_id, :is_unit_manager, :start_date, :end_date, :created_by
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserOrganization(ua)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", userorganizationbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update updates a user organization in the system
func (s *Store) Update(ctx context.Context, ua userorganizationbus.UserOrganization) error {
	const q = `
	UPDATE 
		user_organizations
	SET
		user_id = :user_id,
		organizational_unit_id = :organizational_unit_id,
		role_id = :role_id,
		is_unit_manager = :is_unit_manager,
		start_date = :start_date,
		end_date = :end_date,
		created_by = :created_by
	WHERE
		user_organization_id = :user_organization_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserOrganization(ua)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", userorganizationbus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a user organization from the system
func (s *Store) Delete(ctx context.Context, ua userorganizationbus.UserOrganization) error {
	const q = `
	DELETE FROM
		user_organizations
	WHERE
		user_organization_id = :user_organization_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserOrganization(ua)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of user organizations from the system
func (s *Store) Query(ctx context.Context, filter userorganizationbus.QueryFilter, orderBy order.By, page page.Page) ([]userorganizationbus.UserOrganization, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		user_organization_id, user_id, organizational_unit_id, role_id, is_unit_manager, start_date, end_date, created_by, created_at
	FROM
		user_organizations`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var uas []userOrganization

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &uas); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUserOrganizations(uas), nil
}

// Count returns the number of user organizations in the system
func (s *Store) Count(ctx context.Context, filter userorganizationbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(*)
	FROM
		user_organizations`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryint: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single user organization from the system
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (userorganizationbus.UserOrganization, error) {
	data := struct {
		ID string `db:"user_organization_id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT
		user_organization_id, user_id, organizational_unit_id, role_id, is_unit_manager, start_date, end_date, created_by, created_at
	FROM
		user_organizations
	WHERE
		user_organization_id = :user_organization_id
	`

	var ua userOrganization
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ua); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userorganizationbus.UserOrganization{}, fmt.Errorf("db: %w", userorganizationbus.ErrNotFound)
		}
		return userorganizationbus.UserOrganization{}, fmt.Errorf("db: %w", err)
	}

	return toBusUserOrganization(ua), nil
}

// QueryByUserID retrieves a list of user organizations by user ID
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID) (userorganizationbus.UserOrganization, error) {
	data := struct {
		ID string `db:"user_id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
		user_organization_id, user_id, organizational_unit_id, role_id, is_unit_manager, start_date, end_date, created_by, created_at
	FROM
		user_organizations
	WHERE
		user_id = :user_id
	`

	var ua userOrganization
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ua); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userorganizationbus.UserOrganization{}, fmt.Errorf("namedqueryslice: %w", userorganizationbus.ErrNotFound)
		}
		return userorganizationbus.UserOrganization{}, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUserOrganization(ua), nil
}
