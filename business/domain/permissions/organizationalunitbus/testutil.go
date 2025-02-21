package organizationalunitbus

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// NOTE: Because of the tree structure intrisic to the the level system in org
// units, a helper function can't be used to generate a bunch of them. We need
// the return result of the parents in order to create the tree structure.

// TestSeedOrganizationalUnits is a helper method for testing.
func TestSeedOrganizationalUnits(ctx context.Context, n int, api *Business) ([]OrganizationalUnit, error) {
	createdOUs := make([]OrganizationalUnit, n)

	// First pass: create root node
	rootOU := NewOrganizationalUnit{
		ParentID:              uuid.Nil,
		Name:                  "Name0",
		CanInheritPermissions: true,
		CanRollupData:         true,
		UnitType:              "UnitType0",
		IsActive:              true,
	}

	// Create root and store its returned data
	createdRoot, err := api.Create(ctx, rootOU)
	if err != nil {
		return nil, fmt.Errorf("failed to create root OU: %w", err)
	}
	createdOUs[0] = createdRoot

	// Create remaining nodes level by level
	for i := 1; i < n; i++ {
		parentIndex := (i - 1) / 2 // formula to find parent index

		nou := NewOrganizationalUnit{
			ParentID:              createdOUs[parentIndex].ID, // Use the ID returned from API
			Name:                  fmt.Sprintf("Name%d", i),
			CanInheritPermissions: true,
			CanRollupData:         true,
			UnitType:              fmt.Sprintf("UnitType%d", i),
			IsActive:              true,
		}

		// Create node and store its returned data
		createdOU, err := api.Create(ctx, nou)
		if err != nil {
			return nil, fmt.Errorf("failed to create OU %d: %w", i, err)
		}
		createdOUs[i] = createdOU
	}

	return createdOUs, nil
}
