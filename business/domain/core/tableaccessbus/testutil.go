package tableaccessbus

import (
	"context"

	"github.com/google/uuid"
)

func TestSeedTableAccess(ctx context.Context, roleIDs uuid.UUIDs, api *Business) ([]TableAccess, error) {

	// Full Access
	newTAs := []NewTableAccess{
		{RoleID: uuid.Nil, TableName: "asset_types", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "asset_conditions", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "countries", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "regions", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "cities", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "streets", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "user_approval_status", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "titles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "offices", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "users", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "valid_assets", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "homes", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "approval_status", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "fulfillment_status", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "tags", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "asset_tags", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "reports_to", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "assets", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "user_assets", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "user_approval_comments", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "contact_infos", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "brands", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "product_categories", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "warehouses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "products", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "physical_attributes", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "product_costs", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "purchase_order_line_item_statuses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "purchase_order_statuses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "purchase_orders", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "purchase_order_line_items", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "suppliers", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "cost_history", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "supplier_products", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "quality_metrics", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "lot_trackings", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "zones", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory_locations", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory_items", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "quality_inspections", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "serial_numbers", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory_transactions", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "inventory_adjustments", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "transfer_orders", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "order_fulfillment_statuses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "line_item_fulfillment_statuses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "customers", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "orders", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "order_line_items", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "table_access", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "table_configs", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "forms", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "form_fields", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "formdata", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Permissions
		{RoleID: uuid.Nil, TableName: "roles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "user_roles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "pages", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "role_pages", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
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
