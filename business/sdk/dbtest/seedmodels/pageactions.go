package seedmodels

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
)

// =============================================================================
// PAGE ACTION BUTTON DEFINITIONS
// =============================================================================
// These button actions provide "New" navigation buttons on list pages that
// direct users to the corresponding "new" form pages.

// ButtonActionDefinition defines the properties for a page action button.
// Supports full customization of button appearance and behavior.
type ButtonActionDefinition struct {
	// Required fields
	Label      string // Button text label
	Icon       string // Icon identifier (e.g., "material-symbols:add-circle")
	TargetPath string // Navigation path when clicked

	// Action ordering (required for multiple buttons per page)
	ActionOrder int // Order in which buttons appear (1, 2, 3, etc.)

	// Optional fields with defaults
	Variant            string // Button style: "default", "secondary", "outline", "ghost", "destructive" (defaults to "default")
	Alignment          string // Button position: "left" or "right" (defaults to "right")
	ConfirmationPrompt string // Optional confirmation message before navigation (empty = no confirmation)
	IsActive           bool   // Whether button is enabled (defaults to true)
}

// GetNewButtonActionDefinitions returns a map of page config names to their
// button action definitions. Each page can have multiple buttons.
func GetNewButtonActionDefinitions() map[string][]ButtonActionDefinition {
	return map[string][]ButtonActionDefinition{
		// Admin Module
		"admin_users_page": {
			{
				Label:       "New User",
				Icon:        "material-symbols:person-add",
				TargetPath:  "/admin/users/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},
		"admin_roles_page": {
			{
				Label:       "New Role",
				Icon:        "material-symbols:add-moderator",
				TargetPath:  "/admin/roles/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},

		// Assets Module
		"assets_list_page": {
			{
				Label:       "New Asset",
				Icon:        "material-symbols:add-circle",
				TargetPath:  "/assets/list/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},

		// HR Module
		"hr_employees_page": {
			{
				Label:       "New Employee",
				Icon:        "material-symbols:person-add",
				TargetPath:  "/hr/employees/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},
		"hr_offices_page": {
			{
				Label:       "New Office",
				Icon:        "material-symbols:add-business",
				TargetPath:  "/hr/offices/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},

		// Inventory Module
		"inventory_items_page": {
			{
				Label:       "New Item",
				Icon:        "material-symbols:add-box",
				TargetPath:  "/inventory/items/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},
		"inventory_warehouses_page": {
			{
				Label:       "New Warehouse",
				Icon:        "material-symbols:add-business",
				TargetPath:  "/inventory/warehouses/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},
		"inventory_transfers_page": {
			{
				Label:       "New Transfer",
				Icon:        "material-symbols:add-circle",
				TargetPath:  "/inventory/transfers/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},
		"inventory_adjustments_page": {
			{
				Label:       "New Adjustment",
				Icon:        "material-symbols:tune",
				TargetPath:  "/inventory/adjustments/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},

		// Procurement Module
		"suppliers_page": {
			{
				Label:       "New Supplier",
				Icon:        "material-symbols:add-business",
				TargetPath:  "/procurement/suppliers/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},
		"procurement_purchase_orders": {
			{
				Label:       "New Purchase Order",
				Icon:        "material-symbols:note-add",
				TargetPath:  "/procurement/orders/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},

		// Sales Module
		"sales_customers_page": {
			{
				Label:       "New Customer",
				Icon:        "material-symbols:person-add",
				TargetPath:  "/sales/customers/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},
		"orders_page": {
			{
				Label:       "New Order",
				Icon:        "material-symbols:add-shopping-cart",
				TargetPath:  "/sales/orders/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},

		// Sales Dashboard (multiple buttons)
		"sales_dashboard_page": {
			{
				Label:       "New Customer",
				Icon:        "material-symbols:person-add",
				TargetPath:  "/sales/customers/new",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
			{
				Label:       "New Order",
				Icon:        "material-symbols:add-shopping-cart",
				TargetPath:  "/sales/orders/new",
				ActionOrder: 2,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
		},

		// Main Dashboard (testing ground for all button variants)
		"main_dashboard_page": {
			{
				Label:       "Default Button",
				Icon:        "material-symbols:add-circle",
				TargetPath:  "/test/default",
				ActionOrder: 1,
				Variant:     "default",
				Alignment:   "right",
				IsActive:    true,
			},
			{
				Label:       "Secondary Button",
				Icon:        "material-symbols:edit",
				TargetPath:  "/test/secondary",
				ActionOrder: 2,
				Variant:     "secondary",
				Alignment:   "right",
				IsActive:    true,
			},
			{
				Label:       "Outline Button",
				Icon:        "material-symbols:save",
				TargetPath:  "/test/outline",
				ActionOrder: 3,
				Variant:     "outline",
				Alignment:   "left",
				IsActive:    true,
			},
			{
				Label:       "Ghost Button",
				Icon:        "material-symbols:download",
				TargetPath:  "/test/ghost",
				ActionOrder: 4,
				Variant:     "ghost",
				Alignment:   "left",
				IsActive:    true,
			},
			{
				Label:              "Destructive Button",
				Icon:               "material-symbols:delete-forever",
				TargetPath:         "/test/destructive",
				ActionOrder:        5,
				Variant:            "destructive",
				Alignment:          "right",
				ConfirmationPrompt: "Are you sure you want to perform this destructive action? This is just a test.",
				IsActive:           true,
			},
		},
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// defaultIfEmpty returns the default value if the provided value is empty.
func defaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// defaultIfZero returns the default value if the provided value is the zero value for its type.
func defaultIfZero[T comparable](value, defaultValue T) T {
	var zero T
	if value == zero {
		return defaultValue
	}
	return value
}

// =============================================================================
// VALIDATION FUNCTIONS
// =============================================================================

// ValidateButtonDefinition validates that a button definition has all required fields
// and that optional fields contain valid values.
func ValidateButtonDefinition(def ButtonActionDefinition, index int, pageName string) error {
	// Validate required fields
	if def.Label == "" {
		return fmt.Errorf("button %d for page %s: label is required", index, pageName)
	}
	if def.TargetPath == "" {
		return fmt.Errorf("button %d for page %s: target path is required", index, pageName)
	}
	if def.ActionOrder <= 0 {
		return fmt.Errorf("button %d for page %s: action order must be positive (got %d)",
			index, pageName, def.ActionOrder)
	}

	// Validate variant if specified
	if def.Variant != "" {
		validVariants := map[string]bool{
			"default": true, "secondary": true, "outline": true,
			"ghost": true, "destructive": true,
		}
		if !validVariants[def.Variant] {
			return fmt.Errorf("button %d for page %s: invalid variant %q (must be: default, secondary, outline, ghost, or destructive)",
				index, pageName, def.Variant)
		}
	}

	// Validate alignment if specified
	if def.Alignment != "" && def.Alignment != "left" && def.Alignment != "right" {
		return fmt.Errorf("button %d for page %s: invalid alignment %q (must be 'left' or 'right')",
			index, pageName, def.Alignment)
	}

	return nil
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

// CreateNewButtonAction creates a NewButtonAction from a definition and page config ID.
// Applies sensible defaults for optional fields if not provided.
func CreateNewButtonAction(pageConfigID uuid.UUID, def ButtonActionDefinition) pageactionbus.NewButtonAction {
	return pageactionbus.NewButtonAction{
		PageConfigID:       pageConfigID,
		ActionOrder:        def.ActionOrder, // Required, no default
		IsActive:           defaultIfZero(def.IsActive, true),
		Label:              def.Label,
		Icon:               def.Icon,
		TargetPath:         def.TargetPath,
		Variant:            defaultIfEmpty(def.Variant, "default"),
		Alignment:          defaultIfEmpty(def.Alignment, "right"),
		ConfirmationPrompt: def.ConfirmationPrompt, // Empty string is valid
	}
}
