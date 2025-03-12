package crossunitpermissionsbus

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/testing"
)

// TestSeedCrossUnitPermissions is a helper method for testing.
func TestSeedCrossUnitPermissions(ctx context.Context, sourceUnitIDs uuid.UUIDs, targetUnitIDs uuid.UUIDs, grantedBy uuid.UUID, api *Business) ([]CrossUnitPermission, error) {
	crossUnitPermissions := make([]CrossUnitPermission, len(testing.CrossUnitPermission))

	for i, cupMap := range testing.CrossUnitPermission {
		ncup, err := testing.MapToStruct[NewCrossUnitPermission](cupMap)
		if err != nil {
			return nil, err
		}

		ncup.SourceUnitID = sourceUnitIDs[i]
		ncup.TargetUnitID = targetUnitIDs[i]
		ncup.GrantedBy = grantedBy

		cup, err := api.Create(ctx, ncup)
		if err != nil {
			return nil, err
		}

		crossUnitPermissions[i] = cup
	}

	return crossUnitPermissions, nil
}
