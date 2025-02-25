package organizationalunitbus

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/testing"
)

// TestSeedOrganizationalUnits creates a hierarchy of organizational units for testing.
// It directly sets the paths and levels as expected by the tests.
func TestSeedOrganizationalUnits(ctx context.Context, api *Business) ([]OrganizationalUnit, error) {
	// Create an array to hold all created organizational units
	var allOUs []OrganizationalUnit

	// Create a map to store OUs by their expected path for easy lookup
	ouByPath := make(map[string]OrganizationalUnit)

	// Get the test data from the testing package
	testData := testing.OrganizationalUnits

	// Process OUs level by level to ensure parents are created before children
	for level := 0; level <= 3; level++ {
		for _, ouData := range testData {
			// Skip if this OU isn't at the current level
			if ouData["Level"].(int) != level {
				continue
			}

			// Create a new OU object from the test data
			newOU, err := testing.MapToStruct[NewOrganizationalUnit](ouData)
			if err != nil {
				return nil, fmt.Errorf("converting test data to NewOrganizationalUnit: %w", err)
			}

			// For non-root nodes, find and set the parent ID
			expectedPath := ouData["Path"].(string)
			if level > 0 {
				pathParts := strings.Split(expectedPath, ".")
				parentPath := strings.Join(pathParts[:len(pathParts)-1], ".")

				parent, exists := ouByPath[parentPath]
				if !exists {
					return nil, fmt.Errorf("parent with path %s not found for OU %s",
						parentPath, newOU.Name)
				}

				newOU.ParentID = parent.ID
			}

			// Create the organizational unit
			ou := OrganizationalUnit{
				ID:                    uuid.New(),
				ParentID:              newOU.ParentID,
				Name:                  newOU.Name,
				Level:                 ouData["Level"].(int), // Use exact level from test data
				Path:                  expectedPath,          // Using the expected path from test data
				CanInheritPermissions: newOU.CanInheritPermissions,
				CanRollupData:         newOU.CanRollupData,
				UnitType:              newOU.UnitType,
				IsActive:              newOU.IsActive,
			}

			// Use the storer directly to bypass the path/level generation logic
			if err := api.storer.Create(ctx, ou); err != nil {
				return nil, fmt.Errorf("creating organizational unit %s: %w", ou.Name, err)
			}

			// Store the created OU indexed by its path for easy lookup
			ouByPath[ou.Path] = ou
			allOUs = append(allOUs, ou)
		}
	}

	return allOUs, nil
}
