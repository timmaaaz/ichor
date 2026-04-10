package userpreferencesdb

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
)

type userPreference struct {
	UserID      uuid.UUID       `db:"user_id"`
	Key         string          `db:"key"`
	Value       json.RawMessage `db:"value"`
	UpdatedDate time.Time       `db:"updated_date"`
}

func toBusUserPreference(db userPreference) userpreferencesbus.UserPreference {
	return userpreferencesbus.UserPreference{
		UserID:      db.UserID,
		Key:         db.Key,
		Value:       db.Value,
		UpdatedDate: db.UpdatedDate,
	}
}

func toBusUserPreferences(dbs []userPreference) []userpreferencesbus.UserPreference {
	bus := make([]userpreferencesbus.UserPreference, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusUserPreference(db)
	}
	return bus
}

func toDBUserPreference(bus userpreferencesbus.UserPreference) userPreference {
	return userPreference{
		UserID:      bus.UserID,
		Key:         bus.Key,
		Value:       bus.Value,
		UpdatedDate: bus.UpdatedDate,
	}
}
