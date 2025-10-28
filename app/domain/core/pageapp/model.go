package pageapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
)

type QueryParams struct {
	Page     string
	Rows     string
	OrderBy  string
	ID       string
	Path     string
	Name     string
	Module   string
	IsActive string
}

type Page struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Module    string `json:"module"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sortOrder"`
	IsActive  bool   `json:"isActive"`
}

func (app Page) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

type Pages struct {
	Items []Page `json:"items"`
}

func (app Pages) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppPage(bus pagebus.Page) Page {
	return Page{
		ID:        bus.ID.String(),
		Path:      bus.Path,
		Name:      bus.Name,
		Module:    bus.Module,
		Icon:      bus.Icon,
		SortOrder: bus.SortOrder,
		IsActive:  bus.IsActive,
	}
}

func ToAppPages(bus []pagebus.Page) []Page {
	app := make([]Page, len(bus))
	for i, v := range bus {
		app[i] = ToAppPage(v)
	}
	return app
}

// =============================================================================

type NewPage struct {
	Path      string `json:"path" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Module    string `json:"module" validate:"required"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sortOrder"`
	IsActive  bool   `json:"isActive"`
}

func (app *NewPage) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewPage) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewPage(app NewPage) (pagebus.NewPage, error) {
	return pagebus.NewPage{
		Path:      app.Path,
		Name:      app.Name,
		Module:    app.Module,
		Icon:      app.Icon,
		SortOrder: app.SortOrder,
		IsActive:  app.IsActive,
	}, nil
}

// =============================================================================

type UpdatePage struct {
	Path      *string `json:"path"`
	Name      *string `json:"name"`
	Module    *string `json:"module"`
	Icon      *string `json:"icon"`
	SortOrder *int    `json:"sortOrder"`
	IsActive  *bool   `json:"isActive"`
}

// Decode implements the decoder interface.
func (app *UpdatePage) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdatePage) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdatePage(app UpdatePage) (pagebus.UpdatePage, error) {
	return pagebus.UpdatePage{
		Path:      app.Path,
		Name:      app.Name,
		Module:    app.Module,
		Icon:      app.Icon,
		SortOrder: app.SortOrder,
		IsActive:  app.IsActive,
	}, nil
}
