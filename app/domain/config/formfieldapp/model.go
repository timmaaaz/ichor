package formfieldapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

// QueryParams represents the query parameters for form fields.
type QueryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	ID        string
	FormID    string
	Name      string
	FieldType string
	Required  string
}

// FormField represents a form field for the app layer.
type FormField struct {
	ID         string          `json:"id"`
	FormID     string          `json:"form_id"`
	Name       string          `json:"name"`
	Label      string          `json:"label"`
	FieldType  string          `json:"field_type"`
	FieldOrder int             `json:"field_order"`
	Required   bool            `json:"required"`
	Config     json.RawMessage `json:"config"`
}

type FormFields struct {
	Fields []FormField
}

func (app FormFields) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func (app FormField) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppFormField converts a business form field to an app form field.
func ToAppFormField(bus formfieldbus.FormField) FormField {
	return FormField{
		ID:         bus.ID.String(),
		FormID:     bus.FormID.String(),
		Name:       bus.Name,
		Label:      bus.Label,
		FieldType:  bus.FieldType,
		FieldOrder: bus.FieldOrder,
		Required:   bus.Required,
		Config:     bus.Config,
	}
}

// ToAppFormFields converts business form fields to app form fields.
func ToAppFormFieldSlice(bus []formfieldbus.FormField) []FormField {
	app := make([]FormField, len(bus))
	for i, v := range bus {
		app[i] = ToAppFormField(v)
	}
	return app
}

// ToAppFormFields converts business form fields to app form fields wrapped in FormFields.
func ToAppFormFields(app []FormField) FormFields {
	return FormFields{Fields: app}
}

// NewFormField represents data needed to create a form field.
type NewFormField struct {
	FormID     string          `json:"form_id" validate:"required,uuid"`
	Name       string          `json:"name" validate:"required,min=1,max=255"`
	Label      string          `json:"label" validate:"required,min=1,max=255"`
	FieldType  string          `json:"field_type" validate:"required,min=1,max=50"`
	FieldOrder int             `json:"field_order" validate:"required,min=0"`
	Required   bool            `json:"required"`
	Config     json.RawMessage `json:"config" validate:"required"`
}

func (app *NewFormField) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewFormField) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	// Validate that Config is valid JSON
	var configTest map[string]interface{}
	if err := json.Unmarshal(app.Config, &configTest); err != nil {
		return errs.Newf(errs.InvalidArgument, "invalid config JSON: %s", err)
	}

	return nil
}

func toBusNewFormField(app NewFormField) (formfieldbus.NewFormField, error) {
	formID, err := uuid.Parse(app.FormID)
	if err != nil {
		return formfieldbus.NewFormField{}, errs.NewFieldsError("formID", err)
	}

	// Validate config JSON
	var configTest map[string]interface{}
	if err := json.Unmarshal(app.Config, &configTest); err != nil {
		return formfieldbus.NewFormField{}, errs.NewFieldsError("config", err)
	}

	return formfieldbus.NewFormField{
		FormID:     formID,
		Name:       app.Name,
		Label:      app.Label,
		FieldType:  app.FieldType,
		FieldOrder: app.FieldOrder,
		Required:   app.Required,
		Config:     app.Config,
	}, nil
}

// UpdateFormField represents data needed to update a form field.
type UpdateFormField struct {
	FormID     *string          `json:"form_id" validate:"omitempty,uuid"`
	Name       *string          `json:"name" validate:"omitempty,min=1,max=255"`
	Label      *string          `json:"label" validate:"omitempty,min=1,max=255"`
	FieldType  *string          `json:"field_type" validate:"omitempty,min=1,max=50"`
	FieldOrder *int             `json:"field_order" validate:"omitempty,min=0"`
	Required   *bool            `json:"required"`
	Config     *json.RawMessage `json:"config" validate:"omitempty"`
}

func (app *UpdateFormField) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateFormField) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	// Validate Config if provided
	if app.Config != nil {
		var configTest map[string]interface{}
		if err := json.Unmarshal(*app.Config, &configTest); err != nil {
			return errs.Newf(errs.InvalidArgument, "invalid config JSON: %s", err)
		}
	}

	return nil
}

func toBusUpdateFormField(app UpdateFormField) (formfieldbus.UpdateFormField, error) {
	uff := formfieldbus.UpdateFormField{}

	if app.FormID != nil {
		formID, err := uuid.Parse(*app.FormID)
		if err != nil {
			return formfieldbus.UpdateFormField{}, errs.NewFieldsError("formID", err)
		}
		uff.FormID = &formID
	}

	if app.Name != nil {
		uff.Name = app.Name
	}

	if app.Label != nil {
		uff.Label = app.Label
	}

	if app.FieldType != nil {
		uff.FieldType = app.FieldType
	}

	if app.FieldOrder != nil {
		uff.FieldOrder = app.FieldOrder
	}

	if app.Required != nil {
		uff.Required = app.Required
	}

	if app.Config != nil {
		var configTest map[string]interface{}
		if err := json.Unmarshal(*app.Config, &configTest); err != nil {
			return formfieldbus.UpdateFormField{}, errs.NewFieldsError("config", err)
		}
		uff.Config = app.Config
	}

	return uff, nil
}
