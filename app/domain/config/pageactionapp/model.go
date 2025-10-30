package pageactionapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
)

// QueryParams holds the query parameters for filtering page actions.
type QueryParams struct {
	Page         string
	Rows         string
	OrderBy      string
	ID           string
	PageConfigID string
	ActionType   string
	IsActive     string
}

// =============================================================================
// Response Models
// =============================================================================

// PageAction represents a unified page action response.
type PageAction struct {
	ID           string          `json:"id"`
	PageConfigID string          `json:"pageConfigId"`
	ActionType   string          `json:"actionType"`
	ActionOrder  int             `json:"actionOrder"`
	IsActive     bool            `json:"isActive"`
	Button       *ButtonAction   `json:"button,omitempty"`
	Dropdown     *DropdownAction `json:"dropdown,omitempty"`
}

func (app PageAction) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ButtonAction contains button-specific configuration.
type ButtonAction struct {
	Label              string `json:"label"`
	Icon               string `json:"icon,omitempty"`
	TargetPath         string `json:"targetPath"`
	Variant            string `json:"variant"`
	Alignment          string `json:"alignment"`
	ConfirmationPrompt string `json:"confirmationPrompt,omitempty"`
}

// DropdownAction contains dropdown-specific configuration including items.
type DropdownAction struct {
	Label string         `json:"label"`
	Icon  string         `json:"icon,omitempty"`
	Items []DropdownItem `json:"items"`
}

// DropdownItem represents a single item within a dropdown action.
type DropdownItem struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TargetPath string `json:"targetPath"`
	ItemOrder  int    `json:"itemOrder"`
}

// ActionsGroupedByType represents page actions grouped by their type.
type ActionsGroupedByType struct {
	Buttons    []PageAction `json:"buttons"`
	Dropdowns  []PageAction `json:"dropdowns"`
	Separators []PageAction `json:"separators"`
}

func (app ActionsGroupedByType) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// PageActions is a collection wrapper that implements the Encoder interface.
type PageActions []PageAction

func (app PageActions) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppPageAction converts a business PageAction to app PageAction.
func ToAppPageAction(bus pageactionbus.PageAction) PageAction {
	app := PageAction{
		ID:           bus.ID.String(),
		PageConfigID: bus.PageConfigID.String(),
		ActionType:   string(bus.ActionType),
		ActionOrder:  bus.ActionOrder,
		IsActive:     bus.IsActive,
	}

	if bus.Button != nil {
		app.Button = &ButtonAction{
			Label:              bus.Button.Label,
			Icon:               bus.Button.Icon,
			TargetPath:         bus.Button.TargetPath,
			Variant:            bus.Button.Variant,
			Alignment:          bus.Button.Alignment,
			ConfirmationPrompt: bus.Button.ConfirmationPrompt,
		}
	}

	if bus.Dropdown != nil {
		items := make([]DropdownItem, len(bus.Dropdown.Items))
		for i, item := range bus.Dropdown.Items {
			items[i] = DropdownItem{
				ID:         item.ID.String(),
				Label:      item.Label,
				TargetPath: item.TargetPath,
				ItemOrder:  item.ItemOrder,
			}
		}
		app.Dropdown = &DropdownAction{
			Label: bus.Dropdown.Label,
			Icon:  bus.Dropdown.Icon,
			Items: items,
		}
	}

	return app
}

// ToAppPageActions converts a slice of business PageActions to app PageActions.
func ToAppPageActions(bus []pageactionbus.PageAction) []PageAction {
	app := make([]PageAction, len(bus))
	for i, v := range bus {
		app[i] = ToAppPageAction(v)
	}
	return app
}

// ToAppActionsGroupedByType converts business ActionsGroupedByType to app.
func ToAppActionsGroupedByType(bus pageactionbus.ActionsGroupedByType) ActionsGroupedByType {
	return ActionsGroupedByType{
		Buttons:    ToAppPageActions(bus.Buttons),
		Dropdowns:  ToAppPageActions(bus.Dropdowns),
		Separators: ToAppPageActions(bus.Separators),
	}
}

// =============================================================================
// Create Models
// =============================================================================

// NewButtonAction contains information needed to create a button action.
type NewButtonAction struct {
	PageConfigID       string `json:"pageConfigId" validate:"required,uuid"`
	ActionOrder        int    `json:"actionOrder"`
	IsActive           bool   `json:"isActive"`
	Label              string `json:"label" validate:"required"`
	Icon               string `json:"icon"`
	TargetPath         string `json:"targetPath" validate:"required"`
	Variant            string `json:"variant" validate:"required,oneof=default secondary outline ghost destructive"`
	Alignment          string `json:"alignment" validate:"required,oneof=left right"`
	ConfirmationPrompt string `json:"confirmationPrompt"`
}

func (app *NewButtonAction) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewButtonAction) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewButtonAction(app NewButtonAction) (pageactionbus.NewButtonAction, error) {
	pageConfigID, err := uuid.Parse(app.PageConfigID)
	if err != nil {
		return pageactionbus.NewButtonAction{}, errs.Newf(errs.InvalidArgument, "parse pageConfigId: %s", err)
	}

	return pageactionbus.NewButtonAction{
		PageConfigID:       pageConfigID,
		ActionOrder:        app.ActionOrder,
		IsActive:           app.IsActive,
		Label:              app.Label,
		Icon:               app.Icon,
		TargetPath:         app.TargetPath,
		Variant:            app.Variant,
		Alignment:          app.Alignment,
		ConfirmationPrompt: app.ConfirmationPrompt,
	}, nil
}

// NewDropdownItem contains information needed to create a dropdown item.
type NewDropdownItem struct {
	Label      string `json:"label" validate:"required"`
	TargetPath string `json:"targetPath" validate:"required"`
	ItemOrder  int    `json:"itemOrder"`
}

// NewDropdownAction contains information needed to create a dropdown action.
type NewDropdownAction struct {
	PageConfigID string            `json:"pageConfigId" validate:"required,uuid"`
	ActionOrder  int               `json:"actionOrder"`
	IsActive     bool              `json:"isActive"`
	Label        string            `json:"label" validate:"required"`
	Icon         string            `json:"icon"`
	Items        []NewDropdownItem `json:"items" validate:"required,min=1,dive"`
}

func (app *NewDropdownAction) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewDropdownAction) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewDropdownAction(app NewDropdownAction) (pageactionbus.NewDropdownAction, error) {
	pageConfigID, err := uuid.Parse(app.PageConfigID)
	if err != nil {
		return pageactionbus.NewDropdownAction{}, errs.Newf(errs.InvalidArgument, "parse pageConfigId: %s", err)
	}

	items := make([]pageactionbus.NewDropdownItem, len(app.Items))
	for i, item := range app.Items {
		items[i] = pageactionbus.NewDropdownItem{
			Label:      item.Label,
			TargetPath: item.TargetPath,
			ItemOrder:  item.ItemOrder,
		}
	}

	return pageactionbus.NewDropdownAction{
		PageConfigID: pageConfigID,
		ActionOrder:  app.ActionOrder,
		IsActive:     app.IsActive,
		Label:        app.Label,
		Icon:         app.Icon,
		Items:        items,
	}, nil
}

// NewSeparatorAction contains information needed to create a separator action.
type NewSeparatorAction struct {
	PageConfigID string `json:"pageConfigId" validate:"required,uuid"`
	ActionOrder  int    `json:"actionOrder"`
	IsActive     bool   `json:"isActive"`
}

func (app *NewSeparatorAction) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewSeparatorAction) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewSeparatorAction(app NewSeparatorAction) (pageactionbus.NewSeparatorAction, error) {
	pageConfigID, err := uuid.Parse(app.PageConfigID)
	if err != nil {
		return pageactionbus.NewSeparatorAction{}, errs.Newf(errs.InvalidArgument, "parse pageConfigId: %s", err)
	}

	return pageactionbus.NewSeparatorAction{
		PageConfigID: pageConfigID,
		ActionOrder:  app.ActionOrder,
		IsActive:     app.IsActive,
	}, nil
}

// =============================================================================
// Update Models
// =============================================================================

// UpdateButtonAction contains information needed to update a button action.
type UpdateButtonAction struct {
	PageConfigID       *string `json:"pageConfigId" validate:"omitempty,uuid"`
	ActionOrder        *int    `json:"actionOrder"`
	IsActive           *bool   `json:"isActive"`
	Label              *string `json:"label"`
	Icon               *string `json:"icon"`
	TargetPath         *string `json:"targetPath"`
	Variant            *string `json:"variant" validate:"omitempty,oneof=default secondary outline ghost destructive"`
	Alignment          *string `json:"alignment" validate:"omitempty,oneof=left right"`
	ConfirmationPrompt *string `json:"confirmationPrompt"`
}

func (app *UpdateButtonAction) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateButtonAction) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateButtonAction(app UpdateButtonAction) (pageactionbus.UpdateButtonAction, error) {
	bus := pageactionbus.UpdateButtonAction{
		ActionOrder:        app.ActionOrder,
		IsActive:           app.IsActive,
		Label:              app.Label,
		Icon:               app.Icon,
		TargetPath:         app.TargetPath,
		Variant:            app.Variant,
		Alignment:          app.Alignment,
		ConfirmationPrompt: app.ConfirmationPrompt,
	}

	if app.PageConfigID != nil {
		pageConfigID, err := uuid.Parse(*app.PageConfigID)
		if err != nil {
			return pageactionbus.UpdateButtonAction{}, errs.Newf(errs.InvalidArgument, "parse pageConfigId: %s", err)
		}
		bus.PageConfigID = &pageConfigID
	}

	return bus, nil
}

// UpdateDropdownAction contains information needed to update a dropdown action.
type UpdateDropdownAction struct {
	PageConfigID *string            `json:"pageConfigId" validate:"omitempty,uuid"`
	ActionOrder  *int               `json:"actionOrder"`
	IsActive     *bool              `json:"isActive"`
	Label        *string            `json:"label"`
	Icon         *string            `json:"icon"`
	Items        *[]NewDropdownItem `json:"items" validate:"omitempty,min=1,dive"`
}

func (app *UpdateDropdownAction) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateDropdownAction) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateDropdownAction(app UpdateDropdownAction) (pageactionbus.UpdateDropdownAction, error) {
	bus := pageactionbus.UpdateDropdownAction{
		ActionOrder: app.ActionOrder,
		IsActive:    app.IsActive,
		Label:       app.Label,
		Icon:        app.Icon,
	}

	if app.PageConfigID != nil {
		pageConfigID, err := uuid.Parse(*app.PageConfigID)
		if err != nil {
			return pageactionbus.UpdateDropdownAction{}, errs.Newf(errs.InvalidArgument, "parse pageConfigId: %s", err)
		}
		bus.PageConfigID = &pageConfigID
	}

	if app.Items != nil {
		items := make([]pageactionbus.NewDropdownItem, len(*app.Items))
		for i, item := range *app.Items {
			items[i] = pageactionbus.NewDropdownItem{
				Label:      item.Label,
				TargetPath: item.TargetPath,
				ItemOrder:  item.ItemOrder,
			}
		}
		bus.Items = &items
	}

	return bus, nil
}

// UpdateSeparatorAction contains information needed to update a separator action.
type UpdateSeparatorAction struct {
	PageConfigID *string `json:"pageConfigId" validate:"omitempty,uuid"`
	ActionOrder  *int    `json:"actionOrder"`
	IsActive     *bool   `json:"isActive"`
}

func (app *UpdateSeparatorAction) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateSeparatorAction) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateSeparatorAction(app UpdateSeparatorAction) (pageactionbus.UpdateSeparatorAction, error) {
	bus := pageactionbus.UpdateSeparatorAction{
		ActionOrder: app.ActionOrder,
		IsActive:    app.IsActive,
	}

	if app.PageConfigID != nil {
		pageConfigID, err := uuid.Parse(*app.PageConfigID)
		if err != nil {
			return pageactionbus.UpdateSeparatorAction{}, errs.Newf(errs.InvalidArgument, "parse pageConfigId: %s", err)
		}
		bus.PageConfigID = &pageConfigID
	}

	return bus, nil
}

// =============================================================================
// Batch Operations
// =============================================================================

// BatchActionRequest represents a single action in a batch create request.
type BatchActionRequest struct {
	ActionType string              `json:"actionType" validate:"required,oneof=button dropdown separator"`
	Button     *NewButtonAction    `json:"button" validate:"required_if=ActionType button"`
	Dropdown   *NewDropdownAction  `json:"dropdown" validate:"required_if=ActionType dropdown"`
	Separator  *NewSeparatorAction `json:"separator" validate:"required_if=ActionType separator"`
}

// BatchCreateRequest represents a batch create request for multiple actions.
type BatchCreateRequest struct {
	Actions []BatchActionRequest `json:"actions" validate:"required,min=1,dive"`
}

func (app *BatchCreateRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app BatchCreateRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	// Additional validation: ensure only one action type is set per request
	for i, action := range app.Actions {
		count := 0
		if action.Button != nil {
			count++
		}
		if action.Dropdown != nil {
			count++
		}
		if action.Separator != nil {
			count++
		}
		if count != 1 {
			return errs.Newf(errs.InvalidArgument, "action[%d]: exactly one of button, dropdown, or separator must be set", i)
		}

		// Validate the set action matches the actionType
		switch action.ActionType {
		case "button":
			if action.Button == nil {
				return errs.Newf(errs.InvalidArgument, "action[%d]: actionType is button but button is nil", i)
			}
			if err := action.Button.Validate(); err != nil {
				return errs.Newf(errs.InvalidArgument, "action[%d].button: %s", i, err)
			}
		case "dropdown":
			if action.Dropdown == nil {
				return errs.Newf(errs.InvalidArgument, "action[%d]: actionType is dropdown but dropdown is nil", i)
			}
			if err := action.Dropdown.Validate(); err != nil {
				return errs.Newf(errs.InvalidArgument, "action[%d].dropdown: %s", i, err)
			}
		case "separator":
			if action.Separator == nil {
				return errs.Newf(errs.InvalidArgument, "action[%d]: actionType is separator but separator is nil", i)
			}
			if err := action.Separator.Validate(); err != nil {
				return errs.Newf(errs.InvalidArgument, "action[%d].separator: %s", i, err)
			}
		}
	}

	return nil
}
