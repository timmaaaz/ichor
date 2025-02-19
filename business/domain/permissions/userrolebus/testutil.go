package userrolebus

import (
	"context"

	"github.com/google/uuid"
)

// TestNewUserRoles returns a slice of NewUserRole for testing.
func TestNewUserRoles(n int, userID uuid.UUID, roleIDs uuid.UUIDs) []NewUserRole {
	newUserRoles := make([]NewUserRole, n)

	for i := 0; i < n; i++ {
		nur := NewUserRole{
			UserID: userID,
			RoleID: roleIDs[i], // Should be the same length as user roles.
		}

		newUserRoles[i] = nur
	}

	return newUserRoles
}

// TestSeedRoles is a helper method for testing.
func TestSeedUserRoles(ctx context.Context, n int, userID uuid.UUID, roleIDs uuid.UUIDs, api *Business) ([]UserRole, error) {
	newUserRoles := TestNewUserRoles(n, userID, roleIDs)
	userRoles := make([]UserRole, n)

	for i, nur := range newUserRoles {
		ur, err := api.Create(ctx, nur)
		if err != nil {
			return nil, err
		}

		userRoles[i] = ur
	}

	return userRoles, nil
}
