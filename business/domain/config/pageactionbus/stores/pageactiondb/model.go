package pageactiondb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
)

// dbPageAction represents the base page action in the database.
type dbPageAction struct {
	ID           uuid.UUID `db:"id"`
	PageConfigID uuid.UUID `db:"page_config_id"`
	ActionType   string    `db:"action_type"`
	ActionOrder  int       `db:"action_order"`
	IsActive     bool      `db:"is_active"`
}

// dbButtonAction represents a button action in the database.
type dbButtonAction struct {
	ActionID           uuid.UUID `db:"action_id"`
	Label              string    `db:"label"`
	Icon               string    `db:"icon"`
	TargetPath         string    `db:"target_path"`
	Variant            string    `db:"variant"`
	Alignment          string    `db:"alignment"`
	ConfirmationPrompt string    `db:"confirmation_prompt"`
}

// dbDropdownAction represents a dropdown action in the database.
type dbDropdownAction struct {
	ActionID uuid.UUID `db:"action_id"`
	Label    string    `db:"label"`
	Icon     string    `db:"icon"`
}

// dbDropdownItem represents a dropdown item in the database.
type dbDropdownItem struct {
	ID               uuid.UUID `db:"id"`
	DropdownActionID uuid.UUID `db:"dropdown_action_id"`
	Label            string    `db:"label"`
	TargetPath       string    `db:"target_path"`
	ItemOrder        int       `db:"item_order"`
}

// toDBPageAction converts a business PageAction to database format.
func toDBPageAction(bus pageactionbus.PageAction) dbPageAction {
	return dbPageAction{
		ID:           bus.ID,
		PageConfigID: bus.PageConfigID,
		ActionType:   string(bus.ActionType),
		ActionOrder:  bus.ActionOrder,
		IsActive:     bus.IsActive,
	}
}

// toBusPageAction converts a database base action to business format.
// Note: This creates the shell - type-specific data must be added separately.
func toBusPageAction(db dbPageAction) pageactionbus.PageAction {
	return pageactionbus.PageAction{
		ID:           db.ID,
		PageConfigID: db.PageConfigID,
		ActionType:   pageactionbus.ActionType(db.ActionType),
		ActionOrder:  db.ActionOrder,
		IsActive:     db.IsActive,
	}
}

// toBusPageActions converts a slice of database actions to business format.
func toBusPageActions(dbs []dbPageAction) []pageactionbus.PageAction {
	actions := make([]pageactionbus.PageAction, len(dbs))
	for i, db := range dbs {
		actions[i] = toBusPageAction(db)
	}
	return actions
}

// toDBButtonAction converts a business ButtonAction to database format.
func toDBButtonAction(actionID uuid.UUID, bus pageactionbus.ButtonAction) dbButtonAction {
	return dbButtonAction{
		ActionID:           actionID,
		Label:              bus.Label,
		Icon:               bus.Icon,
		TargetPath:         bus.TargetPath,
		Variant:            bus.Variant,
		Alignment:          bus.Alignment,
		ConfirmationPrompt: bus.ConfirmationPrompt,
	}
}

// toBusButtonAction converts a database button action to business format.
func toBusButtonAction(db dbButtonAction) pageactionbus.ButtonAction {
	return pageactionbus.ButtonAction{
		Label:              db.Label,
		Icon:               db.Icon,
		TargetPath:         db.TargetPath,
		Variant:            db.Variant,
		Alignment:          db.Alignment,
		ConfirmationPrompt: db.ConfirmationPrompt,
	}
}

// toDBDropdownAction converts a business DropdownAction to database format.
func toDBDropdownAction(actionID uuid.UUID, bus pageactionbus.DropdownAction) dbDropdownAction {
	return dbDropdownAction{
		ActionID: actionID,
		Label:    bus.Label,
		Icon:     bus.Icon,
	}
}

// toBusDropdownAction converts a database dropdown action to business format.
func toBusDropdownAction(db dbDropdownAction, items []dbDropdownItem) pageactionbus.DropdownAction {
	dropdownItems := make([]pageactionbus.DropdownItem, len(items))
	for i, item := range items {
		dropdownItems[i] = pageactionbus.DropdownItem{
			ID:         item.ID,
			Label:      item.Label,
			TargetPath: item.TargetPath,
			ItemOrder:  item.ItemOrder,
		}
	}

	return pageactionbus.DropdownAction{
		Label: db.Label,
		Icon:  db.Icon,
		Items: dropdownItems,
	}
}

// toDBDropdownItem converts a business DropdownItem to database format.
func toDBDropdownItem(dropdownActionID uuid.UUID, bus pageactionbus.NewDropdownItem, id uuid.UUID) dbDropdownItem {
	return dbDropdownItem{
		ID:               id,
		DropdownActionID: dropdownActionID,
		Label:            bus.Label,
		TargetPath:       bus.TargetPath,
		ItemOrder:        bus.ItemOrder,
	}
}
