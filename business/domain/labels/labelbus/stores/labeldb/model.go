package labeldb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
)

type labelCatalog struct {
	ID          uuid.UUID      `db:"id"`
	Code        string         `db:"code"`
	Type        string         `db:"type"`
	EntityRef   sql.NullString `db:"entity_ref"`
	PayloadJSON string         `db:"payload_json"`
	CreatedDate time.Time      `db:"created_date"`
}

func toDBLabelCatalog(bus labelbus.LabelCatalog) labelCatalog {
	var entityRef sql.NullString
	if bus.EntityRef != "" {
		entityRef = sql.NullString{String: bus.EntityRef, Valid: true}
	}

	return labelCatalog{
		ID:          bus.ID,
		Code:        bus.Code,
		Type:        bus.Type,
		EntityRef:   entityRef,
		PayloadJSON: bus.PayloadJSON,
		CreatedDate: bus.CreatedDate.UTC(),
	}
}

func toBusLabelCatalog(db labelCatalog) labelbus.LabelCatalog {
	var entityRef string
	if db.EntityRef.Valid {
		entityRef = db.EntityRef.String
	}

	return labelbus.LabelCatalog{
		ID:          db.ID,
		Code:        db.Code,
		Type:        db.Type,
		EntityRef:   entityRef,
		PayloadJSON: db.PayloadJSON,
		CreatedDate: db.CreatedDate.In(time.Local),
	}
}

func toBusLabelCatalogs(dbs []labelCatalog) []labelbus.LabelCatalog {
	out := make([]labelbus.LabelCatalog, len(dbs))
	for i, d := range dbs {
		out[i] = toBusLabelCatalog(d)
	}
	return out
}
