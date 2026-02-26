package alertdb

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
)

// dbAlert represents the database structure for an alert.
type dbAlert struct {
	ID               uuid.UUID       `db:"id"`
	AlertType        string          `db:"alert_type"`
	Severity         string          `db:"severity"`
	Title            string          `db:"title"`
	Message          string          `db:"message"`
	Context          json.RawMessage `db:"context"`
	SourceEntityName sql.NullString  `db:"source_entity_name"`
	SourceEntityID   sql.NullString  `db:"source_entity_id"`
	SourceRuleID     sql.NullString  `db:"source_rule_id"`
	SourceRuleName   sql.NullString  `db:"source_rule_name"`
	Status           string          `db:"status"`
	ExpiresDate      sql.NullTime    `db:"expires_date"`
	CreatedDate      time.Time       `db:"created_date"`
	UpdatedDate      time.Time       `db:"updated_date"`
}

func toDBAlert(a alertbus.Alert) dbAlert {
	db := dbAlert{
		ID:          a.ID,
		AlertType:   a.AlertType,
		Severity:    a.Severity,
		Title:       a.Title,
		Message:     a.Message,
		Context:     a.Context,
		Status:      a.Status,
		CreatedDate: a.CreatedDate,
		UpdatedDate: a.UpdatedDate,
	}

	if a.SourceEntityName != "" {
		db.SourceEntityName = sql.NullString{String: a.SourceEntityName, Valid: true}
	}
	if a.SourceEntityID != uuid.Nil {
		db.SourceEntityID = sql.NullString{String: a.SourceEntityID.String(), Valid: true}
	}
	if a.SourceRuleID != uuid.Nil {
		db.SourceRuleID = sql.NullString{String: a.SourceRuleID.String(), Valid: true}
	}
	if a.ExpiresDate != nil {
		db.ExpiresDate = sql.NullTime{Time: *a.ExpiresDate, Valid: true}
	}

	return db
}

func toBusAlert(db dbAlert) alertbus.Alert {
	a := alertbus.Alert{
		ID:          db.ID,
		AlertType:   db.AlertType,
		Severity:    db.Severity,
		Title:       db.Title,
		Message:     db.Message,
		Context:     db.Context,
		Status:      db.Status,
		CreatedDate: db.CreatedDate,
		UpdatedDate: db.UpdatedDate,
	}

	if db.SourceEntityName.Valid {
		a.SourceEntityName = db.SourceEntityName.String
	}
	if db.SourceEntityID.Valid {
		a.SourceEntityID, _ = uuid.Parse(db.SourceEntityID.String)
	}
	if db.SourceRuleID.Valid {
		a.SourceRuleID, _ = uuid.Parse(db.SourceRuleID.String)
	}
	if db.SourceRuleName.Valid {
		a.SourceRuleName = db.SourceRuleName.String
	}
	if db.ExpiresDate.Valid {
		a.ExpiresDate = &db.ExpiresDate.Time
	}

	return a
}

func toBusAlerts(dbs []dbAlert) []alertbus.Alert {
	alerts := make([]alertbus.Alert, len(dbs))
	for i, db := range dbs {
		alerts[i] = toBusAlert(db)
	}
	return alerts
}

// dbAlertRecipient represents the database structure for an alert recipient.
type dbAlertRecipient struct {
	ID            uuid.UUID `db:"id"`
	AlertID       uuid.UUID `db:"alert_id"`
	RecipientType string    `db:"recipient_type"`
	RecipientID   uuid.UUID `db:"recipient_id"`
	CreatedDate   time.Time `db:"created_date"`
}

func toDBAlertRecipient(ar alertbus.AlertRecipient) dbAlertRecipient {
	return dbAlertRecipient{
		ID:            ar.ID,
		AlertID:       ar.AlertID,
		RecipientType: ar.RecipientType,
		RecipientID:   ar.RecipientID,
		CreatedDate:   ar.CreatedDate,
	}
}

func toBusAlertRecipient(db dbAlertRecipient) alertbus.AlertRecipient {
	return alertbus.AlertRecipient{
		ID:            db.ID,
		AlertID:       db.AlertID,
		RecipientType: db.RecipientType,
		RecipientID:   db.RecipientID,
		CreatedDate:   db.CreatedDate,
	}
}

func toBusAlertRecipients(dbs []dbAlertRecipient) []alertbus.AlertRecipient {
	recipients := make([]alertbus.AlertRecipient, len(dbs))
	for i, db := range dbs {
		recipients[i] = toBusAlertRecipient(db)
	}
	return recipients
}

// dbAlertAcknowledgment represents the database structure for an alert acknowledgment.
type dbAlertAcknowledgment struct {
	ID               uuid.UUID      `db:"id"`
	AlertID          uuid.UUID      `db:"alert_id"`
	AcknowledgedBy   uuid.UUID      `db:"acknowledged_by"`
	AcknowledgerName sql.NullString `db:"acknowledger_name"`
	AcknowledgedDate time.Time      `db:"acknowledged_date"`
	Notes            sql.NullString `db:"notes"`
}

func toDBAlertAcknowledgment(ack alertbus.AlertAcknowledgment) dbAlertAcknowledgment {
	db := dbAlertAcknowledgment{
		ID:               ack.ID,
		AlertID:          ack.AlertID,
		AcknowledgedBy:   ack.AcknowledgedBy,
		AcknowledgedDate: ack.AcknowledgedDate,
	}
	if ack.Notes != "" {
		db.Notes = sql.NullString{String: ack.Notes, Valid: true}
	}
	return db
}

func toBusAlertAcknowledgment(db dbAlertAcknowledgment) alertbus.AlertAcknowledgment {
	ack := alertbus.AlertAcknowledgment{
		ID:               db.ID,
		AlertID:          db.AlertID,
		AcknowledgedBy:   db.AcknowledgedBy,
		AcknowledgedDate: db.AcknowledgedDate,
	}
	if db.AcknowledgerName.Valid {
		ack.AcknowledgerName = db.AcknowledgerName.String
	}
	if db.Notes.Valid {
		ack.Notes = db.Notes.String
	}
	return ack
}

func toBusAlertAcknowledgments(dbs []dbAlertAcknowledgment) []alertbus.AlertAcknowledgment {
	acks := make([]alertbus.AlertAcknowledgment, len(dbs))
	for i, db := range dbs {
		acks[i] = toBusAlertAcknowledgment(db)
	}
	return acks
}
