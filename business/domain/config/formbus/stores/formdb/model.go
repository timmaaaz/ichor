package formdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
)

type form struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

func toDBForm(bus formbus.Form) form {
	return form{
		ID:   bus.ID,
		Name: bus.Name,
	}
}

func toBusForm(db form) formbus.Form {
	return formbus.Form{
		ID:   db.ID,
		Name: db.Name,
	}
}

func toBusForms(dbs []form) []formbus.Form {
	bus := make([]formbus.Form, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusForm(db)
	}
	return bus
}