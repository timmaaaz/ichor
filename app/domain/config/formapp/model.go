package formapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
)

// QueryParams represents the query parameters for forms.
type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string
	ID      string
	Name    string
}

// Form represents a form for the app layer.
type Form struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (app Form) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppForm converts a business form to an app form.
func ToAppForm(bus formbus.Form) Form {
	return Form{
		ID:   bus.ID.String(),
		Name: bus.Name,
	}
}

// ToAppForms converts business forms to app forms.
func ToAppForms(bus []formbus.Form) []Form {
	app := make([]Form, len(bus))
	for i, v := range bus {
		app[i] = ToAppForm(v)
	}
	return app
}

// NewForm represents data needed to create a form.
type NewForm struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

func (app *NewForm) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewForm) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewForm(app NewForm) formbus.NewForm {
	return formbus.NewForm{
		Name: app.Name,
	}
}

// UpdateForm represents data needed to update a form.
type UpdateForm struct {
	Name *string `json:"name" validate:"omitempty,min=1,max=255"`
}

func (app *UpdateForm) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateForm) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateForm(app UpdateForm) formbus.UpdateForm {
	return formbus.UpdateForm{
		Name: app.Name,
	}
}
