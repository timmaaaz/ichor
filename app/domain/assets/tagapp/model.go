package tagapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assets/tagbus"
)

type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Name        string
	Description string
}

type Tag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Encode implements the encoder interface.
func (app Tag) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppTag(bus tagbus.Tag) Tag {
	return Tag{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
	}
}

func ToAppTags(bus []tagbus.Tag) []Tag {
	app := make([]Tag, len(bus))
	for i, v := range bus {
		app[i] = ToAppTag(v)
	}
	return app
}

// =============================================================================

type NewTag struct {
	Name        string `json:"name" validate:"required,min=3,max=31"`
	Description string `json:"description" validate:"required,min=3,max=127"`
}

// Decode implements the decoder interface.
func (app *NewTag) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewTag) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusNewTag(app NewTag) tagbus.NewTag {
	return tagbus.NewTag{
		Name:        app.Name,
		Description: app.Description,
	}
}

// =============================================================================

type UpdateTag struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=31"`
	Description *string `json:"description" validate:"omitempty,min=3,max=127"`
}

// Decode implements the decoder interface.
func (app *UpdateTag) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateTag) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func ToBusUpdateTag(app UpdateTag) tagbus.UpdateTag {
	return tagbus.UpdateTag{
		Name:        app.Name,
		Description: app.Description,
	}
}
