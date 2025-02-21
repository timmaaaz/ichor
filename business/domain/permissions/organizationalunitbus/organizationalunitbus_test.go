package organizationalunitbus_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_OrganizationalUnit(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_OrganizationalUnit")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}
	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, updatePathPropagation(db.BusDomain, sd), "updatePathPropagation")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	orgUnits, err := organizationalunitbus.TestSeedOrganizationalUnits(ctx, 5, busDomain.OrganizationalUnit)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding organizational units : %w", err)
	}

	return unitest.SeedData{
		OrgUnits: orgUnits,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []organizationalunitbus.OrganizationalUnit{
				sd.OrgUnits[0],
				sd.OrgUnits[1],
				sd.OrgUnits[2],
			},
			ExcFunc: func(ctx context.Context) any {
				got, err := busDomain.OrganizationalUnit.Query(ctx, organizationalunitbus.QueryFilter{}, organizationalunitbus.DefaultOrderBy, page.MustParse("1", "3"))
				if err != nil {
					return err
				}
				return got
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.([]organizationalunitbus.OrganizationalUnit)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(gotResp, exp.([]organizationalunitbus.OrganizationalUnit))
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Create",
			ExpResp: organizationalunitbus.OrganizationalUnit{
				ParentID:              sd.OrgUnits[0].ID,
				Name:                  "Name5",
				Level:                 1,
				Path:                  strings.Join([]string{sd.OrgUnits[0].Path, "Name5"}, "."),
				CanInheritPermissions: true,
				CanRollupData:         true,
				UnitType:              "UnitType5",
				IsActive:              true,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.OrganizationalUnit.Create(ctx, organizationalunitbus.NewOrganizationalUnit{
					ParentID:              sd.OrgUnits[0].ID,
					Name:                  "Name5",
					CanInheritPermissions: true,
					CanRollupData:         true,
					UnitType:              "UnitType5",
					IsActive:              true,
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(organizationalunitbus.OrganizationalUnit)
				if !exists {
					return "error occurred"
				}
				expResp, exists := exp.(organizationalunitbus.OrganizationalUnit)
				if !exists {
					return "error occurred"
				}

				expResp.ID = gotResp.ID

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: organizationalunitbus.OrganizationalUnit{
				ID:                    sd.OrgUnits[0].ID,
				ParentID:              sd.OrgUnits[0].ParentID,
				Name:                  "NewName0",
				Level:                 sd.OrgUnits[0].Level,
				Path:                  "NewName0",
				CanInheritPermissions: sd.OrgUnits[0].CanInheritPermissions,
				CanRollupData:         sd.OrgUnits[0].CanRollupData,
				UnitType:              sd.OrgUnits[0].UnitType,
				IsActive:              sd.OrgUnits[0].IsActive,
			},
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.OrganizationalUnit.Update(ctx, sd.OrgUnits[0], organizationalunitbus.UpdateOrganizationalUnit{
					Name: dbtest.StringPointer("NewName0"),
				})
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(organizationalunitbus.OrganizationalUnit)
				if !exists {
					return "error occurred"
				}

				expResp, exists := exp.(organizationalunitbus.OrganizationalUnit)
				if !exists {
					return "error occurred"
				}

				return cmp.Diff(expResp, gotResp)
			},
		},
	}
}

func updatePathPropagation(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Update_RootNodeName_UpdatesAllChildPaths",
			ExpResp: true, // We expect success verification
			ExcFunc: func(ctx context.Context) any {
				// 1. First rename the root node (index 0)
				newRootName := "RenamedRoot"
				_, err := busDomain.OrganizationalUnit.Update(ctx, sd.OrgUnits[0], organizationalunitbus.UpdateOrganizationalUnit{
					Name: dbtest.StringPointer(newRootName),
				})
				if err != nil {
					return err
				}

				// 2. Query all units to verify paths were updated correctly
				allUnits := make([]organizationalunitbus.OrganizationalUnit, 0, len(sd.OrgUnits))
				for _, ou := range sd.OrgUnits {
					updated, err := busDomain.OrganizationalUnit.QueryByID(ctx, ou.ID)
					if err != nil {
						return err
					}
					allUnits = append(allUnits, updated)
				}

				// 3. Verify each unit's path was updated correctly
				// Root node should have the new name as its path
				if allUnits[0].Path != newRootName {
					return fmt.Errorf("root path not updated correctly, got %s, want %s",
						allUnits[0].Path, newRootName)
				}

				// Level 1 nodes (indices 1 and 2) should have paths prefixed with new root name
				expectedLevel1Prefix := newRootName + "."
				for i := 1; i <= 2; i++ {
					if !strings.HasPrefix(allUnits[i].Path, expectedLevel1Prefix) {
						return fmt.Errorf("level 1 node (index %d) path not updated correctly: %s",
							i, allUnits[i].Path)
					}
					// Extract the node-specific part (should be unchanged except for prefix)
					nodeName := strings.ReplaceAll(fmt.Sprintf("Name%d", i), " ", "_")
					expectedPath := expectedLevel1Prefix + nodeName
					if allUnits[i].Path != expectedPath {
						return fmt.Errorf("level 1 node path mismatch, got %s, want %s",
							allUnits[i].Path, expectedPath)
					}
				}

				// Level 2 nodes (indices 3 and 4) should have paths with updated prefixes
				for i := 3; i <= 4; i++ {
					parentIndex := (i - 1) / 2
					parentName := strings.ReplaceAll(fmt.Sprintf("Name%d", parentIndex), " ", "_")
					nodeName := strings.ReplaceAll(fmt.Sprintf("Name%d", i), " ", "_")

					expectedPath := fmt.Sprintf("%s.%s.%s", newRootName, parentName, nodeName)
					if allUnits[i].Path != expectedPath {
						return fmt.Errorf("level 2 node (index %d) path not updated correctly: got %s, want %s",
							i, allUnits[i].Path, expectedPath)
					}
				}

				return true // All verifications passed
			},
			CmpFunc: func(got, exp any) string {
				success, ok := got.(bool)
				if !ok {
					errVal, isErr := got.(error)
					if isErr {
						return fmt.Sprintf("error occurred: %v", errVal)
					}
					return "unexpected response type"
				}

				if !success {
					return "path verification failed"
				}

				return "" // Empty string means test passed
			},
		},
		{
			Name:    "Update_MidLevelNodeName_UpdatesChildPaths",
			ExpResp: true, // We expect success verification
			ExcFunc: func(ctx context.Context) any {
				// 1. First rename a mid-level node (index 1, which should have children)
				midLevelNode := sd.OrgUnits[1]
				newMidLevelName := "RenamedMidLevel"

				// Get current root node to determine its actual name
				rootNode, err := busDomain.OrganizationalUnit.QueryByID(ctx, sd.OrgUnits[0].ID)
				if err != nil {
					return err
				}
				rootName := strings.ReplaceAll(rootNode.Name, " ", "_")

				_, err = busDomain.OrganizationalUnit.Update(ctx, midLevelNode, organizationalunitbus.UpdateOrganizationalUnit{
					Name: dbtest.StringPointer(newMidLevelName),
				})
				if err != nil {
					return err
				}

				// 2. Find children of the renamed node
				// According to the tree structure, index 1's children would be at indices 3
				childIndices := []int{3}

				// 3. Query all relevant units
				updatedParent, err := busDomain.OrganizationalUnit.QueryByID(ctx, midLevelNode.ID)
				if err != nil {
					return err
				}

				children := make([]organizationalunitbus.OrganizationalUnit, 0, len(childIndices))
				for _, idx := range childIndices {
					child, err := busDomain.OrganizationalUnit.QueryByID(ctx, sd.OrgUnits[idx].ID)
					if err != nil {
						return err
					}
					children = append(children, child)
				}

				// 4. Verify parent's path was updated correctly
				// We use the actual root name queried above
				expectedParentPath := fmt.Sprintf("%s.%s", rootName, newMidLevelName)

				if updatedParent.Path != expectedParentPath {
					return fmt.Errorf("mid-level node path not updated correctly: got %s, want %s",
						updatedParent.Path, expectedParentPath)
				}

				// 5. Verify each child's path reflects the parent's new name
				for i, child := range children {
					childIdx := childIndices[i]
					childName := strings.ReplaceAll(fmt.Sprintf("Name%d", childIdx), " ", "_")
					expectedChildPath := fmt.Sprintf("%s.%s", expectedParentPath, childName)

					if child.Path != expectedChildPath {
						return fmt.Errorf("child path (index %d) not updated correctly: got %s, want %s",
							childIdx, child.Path, expectedChildPath)
					}
				}

				return true // All verifications passed
			},
			CmpFunc: func(got, exp any) string {
				success, ok := got.(bool)
				if !ok {
					errVal, isErr := got.(error)
					if isErr {
						return fmt.Sprintf("error occurred: %v", errVal)
					}
					return "unexpected response type"
				}

				if !success {
					return "path verification failed"
				}

				return "" // Empty string means test passed
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "Delete_LeafNode_Simple",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// Delete a leaf node (index 3)
				leafNode := sd.OrgUnits[3]
				leafID := leafNode.ID

				// Get parent ID for verification
				parentID := leafNode.ParentID

				// Delete the leaf node
				err := busDomain.OrganizationalUnit.Delete(ctx, leafNode)
				if err != nil {
					return fmt.Errorf("failed to delete leaf node: %w", err)
				}

				// Verify leaf is deleted
				_, err = busDomain.OrganizationalUnit.QueryByID(ctx, leafID)
				if err == nil {
					return fmt.Errorf("leaf node still exists after deletion")
				}

				// Verify parent still exists
				_, err = busDomain.OrganizationalUnit.QueryByID(ctx, parentID)
				if err != nil {
					return fmt.Errorf("parent unexpectedly deleted: %w", err)
				}

				return true
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "Delete_MidLevel_CascadesDown",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// Get a mid-level node (index 1) and its children
				midLevelNode := sd.OrgUnits[1]
				midLevelID := midLevelNode.ID

				// Find all children before deletion
				children, err := busDomain.OrganizationalUnit.QueryByParentID(ctx, midLevelID)
				if err != nil {
					return fmt.Errorf("failed to query children: %w", err)
				}

				if len(children) == 0 {
					return fmt.Errorf("expected mid-level node to have children")
				}

				// Store child IDs for verification
				childIDs := make([]uuid.UUID, len(children))
				for i, child := range children {
					childIDs[i] = child.ID
				}

				// Delete the mid-level node
				err = busDomain.OrganizationalUnit.Delete(ctx, midLevelNode)
				if err != nil {
					return fmt.Errorf("failed to delete mid-level node: %w", err)
				}

				// Verify mid-level node is deleted
				_, err = busDomain.OrganizationalUnit.QueryByID(ctx, midLevelID)
				if err == nil {
					return fmt.Errorf("mid-level node still exists after deletion")
				}

				// Verify all children are deleted
				for _, childID := range childIDs {
					_, err = busDomain.OrganizationalUnit.QueryByID(ctx, childID)
					if err == nil {
						return fmt.Errorf("child %s still exists after parent deletion", childID)
					}
				}

				return true
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "Delete_Root_CascadesAll",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				// Get root node (index 0)
				rootNode := sd.OrgUnits[0]

				// Get all remaining nodes in the tree before deletion
				var allIDs []uuid.UUID
				for _, ou := range sd.OrgUnits {
					// Skip indices we've already deleted in previous tests
					if ou.ID == sd.OrgUnits[1].ID || ou.ID == sd.OrgUnits[3].ID {
						continue
					}
					// Verify node still exists before adding to our check list
					_, err := busDomain.OrganizationalUnit.QueryByID(ctx, ou.ID)
					if err == nil {
						allIDs = append(allIDs, ou.ID)
					}
				}

				// Delete the root node
				err := busDomain.OrganizationalUnit.Delete(ctx, rootNode)
				if err != nil {
					return fmt.Errorf("failed to delete root node: %w", err)
				}

				// Verify all remaining nodes are deleted
				for _, id := range allIDs {
					_, err = busDomain.OrganizationalUnit.QueryByID(ctx, id)
					if err == nil {
						return fmt.Errorf("node %s still exists after root deletion", id)
					}
				}

				return true
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
	}
}
