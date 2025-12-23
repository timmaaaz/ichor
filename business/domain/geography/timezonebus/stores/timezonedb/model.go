package timezonedb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
)

type timezone struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	DisplayName string    `db:"display_name"`
	UTCOffset   string    `db:"utc_offset"`
	IsActive    bool      `db:"is_active"`
}

func toDBTimezone(bus timezonebus.Timezone) timezone {
	return timezone{
		ID:          bus.ID,
		Name:        bus.Name,
		DisplayName: bus.DisplayName,
		UTCOffset:   bus.UTCOffset,
		IsActive:    bus.IsActive,
	}
}

func toBusTimezone(dbTz timezone) timezonebus.Timezone {
	return timezonebus.Timezone{
		ID:          dbTz.ID,
		Name:        dbTz.Name,
		DisplayName: dbTz.DisplayName,
		UTCOffset:   dbTz.UTCOffset,
		IsActive:    dbTz.IsActive,
	}
}

func toBusTimezones(dbTzs []timezone) []timezonebus.Timezone {
	tzs := make([]timezonebus.Timezone, len(dbTzs))
	for i, tz := range dbTzs {
		tzs[i] = toBusTimezone(tz)
	}
	return tzs
}
