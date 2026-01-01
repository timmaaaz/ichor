package tableaccessbus

import (
	"context"

	"github.com/google/uuid"
)

func TestSeedTableAccess(ctx context.Context, roleIDs uuid.UUIDs, api *Business) ([]TableAccess, error) {

	// Full Access
	newTAs := []NewTableAccess{
		// Assets schema
		{RoleID: uuid.Nil, TableName: "assets.asset_types", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "assets.asset_conditions", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "assets.tags", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "assets.asset_tags", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "assets.assets", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "assets.user_assets", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "assets.valid_assets", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Geography schema
		{RoleID: uuid.Nil, TableName: "geography.countries", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "geography.regions", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "geography.cities", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "geography.streets", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "geography.timezones", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// HR schema
		{RoleID: uuid.Nil, TableName: "hr.user_approval_status", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "hr.titles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "hr.offices", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "hr.homes", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "assets.approval_status", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "hr.reports_to", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "hr.user_approval_comments", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Core schema
		{RoleID: uuid.Nil, TableName: "core.users", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "core.contact_infos", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "core.roles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "core.user_roles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "core.pages", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "core.role_pages", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "core.table_access", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Sales schema
		{RoleID: uuid.Nil, TableName: "assets.fulfillment_status", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "sales.customers", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "sales.orders", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "sales.order_line_items", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "sales.order_fulfillment_statuses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "sales.line_item_fulfillment_statuses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Products schema
		{RoleID: uuid.Nil, TableName: "products.brands", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "products.product_categories", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "products.products", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "products.physical_attributes", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "products.product_costs", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "products.cost_history", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "products.quality_metrics", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Inventory schema
		{RoleID: uuid.Nil, TableName: "inventory.warehouses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory.zones", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory.inventory_locations", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory.inventory_items", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory.lot_trackings", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory.quality_inspections", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory.serial_numbers", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory.inventory_transactions", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory.inventory_adjustments", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory.transfer_orders", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Procurement schema
		{RoleID: uuid.Nil, TableName: "procurement.suppliers", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "procurement.supplier_products", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "procurement.purchase_order_statuses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "procurement.purchase_order_line_item_statuses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "procurement.purchase_orders", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "procurement.purchase_order_line_items", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Config schema
		{RoleID: uuid.Nil, TableName: "config.table_configs", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "config.forms", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "config.form_fields", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "config.page_configs", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "config.page_content", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "config.page_actions", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "config.page_action_buttons", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "config.page_action_dropdowns", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "config.page_action_dropdown_items", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Workflow schema
		{RoleID: uuid.Nil, TableName: "workflow.alerts", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Special tables (kept without schema prefix)
		{RoleID: uuid.Nil, TableName: "formdata", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Virtual Tables (introspection, etc.)
		{RoleID: uuid.Nil, TableName: "introspection", CanCreate: false, CanRead: true, CanUpdate: false, CanDelete: false},
	}

	ret := make([]TableAccess, 0)

	for _, roleID := range roleIDs {
		for _, newTA := range newTAs {
			newTA.RoleID = roleID
			t, err := api.Create(ctx, newTA)
			if err != nil {
				return nil, err
			}
			ret = append(ret, t)
		}
	}

	return ret, nil
}
