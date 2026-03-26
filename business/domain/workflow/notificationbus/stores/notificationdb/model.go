package notificationdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
)

type dbNotification struct {
	ID               uuid.UUID      `db:"id"`
	UserID           uuid.UUID      `db:"user_id"`
	Title            string         `db:"title"`
	Message          sql.NullString `db:"message"`
	Priority         string         `db:"priority"`
	IsRead           bool           `db:"is_read"`
	ReadDate         sql.NullTime   `db:"read_date"`
	SourceEntityName sql.NullString `db:"source_entity_name"`
	SourceEntityID   sql.NullString `db:"source_entity_id"`
	ActionURL        sql.NullString `db:"action_url"`
	CreatedDate      time.Time      `db:"created_date"`
}

func toDBNotification(n notificationbus.Notification) dbNotification {
	db := dbNotification{
		ID:          n.ID,
		UserID:      n.UserID,
		Title:       n.Title,
		Priority:    n.Priority,
		IsRead:      n.IsRead,
		CreatedDate: n.CreatedDate,
	}

	if n.Message != "" {
		db.Message = sql.NullString{String: n.Message, Valid: true}
	}
	if n.ReadDate != nil {
		db.ReadDate = sql.NullTime{Time: *n.ReadDate, Valid: true}
	}
	if n.SourceEntityName != "" {
		db.SourceEntityName = sql.NullString{String: n.SourceEntityName, Valid: true}
	}
	if n.SourceEntityID != uuid.Nil {
		db.SourceEntityID = sql.NullString{String: n.SourceEntityID.String(), Valid: true}
	}
	if n.ActionURL != "" {
		db.ActionURL = sql.NullString{String: n.ActionURL, Valid: true}
	}

	return db
}

func toBusNotification(db dbNotification) notificationbus.Notification {
	n := notificationbus.Notification{
		ID:          db.ID,
		UserID:      db.UserID,
		Title:       db.Title,
		Priority:    db.Priority,
		IsRead:      db.IsRead,
		CreatedDate: db.CreatedDate,
	}

	if db.Message.Valid {
		n.Message = db.Message.String
	}
	if db.ReadDate.Valid {
		n.ReadDate = &db.ReadDate.Time
	}
	if db.SourceEntityName.Valid {
		n.SourceEntityName = db.SourceEntityName.String
	}
	if db.SourceEntityID.Valid {
		n.SourceEntityID, _ = uuid.Parse(db.SourceEntityID.String)
	}
	if db.ActionURL.Valid {
		n.ActionURL = db.ActionURL.String
	}

	return n
}

func toBusNotifications(dbs []dbNotification) []notificationbus.Notification {
	notifications := make([]notificationbus.Notification, len(dbs))
	for i, db := range dbs {
		notifications[i] = toBusNotification(db)
	}
	return notifications
}
