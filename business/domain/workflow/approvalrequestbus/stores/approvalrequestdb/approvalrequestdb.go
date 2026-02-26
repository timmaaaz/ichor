// Package approvalrequestdb provides database operations for workflow approval requests.
package approvalrequestdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for approval request database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (approvalrequestbus.Storer, error) {
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

// Create adds a new approval request to the database.
func (s *Store) Create(ctx context.Context, req approvalrequestbus.ApprovalRequest) error {
	const q = `
	INSERT INTO workflow.approval_requests (
		approval_request_id, execution_id, rule_id, action_name,
		approvers, approval_type, status, timeout_hours,
		task_token, approval_message, created_date
	) VALUES (
		:approval_request_id, :execution_id, :rule_id, :action_name,
		:approvers, :approval_type, :status, :timeout_hours,
		:task_token, :approval_message, :created_date
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBApprovalRequest(req)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryByID returns a single approval request by its ID.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (approvalrequestbus.ApprovalRequest, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT ar.*, r.name AS rule_name
	FROM workflow.approval_requests ar
	LEFT JOIN workflow.automation_rules r ON ar.rule_id = r.id
	WHERE ar.approval_request_id = :id`

	var dbReq dbApprovalRequest
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbReq); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return approvalrequestbus.ApprovalRequest{}, approvalrequestbus.ErrNotFound
		}
		return approvalrequestbus.ApprovalRequest{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusApprovalRequest(dbReq)
}

// Resolve atomically transitions a pending approval request to a resolved state.
// Uses conditional UPDATE (WHERE status = 'pending') to prevent race conditions.
// Returns ErrAlreadyResolved if zero rows updated (request was already resolved).
func (s *Store) Resolve(ctx context.Context, id, resolvedBy uuid.UUID, status, reason string) (approvalrequestbus.ApprovalRequest, error) {
	data := struct {
		ID               string `db:"id"`
		Status           string `db:"status"`
		ResolvedBy       string `db:"resolved_by"`
		ResolutionReason string `db:"resolution_reason"`
	}{
		ID:               id.String(),
		Status:           status,
		ResolvedBy:       resolvedBy.String(),
		ResolutionReason: reason,
	}

	const q = `
	UPDATE workflow.approval_requests
	SET status = :status, resolved_by = :resolved_by, resolution_reason = :resolution_reason, resolved_date = NOW()
	WHERE approval_request_id = :id AND status = 'pending'
	RETURNING *`

	var dbReq dbApprovalRequest
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbReq); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return approvalrequestbus.ApprovalRequest{}, approvalrequestbus.ErrAlreadyResolved
		}
		return approvalrequestbus.ApprovalRequest{}, fmt.Errorf("resolve: %w", err)
	}

	return toBusApprovalRequest(dbReq)
}

// Query returns approval requests based on filter criteria.
func (s *Store) Query(ctx context.Context, filter approvalrequestbus.QueryFilter, orderBy order.By, pg page.Page) ([]approvalrequestbus.ApprovalRequest, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
	SELECT ar.*, r.name AS rule_name
	FROM workflow.approval_requests ar
	LEFT JOIN workflow.automation_rules r ON ar.rule_id = r.id
	WHERE TRUE`

	buf := bytes.NewBufferString(q)
	applyFilterWithJoin(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}
	buf.WriteString(orderByClause)

	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbReqs []dbApprovalRequest
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbReqs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusApprovalRequests(dbReqs)
}

// Count returns the total count of approval requests matching the filter.
func (s *Store) Count(ctx context.Context, filter approvalrequestbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `SELECT COUNT(*) FROM workflow.approval_requests`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// IsApprover checks whether a given user is in the approvers array for the request.
func (s *Store) IsApprover(ctx context.Context, approvalID, userID uuid.UUID) (bool, error) {
	data := struct {
		ApprovalID string `db:"approval_id"`
		UserID     string `db:"user_id"`
	}{
		ApprovalID: approvalID.String(),
		UserID:     userID.String(),
	}

	const q = `SELECT EXISTS(SELECT 1 FROM workflow.approval_requests WHERE approval_request_id = :approval_id AND CAST(:user_id AS uuid) = ANY(approvers)) AS exists`

	var result struct {
		Exists bool `db:"exists"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &result); err != nil {
		return false, fmt.Errorf("is approver: %w", err)
	}
	return result.Exists, nil
}
