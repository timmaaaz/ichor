package rolebus

import (
	"context"

	"github.com/timmaaaz/ichor/business/domain/permissions/testing"
)

// TestSeedRoles is a helper method for testing.
func TestSeedRoles(ctx context.Context, n int, api *Business) ([]Role, error) {
	roles := make([]Role, len(testing.Roles))

	for i, roleMap := range testing.Roles {
		nr, err := testing.MapToStruct[NewRole](roleMap)
		if err != nil {
			return nil, err
		}

		r, err := api.Create(ctx, nr)
		if err != nil {
			return nil, err
		}

		roles[i] = r
	}

	return roles, nil
}
