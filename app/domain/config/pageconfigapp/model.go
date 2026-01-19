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
	UserID    string `json:"user_id,omitempty"`
	IsDefault bool   `json:"is_default"`
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
	UserID    string `json:"user_id" validate:"omitempty,uuid"`
	IsDefault bool   `json:"is_default"`
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
	UserID    *string `json:"user_id" validate:"omitempty,uuid"`
	IsDefault *bool   `json:"is_default"`
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

// =============================================================================
// Export/Import Models
// =============================================================================

// ExportPackage represents a JSON export package for page configs.
type ExportPackage struct {
	Version    string              `json:"version"`
	Type       string              `json:"type"`
	ExportedAt string              `json:"exported_at"`
	Count      int                 `json:"count"`
	Data       []PageConfigPackage `json:"data"`
}

// Encode implements the encoder interface.
func (app ExportPackage) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// PageConfigPackage represents a page config with its content and actions.
type PageConfigPackage struct {
	PageConfig PageConfig       `json:"page_config"`
	Contents   []PageContentApp `json:"contents"`
	Actions    PageActionsApp   `json:"actions"`
}

// Encode implements the encoder interface for PageConfigPackage.
func (app PageConfigPackage) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// PageContentApp represents page content for export (app layer).
type PageContentApp struct {
	ID            string `json:"id"`
	PageConfigID  string `json:"page_config_id"`
	ContentType   string `json:"content_type"`
	Label         string `json:"label"`
	TableConfigID string `json:"table_config_id,omitempty"`
	FormID        string `json:"form_id,omitempty"`
	OrderIndex    int    `json:"order_index"`
	ParentID      string `json:"parent_id,omitempty"`
	Layout        string `json:"layout,omitempty"`
	IsVisible     bool   `json:"is_visible"`
	IsDefault     bool   `json:"is_default"`
}

// PageActionsApp represents page actions for export (app layer).
type PageActionsApp struct {
	Buttons    []PageActionApp `json:"buttons"`
	Dropdowns  []PageActionApp `json:"dropdowns"`
	Separators []PageActionApp `json:"separators"`
}

// PageActionApp represents a single page action for export (app layer).
type PageActionApp struct {
	ID           string             `json:"id"`
	PageConfigID string             `json:"page_config_id"`
	ActionType   string             `json:"action_type"`
	ActionOrder  int                `json:"action_order"`
	IsActive     bool               `json:"is_active"`
	Button       *ButtonActionApp   `json:"button,omitempty"`
	Dropdown     *DropdownActionApp `json:"dropdown,omitempty"`
}

// ButtonActionApp represents button-specific data for export (app layer).
type ButtonActionApp struct {
	Label              string `json:"label"`
	Icon               string `json:"icon"`
	TargetPath         string `json:"target_path"`
	Variant            string `json:"variant"`
	Alignment          string `json:"alignment"`
	ConfirmationPrompt string `json:"confirmation_prompt,omitempty"`
}

// DropdownActionApp represents dropdown-specific data for export (app layer).
type DropdownActionApp struct {
	Label string              `json:"label"`
	Icon  string              `json:"icon"`
	Items []DropdownItemApp   `json:"items"`
}

// DropdownItemApp represents a dropdown item for export (app layer).
type DropdownItemApp struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TargetPath string `json:"target_path"`
	ItemOrder  int    `json:"item_order"`
}

// ImportPackage represents a JSON import package for page configs.
type ImportPackage struct {
	Mode string              `json:"mode"` // "merge", "skip", "replace"
	Data []PageConfigPackage `json:"data"`
}

// Decode implements the decoder interface.
func (app *ImportPackage) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app ImportPackage) Validate() error {
	if app.Mode != "merge" && app.Mode != "skip" && app.Mode != "replace" {
		return errs.Newf(errs.InvalidArgument, "mode must be 'merge', 'skip', or 'replace'")
	}

	if len(app.Data) == 0 {
		return errs.Newf(errs.InvalidArgument, "data cannot be empty")
	}

	for i, pkg := range app.Data {
		if pkg.PageConfig.Name == "" {
			return errs.Newf(errs.InvalidArgument, "page config %d: name is required", i)
		}
	}

	return nil
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	ImportedCount int      `json:"imported_count"`
	SkippedCount  int      `json:"skipped_count"`
	UpdatedCount  int      `json:"updated_count"`
	Errors        []string `json:"errors,omitempty"`
}

// Encode implements the encoder interface.
func (app ImportResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// =============================================================================
// Export/Import Conversion Functions
// =============================================================================

// toAppPageConfigWithRelations converts business PageConfigWithRelations to app layer.
func toAppPageConfigWithRelations(bus pageconfigbus.PageConfigWithRelations) PageConfigPackage {
	return PageConfigPackage{
		PageConfig: ToAppPageConfig(bus.PageConfig),
		Contents:   toAppPageContents(bus.Contents),
		Actions:    toAppPageActions(bus.Actions),
	}
}

func toAppPageContents(bus []pageconfigbus.PageContentExport) []PageContentApp {
	app := make([]PageContentApp, len(bus))
	for i, b := range bus {
		content := PageContentApp{
			ID:           b.ID.String(),
			PageConfigID: b.PageConfigID.String(),
			ContentType:  b.ContentType,
			Label:        b.Label,
			OrderIndex:   b.OrderIndex,
			Layout:       string(b.Layout),
			IsVisible:    b.IsVisible,
			IsDefault:    b.IsDefault,
		}
		if b.TableConfigID != uuid.Nil {
			content.TableConfigID = b.TableConfigID.String()
		}
		if b.FormID != uuid.Nil {
			content.FormID = b.FormID.String()
		}
		if b.ParentID != uuid.Nil {
			content.ParentID = b.ParentID.String()
		}
		app[i] = content
	}
	return app
}

func toAppPageActions(bus pageconfigbus.PageActionsExport) PageActionsApp {
	return PageActionsApp{
		Buttons:    toAppPageActionsList(bus.Buttons),
		Dropdowns:  toAppPageActionsList(bus.Dropdowns),
		Separators: toAppPageActionsList(bus.Separators),
	}
}

func toAppPageActionsList(bus []pageconfigbus.PageActionExport) []PageActionApp {
	app := make([]PageActionApp, len(bus))
	for i, b := range bus {
		action := PageActionApp{
			ID:           b.ID.String(),
			PageConfigID: b.PageConfigID.String(),
			ActionType:   b.ActionType,
			ActionOrder:  b.ActionOrder,
			IsActive:     b.IsActive,
		}
		if b.Button != nil {
			action.Button = &ButtonActionApp{
				Label:              b.Button.Label,
				Icon:               b.Button.Icon,
				TargetPath:         b.Button.TargetPath,
				Variant:            b.Button.Variant,
				Alignment:          b.Button.Alignment,
				ConfirmationPrompt: b.Button.ConfirmationPrompt,
			}
		}
		if b.Dropdown != nil {
			items := make([]DropdownItemApp, len(b.Dropdown.Items))
			for j, item := range b.Dropdown.Items {
				items[j] = DropdownItemApp{
					ID:         item.ID.String(),
					Label:      item.Label,
					TargetPath: item.TargetPath,
					ItemOrder:  item.ItemOrder,
				}
			}
			action.Dropdown = &DropdownActionApp{
				Label: b.Dropdown.Label,
				Icon:  b.Dropdown.Icon,
				Items: items,
			}
		}
		app[i] = action
	}
	return app
}

// ToBusPageConfigWithRelations converts an app PageConfigPackage to business PageConfigWithRelations.
func ToBusPageConfigWithRelations(app PageConfigPackage) (pageconfigbus.PageConfigWithRelations, error) {
	configID, err := uuid.Parse(app.PageConfig.ID)
	if err != nil {
		configID = uuid.New()
	}

	var userID uuid.UUID
	if app.PageConfig.UserID != "" {
		userID, err = uuid.Parse(app.PageConfig.UserID)
		if err != nil {
			return pageconfigbus.PageConfigWithRelations{}, fmt.Errorf("parse user id: %w", err)
		}
	}

	config := pageconfigbus.PageConfig{
		ID:        configID,
		Name:      app.PageConfig.Name,
		UserID:    userID,
		IsDefault: app.PageConfig.IsDefault,
	}

	// Convert contents
	contents := make([]pageconfigbus.PageContentExport, len(app.Contents))
	for i, appContent := range app.Contents {
		contentID, _ := uuid.Parse(appContent.ID)
		if contentID == uuid.Nil {
			contentID = uuid.New()
		}

		pageConfigID, _ := uuid.Parse(appContent.PageConfigID)
		tableConfigID, _ := uuid.Parse(appContent.TableConfigID)
		formID, _ := uuid.Parse(appContent.FormID)
		parentID, _ := uuid.Parse(appContent.ParentID)

		contents[i] = pageconfigbus.PageContentExport{
			ID:            contentID,
			PageConfigID:  pageConfigID,
			ContentType:   appContent.ContentType,
			Label:         appContent.Label,
			TableConfigID: tableConfigID,
			FormID:        formID,
			OrderIndex:    appContent.OrderIndex,
			ParentID:      parentID,
			Layout:        []byte(appContent.Layout),
			IsVisible:     appContent.IsVisible,
			IsDefault:     appContent.IsDefault,
		}
	}

	// Convert actions
	actions := pageconfigbus.PageActionsExport{
		Buttons:    toBusPageActionsList(app.Actions.Buttons),
		Dropdowns:  toBusPageActionsList(app.Actions.Dropdowns),
		Separators: toBusPageActionsList(app.Actions.Separators),
	}

	return pageconfigbus.PageConfigWithRelations{
		PageConfig: config,
		Contents:   contents,
		Actions:    actions,
	}, nil
}

func toBusPageActionsList(app []PageActionApp) []pageconfigbus.PageActionExport {
	bus := make([]pageconfigbus.PageActionExport, len(app))
	for i, appAction := range app {
		actionID, _ := uuid.Parse(appAction.ID)
		if actionID == uuid.Nil {
			actionID = uuid.New()
		}

		pageConfigID, _ := uuid.Parse(appAction.PageConfigID)

		action := pageconfigbus.PageActionExport{
			ID:           actionID,
			PageConfigID: pageConfigID,
			ActionType:   appAction.ActionType,
			ActionOrder:  appAction.ActionOrder,
			IsActive:     appAction.IsActive,
		}

		if appAction.Button != nil {
			action.Button = &pageconfigbus.ButtonActionExport{
				Label:              appAction.Button.Label,
				Icon:               appAction.Button.Icon,
				TargetPath:         appAction.Button.TargetPath,
				Variant:            appAction.Button.Variant,
				Alignment:          appAction.Button.Alignment,
				ConfirmationPrompt: appAction.Button.ConfirmationPrompt,
			}
		}

		if appAction.Dropdown != nil {
			items := make([]pageconfigbus.DropdownItemExport, len(appAction.Dropdown.Items))
			for j, appItem := range appAction.Dropdown.Items {
				itemID, _ := uuid.Parse(appItem.ID)
				if itemID == uuid.Nil {
					itemID = uuid.New()
				}
				items[j] = pageconfigbus.DropdownItemExport{
					ID:         itemID,
					Label:      appItem.Label,
					TargetPath: appItem.TargetPath,
					ItemOrder:  appItem.ItemOrder,
				}
			}
			action.Dropdown = &pageconfigbus.DropdownActionExport{
				Label: appAction.Dropdown.Label,
				Icon:  appAction.Dropdown.Icon,
				Items: items,
			}
		}

		bus[i] = action
	}
	return bus
}

// ValidationResult represents validation results for HTTP responses.
// Implements web.Encoder for API layer.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidationError represents a validation error in HTTP responses.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Encode implements web.Encoder for ValidationResult.
func (v ValidationResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(v)
	return data, "application/json", err
}

// ImportStats represents import results for HTTP responses.
type ImportStats struct {
	ImportedCount int `json:"imported_count"`
	UpdatedCount  int `json:"updated_count"`
	SkippedCount  int `json:"skipped_count"`
}

// Encode implements web.Encoder for ImportStats.
func (s ImportStats) Encode() ([]byte, string, error) {
	data, err := json.Marshal(s)
	return data, "application/json", err
}

// toAppValidationResult converts business types to app types.
func toAppValidationResult(bus pageconfigbus.ValidationResult) ValidationResult {
	errors := make([]ValidationError, len(bus.Errors))
	for i, e := range bus.Errors {
		errors[i] = ValidationError{
			Field:   e.Field,
			Message: e.Message,
			Code:    e.Code,
		}
	}
	return ValidationResult{
		Valid:  bus.Valid,
		Errors: errors,
	}
}

// toAppImportStats converts business types to app types.
func toAppImportStats(bus pageconfigbus.ImportStats) ImportStats {
	return ImportStats{
		ImportedCount: bus.ImportedCount,
		UpdatedCount:  bus.UpdatedCount,
		SkippedCount:  bus.SkippedCount,
	}
}
