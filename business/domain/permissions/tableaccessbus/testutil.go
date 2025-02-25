package tableaccessbus

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/testing"
)

// TestSeedTableAccesses is a helper method for testing.
func TestSeedTableAccesses(ctx context.Context, roleID uuid.UUID, tables []string, api *Business) ([]TableAccess, error) {
	tableAccesses := make([]TableAccess, len(testing.TableAccess))

	for i, taMap := range testing.TableAccess {
		nta, err := testing.MapToStruct[NewTableAccess](taMap)
		if err != nil {
			return nil, err
		}
		nta.RoleID = roleID

		ta, err := api.Create(ctx, nta)
		if err != nil {
			return nil, err
		}

		tableAccesses[i] = ta
	}

	return tableAccesses, nil
}
