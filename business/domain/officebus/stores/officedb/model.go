package officedb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/officebus"
)

type office struct {
	ID       uuid.UUID `db:"office_id"`
	Name     string    `db:"name"`
	StreetID uuid.UUID `db:"street_id"`
}

func toDBOffice(bus officebus.Office) office {
	return office{
		ID:       bus.ID,
		Name:     bus.Name,
		StreetID: bus.StreetID,
	}
}

func toBusOffice(dbOffice office) officebus.Office {
	return officebus.Office{
		ID:       dbOffice.ID,
		Name:     dbOffice.Name,
		StreetID: dbOffice.StreetID,
	}
}

func toBusOffices(dbOffices []office) []officebus.Office {
	offices := make([]officebus.Office, len(dbOffices))
	for i, at := range dbOffices {
		offices[i] = toBusOffice(at)
	}
	return offices
}
