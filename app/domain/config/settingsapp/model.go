package settingsapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	Key     string
	Prefix  string
}

type Setting struct {
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description"`
	CreatedDate string          `json:"created_date"`
	UpdatedDate string          `json:"updated_date"`
}

func (app Setting) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppSetting(bus settingsbus.Setting) Setting {
	return Setting{
		Key:         bus.Key,
		Value:       bus.Value,
		Description: bus.Description,
		CreatedDate: bus.CreatedDate.Format("2006-01-02T15:04:05Z"),
		UpdatedDate: bus.UpdatedDate.Format("2006-01-02T15:04:05Z"),
	}
}

func ToAppSettings(bus []settingsbus.Setting) []Setting {
	app := make([]Setting, len(bus))
	for i, v := range bus {
		app[i] = ToAppSetting(v)
	}
	return app
}

type NewSetting struct {
	Key         string          `json:"key" validate:"required"`
	Value       json.RawMessage `json:"value" validate:"required"`
	Description string          `json:"description"`
}

func (app *NewSetting) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app NewSetting) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewSetting(app NewSetting) settingsbus.NewSetting {
	return settingsbus.NewSetting{
		Key:         app.Key,
		Value:       app.Value,
		Description: app.Description,
	}
}

type UpdateSetting struct {
	Value       json.RawMessage `json:"value"`
	Description *string         `json:"description"`
}

func (app *UpdateSetting) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app UpdateSetting) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateSetting(app UpdateSetting) settingsbus.UpdateSetting {
	return settingsbus.UpdateSetting{
		Value:       app.Value,
		Description: app.Description,
	}
}
