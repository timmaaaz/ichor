package formapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/formfieldapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
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
	ID                string `json:"id"`
	Name              string `json:"name"`
	IsReferenceData   bool   `json:"isReferenceData"`
	AllowInlineCreate bool   `json:"allowInlineCreate"`
}

func (app Form) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppForm converts a business form to an app form.
func ToAppForm(bus formbus.Form) Form {
	return Form{
		ID:                bus.ID.String(),
		Name:              bus.Name,
		IsReferenceData:   bus.IsReferenceData,
		AllowInlineCreate: bus.AllowInlineCreate,
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

// Forms is a collection wrapper that implements the Encoder interface.
type Forms []Form

func (app Forms) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// NewForm represents data needed to create a form.
type NewForm struct {
	Name              string `json:"name" validate:"required,min=1,max=255"`
	IsReferenceData   bool   `json:"isReferenceData"`
	AllowInlineCreate bool   `json:"allowInlineCreate"`
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
		Name:              app.Name,
		IsReferenceData:   app.IsReferenceData,
		AllowInlineCreate: app.AllowInlineCreate,
	}
}

// UpdateForm represents data needed to update a form.
type UpdateForm struct {
	Name              *string `json:"name" validate:"omitempty,min=1,max=255"`
	IsReferenceData   *bool   `json:"isReferenceData"`
	AllowInlineCreate *bool   `json:"allowInlineCreate"`
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
		Name:              app.Name,
		IsReferenceData:   app.IsReferenceData,
		AllowInlineCreate: app.AllowInlineCreate,
	}
}

// FormFull represents a form with all its associated fields.
type FormFull struct {
	ID                string                   `json:"id"`
	Name              string                   `json:"name"`
	IsReferenceData   bool                     `json:"isReferenceData"`
	AllowInlineCreate bool                     `json:"allowInlineCreate"`
	Fields            []formfieldapp.FormField `json:"fields"`
}

func (app FormFull) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppFormFull converts a business form and its fields to an app FormFull.
func ToAppFormFull(form formbus.Form, fields []formfieldapp.FormField) FormFull {
	return FormFull{
		ID:                form.ID.String(),
		Name:              form.Name,
		IsReferenceData:   form.IsReferenceData,
		AllowInlineCreate: form.AllowInlineCreate,
		Fields:            fields,
	}
}

// ExportPackage represents a JSON export package for forms.
type ExportPackage struct {
	Version    string        `json:"version"`
	Type       string        `json:"type"`
	ExportedAt string        `json:"exportedAt"`
	Count      int           `json:"count"`
	Data       []FormPackage `json:"data"`
}

func (app ExportPackage) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// FormPackage represents a single form with its fields for export/import.
type FormPackage struct {
	Form   Form                     `json:"form"`
	Fields []formfieldapp.FormField `json:"fields"`
}

// ImportPackage represents a JSON import package for forms.
type ImportPackage struct {
	Mode string        `json:"mode"` // "merge", "skip", "replace"
	Data []FormPackage `json:"data"`
}

func (app *ImportPackage) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app ImportPackage) Validate() error {
	if app.Mode != "merge" && app.Mode != "skip" && app.Mode != "replace" {
		return errs.Newf(errs.InvalidArgument, "mode must be 'merge', 'skip', or 'replace'")
	}

	if len(app.Data) == 0 {
		return errs.Newf(errs.InvalidArgument, "data cannot be empty")
	}

	// Validate each form package
	for i, pkg := range app.Data {
		if pkg.Form.Name == "" {
			return errs.Newf(errs.InvalidArgument, "form %d: name is required", i)
		}
	}

	return nil
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	ImportedCount int      `json:"importedCount"`
	SkippedCount  int      `json:"skippedCount"`
	UpdatedCount  int      `json:"updatedCount"`
	Errors        []string `json:"errors,omitempty"`
}

func (app ImportResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToBusFormWithFields converts an app FormPackage to a business FormWithFields.
func ToBusFormWithFields(app FormPackage) (formbus.FormWithFields, error) {
	formID, err := uuid.Parse(app.Form.ID)
	if err != nil {
		// Generate new ID if parsing fails (import scenario)
		formID = uuid.New()
	}

	form := formbus.Form{
		ID:                formID,
		Name:              app.Form.Name,
		IsReferenceData:   app.Form.IsReferenceData,
		AllowInlineCreate: app.Form.AllowInlineCreate,
	}

	fields := make([]formfieldbus.FormField, len(app.Fields))
	for i, appField := range app.Fields {
		fieldID, err := uuid.Parse(appField.ID)
		if err != nil {
			// Generate new ID if parsing fails (import scenario)
			fieldID = uuid.New()
		}
		fFormID, err := uuid.Parse(appField.FormID)
		if err != nil {
			fFormID = formID // Use the form's ID if parsing fails
		}
		entityID, err := uuid.Parse(appField.EntityID)
		if err != nil {
			return formbus.FormWithFields{}, errs.Newf(errs.InvalidArgument, "field %d: invalid entity_id: %s", i, err)
		}

		fields[i] = formfieldbus.FormField{
			ID:           fieldID,
			FormID:       fFormID,
			EntityID:     entityID,
			EntitySchema: appField.EntitySchema,
			EntityTable:  appField.EntityTable,
			Name:         appField.Name,
			Label:        appField.Label,
			FieldType:    appField.FieldType,
			FieldOrder:   appField.FieldOrder,
			Required:     appField.Required,
			Config:       appField.Config,
		}
	}

	return formbus.FormWithFields{
		Form:   form,
		Fields: fields,
	}, nil
}
