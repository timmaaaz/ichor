package pageconfigapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
)

// QueryParams represents query parameters for filtering page configs
type QueryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	ID        string
	Name      string
	UserID    string
	IsDefault string
}

// PageConfig represents the application layer model for page configuration
type PageConfig struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UserID    string `json:"userId,omitempty"`
	IsDefault bool   `json:"isDefault"`
}

// Encode implements the encoder interface for PageConfig
func (app PageConfig) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// PageConfigs is a collection wrapper that implements the Encoder interface.
type PageConfigs []PageConfig

// Encode implements the encoder interface for PageConfigs
func (app PageConfigs) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// NewPageConfig contains data required to create a new page configuration
type NewPageConfig struct {
	Name      string `json:"name" validate:"required"`
	UserID    string `json:"userId" validate:"omitempty,uuid"`
	IsDefault bool   `json:"isDefault"`
}

// Decode implements the decoder interface for NewPageConfig
func (app *NewPageConfig) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate performs business rule validation on NewPageConfig
func (app NewPageConfig) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// UpdatePageConfig contains data for updating an existing page configuration
type UpdatePageConfig struct {
	Name      *string `json:"name"`
	UserID    *string `json:"userId" validate:"omitempty,uuid"`
	IsDefault *bool   `json:"isDefault"`
}

// Decode implements the decoder interface for UpdatePageConfig
func (app *UpdatePageConfig) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate performs business rule validation on UpdatePageConfig
func (app UpdatePageConfig) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// =============================================================================
// Conversion Functions
// =============================================================================

// ToAppPageConfig converts a business layer PageConfig to app layer
func ToAppPageConfig(bus pageconfigbus.PageConfig) PageConfig {
	app := PageConfig{
		ID:        bus.ID.String(),
		Name:      bus.Name,
		IsDefault: bus.IsDefault,
	}

	if bus.UserID != uuid.Nil {
		app.UserID = bus.UserID.String()
	}

	return app
}

// ToAppPageConfigs converts a slice of business layer PageConfig to app layer
func ToAppPageConfigs(bus []pageconfigbus.PageConfig) []PageConfig {
	app := make([]PageConfig, len(bus))
	for i, b := range bus {
		app[i] = ToAppPageConfig(b)
	}
	return app
}

// toBusNewPageConfig converts app layer NewPageConfig to business layer
func toBusNewPageConfig(app NewPageConfig) (pageconfigbus.NewPageConfig, error) {
	bus := pageconfigbus.NewPageConfig{
		Name:      app.Name,
		IsDefault: app.IsDefault,
	}

	if app.UserID != "" {
		userID, err := uuid.Parse(app.UserID)
		if err != nil {
			return pageconfigbus.NewPageConfig{}, fmt.Errorf("parse user id: %w", err)
		}
		bus.UserID = userID
	}

	return bus, nil
}

// toBusUpdatePageConfig converts app layer UpdatePageConfig to business layer
func toBusUpdatePageConfig(app UpdatePageConfig) (pageconfigbus.UpdatePageConfig, error) {
	bus := pageconfigbus.UpdatePageConfig{
		Name:      app.Name,
		IsDefault: app.IsDefault,
	}

	if app.UserID != nil && *app.UserID != "" {
		userID, err := uuid.Parse(*app.UserID)
		if err != nil {
			return pageconfigbus.UpdatePageConfig{}, fmt.Errorf("parse user id: %w", err)
		}
		bus.UserID = &userID
	}

	return bus, nil
}
