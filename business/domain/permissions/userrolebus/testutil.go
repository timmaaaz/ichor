package userrolebus

import (
	"context"

	"github.com/google/uuid"
)

// TestNewUserRoles returns a slice of NewUserRole for testing.
func TestNewUserRoles(n int, userID uuid.UUIDs, roleIDs uuid.UUIDs) []NewUserRole {
	newUserRoles := make([]NewUserRole, n)

	for i := 0; i < n; i++ {
		nur := NewUserRole{
			UserID: userID[i],
			RoleID: roleIDs[i%len(roleIDs)], // Should be the same length as user roles.
		}

		newUserRoles[i] = nur
	}

	return newUserRoles
}

// TestSeedRoles is a helper method for testing.
func TestSeedUserRoles(ctx context.Context, userIDs uuid.UUIDs, roleIDs uuid.UUIDs, api *Business) ([]UserRole, error) {
	newUserRoles := TestNewUserRoles(len(userIDs), userIDs, roleIDs)
	userRoles := make([]UserRole, len(newUserRoles))

	for i, nur := range newUserRoles {
		ur, err := api.Create(ctx, nur)
		if err != nil {
			return nil, err
		}

		userRoles[i] = ur
	}

	return userRoles, nil
}
