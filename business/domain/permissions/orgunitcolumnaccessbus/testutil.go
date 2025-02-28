package orgunitcolumnaccessbus

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/testing"
)

// TestSeedOrgUnitColumnAccesses is a helper method for testing.
func TestSeedOrgUnitColumnAccesses(ctx context.Context, orgIDs uuid.UUIDs, api *Business) ([]OrgUnitColumnAccess, error) {
	orgUnitColumnAccesses := make([]OrgUnitColumnAccess, len(testing.OrgUnitColumnAccess))

	for i, oucaMap := range testing.OrgUnitColumnAccess {
		nouca, err := testing.MapToStruct[NewOrgUnitColumnAccess](oucaMap)
		if err != nil {
			return nil, err
		}

		nouca.OrganizationalUnitID = orgIDs[i]

		ouca, err := api.Create(ctx, nouca)
		if err != nil {
			return nil, err
		}

		orgUnitColumnAccesses[i] = ouca
	}

	return orgUnitColumnAccesses, nil
}
