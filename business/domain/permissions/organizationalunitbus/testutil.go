package organizationalunitbus

import (
	"context"
	"fmt"
	"strings"

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

	// var parentID uuid.UUID

	// Process OUs level by level to ensure parents are created before children
	for level := 0; level <= 3; level++ {
		for _, ouData := range testData {

			if level != ouData["Level"].(int) {
				continue
			}

			// Create a new OU object from the test data
			newOU, err := testing.MapToStruct[NewOrganizationalUnit](ouData)
			if err != nil {
				return nil, fmt.Errorf("converting test data to NewOrganizationalUnit: %w", err)
			}

			// For non-root nodes, find and set the parent ID
			expectedPath := strings.ReplaceAll(ouData["Path"].(string), " ", "_")
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

			// Use the storer directly to bypass the path/level generation logic
			var ou OrganizationalUnit
			if ou, err = api.Create(ctx, newOU); err != nil {
				return nil, fmt.Errorf("creating organizational unit %s: %w", ou.Name, err)
			}

			// Store the created OU indexed by its path for easy lookup
			ouByPath[ou.Path] = ou
			allOUs = append(allOUs, ou)
		}
	}

	return allOUs, nil
}
