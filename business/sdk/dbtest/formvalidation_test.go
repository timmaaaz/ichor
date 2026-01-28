package dbtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
)

// TestFormConfigsAgainstSchema validates all registered form configurations
// against the actual database schema. This catches:
// - Typos in entity/table names
// - References to non-existent columns
// - Invalid dropdown/enum references
func TestFormConfigsAgainstSchema(t *testing.T) {
	// Spin up test database with migrations + seed
	db := NewDatabase(t, "form_schema_validation")
	ctx := context.Background()

	// Dummy IDs for form generation
	dummyFormID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	dummyEntityID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	// Track overall results
	var totalErrors, totalWarnings int
	var failedForms []string

	t.Logf("\n=== Validating %d Registered Forms Against Schema ===\n", len(seedmodels.FormRegistry))

	for _, entry := range seedmodels.FormRegistry {
		t.Run(entry.Name, func(t *testing.T) {
			// Generate form fields
			fields := entry.Generator(dummyFormID, dummyEntityID)

			opts := formfieldbus.FormValidationOptions{
				SupportsUpdate: entry.SupportsUpdate,
				FormName:       entry.Name,
			}

			// 1. Run structural validation (Part 2)
			structuralResult := formfieldbus.ValidateFormFields(fields, opts)

			// 2. Run deep schema validation (Part 3)
			deepResult := formfieldbus.ValidateFormFieldsAgainstSchema(
				ctx,
				db.BusDomain.Introspection,
				fields,
				opts,
			)

			// Track if this form had errors
			hasErrors := false

			// Report structural errors
			for _, err := range structuralResult.Errors {
				t.Errorf("[STRUCTURAL] %s: %s (%s)", err.Field, err.Message, err.Code)
				totalErrors++
				hasErrors = true
			}

			// Report deep validation errors
			for _, err := range deepResult.Errors {
				t.Errorf("[SCHEMA] %s: %s (%s)", err.Field, err.Message, err.Code)
				totalErrors++
				hasErrors = true
			}

			// Report warnings (non-fatal)
			for _, warn := range structuralResult.Warnings {
				t.Logf("[WARNING] %s: %s", warn.Field, warn.Message)
				totalWarnings++
			}
			for _, warn := range deepResult.Warnings {
				t.Logf("[WARNING] %s: %s", warn.Field, warn.Message)
				totalWarnings++
			}

			if hasErrors {
				failedForms = append(failedForms, entry.Name)
			}
		})
	}

	// Print summary
	t.Logf("\n=== Form Validation Summary ===")
	t.Logf("Forms validated: %d", len(seedmodels.FormRegistry))
	t.Logf("Total errors: %d", totalErrors)
	t.Logf("Total warnings: %d", totalWarnings)

	if len(failedForms) > 0 {
		t.Logf("\nFailed forms:")
		for _, name := range failedForms {
			t.Logf("  - %s", name)
		}
	}
}

// TestDeepValidationCatchesTypos verifies that deep validation catches intentional errors.
// This test creates form fields with known errors and verifies they are detected.
func TestDeepValidationCatchesTypos(t *testing.T) {
	// Spin up test database with migrations + seed
	db := NewDatabase(t, "form_typo_detection")
	ctx := context.Background()

	dummyFormID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	dummyEntityID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	testCases := []struct {
		name          string
		fields        []formfieldbus.NewFormField
		expectedCode  string
		expectedField string
	}{
		{
			name: "non-existent table",
			fields: []formfieldbus.NewFormField{
				{
					FormID:       dummyFormID,
					EntityID:     dummyEntityID,
					EntitySchema: "sales",
					EntityTable:  "orderz", // typo: should be "orders"
					Name:         "id",
					Label:        "ID",
					FieldType:    "hidden",
					Config:       []byte(`{"hidden": true}`),
				},
			},
			expectedCode:  "TABLE_NOT_FOUND",
			expectedField: "fields[0]",
		},
		{
			name: "non-existent column",
			fields: []formfieldbus.NewFormField{
				{
					FormID:       dummyFormID,
					EntityID:     dummyEntityID,
					EntitySchema: "sales",
					EntityTable:  "orders",
					Name:         "customer_idd", // typo: should be "customer_id"
					Label:        "Customer",
					FieldType:    "text",
					Config:       []byte(`{}`),
				},
			},
			expectedCode:  "COLUMN_NOT_FOUND",
			expectedField: "fields[0]",
		},
		{
			name: "dropdown with non-existent entity",
			fields: []formfieldbus.NewFormField{
				{
					FormID:       dummyFormID,
					EntityID:     dummyEntityID,
					EntitySchema: "sales",
					EntityTable:  "orders",
					Name:         "customer_id",
					Label:        "Customer",
					FieldType:    "smart-combobox",
					Config:       []byte(`{"entity": "sales.customerz", "label_column": "name", "value_column": "id"}`), // typo
				},
			},
			expectedCode:  "DROPDOWN_TABLE_NOT_FOUND",
			expectedField: "fields[0].config.entity",
		},
		{
			name: "dropdown with non-existent column",
			fields: []formfieldbus.NewFormField{
				{
					FormID:       dummyFormID,
					EntityID:     dummyEntityID,
					EntitySchema: "sales",
					EntityTable:  "orders",
					Name:         "customer_id",
					Label:        "Customer",
					FieldType:    "smart-combobox",
					Config:       []byte(`{"entity": "sales.customers", "label_column": "namee", "value_column": "id"}`), // typo in label_column
				},
			},
			expectedCode:  "DROPDOWN_COLUMN_NOT_FOUND",
			expectedField: "fields[0].config.label_column",
		},
		{
			name: "enum with non-existent type",
			fields: []formfieldbus.NewFormField{
				{
					FormID:       dummyFormID,
					EntityID:     dummyEntityID,
					EntitySchema: "sales",
					EntityTable:  "orders",
					Name:         "status",
					Label:        "Status",
					FieldType:    "enum",
					Config:       []byte(`{"enum_name": "sales.order_statuz"}`), // typo
				},
			},
			expectedCode:  "ENUM_NOT_FOUND",
			expectedField: "fields[0].config.enum_name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := formfieldbus.FormValidationOptions{
				SupportsUpdate: false,
				FormName:       tc.name,
			}

			result := formfieldbus.ValidateFormFieldsAgainstSchema(
				ctx,
				db.BusDomain.Introspection,
				tc.fields,
				opts,
			)

			// Verify the expected error was caught
			found := false
			for _, err := range result.Errors {
				if err.Code == tc.expectedCode && err.Field == tc.expectedField {
					found = true
					t.Logf("Correctly caught error: %s at %s", err.Code, err.Field)
					break
				}
			}

			if !found {
				t.Errorf("Expected error with code %q at field %q was not found", tc.expectedCode, tc.expectedField)
				t.Logf("Actual errors found:")
				for _, err := range result.Errors {
					t.Logf("  - %s: %s (%s)", err.Field, err.Message, err.Code)
				}
			}
		})
	}
}
