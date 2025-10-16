package formfielddb

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

type formField struct {
	ID         uuid.UUID       `db:"id"`
	FormID     uuid.UUID       `db:"form_id"`
	Name       string          `db:"name"`
	Label      string          `db:"label"`
	FieldType  string          `db:"field_type"`
	FieldOrder int             `db:"field_order"`
	Required   bool            `db:"required"`
	Config     json.RawMessage `db:"config"`
}

func toDBFormField(bus formfieldbus.FormField) formField {
	return formField{
		ID:         bus.ID,
		FormID:     bus.FormID,
		Name:       bus.Name,
		Label:      bus.Label,
		FieldType:  bus.FieldType,
		FieldOrder: bus.FieldOrder,
		Required:   bus.Required,
		Config:     bus.Config,
	}
}

func toBusFormField(db formField) formfieldbus.FormField {
	return formfieldbus.FormField{
		ID:         db.ID,
		FormID:     db.FormID,
		Name:       db.Name,
		Label:      db.Label,
		FieldType:  db.FieldType,
		FieldOrder: db.FieldOrder,
		Required:   db.Required,
		Config:     db.Config,
	}
}

func toBusFormFields(dbs []formField) []formfieldbus.FormField {
	bus := make([]formfieldbus.FormField, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusFormField(db)
	}
	return bus
}