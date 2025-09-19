package commentdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (commentbus.Storer, error) {
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
func (s *Store) Create(ctx context.Context, as commentbus.UserApprovalComment) error {
	const q = `
    INSERT INTO hr.user_approval_comments (
        id, commenter_id, user_id, comment, created_date
    ) VALUES (
        :id, :commenter_id, :user_id, :comment, :created_date
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserApprovalComment(as)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", commentbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces an approval status document in the database.
func (s *Store) Update(ctx context.Context, as commentbus.UserApprovalComment) error {
	const q = `
	UPDATE hr.user_approval_comments
	SET 
	    comment = :comment
	WHERE 
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserApprovalComment(as)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", commentbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an approval status from the database.
func (s *Store) Delete(ctx context.Context, as commentbus.UserApprovalComment) error {
	const q = `
	DELETE FROM
		hr.user_approval_comments
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserApprovalComment(as)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of approval statuses from the database.
func (s *Store) Query(ctx context.Context, filter commentbus.QueryFilter, orderBy order.By, page page.Page) ([]commentbus.UserApprovalComment, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT 
		id, comment, user_id, commenter_id, created_date
	FROM
		hr.user_approval_comments
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var comments []userApprovalComment

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &comments); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusUserApprovalComments(comments), nil
}

// Count returns the total number of approval statuses
func (s *Store) Count(ctx context.Context, filter commentbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        hr.user_approval_comments`

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
func (s *Store) QueryByID(ctx context.Context, aprvlStatusID uuid.UUID) (commentbus.UserApprovalComment, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: aprvlStatusID.String(),
	}

	const q = `
    SELECT
        id, comment, user_id, commenter_id, created_date
    FROM
        hr.user_approval_comments
    WHERE
        id = :id
    `

	var comment userApprovalComment
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &comment); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return commentbus.UserApprovalComment{}, fmt.Errorf("db: %w", commentbus.ErrNotFound)
		}
		return commentbus.UserApprovalComment{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusUserApprovalComment(comment), nil
}
