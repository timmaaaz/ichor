package userpreferencesbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserPreference represents a single user preference key-value pair.
type UserPreference struct {
	UserID      uuid.UUID       `json:"user_id"`
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	UpdatedDate time.Time       `json:"updated_date"`
}

// NewUserPreference contains fields required to set a preference.
type NewUserPreference struct {
	UserID uuid.UUID       `json:"user_id"`
	Key    string          `json:"key"`
	Value  json.RawMessage `json:"value"`
}
