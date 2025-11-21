package formfielddb

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

type formField struct {
	ID           uuid.UUID       `db:"id"`
	FormID       uuid.UUID       `db:"form_id"`
	EntityID     uuid.UUID       `db:"entity_id"`
	EntitySchema string          `db:"entity_schema"`
	EntityTable  string          `db:"entity_table"`
	Name         string          `db:"name"`
	Label        string          `db:"label"`
	FieldType    string          `db:"field_type"`
	FieldOrder   int             `db:"field_order"`
	Required     bool            `db:"required"`
	Config       json.RawMessage `db:"config"`
}

func toDBFormField(bus formfieldbus.FormField) formField {
	config := bus.Config
	if len(config) == 0 {
		config = json.RawMessage([]byte("{}"))
	}

	return formField{
		ID:           bus.ID,
		FormID:       bus.FormID,
		EntityID:     bus.EntityID,
		EntitySchema: bus.EntitySchema,
		EntityTable:  bus.EntityTable,
		Name:         bus.Name,
		Label:        bus.Label,
		FieldType:    bus.FieldType,
		FieldOrder:   bus.FieldOrder,
		Required:     bus.Required,
		Config:       config,
	}
}

func toBusFormField(db formField) formfieldbus.FormField {
	return formfieldbus.FormField{
		ID:           db.ID,
		FormID:       db.FormID,
		EntityID:     db.EntityID,
		EntitySchema: db.EntitySchema,
		EntityTable:  db.EntityTable,
		Name:         db.Name,
		Label:        db.Label,
		FieldType:    db.FieldType,
		FieldOrder:   db.FieldOrder,
		Required:     db.Required,
		Config:       db.Config,
	}
}

func toBusFormFields(dbs []formField) []formfieldbus.FormField {
	bus := make([]formfieldbus.FormField, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusFormField(db)
	}
	return bus
}
