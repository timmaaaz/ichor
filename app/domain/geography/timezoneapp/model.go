package timezoneapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
)

// QueryParams represents the query parameters that can be used.
type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Name        string
	DisplayName string
	IsActive    string
}

// Timezone represents a timezone in the app layer.
type Timezone struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	UTCOffset   string `json:"utc_offset"`
	IsActive    bool   `json:"is_active"`
}

// Encode implements the encoder interface.
func (app Timezone) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppTimezone converts a business timezone to an app timezone.
func ToAppTimezone(bus timezonebus.Timezone) Timezone {
	return Timezone{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		DisplayName: bus.DisplayName,
		UTCOffset:   bus.UTCOffset,
		IsActive:    bus.IsActive,
	}
}

// ToAppTimezones converts a slice of business timezones to app timezones.
func ToAppTimezones(bus []timezonebus.Timezone) []Timezone {
	app := make([]Timezone, len(bus))
	for i, v := range bus {
		app[i] = ToAppTimezone(v)
	}
	return app
}

// =============================================================================

// NewTimezone defines the data needed to add a timezone.
type NewTimezone struct {
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
	UTCOffset   string `json:"utc_offset" validate:"required"`
	IsActive    bool   `json:"is_active"`
}

// Decode implements the decoder interface.
func (app *NewTimezone) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewTimezone) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewTimezone(app NewTimezone) timezonebus.NewTimezone {
	return timezonebus.NewTimezone{
		Name:        app.Name,
		DisplayName: app.DisplayName,
		UTCOffset:   app.UTCOffset,
		IsActive:    app.IsActive,
	}
}

// =============================================================================

// UpdateTimezone defines the data needed to update a timezone.
type UpdateTimezone struct {
	Name        *string `json:"name"`
	DisplayName *string `json:"display_name"`
	UTCOffset   *string `json:"utc_offset"`
	IsActive    *bool   `json:"is_active"`
}

// Decode implements the decoder interface.
func (app *UpdateTimezone) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateTimezone) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateTimezone(app UpdateTimezone) (timezonebus.UpdateTimezone, error) {
	return timezonebus.UpdateTimezone{
		Name:        app.Name,
		DisplayName: app.DisplayName,
		UTCOffset:   app.UTCOffset,
		IsActive:    app.IsActive,
	}, nil
}

// =============================================================================

// QueryByIDsRequest represents a batch query by IDs.
type QueryByIDsRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// Decode implements the decoder interface.
func (app *QueryByIDsRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app QueryByIDsRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	for _, id := range app.IDs {
		if _, err := uuid.Parse(id); err != nil {
			return errs.Newf(errs.InvalidArgument, "validate: invalid uuid %q", id)
		}
	}
	return nil
}

// Timezones is a collection wrapper that implements the Encoder interface.
type Timezones []Timezone

// Encode implements the encoder interface.
func (app Timezones) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}
