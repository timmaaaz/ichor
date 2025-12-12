package pageconfigbus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ExportImport(t *testing.T) {
	t.Parallel()

	// Create isolated test database
	db := dbtest.NewDatabase(t, "Test_ExportImport")

	// Seed test data
	sd, err := insertExportImportSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Run table-driven tests
	// Run export tests first before any imports modify the data
	unitest.Run(t, exportByIDs(db.BusDomain, sd), "export-by-ids")
	unitest.Run(t, exportMultipleIDs(db.BusDomain, sd), "export-multiple")

	// Run import tests (these may modify/delete seed data)
	unitest.Run(t, importSkipMode(db.BusDomain, sd), "import-skip")
	unitest.Run(t, importReplaceMode(db.BusDomain, sd), "import-replace")
	unitest.Run(t, importMergeMode(db.BusDomain, sd), "import-merge")
	unitest.Run(t, importWithNestedContent(db.BusDomain, sd), "import-nested")
}

// insertExportImportSeedData creates test page configs for export/import testing.
func insertExportImportSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	// Seed page configs using existing test utilities
	configs, err := pageconfigbus.TestSeedPageConfigs(ctx, 5, busDomain.PageConfig)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding page configs : %w", err)
	}

	return unitest.SeedData{
		PageConfigs: configs,
	}, nil
}

// exportByIDs tests ExportByIDs() with a single ID.
func exportByIDs(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "single-id",
			ExcFunc: func(ctx context.Context) any {
				ids := []uuid.UUID{sd.PageConfigs[0].ID}
				result, err := busDomain.PageConfig.ExportByIDs(ctx, ids)
				if err != nil {
					return err
				}
				return result
			},
			CmpFunc: func(got any, exp any) string {
				gotConfigs := got.([]pageconfigbus.PageConfigWithRelations)

				// Verify we got exactly one config
				if len(gotConfigs) != 1 {
					return fmt.Sprintf("expected 1 config, got %d", len(gotConfigs))
				}

				// Verify the config ID matches
				if gotConfigs[0].PageConfig.ID != sd.PageConfigs[0].ID {
					return fmt.Sprintf("expected config ID %s, got %s", sd.PageConfigs[0].ID, gotConfigs[0].PageConfig.ID)
				}

				// Verify the config name matches
				if gotConfigs[0].PageConfig.Name != sd.PageConfigs[0].Name {
					return fmt.Sprintf("expected config name %s, got %s", sd.PageConfigs[0].Name, gotConfigs[0].PageConfig.Name)
				}

				return ""
			},
		},
		{
			Name: "nonexistent-id",
			ExcFunc: func(ctx context.Context) any {
				ids := []uuid.UUID{uuid.New()} // Random UUID that doesn't exist
				result, err := busDomain.PageConfig.ExportByIDs(ctx, ids)
				if err != nil {
					return err
				}
				return result
			},
			CmpFunc: func(got any, exp any) string {
				// Check if we got an error instead of results
				if _, isErr := got.(error); isErr {
					// This is expected - no rows found is acceptable
					return ""
				}

				gotConfigs := got.([]pageconfigbus.PageConfigWithRelations)

				// Verify we got zero configs
				if len(gotConfigs) != 0 {
					return fmt.Sprintf("expected 0 configs for nonexistent ID, got %d", len(gotConfigs))
				}

				return ""
			},
		},
	}

	return table
}

// exportMultipleIDs tests ExportByIDs() with multiple IDs.
func exportMultipleIDs(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "multiple-ids",
			ExcFunc: func(ctx context.Context) any {
				ids := []uuid.UUID{
					sd.PageConfigs[0].ID,
					sd.PageConfigs[1].ID,
					sd.PageConfigs[2].ID,
				}
				result, err := busDomain.PageConfig.ExportByIDs(ctx, ids)
				if err != nil {
					return err
				}
				return result
			},
			CmpFunc: func(got any, exp any) string {
				// Check if we got an error instead of results
				if err, isErr := got.(error); isErr {
					return fmt.Sprintf("unexpected error: %v", err)
				}

				gotConfigs := got.([]pageconfigbus.PageConfigWithRelations)

				// Verify we got exactly three configs
				if len(gotConfigs) != 3 {
					return fmt.Sprintf("expected 3 configs, got %d", len(gotConfigs))
				}

				// Verify the config IDs are in the result
				expectedIDs := map[uuid.UUID]bool{
					sd.PageConfigs[0].ID: false,
					sd.PageConfigs[1].ID: false,
					sd.PageConfigs[2].ID: false,
				}

				for _, config := range gotConfigs {
					if _, exists := expectedIDs[config.PageConfig.ID]; exists {
						expectedIDs[config.PageConfig.ID] = true
					}
				}

				for id, found := range expectedIDs {
					if !found {
						return fmt.Sprintf("expected config ID %s not found in results", id)
					}
				}

				return ""
			},
		},
	}

	return table
}

// importSkipMode tests ImportPageConfigs() with skip mode.
func importSkipMode(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "skip-existing",
			ExpResp: pageconfigbus.ImportStats{
				ImportedCount: 0,
				SkippedCount:  1,
				UpdatedCount:  0,
			},
			ExcFunc: func(ctx context.Context) any {
				// Create export package with existing page
				pkg := []pageconfigbus.PageConfigWithRelations{
					{
						PageConfig: sd.PageConfigs[0],
						Contents:   []pageconfigbus.PageContentExport{},
						Actions: pageconfigbus.PageActionsExport{
							Buttons:    []pageconfigbus.PageActionExport{},
							Dropdowns:  []pageconfigbus.PageActionExport{},
							Separators: []pageconfigbus.PageActionExport{},
						},
					},
				}

				// Import with skip mode
				stats, err := busDomain.PageConfig.ImportPageConfigs(ctx, pkg, "skip")
				if err != nil {
					return err
				}
				return stats
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name: "skip-new",
			ExpResp: pageconfigbus.ImportStats{
				ImportedCount: 1,
				SkippedCount:  0,
				UpdatedCount:  0,
			},
			ExcFunc: func(ctx context.Context) any {
				// Create export package with new page (new name)
				newConfig := sd.PageConfigs[0]
				newConfig.Name = "New Skip Test Page " + uuid.New().String()
				pkg := []pageconfigbus.PageConfigWithRelations{
					{
						PageConfig: newConfig,
						Contents:   []pageconfigbus.PageContentExport{},
						Actions: pageconfigbus.PageActionsExport{
							Buttons:    []pageconfigbus.PageActionExport{},
							Dropdowns:  []pageconfigbus.PageActionExport{},
							Separators: []pageconfigbus.PageActionExport{},
						},
					},
				}

				// Import with skip mode
				stats, err := busDomain.PageConfig.ImportPageConfigs(ctx, pkg, "skip")
				if err != nil {
					return err
				}
				return stats
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

// importReplaceMode tests ImportPageConfigs() with replace mode.
func importReplaceMode(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "replace-existing",
			ExpResp: pageconfigbus.ImportStats{
				ImportedCount: 0,
				SkippedCount:  0,
				UpdatedCount:  1,
			},
			ExcFunc: func(ctx context.Context) any {
				// Create export package with existing page (same name, is_default=true)
				modifiedConfig := sd.PageConfigs[1]
				modifiedConfig.IsDefault = true // Ensure it matches QueryByName logic
				pkg := []pageconfigbus.PageConfigWithRelations{
					{
						PageConfig: modifiedConfig,
						Contents:   []pageconfigbus.PageContentExport{},
						Actions: pageconfigbus.PageActionsExport{
							Buttons:    []pageconfigbus.PageActionExport{},
							Dropdowns:  []pageconfigbus.PageActionExport{},
							Separators: []pageconfigbus.PageActionExport{},
						},
					},
				}

				// Import with replace mode
				stats, err := busDomain.PageConfig.ImportPageConfigs(ctx, pkg, "replace")
				if err != nil {
					return err
				}
				return stats
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name: "replace-new",
			ExpResp: pageconfigbus.ImportStats{
				ImportedCount: 1,
				SkippedCount:  0,
				UpdatedCount:  0,
			},
			ExcFunc: func(ctx context.Context) any {
				// Create export package with new page
				newConfig := sd.PageConfigs[0]
				newConfig.Name = "New Replace Test Page " + uuid.New().String()
				pkg := []pageconfigbus.PageConfigWithRelations{
					{
						PageConfig: newConfig,
						Contents:   []pageconfigbus.PageContentExport{},
						Actions: pageconfigbus.PageActionsExport{
							Buttons:    []pageconfigbus.PageActionExport{},
							Dropdowns:  []pageconfigbus.PageActionExport{},
							Separators: []pageconfigbus.PageActionExport{},
						},
					},
				}

				// Import with replace mode
				stats, err := busDomain.PageConfig.ImportPageConfigs(ctx, pkg, "replace")
				if err != nil {
					return err
				}
				return stats
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

// importMergeMode tests ImportPageConfigs() with merge mode.
func importMergeMode(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "merge-existing",
			ExpResp: pageconfigbus.ImportStats{
				ImportedCount: 0,
				SkippedCount:  0,
				UpdatedCount:  1,
			},
			ExcFunc: func(ctx context.Context) any {
				// Create export package with existing page (same name, is_default=true)
				mergedConfig := sd.PageConfigs[2]
				mergedConfig.IsDefault = true // Ensure it matches QueryByName logic
				pkg := []pageconfigbus.PageConfigWithRelations{
					{
						PageConfig: mergedConfig,
						Contents:   []pageconfigbus.PageContentExport{},
						Actions: pageconfigbus.PageActionsExport{
							Buttons:    []pageconfigbus.PageActionExport{},
							Dropdowns:  []pageconfigbus.PageActionExport{},
							Separators: []pageconfigbus.PageActionExport{},
						},
					},
				}

				// Import with merge mode
				stats, err := busDomain.PageConfig.ImportPageConfigs(ctx, pkg, "merge")
				if err != nil {
					return err
				}
				return stats
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name: "merge-new",
			ExpResp: pageconfigbus.ImportStats{
				ImportedCount: 1,
				SkippedCount:  0,
				UpdatedCount:  0,
			},
			ExcFunc: func(ctx context.Context) any {
				// Create export package with new page
				newConfig := sd.PageConfigs[0]
				newConfig.Name = "New Merge Test Page " + uuid.New().String()
				pkg := []pageconfigbus.PageConfigWithRelations{
					{
						PageConfig: newConfig,
						Contents:   []pageconfigbus.PageContentExport{},
						Actions: pageconfigbus.PageActionsExport{
							Buttons:    []pageconfigbus.PageActionExport{},
							Dropdowns:  []pageconfigbus.PageActionExport{},
							Separators: []pageconfigbus.PageActionExport{},
						},
					},
				}

				// Import with merge mode
				stats, err := busDomain.PageConfig.ImportPageConfigs(ctx, pkg, "merge")
				if err != nil {
					return err
				}
				return stats
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

// importWithNestedContent tests ImportPageConfigs() with nested content structures.
func importWithNestedContent(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "import-with-content",
			ExpResp: pageconfigbus.ImportStats{
				ImportedCount: 1,
				SkippedCount:  0,
				UpdatedCount:  0,
			},
			ExcFunc: func(ctx context.Context) any {
				// Create a new page config with nested content
				newConfig := sd.PageConfigs[0]
				newConfig.Name = "Nested Content Test Page " + uuid.New().String()

				// Create parent content item
				parentContent := pageconfigbus.PageContentExport{
					ID:            uuid.New(),
					PageConfigID:  newConfig.ID,
					ContentType:   "container",
					Label:         "Parent Container",
					TableConfigID: uuid.Nil,
					FormID:        uuid.Nil,
					ChartConfigID: uuid.Nil,
					OrderIndex:    0,
					ParentID:      uuid.Nil,
					Layout:        json.RawMessage(`{"colSpan": {"default": 12}}`),
					IsVisible:     true,
					IsDefault:     false,
				}

				// Create child content item
				childContent := pageconfigbus.PageContentExport{
					ID:            uuid.New(),
					PageConfigID:  newConfig.ID,
					ContentType:   "text",
					Label:         "Child Text",
					TableConfigID: uuid.Nil,
					FormID:        uuid.Nil,
					ChartConfigID: uuid.Nil,
					OrderIndex:    0,
					ParentID:      parentContent.ID,
					Layout:        json.RawMessage(`{"colSpan": {"default": 12}}`),
					IsVisible:     true,
					IsDefault:     false,
				}

				pkg := []pageconfigbus.PageConfigWithRelations{
					{
						PageConfig: newConfig,
						Contents:   []pageconfigbus.PageContentExport{parentContent, childContent},
						Actions: pageconfigbus.PageActionsExport{
							Buttons:    []pageconfigbus.PageActionExport{},
							Dropdowns:  []pageconfigbus.PageActionExport{},
							Separators: []pageconfigbus.PageActionExport{},
						},
					},
				}

				// Import with skip mode (new page)
				stats, err := busDomain.PageConfig.ImportPageConfigs(ctx, pkg, "skip")
				if err != nil {
					return err
				}
				return stats
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
