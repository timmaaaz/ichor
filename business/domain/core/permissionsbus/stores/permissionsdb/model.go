package permissionsdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/dbarray"
)

// First, create a new struct to match the query result structure
type userPermissionsRow struct {
	UserID    uuid.UUID      `db:"id"`
	Username  string         `db:"username"`
	TableName string         `db:"table_name"`
	CanCreate bool           `db:"can_create"`
	CanRead   bool           `db:"can_read"`
	CanUpdate bool           `db:"can_update"`
	CanDelete bool           `db:"can_delete"`
	Roles     dbarray.String `db:"roles"`
}

// Convert the query result rows to the expected business structure
func convertRowsToUserPermissions(rows []userPermissionsRow) permissionsbus.UserPermissions {
	// if len(rows) == 0 {
	// 	return permissionsbus.UserPermissions{}
	// }

	// // All rows have the same user ID and username
	// result := permissionsbus.UserPermissions{
	// 	UserID:   rows[0].UserID,
	// 	Username: rows[0].Username,
	// }

	// // Collect unique role names
	// roleMap := make(map[string]bool)
	// for _, row := range rows {
	// 	for _, role := range row.Roles {
	// 		roleMap[role] = true
	// 	}
	// }

	// // Create role objects
	// // Note: Since the query doesn't return role IDs or descriptions,
	// // we're creating simplified role objects with just the name
	// roles := make([]permissionsbus.UserRole, 0, len(roleMap))
	// for roleName := range roleMap {
	// 	roles = append(roles, permissionsbus.UserRole{
	// 		Name: roleName,
	// 		// Other fields will be zero values
	// 	})
	// }
	// result.Roles = roles

	// // Add table access permissions to each role
	// // This is a simplification since the query aggregates permissions by table
	// // across all roles. For a more accurate representation, you'd need to
	// // modify the query to maintain the role-table relationship.
	// tableAccesses := make([]permissionsbus.TableAccess, 0, len(rows))
	// for _, row := range rows {
	// 	tableAccesses = append(tableAccesses, permissionsbus.TableAccess{
	// 		TableName: row.TableName,
	// 		CanCreate: row.CanCreate,
	// 		CanRead:   row.CanRead,
	// 		CanUpdate: row.CanUpdate,
	// 		CanDelete: row.CanDelete,
	// 	})
	// }

	// // For this simplified example, we'll just attach all table permissions to the first role
	// // In a real implementation, you'd need to redesign your query or do additional processing
	// if len(result.Roles) > 0 {
	// 	result.Roles[0].Tables = tableAccesses
	// }

	return permissionsbus.UserPermissions{}
}
