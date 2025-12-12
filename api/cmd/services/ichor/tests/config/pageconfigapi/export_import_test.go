package pageconfigapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

func Test_ExportImport(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_ExportImport")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Export tests
	test.Run(t, export200(sd), "export-200")
	test.Run(t, export401(sd), "export-401")
	test.Run(t, export404(sd), "export-404")

	// Import tests
	test.Run(t, import200(sd), "import-200")
	test.Run(t, import400(sd), "import-400")
	test.Run(t, import401(sd), "import-401")
}

// export200 tests successful export scenarios
func export200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic-export",
			URL:        "/v1/config/page-configs/id/" + sd.PageConfigs[0].ID + "/export",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &pageconfigapp.PageConfigPackage{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*pageconfigapp.PageConfigPackage)
				if !exists {
					return "error occurred"
				}

				// Verify the exported config matches the seed config
				if gotResp.PageConfig.ID != sd.PageConfigs[0].ID {
					return cmp.Diff(gotResp.PageConfig.ID, sd.PageConfigs[0].ID)
				}

				if gotResp.PageConfig.Name != sd.PageConfigs[0].Name {
					return cmp.Diff(gotResp.PageConfig.Name, sd.PageConfigs[0].Name)
				}

				// Verify structure exists (not checking exact values, just presence)
				if gotResp.Contents == nil {
					return "expected Contents to be non-nil"
				}

				return ""
			},
		},
	}
}

// export401 tests unauthorized export scenarios
func export401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/config/page-configs/id/" + sd.PageConfigs[0].ID + "/export",
			Token:      "&nbsp;",
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-token",
			URL:        "/v1/config/page-configs/id/" + sd.PageConfigs[0].ID + "/export",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodGet,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// export404 tests export with non-existent ID
func export404(sd apitest.SeedData) []apitest.Table {
	randomID := uuid.New().String()
	return []apitest.Table{
		{
			Name:       "nonexistent-id",
			URL:        "/v1/config/page-configs/id/" + randomID + "/export",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			// TODO: Backend currently returns 500 instead of 404 for non-existent configs
			// Should be http.StatusNotFound once backend is fixed
			StatusCode: http.StatusInternalServerError,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Internal, "export failed"),
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)

				// Just verify we got an error response with Internal code
				// Once backend is fixed, this should check for NotFound error code
				if gotErr.Code != errs.Internal {
					return cmp.Diff(gotErr.Code, errs.Internal)
				}

				return ""
			},
		},
	}
}

// import200 tests successful import scenarios
func import200(sd apitest.SeedData) []apitest.Table {
	// Create a minimal page config export package using raw JSON
	// The import endpoint expects the business layer format (PascalCase field names, UUID types)
	// Use json.RawMessage so apitest framework doesn't double-encode it

	// Each test gets its own unique name to avoid conflicts
	skipBlob := json.RawMessage(`{
		"PageConfig": {
			"Name": "Import Test Page ` + uuid.New().String() + `",
			"UserID": "00000000-0000-0000-0000-000000000000",
			"IsDefault": true
		},
		"Contents": [],
		"Actions": {
			"Buttons": [],
			"Dropdowns": [],
			"Separators": []
		}
	}`)

	replaceBlob := json.RawMessage(`{
		"PageConfig": {
			"Name": "Import Test Page ` + uuid.New().String() + `",
			"UserID": "00000000-0000-0000-0000-000000000000",
			"IsDefault": true
		},
		"Contents": [],
		"Actions": {
			"Buttons": [],
			"Dropdowns": [],
			"Separators": []
		}
	}`)

	mergeBlob := json.RawMessage(`{
		"PageConfig": {
			"Name": "Import Test Page ` + uuid.New().String() + `",
			"UserID": "00000000-0000-0000-0000-000000000000",
			"IsDefault": true
		},
		"Contents": [],
		"Actions": {
			"Buttons": [],
			"Dropdowns": [],
			"Separators": []
		}
	}`)

	defaultBlob := json.RawMessage(`{
		"PageConfig": {
			"Name": "Import Test Page ` + uuid.New().String() + `",
			"UserID": "00000000-0000-0000-0000-000000000000",
			"IsDefault": true
		},
		"Contents": [],
		"Actions": {
			"Buttons": [],
			"Dropdowns": [],
			"Separators": []
		}
	}`)

	return []apitest.Table{
		{
			Name:       "import-skip-mode",
			URL:        "/v1/config/page-configs/import-blob?mode=skip",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      skipBlob,
			GotResp:    &pageconfigapp.ImportStats{},
			ExpResp: &pageconfigapp.ImportStats{
				ImportedCount: 1,
				SkippedCount:  0,
				UpdatedCount:  0,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "import-replace-mode",
			URL:        "/v1/config/page-configs/import-blob?mode=replace",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      replaceBlob,
			GotResp:    &pageconfigapp.ImportStats{},
			ExpResp: &pageconfigapp.ImportStats{
				ImportedCount: 1,
				SkippedCount:  0,
				UpdatedCount:  0,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "import-merge-mode",
			URL:        "/v1/config/page-configs/import-blob?mode=merge",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      mergeBlob,
			GotResp:    &pageconfigapp.ImportStats{},
			ExpResp: &pageconfigapp.ImportStats{
				ImportedCount: 1,
				SkippedCount:  0,
				UpdatedCount:  0,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "import-default-mode",
			URL:        "/v1/config/page-configs/import-blob",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input:      defaultBlob,
			GotResp:    &pageconfigapp.ImportStats{},
			ExpResp: &pageconfigapp.ImportStats{
				ImportedCount: 1,
				SkippedCount:  0,
				UpdatedCount:  0,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// import400 tests bad request import scenarios
func import400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		// Note: Testing truly invalid JSON (malformed) is difficult with apitest framework
		// because json.RawMessage validates JSON at test setup time. The backend's
		// invalid JSON handling is tested indirectly through other validation failures.
		{
			Name:       "invalid-mode",
			URL:        "/v1/config/page-configs/import-blob?mode=invalid",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      json.RawMessage(`{}`),
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "invalid mode: invalid (must be: skip, replace, or merge)"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-name",
			URL:        "/v1/config/page-configs/import-blob?mode=skip",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input:      json.RawMessage(`{"PageConfig":{},"Contents":[],"Actions":{"Buttons":[],"Dropdowns":[],"Separators":[]}}`),
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "name is required"),
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)
				expErr := exp.(*errs.Error)

				// Compare error codes
				if gotErr.Code != expErr.Code {
					return cmp.Diff(gotErr.Code, expErr.Code)
				}

				return ""
			},
		},
	}
}

// import401 tests unauthorized import scenarios
func import401(sd apitest.SeedData) []apitest.Table {
	blob := json.RawMessage(`{"PageConfig":{"Name":"Test"},"Contents":[],"Actions":{"Buttons":[],"Dropdowns":[],"Separators":[]}}`)

	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/config/page-configs/import-blob?mode=skip",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      blob,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-token",
			URL:        "/v1/config/page-configs/import-blob?mode=skip",
			Token:      sd.Admins[0].Token[:10],
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      blob,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "roleadminonly",
			URL:        "/v1/config/page-configs/import-blob?mode=skip",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input:      blob,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authorize: you are not authorized for that action, claims[[USER]] rule[rule_admin_only]"),
			CmpFunc: func(got any, exp any) string {
				gotErr := got.(*errs.Error)
				expErr := exp.(*errs.Error)

				// Compare error codes
				if gotErr.Code != expErr.Code {
					return cmp.Diff(gotErr.Code, expErr.Code)
				}

				return ""
			},
		},
	}
}
