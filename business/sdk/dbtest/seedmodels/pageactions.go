package seedmodels

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
)

// =============================================================================
// PAGE ACTION BUTTON DEFINITIONS
// =============================================================================
// These button actions provide "New" navigation buttons on list pages that
// direct users to the corresponding "new" form pages.

// ButtonActionDefinition defines the properties for a page action button
type ButtonActionDefinition struct {
	Label      string
	Icon       string
	TargetPath string
}

// GetNewButtonActionDefinitions returns a map of page config names to their
// "New" button action definitions
func GetNewButtonActionDefinitions() map[string]ButtonActionDefinition {
	return map[string]ButtonActionDefinition{
		// Admin Module
		"admin_users_page": {
			Label:      "New User",
			Icon:       "material-symbols:person-add",
			TargetPath: "/admin/users/new",
		},
		"admin_roles_page": {
			Label:      "New Role",
			Icon:       "material-symbols:add-moderator",
			TargetPath: "/admin/roles/new",
		},

		// Assets Module
		"assets_list_page": {
			Label:      "New Asset",
			Icon:       "material-symbols:add-circle",
			TargetPath: "/assets/list/new",
		},

		// HR Module
		"hr_employees_page": {
			Label:      "New Employee",
			Icon:       "material-symbols:person-add",
			TargetPath: "/hr/employees/new",
		},
		"hr_offices_page": {
			Label:      "New Office",
			Icon:       "material-symbols:add-business",
			TargetPath: "/hr/offices/new",
		},

		// Inventory Module
		"inventory_items_page": {
			Label:      "New Item",
			Icon:       "material-symbols:add-box",
			TargetPath: "/inventory/items/new",
		},
		"inventory_warehouses_page": {
			Label:      "New Warehouse",
			Icon:       "material-symbols:add-business",
			TargetPath: "/inventory/warehouses/new",
		},
		"inventory_transfers_page": {
			Label:      "New Transfer",
			Icon:       "material-symbols:add-circle",
			TargetPath: "/inventory/transfers/new",
		},
		"inventory_adjustments_page": {
			Label:      "New Adjustment",
			Icon:       "material-symbols:tune",
			TargetPath: "/inventory/adjustments/new",
		},

		// Procurement Module
		"suppliers_page": {
			Label:      "New Supplier",
			Icon:       "material-symbols:add-business",
			TargetPath: "/procurement/suppliers/new",
		},
		"procurement_purchase_orders": {
			Label:      "New Purchase Order",
			Icon:       "material-symbols:note-add",
			TargetPath: "/procurement/orders/new",
		},

		// Sales Module
		"sales_customers_page": {
			Label:      "New Customer",
			Icon:       "material-symbols:person-add",
			TargetPath: "/sales/customers/new",
		},
		"orders_page": {
			Label:      "New Order",
			Icon:       "material-symbols:add-shopping-cart",
			TargetPath: "/sales/orders/new",
		},
	}
}

// CreateNewButtonAction creates a NewButtonAction from a definition and page config ID
func CreateNewButtonAction(pageConfigID uuid.UUID, def ButtonActionDefinition) pageactionbus.NewButtonAction {
	return pageactionbus.NewButtonAction{
		PageConfigID:       pageConfigID,
		ActionOrder:        1,
		IsActive:           true,
		Label:              def.Label,
		Icon:               def.Icon,
		TargetPath:         def.TargetPath,
		Variant:            "default",
		Alignment:          "right",
		ConfirmationPrompt: "",
	}
}
