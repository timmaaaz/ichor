package settingsdb

import (
	"encoding/json"
	"time"

	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
)

type setting struct {
	Key         string          `db:"key"`
	Value       json.RawMessage `db:"value"`
	Description string          `db:"description"`
	CreatedDate time.Time       `db:"created_date"`
	UpdatedDate time.Time       `db:"updated_date"`
}

func toBusSetting(s setting) settingsbus.Setting {
	return settingsbus.Setting{
		Key:         s.Key,
		Value:       s.Value,
		Description: s.Description,
		CreatedDate: s.CreatedDate,
		UpdatedDate: s.UpdatedDate,
	}
}

func toBusSettings(ss []setting) []settingsbus.Setting {
	bus := make([]settingsbus.Setting, len(ss))
	for i, s := range ss {
		bus[i] = toBusSetting(s)
	}
	return bus
}

func toDBSetting(s settingsbus.Setting) setting {
	return setting{
		Key:         s.Key,
		Value:       s.Value,
		Description: s.Description,
		CreatedDate: s.CreatedDate,
		UpdatedDate: s.UpdatedDate,
	}
}
