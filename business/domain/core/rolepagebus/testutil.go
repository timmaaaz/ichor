package rolepagebus

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

// TestSeedRolePages creates a specified number of role page mappings for testing
func TestSeedRolePages(ctx context.Context, n int, api *Business, roleID uuid.UUID, pageID uuid.UUID) ([]RolePage, error) {
	rolePages := make([]RolePage, n)

	for i := 0; i < n; i++ {
		canAccess := i%2 == 0 // Alternate between true and false
		rolePage, err := api.Create(ctx, NewRolePage{
			RoleID:    roleID,
			PageID:    pageID,
			CanAccess: canAccess,
		})
		if err != nil {
			return nil, fmt.Errorf("seeding role page: %w", err)
		}
		rolePages[i] = rolePage
	}

	return rolePages, nil
}

// TestGenerateNewRolePages generates a slice of NewRolePage for testing with alternating access permissions
func TestGenerateNewRolePages(n int, roleIDs []uuid.UUID, pageIDs []uuid.UUID) []NewRolePage {
	newRolePages := make([]NewRolePage, n)

	for i := 0; i < n; i++ {
		roleIdx := rand.Intn(len(roleIDs))
		pageIdx := rand.Intn(len(pageIDs))

		newRolePages[i] = NewRolePage{
			RoleID:    roleIDs[roleIdx],
			PageID:    pageIDs[pageIdx],
			CanAccess: i%2 == 0, // Alternate between true and false
		}
	}

	return newRolePages
}

// TestGenerateSeedRolePages creates multiple role page mappings across different roles and pages
func TestGenerateSeedRolePages(ctx context.Context, api *Business, roleIDs []uuid.UUID, pageIDs []uuid.UUID) ([]RolePage, error) {
	var rolePages []RolePage

	// Create one mapping per role-page combination
	for _, roleID := range roleIDs {
		for i, pageID := range pageIDs {
			// Give access to even-indexed pages
			canAccess := i%2 == 0

			rolePage, err := api.Create(ctx, NewRolePage{
				RoleID:    roleID,
				PageID:    pageID,
				CanAccess: canAccess,
			})
			if err != nil {
				return nil, fmt.Errorf("seeding role page: %w", err)
			}
			rolePages = append(rolePages, rolePage)
		}
	}

	return rolePages, nil
}
