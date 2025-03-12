package userorganizationbus

import (
	"context"

	"github.com/google/uuid"
)

// TestSeedUserOrganizations is a helper method for testing.
func TestSeedUserOrganizations(ctx context.Context, userIDs uuid.UUIDs, orgIDs uuid.UUIDs, roleIDs uuid.UUIDs, api *Business) ([]UserOrganization, error) {
	userOrgs := make([]UserOrganization, len(orgIDs))

	for i, orgID := range orgIDs {
		newOrg := NewUserOrganization{
			OrganizationalUnitID: orgID,
			UserID:               userIDs[i%len(userIDs)],
			RoleID:               roleIDs[i%len(roleIDs)],
			CreatedBy:            userIDs[0],
		}

		uo, err := api.Create(ctx, newOrg)
		if err != nil {
			return []UserOrganization{}, err
		}
		userOrgs[i] = uo
	}
	return userOrgs, nil
}
