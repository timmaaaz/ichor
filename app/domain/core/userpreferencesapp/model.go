package userpreferencesapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/userpreferencesbus"
)

// UserPreference represents a user preference for API responses.
type UserPreference struct {
	UserID      string          `json:"user_id"`
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	UpdatedDate string          `json:"updated_date"`
}

// Encode implements the web.Encoder interface.
func (app UserPreference) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppUserPreference converts a bus UserPreference to an app UserPreference.
func ToAppUserPreference(bus userpreferencesbus.UserPreference) UserPreference {
	return UserPreference{
		UserID:      bus.UserID.String(),
		Key:         bus.Key,
		Value:       bus.Value,
		UpdatedDate: bus.UpdatedDate.Format("2006-01-02T15:04:05Z"),
	}
}

// ToAppUserPreferences converts a slice of bus UserPreferences to app UserPreferences.
func ToAppUserPreferences(bus []userpreferencesbus.UserPreference) []UserPreference {
	app := make([]UserPreference, len(bus))
	for i, v := range bus {
		app[i] = ToAppUserPreference(v)
	}
	return app
}

// UserPreferences is a collection wrapper that implements the Encoder interface.
type UserPreferences struct {
	Items []UserPreference `json:"items"`
}

// Encode implements the web.Encoder interface.
func (app UserPreferences) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// NewUserPreference contains the value for setting a preference.
// The user_id and key come from path parameters, not the request body.
type NewUserPreference struct {
	Value json.RawMessage `json:"value" validate:"required"`
}

// Decode implements the web.Decoder interface.
func (app *NewUserPreference) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewUserPreference) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}
