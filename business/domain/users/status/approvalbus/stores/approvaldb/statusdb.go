package approvaldb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for approval status database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (approvalbus.Storer, error) {
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

// Create inserts a new approval status into the database.
func (s *Store) Create(ctx context.Context, as approvalbus.UserApprovalStatus) error {
	const q = `
    INSERT INTO user_approval_status (
        user_approval_status_id, icon_id, name
    ) VALUES (
        :user_approval_status_id, :icon_id, :name
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserApprovalStatus(as)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", approvalbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces an approval status document in the database.
func (s *Store) Update(ctx context.Context, as approvalbus.UserApprovalStatus) error {
	const q = `
	UPDATE user_approval_status
	SET 
	    icon_id = :icon_id,
        name = :name
	WHERE 
		user_approval_status_id = :user_approval_status_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserApprovalStatus(as)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", approvalbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an approval status from the database.
func (s *Store) Delete(ctx context.Context, as approvalbus.UserApprovalStatus) error {
	const q = `
	DELETE FROM
		user_approval_status
	WHERE
		user_approval_status_id = :user_approval_status_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserApprovalStatus(as)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of approval statuses from the database.
func (s *Store) Query(ctx context.Context, filter approvalbus.QueryFilter, orderBy order.By, page page.Page) ([]approvalbus.UserApprovalStatus, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT 
		user_approval_status_id, icon_id, name
	FROM
		user_approval_status
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var aprvlStatuses []userApprovalStatus

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &aprvlStatuses); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUserApprovalStatuses(aprvlStatuses), nil
}

// Count returns the total number of approval statuses
func (s *Store) Count(ctx context.Context, filter approvalbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        user_approval_status`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerysingle: %w", err)
	}

	return count.Count, nil
}

// QueryByID finds the approval status by the specified ID.
func (s *Store) QueryByID(ctx context.Context, aprvlStatusID uuid.UUID) (approvalbus.UserApprovalStatus, error) {
	data := struct {
		ID string `db:"user_approval_status_id"`
	}{
		ID: aprvlStatusID.String(),
	}

	const q = `
    SELECT
        user_approval_status_id, icon_id, name
    FROM
        user_approval_status
    WHERE
        user_approval_status_id = :user_approval_status_id
    `

	var userApprovalStatus userApprovalStatus
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &userApprovalStatus); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return approvalbus.UserApprovalStatus{}, fmt.Errorf("db: %w", approvalbus.ErrNotFound)
		}
		return approvalbus.UserApprovalStatus{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusUserApprovalStatus(userApprovalStatus), nil
}
