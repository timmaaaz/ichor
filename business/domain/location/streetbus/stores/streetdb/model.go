package streetdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
)

type street struct {
	ID         uuid.UUID `db:"id"`
	CityID     uuid.UUID `db:"city_id"`
	Line1      string    `db:"line_1"`
	Line2      string    `db:"line_2"`
	PostalCode string    `db:"postal_code"`
}

func toDBStreet(bus streetbus.Street) street {
	return street{
		ID:         bus.ID,
		CityID:     bus.CityID,
		Line1:      bus.Line1,
		Line2:      bus.Line2,
		PostalCode: bus.PostalCode,
	}
}

func toBusStreet(dbStreet street) streetbus.Street {
	return streetbus.Street{
		ID:         dbStreet.ID,
		CityID:     dbStreet.CityID,
		Line1:      dbStreet.Line1,
		Line2:      dbStreet.Line2,
		PostalCode: dbStreet.PostalCode,
	}
}

func toBusStreets(dbStreets []street) []streetbus.Street {
	streets := make([]streetbus.Street, len(dbStreets))
	for i, str := range dbStreets {
		streets[i] = toBusStreet(str)
	}
	return streets
}
