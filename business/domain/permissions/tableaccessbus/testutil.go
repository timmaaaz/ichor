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
		{RoleID: uuid.Nil, TableName: "contact_info", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "brands", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "product_categories", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "warehouses", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "products", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "physical_attributes", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "product_costs", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "suppliers", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "cost_history", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "supplier_products", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "quality_metrics", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "lot_tracking", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "zones", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},

		// Permissions
		{RoleID: uuid.Nil, TableName: "roles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "user_roles", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
		{RoleID: uuid.Nil, TableName: "table_access", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
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
