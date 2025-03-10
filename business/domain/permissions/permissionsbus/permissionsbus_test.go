package permissionsbus_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

// User roles
// users[0]: admin

func Test_Permissions(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_Permissions")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, query(db.BusDomain, sd), "query")
	// unitest.Run(t, create(db.BusDomain, sd), "create")
	// unitest.Run(t, update(db.BusDomain, sd), "update")
	// unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	users, err := busDomain.User.Query(ctx, userbus.QueryFilter{}, order.NewBy(userbus.OrderByUsername, order.ASC), page.MustParse("1", "100"))
	if err != nil {
		return unitest.SeedData{}, err
	}
	seedUsers := make([]unitest.User, len(users))
	for i, u := range users {
		seedUsers[i] = unitest.User{
			User: u,
		}
	}

	// CONSTRUCT SEED DATA
	return unitest.SeedData{
		Users: seedUsers,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	// Create the Role
	role := &userrolebus.UserRole{
		ID:     uuid.Nil, // Will be set later with got.Role.ID
		UserID: uuid.Nil,
		RoleID: uuid.Nil,
	}

	// Create TableAccess map with all entries from the JSON
	tableAccess := make(map[string]tableaccessbus.TableAccess)

	// Add all table access entries
	tableAccess["approval_status"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "approval_status",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["asset_conditions"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "asset_conditions",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["asset_tags"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "asset_tags",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["asset_types"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "asset_types",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["assets"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "assets",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["brands"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "brands",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["cities"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "cities",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["contact_info"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "contact_info",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["countries"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "countries",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["cross_unit_permissions"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "cross_unit_permissions",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["fulfillment_status"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "fulfillment_status",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["homes"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "homes",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["offices"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "offices",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["org_unit_column_access"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "org_unit_column_access",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["organizational_units"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "organizational_units",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["permission_overrides"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "permission_overrides",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["product_categories"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "product_categories",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["products"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "products",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["regions"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "regions",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["reports_to"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "reports_to",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["restricted_columns"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "restricted_columns",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["roles"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "roles",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["streets"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "streets",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["table_access"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "table_access",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["tags"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "tags",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["temporary_unit_access"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "temporary_unit_access",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["titles"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "titles",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["user_approval_comments"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "user_approval_comments",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["user_approval_status"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "user_approval_status",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["user_assets"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "user_assets",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["user_organizations"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "user_organizations",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["user_roles"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "user_roles",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["users"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "users",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["valid_assets"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "valid_assets",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	tableAccess["view_products"] = tableaccessbus.TableAccess{
		ID:        uuid.Nil,
		RoleID:    uuid.Nil,
		TableName: "view_products",
		CanCreate: true,
		CanRead:   true,
		CanUpdate: true,
		CanDelete: true,
	}

	// Create UserPermissions instance
	exp := permissionsbus.UserPermissions{
		UserID:      uuid.Nil,
		Username:    "", // Empty in the JSON
		RoleName:    "ADMIN",
		Role:        role,
		TableAccess: tableAccess,
	}

	return []unitest.Table{
		{
			Name:    "Query",
			ExpResp: exp,
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.Permissions.QueryUserPermissions(ctx, sd.Users[1].ID)
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(exp, got any) string {
				gotResp, exists := got.(permissionsbus.UserPermissions)
				if !exists {
					return "got is not a *permissionsbus.UserPermissions"
				}

				expResp, exists := exp.(permissionsbus.UserPermissions)
				if !exists {
					return "exp is not a *permissionsbus.UserPermissions"
				}

				expResp.UserID = gotResp.UserID

				expResp.Role.ID = gotResp.Role.ID
				expResp.Role.UserID = gotResp.Role.UserID
				expResp.Role.RoleID = gotResp.Role.RoleID

				for k, v := range expResp.TableAccess {
					v.ID = gotResp.TableAccess[k].ID
					v.RoleID = gotResp.TableAccess[k].RoleID
					expResp.TableAccess[k] = v
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}
