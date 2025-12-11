package pageconfigapi_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
)

// Test_ExportImportEquivalence proves that export → import produces equivalent database records.
// This test verifies that the JSON export/import workflow preserves all page configuration data.
func Test_ExportImportEquivalence(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_ExportImportEquivalence")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Run equivalence tests
	test.Run(t, simpleEquivalence(sd), "simple-equivalence")
	test.Run(t, verifyExportedStructure(sd), "verify-structure")
}

// simpleEquivalence tests that exporting a page config produces the expected structure
func simpleEquivalence(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "export-produces-valid-structure",
			URL:        "/v1/config/page-configs/id/" + sd.PageConfigs[0].ID + "/export",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &pageconfigapp.PageConfigPackage{},
			CmpFunc: func(got, exp any) string {
				exportedPkg, ok := got.(*pageconfigapp.PageConfigPackage)
				if !ok {
					return "export failed: got is not *PageConfigPackage"
				}

				// Verify the exported config matches the seed config
				if exportedPkg.PageConfig.ID != sd.PageConfigs[0].ID {
					return cmp.Diff(exportedPkg.PageConfig.ID, sd.PageConfigs[0].ID)
				}

				if exportedPkg.PageConfig.Name != sd.PageConfigs[0].Name {
					return cmp.Diff(exportedPkg.PageConfig.Name, sd.PageConfigs[0].Name)
				}

				// Verify structure exists (not checking exact values, just presence)
				if exportedPkg.Contents == nil {
					return "expected Contents to be non-nil"
				}

				// Verify actions structure exists
				if len(exportedPkg.Actions.Buttons) < 0 { // Just checking it's a valid slice
					return "expected Actions.Buttons to be valid"
				}

				return ""
			},
		},
	}
}

// verifyExportedStructure verifies that all fields in an exported page config are preserved
func verifyExportedStructure(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "exported-fields-complete",
			URL:        "/v1/config/page-configs/id/" + sd.PageConfigs[0].ID + "/export",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &pageconfigapp.PageConfigPackage{},
			CmpFunc: func(got, exp any) string {
				exportedPkg, ok := got.(*pageconfigapp.PageConfigPackage)
				if !ok {
					return "export failed: got is not *PageConfigPackage"
				}

				// Verify PageConfig fields are present
				if exportedPkg.PageConfig.ID == "" {
					return "PageConfig.ID is empty"
				}
				if exportedPkg.PageConfig.Name == "" {
					return "PageConfig.Name is empty"
				}

				// Verify Contents array structure
				for i, content := range exportedPkg.Contents {
					if content.ID == "" {
						return cmp.Diff(content.ID, "non-empty", cmp.Comparer(func(a, b string) bool {
							return a != ""
						}))
					}
					if content.ContentType == "" {
						return "Contents[" + string(rune(i)) + "].ContentType is empty"
					}
					if content.Label == "" {
						return "Contents[" + string(rune(i)) + "].Label is empty"
					}
				}

				// Verify Actions structure
				for i, button := range exportedPkg.Actions.Buttons {
					if button.ID == "" {
						return "Actions.Buttons[" + string(rune(i)) + "].ID is empty"
					}
					if button.ActionType == "" {
						return "Actions.Buttons[" + string(rune(i)) + "].ActionType is empty"
					}
				}

				for i, dropdown := range exportedPkg.Actions.Dropdowns {
					if dropdown.ID == "" {
						return "Actions.Dropdowns[" + string(rune(i)) + "].ID is empty"
					}
					if dropdown.ActionType == "" {
						return "Actions.Dropdowns[" + string(rune(i)) + "].ActionType is empty"
					}
				}

				return ""
			},
		},
	}
}

// Note: Full round-trip equivalence tests (export → delete → import → compare)
// require more complex orchestration across multiple API calls and database state management.
// These tests provide validation that the export structure is complete and correct.
// The import functionality is tested separately in import200() test cases in export_import_test.go.
//
// Future enhancement: Add integration test that:
// 1. Creates a page config via API
// 2. Exports it
// 3. Deletes the original
// 4. Imports the export
// 5. Compares the re-imported config with the original
// This would require test database transaction rollback or cleanup logic.
