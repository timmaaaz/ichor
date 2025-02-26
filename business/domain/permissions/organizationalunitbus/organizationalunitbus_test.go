package organizationalunitbus_test

import (
	"context"
	"fmt"
	"sort"
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

	orgUnits, err := organizationalunitbus.TestSeedOrganizationalUnits(ctx, busDomain.OrganizationalUnit)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding organizational units : %w", err)
	}

	return unitest.SeedData{
		OrgUnits: orgUnits,
	}, nil
}

func query(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	// Make a copy of the OrgUnits slice
	sortedUnits := make([]organizationalunitbus.OrganizationalUnit, len(sd.OrgUnits))
	copy(sortedUnits, sd.OrgUnits)

	// Sort the copy
	sort.Slice(sortedUnits, func(i, j int) bool {
		return sortedUnits[i].Name < sortedUnits[j].Name
	})

	return []unitest.Table{
		{
			Name: "Query",
			ExpResp: []organizationalunitbus.OrganizationalUnit{
				sortedUnits[0],
				sortedUnits[1],
				sortedUnits[2],
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
				Level:                 sd.OrgUnits[0].Level + 1, // Calculate level based on parent's level
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

	p := sd.OrgUnits[0].Path
	parts := strings.Split(p, ".")
	parts[len(parts)-1] = "NewName0"
	newPath := strings.Join(parts, ".")

	return []unitest.Table{
		{
			Name: "Update",
			ExpResp: organizationalunitbus.OrganizationalUnit{
				ID:                    sd.OrgUnits[0].ID,
				ParentID:              sd.OrgUnits[0].ParentID,
				Name:                  "NewName0",
				Level:                 sd.OrgUnits[0].Level,
				Path:                  newPath,
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
				// Store original values for verification
				originalPaths := make(map[uuid.UUID]string)
				for _, ou := range sd.OrgUnits {
					originalPaths[ou.ID] = ou.Path
				}

				// 1. Rename the root node (index 0)
				rootNode := sd.OrgUnits[0]
				newRootName := "RenamedRoot"
				updatedRoot, err := busDomain.OrganizationalUnit.Update(ctx, rootNode, organizationalunitbus.UpdateOrganizationalUnit{
					Name: dbtest.StringPointer(newRootName),
				})
				if err != nil {
					return fmt.Errorf("failed to update root node: %w", err)
				}

				// Verify root's path was updated correctly
				expectedRootPath := strings.ReplaceAll(newRootName, " ", "_")
				if updatedRoot.Path != expectedRootPath {
					return fmt.Errorf("root path not updated correctly: got %s, want %s",
						updatedRoot.Path, expectedRootPath)
				}

				// 2. Fetch all children to verify their paths were updated
				for i := 1; i < len(sd.OrgUnits); i++ {
					updated, err := busDomain.OrganizationalUnit.QueryByID(ctx, sd.OrgUnits[i].ID)
					if err != nil {
						return fmt.Errorf("failed to query unit %s: %w", sd.OrgUnits[i].ID, err)
					}

					// The path should now start with the new root name instead of the old one
					originalPath := originalPaths[sd.OrgUnits[i].ID]
					rootPathPrefix := originalPaths[rootNode.ID]

					if !strings.HasPrefix(originalPath, rootPathPrefix) {
						return fmt.Errorf("original hierarchy is not valid: %s should start with %s",
							originalPath, rootPathPrefix)
					}

					pathSuffix := strings.TrimPrefix(originalPath, rootPathPrefix)
					expectedPath := updatedRoot.Path + pathSuffix

					if updated.Path != expectedPath {
						return fmt.Errorf("child path not updated correctly: got %s, want %s",
							updated.Path, expectedPath)
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
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				var midLevelNode organizationalunitbus.OrganizationalUnit
				var midLevelChildren []organizationalunitbus.OrganizationalUnit

				// Locate a mid-level node (level == 1) in seed data, but re-query from DB for its *updated* info.
				var midLevelCandidates []organizationalunitbus.OrganizationalUnit
				for _, ou := range sd.OrgUnits {
					if ou.Level == 1 {
						midLevelCandidates = append(midLevelCandidates, ou)
					}
				}
				if len(midLevelCandidates) == 0 {
					return fmt.Errorf("no mid-level nodes found in seed data")
				}

				for _, candidate := range midLevelCandidates {
					// Re-query the candidate from the DB to get any updated path from the first test.
					upCandidate, err := busDomain.OrganizationalUnit.QueryByID(ctx, candidate.ID)
					if err != nil {
						return fmt.Errorf("error re-querying candidate: %w", err)
					}

					children, err := busDomain.OrganizationalUnit.QueryByParentID(ctx, upCandidate.ID)
					if err != nil {
						return fmt.Errorf("error querying children for candidate %s: %w", upCandidate.ID, err)
					}

					if len(children) > 0 {
						midLevelNode = upCandidate
						midLevelChildren = children
						break
					}
				}

				if midLevelNode.ID == uuid.Nil {
					return fmt.Errorf("no mid-level node with children found after re-query")
				}

				// Keep track of original child paths before the mid-level rename.
				originalPaths := make(map[uuid.UUID]string)
				for _, child := range midLevelChildren {
					originalPaths[child.ID] = child.Path
				}

				originalMidLevelPath := midLevelNode.Path

				// 2. Rename the mid-level node.
				newMidLevelName := "RenamedMidLevel"
				updatedMidLevel, err := busDomain.OrganizationalUnit.Update(
					ctx,
					midLevelNode,
					organizationalunitbus.UpdateOrganizationalUnit{
						Name: dbtest.StringPointer(newMidLevelName),
					},
				)
				if err != nil {
					return fmt.Errorf("failed to update mid-level node: %w", err)
				}

				// 3. Verify the mid-level node's path was updated correctly.
				parent, err := busDomain.OrganizationalUnit.QueryByID(ctx, midLevelNode.ParentID)
				if err != nil {
					return fmt.Errorf("failed to query parent of mid-level node: %w", err)
				}
				expectedMidLevelPath := fmt.Sprintf("%s.%s",
					parent.Path,
					strings.ReplaceAll(newMidLevelName, " ", "_"),
				)

				if updatedMidLevel.Path != expectedMidLevelPath {
					return fmt.Errorf("mid-level path not updated correctly: got %s, want %s",
						updatedMidLevel.Path, expectedMidLevelPath)
				}

				// 4. Verify all child paths got updated by replacing the old mid-level path prefix.
				for _, child := range midLevelChildren {
					updatedChild, err := busDomain.OrganizationalUnit.QueryByID(ctx, child.ID)
					if err != nil {
						return fmt.Errorf("failed to query child %s: %w", child.ID, err)
					}

					originalChildPath := originalPaths[child.ID]
					if !strings.HasPrefix(originalChildPath, originalMidLevelPath) {
						return fmt.Errorf(
							"original child path %s doesn't start with the old mid-level path %s",
							originalChildPath,
							originalMidLevelPath,
						)
					}

					pathSuffix := strings.TrimPrefix(originalChildPath, originalMidLevelPath)
					expectedChildPath := updatedMidLevel.Path + pathSuffix

					if updatedChild.Path != expectedChildPath {
						return fmt.Errorf("child path not updated correctly: got %s, want %s",
							updatedChild.Path, expectedChildPath)
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
				return "" // Empty means success
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
