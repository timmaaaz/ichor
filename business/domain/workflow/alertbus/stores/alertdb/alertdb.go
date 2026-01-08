// Package alertdb provides database operations for workflow alerts.
package alertdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for alert database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (alertbus.Storer, error) {
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

// namedExecContextUsingIn executes a query with named parameters that includes IN clauses.
func namedExecContextUsingIn(ctx context.Context, log *logger.Logger, db sqlx.ExtContext, query string, data any) error {
	named, args, err := sqlx.Named(query, data)
	if err != nil {
		return fmt.Errorf("sqlx.Named: %w", err)
	}

	query, args, err = sqlx.In(named, args...)
	if err != nil {
		return fmt.Errorf("sqlx.In: %w", err)
	}

	query = db.Rebind(query)

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("execcontext: %w", err)
	}

	return nil
}

// Create adds a new alert to the system.
func (s *Store) Create(ctx context.Context, a alertbus.Alert) error {
	const q = `
	INSERT INTO workflow.alerts (
		id, alert_type, severity, title, message, context,
		source_entity_name, source_entity_id, source_rule_id,
		status, expires_date, created_date, updated_date
	) VALUES (
		:id, :alert_type, :severity, :title, :message, :context,
		:source_entity_name, :source_entity_id, :source_rule_id,
		:status, :expires_date, :created_date, :updated_date
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAlert(a)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// CreateRecipients adds multiple recipients to an alert (batch insert).
func (s *Store) CreateRecipients(ctx context.Context, recipients []alertbus.AlertRecipient) error {
	if len(recipients) == 0 {
		return nil
	}

	const q = `
	INSERT INTO workflow.alert_recipients (
		id, alert_id, recipient_type, recipient_id, created_date
	) VALUES (
		:id, :alert_id, :recipient_type, :recipient_id, :created_date
	)`

	for _, r := range recipients {
		if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAlertRecipient(r)); err != nil {
			if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
				// Skip duplicate recipients silently
				continue
			}
			return fmt.Errorf("namedexeccontext: %w", err)
		}
	}

	return nil
}

// CreateAcknowledgment adds an acknowledgment record for an alert.
func (s *Store) CreateAcknowledgment(ctx context.Context, ack alertbus.AlertAcknowledgment) error {
	const q = `
	INSERT INTO workflow.alert_acknowledgments (
		id, alert_id, acknowledged_by, acknowledged_date, notes
	) VALUES (
		:id, :alert_id, :acknowledged_by, :acknowledged_date, :notes
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAlertAcknowledgment(ack)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", alertbus.ErrAlreadyAcked)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryByID retrieves a single alert from the system by its ID.
func (s *Store) QueryByID(ctx context.Context, alertID uuid.UUID) (alertbus.Alert, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: alertID.String(),
	}

	const q = `
	SELECT
		id, alert_type, severity, title, message, context,
		source_entity_name, source_entity_id, source_rule_id,
		status, expires_date, created_date, updated_date
	FROM
		workflow.alerts
	WHERE
		id = :id`

	var dbAl dbAlert
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbAl); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return alertbus.Alert{}, fmt.Errorf("db: %w", alertbus.ErrNotFound)
		}
		return alertbus.Alert{}, fmt.Errorf("db: %w", err)
	}

	return toBusAlert(dbAl), nil
}

// Query retrieves a list of alerts from the system (admin only).
func (s *Store) Query(ctx context.Context, filter alertbus.QueryFilter, orderBy order.By, pg page.Page) ([]alertbus.Alert, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
	SELECT
		id, alert_type, severity, title, message, context,
		source_entity_name, source_entity_id, source_rule_id,
		status, expires_date, created_date, updated_date
	FROM
		workflow.alerts`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbAlerts []dbAlert
	if hasSeverities(filter) {
		if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, buf.String(), data, &dbAlerts); err != nil {
			return nil, fmt.Errorf("namedqueryslice: %w", err)
		}
	} else {
		if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbAlerts); err != nil {
			return nil, fmt.Errorf("namedqueryslice: %w", err)
		}
	}

	return toBusAlerts(dbAlerts), nil
}

// QueryByUserID retrieves alerts for a user (directly or via roles).
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter alertbus.QueryFilter, orderBy order.By, pg page.Page) ([]alertbus.Alert, error) {
	data := map[string]any{
		"user_id":       userID.String(),
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	// Build role IDs list for IN clause
	roleIDStrings := make([]string, len(roleIDs))
	for i, id := range roleIDs {
		roleIDStrings[i] = id.String()
	}
	data["role_ids"] = roleIDStrings

	var q string
	if len(roleIDs) > 0 {
		q = `
		SELECT DISTINCT
			a.id, a.alert_type, a.severity, a.title, a.message, a.context,
			a.source_entity_name, a.source_entity_id, a.source_rule_id,
			a.status, a.expires_date, a.created_date, a.updated_date
		FROM
			workflow.alerts a
		INNER JOIN workflow.alert_recipients ar ON a.id = ar.alert_id
		WHERE (
			(ar.recipient_type = 'user' AND ar.recipient_id = :user_id)
			OR (ar.recipient_type = 'role' AND ar.recipient_id IN (:role_ids))
		)`
	} else {
		q = `
		SELECT DISTINCT
			a.id, a.alert_type, a.severity, a.title, a.message, a.context,
			a.source_entity_name, a.source_entity_id, a.source_rule_id,
			a.status, a.expires_date, a.created_date, a.updated_date
		FROM
			workflow.alerts a
		INNER JOIN workflow.alert_recipients ar ON a.id = ar.alert_id
		WHERE
			(ar.recipient_type = 'user' AND ar.recipient_id = :user_id)`
	}

	buf := bytes.NewBufferString(q)
	applyFilterWithJoin(filter, data, buf)

	orderByClause, err := orderByClauseWithPrefix(orderBy, "a")
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbAlerts []dbAlert
	// Use NamedQuerySliceUsingIn when we have role_ids or severities (IN clauses)
	if len(roleIDs) > 0 || hasSeverities(filter) {
		if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, buf.String(), data, &dbAlerts); err != nil {
			return nil, fmt.Errorf("namedqueryslice: %w", err)
		}
	} else {
		if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbAlerts); err != nil {
			return nil, fmt.Errorf("namedqueryslice: %w", err)
		}
	}

	return toBusAlerts(dbAlerts), nil
}

// Count returns the total number of alerts in the DB (admin only).
func (s *Store) Count(ctx context.Context, filter alertbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		workflow.alerts`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	if hasSeverities(filter) {
		if err := sqldb.NamedQueryStructUsingIn(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
			return 0, fmt.Errorf("db: %w", err)
		}
	} else {
		if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
			return 0, fmt.Errorf("db: %w", err)
		}
	}

	return count.Count, nil
}

// CountByUserID returns the count of alerts for a user.
func (s *Store) CountByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID, filter alertbus.QueryFilter) (int, error) {
	data := map[string]any{
		"user_id": userID.String(),
	}

	// Build role IDs list for IN clause
	roleIDStrings := make([]string, len(roleIDs))
	for i, id := range roleIDs {
		roleIDStrings[i] = id.String()
	}
	data["role_ids"] = roleIDStrings

	var q string
	if len(roleIDs) > 0 {
		q = `
		SELECT
			COUNT(DISTINCT a.id) AS count
		FROM
			workflow.alerts a
		INNER JOIN workflow.alert_recipients ar ON a.id = ar.alert_id
		WHERE (
			(ar.recipient_type = 'user' AND ar.recipient_id = :user_id)
			OR (ar.recipient_type = 'role' AND ar.recipient_id IN (:role_ids))
		)`
	} else {
		q = `
		SELECT
			COUNT(DISTINCT a.id) AS count
		FROM
			workflow.alerts a
		INNER JOIN workflow.alert_recipients ar ON a.id = ar.alert_id
		WHERE
			(ar.recipient_type = 'user' AND ar.recipient_id = :user_id)`
	}

	buf := bytes.NewBufferString(q)
	applyFilterWithJoin(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	// Use NamedQueryStructUsingIn when we have role_ids or severities (IN clauses)
	if len(roleIDs) > 0 || hasSeverities(filter) {
		if err := sqldb.NamedQueryStructUsingIn(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
			return 0, fmt.Errorf("db: %w", err)
		}
	} else {
		if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
			return 0, fmt.Errorf("db: %w", err)
		}
	}

	return count.Count, nil
}

// UpdateStatus updates the status of an alert.
func (s *Store) UpdateStatus(ctx context.Context, alertID uuid.UUID, status string, now time.Time) error {
	data := struct {
		ID          string    `db:"id"`
		Status      string    `db:"status"`
		UpdatedDate time.Time `db:"updated_date"`
	}{
		ID:          alertID.String(),
		Status:      status,
		UpdatedDate: now,
	}

	const q = `
	UPDATE
		workflow.alerts
	SET
		status = :status,
		updated_date = :updated_date
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// IsRecipient checks if a user is a recipient of an alert (directly or via roles).
func (s *Store) IsRecipient(ctx context.Context, alertID, userID uuid.UUID, roleIDs []uuid.UUID) (bool, error) {
	data := map[string]any{
		"alert_id": alertID.String(),
		"user_id":  userID.String(),
	}

	// Build role IDs list for IN clause
	roleIDStrings := make([]string, len(roleIDs))
	for i, id := range roleIDs {
		roleIDStrings[i] = id.String()
	}
	data["role_ids"] = roleIDStrings

	var q string
	if len(roleIDs) > 0 {
		q = `
		SELECT
			COUNT(1) AS count
		FROM
			workflow.alert_recipients
		WHERE
			alert_id = :alert_id
			AND (
				(recipient_type = 'user' AND recipient_id = :user_id)
				OR (recipient_type = 'role' AND recipient_id IN (:role_ids))
			)`
	} else {
		q = `
		SELECT
			COUNT(1) AS count
		FROM
			workflow.alert_recipients
		WHERE
			alert_id = :alert_id
			AND recipient_type = 'user'
			AND recipient_id = :user_id`
	}

	var count struct {
		Count int `db:"count"`
	}

	if len(roleIDs) > 0 {
		if err := sqldb.NamedQueryStructUsingIn(ctx, s.log, s.db, q, data, &count); err != nil {
			return false, fmt.Errorf("db: %w", err)
		}
	} else {
		if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
			return false, fmt.Errorf("db: %w", err)
		}
	}

	return count.Count > 0, nil
}

// FilterRecipientAlerts returns the subset of alertIDs where the user is a recipient.
func (s *Store) FilterRecipientAlerts(ctx context.Context, alertIDs []uuid.UUID, userID uuid.UUID, roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(alertIDs) == 0 {
		return nil, nil
	}

	alertIDStrings := make([]string, len(alertIDs))
	for i, id := range alertIDs {
		alertIDStrings[i] = id.String()
	}

	roleIDStrings := make([]string, len(roleIDs))
	for i, id := range roleIDs {
		roleIDStrings[i] = id.String()
	}

	data := map[string]any{
		"alert_ids": alertIDStrings,
		"user_id":   userID.String(),
		"role_ids":  roleIDStrings,
	}

	var q string
	if len(roleIDs) > 0 {
		q = `
		SELECT DISTINCT
			alert_id
		FROM
			workflow.alert_recipients
		WHERE
			alert_id IN (:alert_ids)
			AND (
				(recipient_type = 'user' AND recipient_id = :user_id)
				OR (recipient_type = 'role' AND recipient_id IN (:role_ids))
			)`
	} else {
		q = `
		SELECT DISTINCT
			alert_id
		FROM
			workflow.alert_recipients
		WHERE
			alert_id IN (:alert_ids)
			AND recipient_type = 'user'
			AND recipient_id = :user_id`
	}

	var results []struct {
		AlertID string `db:"alert_id"`
	}

	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &results); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	ids := make([]uuid.UUID, len(results))
	for i, r := range results {
		id, err := uuid.Parse(r.AlertID)
		if err != nil {
			return nil, fmt.Errorf("parse uuid: %w", err)
		}
		ids[i] = id
	}

	return ids, nil
}

// QueryActiveByUserID returns the IDs of all active alerts for a user.
func (s *Store) QueryActiveByUserID(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	roleIDStrings := make([]string, len(roleIDs))
	for i, id := range roleIDs {
		roleIDStrings[i] = id.String()
	}

	data := map[string]any{
		"user_id":  userID.String(),
		"role_ids": roleIDStrings,
		"status":   alertbus.StatusActive,
	}

	var q string
	if len(roleIDs) > 0 {
		q = `
		SELECT DISTINCT
			a.id
		FROM
			workflow.alerts a
		INNER JOIN workflow.alert_recipients ar ON a.id = ar.alert_id
		WHERE
			a.status = :status
			AND (
				(ar.recipient_type = 'user' AND ar.recipient_id = :user_id)
				OR (ar.recipient_type = 'role' AND ar.recipient_id IN (:role_ids))
			)`
	} else {
		q = `
		SELECT DISTINCT
			a.id
		FROM
			workflow.alerts a
		INNER JOIN workflow.alert_recipients ar ON a.id = ar.alert_id
		WHERE
			a.status = :status
			AND ar.recipient_type = 'user'
			AND ar.recipient_id = :user_id`
	}

	var results []struct {
		ID string `db:"id"`
	}

	if len(roleIDs) > 0 {
		if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &results); err != nil {
			return nil, fmt.Errorf("namedqueryslice: %w", err)
		}
	} else {
		if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &results); err != nil {
			return nil, fmt.Errorf("namedqueryslice: %w", err)
		}
	}

	ids := make([]uuid.UUID, len(results))
	for i, r := range results {
		id, err := uuid.Parse(r.ID)
		if err != nil {
			return nil, fmt.Errorf("parse uuid: %w", err)
		}
		ids[i] = id
	}

	return ids, nil
}

// AcknowledgeMultiple creates acknowledgment records and updates status for multiple alerts.
func (s *Store) AcknowledgeMultiple(ctx context.Context, alertIDs []uuid.UUID, userID uuid.UUID, notes string, now time.Time) (int, error) {
	if len(alertIDs) == 0 {
		return 0, nil
	}

	// Insert acknowledgment records (skip duplicates)
	const ackQ = `
	INSERT INTO workflow.alert_acknowledgments (
		id, alert_id, acknowledged_by, acknowledged_date, notes
	) VALUES (
		:id, :alert_id, :acknowledged_by, :acknowledged_date, :notes
	) ON CONFLICT (alert_id, acknowledged_by) DO NOTHING`

	for _, alertID := range alertIDs {
		ack := struct {
			ID               string    `db:"id"`
			AlertID          string    `db:"alert_id"`
			AcknowledgedBy   string    `db:"acknowledged_by"`
			AcknowledgedDate time.Time `db:"acknowledged_date"`
			Notes            string    `db:"notes"`
		}{
			ID:               uuid.New().String(),
			AlertID:          alertID.String(),
			AcknowledgedBy:   userID.String(),
			AcknowledgedDate: now,
			Notes:            notes,
		}

		if err := sqldb.NamedExecContext(ctx, s.log, s.db, ackQ, ack); err != nil {
			return 0, fmt.Errorf("namedexeccontext: %w", err)
		}
	}

	// Update status to acknowledged for all alerts
	alertIDStrings := make([]string, len(alertIDs))
	for i, id := range alertIDs {
		alertIDStrings[i] = id.String()
	}

	data := map[string]any{
		"alert_ids":    alertIDStrings,
		"status":       alertbus.StatusAcknowledged,
		"updated_date": now,
	}

	const updateQ = `
	UPDATE
		workflow.alerts
	SET
		status = :status,
		updated_date = :updated_date
	WHERE
		id IN (:alert_ids)`

	if err := namedExecContextUsingIn(ctx, s.log, s.db, updateQ, data); err != nil {
		return 0, fmt.Errorf("namedexeccontext: %w", err)
	}

	return len(alertIDs), nil
}

// DismissMultiple updates status to dismissed for multiple alerts.
func (s *Store) DismissMultiple(ctx context.Context, alertIDs []uuid.UUID, now time.Time) (int, error) {
	if len(alertIDs) == 0 {
		return 0, nil
	}

	alertIDStrings := make([]string, len(alertIDs))
	for i, id := range alertIDs {
		alertIDStrings[i] = id.String()
	}

	data := map[string]any{
		"alert_ids":     alertIDStrings,
		"status":        alertbus.StatusDismissed,
		"active_status": alertbus.StatusActive,
		"updated_date":  now,
	}

	const q = `
	UPDATE
		workflow.alerts
	SET
		status = :status,
		updated_date = :updated_date
	WHERE
		id IN (:alert_ids)
		AND status = :active_status`

	if err := namedExecContextUsingIn(ctx, s.log, s.db, q, data); err != nil {
		return 0, fmt.Errorf("namedexeccontext: %w", err)
	}

	return len(alertIDs), nil
}
