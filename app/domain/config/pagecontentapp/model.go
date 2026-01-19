package pagecontentapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
)

// QueryParams represents query parameters for filtering page content
type QueryParams struct {
	Page         string
	Rows         string
	OrderBy      string
	ID           string
	PageConfigID string
	ContentType  string
	ParentID     string
	IsVisible    string
}

// PageContent represents the application layer model for page content
type PageContent struct {
	ID            string          `json:"id"`
	PageConfigID  string          `json:"page_config_id"`
	ContentType   string          `json:"content_type"`
	Label         string          `json:"label,omitempty"`
	TableConfigID string          `json:"table_config_id,omitempty"`
	FormID        string          `json:"form_id,omitempty"`
	ChartConfigID string          `json:"chart_config_id,omitempty"`
	OrderIndex    int             `json:"order_index"`
	ParentID      string          `json:"parent_id,omitempty"`
	Layout        json.RawMessage `json:"layout"`
	IsVisible     bool            `json:"is_visible"`
	IsDefault     bool            `json:"is_default"`
	Children      []PageContent   `json:"children,omitempty"`
}

// Encode implements the encoder interface for PageContent
func (app PageContent) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// PageContents is a collection wrapper that implements the Encoder interface.
type PageContents []PageContent

// Encode implements the encoder interface for PageContents
func (app PageContents) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// NewPageContent contains data required to create a new page content block
type NewPageContent struct {
	PageConfigID  string          `json:"page_config_id" validate:"required,uuid"`
	ContentType   string          `json:"content_type" validate:"required,oneof=table form tabs container text chart"`
	Label         string          `json:"label"`
	TableConfigID string          `json:"table_config_id" validate:"omitempty,uuid"`
	FormID        string          `json:"form_id" validate:"omitempty,uuid"`
	ChartConfigID string          `json:"chart_config_id" validate:"omitempty,uuid"`
	OrderIndex    int             `json:"order_index"`
	ParentID      string          `json:"parent_id" validate:"omitempty,uuid"`
	Layout        json.RawMessage `json:"layout"`
	IsVisible     bool            `json:"is_visible"`
	IsDefault     bool            `json:"is_default"`
}

// Decode implements the decoder interface for NewPageContent
func (app *NewPageContent) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate performs business rule validation on NewPageContent
func (app NewPageContent) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	// Business rule: content type must match reference
	if app.ContentType == "table" && app.TableConfigID == "" {
		return errs.Newf(errs.InvalidArgument, "table content type requires tableConfigId")
	}
	if app.ContentType == "form" && app.FormID == "" {
		return errs.Newf(errs.InvalidArgument, "form content type requires formId")
	}
	if app.ContentType == "chart" && app.ChartConfigID == "" {
		return errs.Newf(errs.InvalidArgument, "chart content type requires chartConfigId")
	}

	return nil
}

// UpdatePageContent contains data for updating an existing page content block
type UpdatePageContent struct {
	Label      *string          `json:"label"`
	OrderIndex *int             `json:"order_index"`
	Layout     *json.RawMessage `json:"layout"`
	IsVisible  *bool            `json:"is_visible"`
	IsDefault  *bool            `json:"is_default"`
}

// Decode implements the decoder interface for UpdatePageContent
func (app *UpdatePageContent) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate performs business rule validation on UpdatePageContent
func (app UpdatePageContent) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// =============================================================================
// Conversion Functions
// =============================================================================

// ToAppPageContent converts a business layer PageContent to app layer
func ToAppPageContent(bus pagecontentbus.PageContent) PageContent {
	app := PageContent{
		ID:           bus.ID.String(),
		PageConfigID: bus.PageConfigID.String(),
		ContentType:  bus.ContentType,
		Label:        bus.Label,
		OrderIndex:   bus.OrderIndex,
		Layout:       bus.Layout,
		IsVisible:    bus.IsVisible,
		IsDefault:    bus.IsDefault,
	}

	if bus.TableConfigID != uuid.Nil {
		app.TableConfigID = bus.TableConfigID.String()
	}

	if bus.FormID != uuid.Nil {
		app.FormID = bus.FormID.String()
	}

	if bus.ChartConfigID != uuid.Nil {
		app.ChartConfigID = bus.ChartConfigID.String()
	}

	if bus.ParentID != uuid.Nil {
		app.ParentID = bus.ParentID.String()
	}

	// Convert children recursively
	if len(bus.Children) > 0 {
		app.Children = make([]PageContent, len(bus.Children))
		for i, child := range bus.Children {
			app.Children[i] = ToAppPageContent(child)
		}
	}

	return app
}

// ToAppPageContents converts a slice of business layer PageContent to app layer
func ToAppPageContents(bus []pagecontentbus.PageContent) []PageContent {
	app := make([]PageContent, len(bus))
	for i, b := range bus {
		app[i] = ToAppPageContent(b)
	}
	return app
}

// toBusNewPageContent converts app layer NewPageContent to business layer
func toBusNewPageContent(app NewPageContent) (pagecontentbus.NewPageContent, error) {
	pageConfigID, err := uuid.Parse(app.PageConfigID)
	if err != nil {
		return pagecontentbus.NewPageContent{}, fmt.Errorf("parse page config id: %w", err)
	}

	bus := pagecontentbus.NewPageContent{
		PageConfigID: pageConfigID,
		ContentType:  app.ContentType,
		Label:        app.Label,
		OrderIndex:   app.OrderIndex,
		Layout:       app.Layout,
		IsVisible:    app.IsVisible,
		IsDefault:    app.IsDefault,
	}

	if app.TableConfigID != "" {
		tableConfigID, err := uuid.Parse(app.TableConfigID)
		if err != nil {
			return pagecontentbus.NewPageContent{}, fmt.Errorf("parse table config id: %w", err)
		}
		bus.TableConfigID = tableConfigID
	}

	if app.FormID != "" {
		formID, err := uuid.Parse(app.FormID)
		if err != nil {
			return pagecontentbus.NewPageContent{}, fmt.Errorf("parse form id: %w", err)
		}
		bus.FormID = formID
	}

	if app.ChartConfigID != "" {
		chartConfigID, err := uuid.Parse(app.ChartConfigID)
		if err != nil {
			return pagecontentbus.NewPageContent{}, fmt.Errorf("parse chart config id: %w", err)
		}
		bus.ChartConfigID = chartConfigID
	}

	if app.ParentID != "" {
		parentID, err := uuid.Parse(app.ParentID)
		if err != nil {
			return pagecontentbus.NewPageContent{}, fmt.Errorf("parse parent id: %w", err)
		}
		bus.ParentID = parentID
	}

	return bus, nil
}

// toBusUpdatePageContent converts app layer UpdatePageContent to business layer
func toBusUpdatePageContent(app UpdatePageContent) pagecontentbus.UpdatePageContent {
	return pagecontentbus.UpdatePageContent{
		Label:      app.Label,
		OrderIndex: app.OrderIndex,
		Layout:     app.Layout,
		IsVisible:  app.IsVisible,
		IsDefault:  app.IsDefault,
	}
}
