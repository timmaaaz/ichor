package approvalrequestdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/dbarray"
)

type dbApprovalRequest struct {
	ID               uuid.UUID      `db:"approval_request_id"`
	ExecutionID      uuid.UUID      `db:"execution_id"`
	RuleID           uuid.UUID      `db:"rule_id"`
	ActionName       string         `db:"action_name"`
	Approvers        dbarray.String `db:"approvers"`
	ApprovalType     string         `db:"approval_type"`
	Status           string         `db:"status"`
	TimeoutHours     int            `db:"timeout_hours"`
	TaskToken        string         `db:"task_token"`
	ApprovalMessage  sql.NullString `db:"approval_message"`
	ResolvedBy       sql.NullString `db:"resolved_by"`
	ResolutionReason sql.NullString `db:"resolution_reason"`
	CreatedDate      time.Time      `db:"created_date"`
	ResolvedDate     sql.NullTime   `db:"resolved_date"`
}

func toDBApprovalRequest(req approvalrequestbus.ApprovalRequest) dbApprovalRequest {
	approvers := make(dbarray.String, len(req.Approvers))
	for i, id := range req.Approvers {
		approvers[i] = id.String()
	}

	db := dbApprovalRequest{
		ID:           req.ID,
		ExecutionID:  req.ExecutionID,
		RuleID:       req.RuleID,
		ActionName:   req.ActionName,
		Approvers:    approvers,
		ApprovalType: req.ApprovalType,
		Status:       req.Status,
		TimeoutHours: req.TimeoutHours,
		TaskToken:    req.TaskToken,
		CreatedDate:  req.CreatedDate,
	}

	if req.ApprovalMessage != "" {
		db.ApprovalMessage = sql.NullString{String: req.ApprovalMessage, Valid: true}
	}
	if req.ResolvedBy != nil {
		db.ResolvedBy = sql.NullString{String: req.ResolvedBy.String(), Valid: true}
	}
	if req.ResolutionReason != "" {
		db.ResolutionReason = sql.NullString{String: req.ResolutionReason, Valid: true}
	}
	if req.ResolvedDate != nil {
		db.ResolvedDate = sql.NullTime{Time: *req.ResolvedDate, Valid: true}
	}

	return db
}

func toBusApprovalRequest(db dbApprovalRequest) (approvalrequestbus.ApprovalRequest, error) {
	approvers := make([]uuid.UUID, len(db.Approvers))
	for i, s := range db.Approvers {
		id, err := uuid.Parse(s)
		if err != nil {
			return approvalrequestbus.ApprovalRequest{}, fmt.Errorf("parse approver UUID at index %d: %w", i, err)
		}
		approvers[i] = id
	}

	req := approvalrequestbus.ApprovalRequest{
		ID:           db.ID,
		ExecutionID:  db.ExecutionID,
		RuleID:       db.RuleID,
		ActionName:   db.ActionName,
		Approvers:    approvers,
		ApprovalType: db.ApprovalType,
		Status:       db.Status,
		TimeoutHours: db.TimeoutHours,
		TaskToken:    db.TaskToken,
		CreatedDate:  db.CreatedDate,
	}

	if db.ApprovalMessage.Valid {
		req.ApprovalMessage = db.ApprovalMessage.String
	}
	if db.ResolvedBy.Valid {
		id, err := uuid.Parse(db.ResolvedBy.String)
		if err != nil {
			return approvalrequestbus.ApprovalRequest{}, fmt.Errorf("parse resolved_by UUID: %w", err)
		}
		req.ResolvedBy = &id
	}
	if db.ResolutionReason.Valid {
		req.ResolutionReason = db.ResolutionReason.String
	}
	if db.ResolvedDate.Valid {
		req.ResolvedDate = &db.ResolvedDate.Time
	}

	return req, nil
}

func toBusApprovalRequests(dbs []dbApprovalRequest) ([]approvalrequestbus.ApprovalRequest, error) {
	reqs := make([]approvalrequestbus.ApprovalRequest, len(dbs))
	for i, db := range dbs {
		var err error
		reqs[i], err = toBusApprovalRequest(db)
		if err != nil {
			return nil, err
		}
	}
	return reqs, nil
}
