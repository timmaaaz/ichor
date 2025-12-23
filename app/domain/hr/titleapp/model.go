package titleapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
)

type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Description string
	Name        string
}

type Title struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

func (app Title) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppTitle(bus titlebus.Title) Title {
	return Title{
		ID:          bus.ID.String(),
		Description: bus.Description,
		Name:        bus.Name,
	}
}

func ToAppTitles(bus []titlebus.Title) []Title {
	app := make([]Title, len(bus))
	for i, v := range bus {
		app[i] = ToAppTitle(v)
	}
	return app
}

// =============================================================================

type NewTitle struct {
	Description string `json:"description" validate:"omitempty"`
	Name        string `json:"name" validate:"required,min=3,max=100"`
}

func (app *NewTitle) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewTitle) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewTitle(app NewTitle) (titlebus.NewTitle, error) {
	return titlebus.NewTitle{
		Description: app.Description,
		Name:        app.Name,
	}, nil
}

// =============================================================================

type UpdateTitle struct {
	Description *string `json:"description" validate:"omitempty"`
	Name        *string `json:"name" validate:"omitempty,min=3,max=100"`
}

func (app *UpdateTitle) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateTitle) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateTitle(app UpdateTitle) (titlebus.UpdateTitle, error) {

	var name, description *string

	if app.Name != nil {
		name = app.Name
	}

	if app.Description != nil {
		description = app.Description
	}

	return titlebus.UpdateTitle{
		Description: description,
		Name:        name,
	}, nil
}
