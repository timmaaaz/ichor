package reference_test

import (
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/api/domain/http/workflow/referenceapi"
)

// Test_ActionTypeSchemas_MatchExpectedStructure verifies that all action type schemas
// contain required fields and match the expected structure.
// This test prevents backend schema drift that could break the frontend UI.
func Test_ActionTypeSchemas_MatchExpectedStructure(t *testing.T) {
	t.Parallel()

	actionTypes := referenceapi.GetActionTypes()

	if len(actionTypes) == 0 {
		t.Fatal("Expected action types to be returned, got empty slice")
	}

	// Verify we have all 6 expected action types in alphabetical order
	expectedTypes := []string{
		"allocate_inventory",
		"create_alert",
		"seek_approval",
		"send_email",
		"send_notification",
		"update_field",
	}

	if len(actionTypes) != len(expectedTypes) {
		t.Fatalf("Expected %d action types, got %d", len(expectedTypes), len(actionTypes))
	}

	for i, actionType := range actionTypes {
		// Verify alphabetical order
		if actionType.Type != expectedTypes[i] {
			t.Errorf("Expected action type %q at index %d, got %q", expectedTypes[i], i, actionType.Type)
		}

		// Verify all required fields are present
		if actionType.Type == "" {
			t.Errorf("Action type at index %d has empty Type field", i)
		}
		if actionType.Name == "" {
			t.Errorf("Action type %q has empty Name field", actionType.Type)
		}
		if actionType.Description == "" {
			t.Errorf("Action type %q has empty Description field", actionType.Type)
		}
		if actionType.Category == "" {
			t.Errorf("Action type %q has empty Category field", actionType.Type)
		}

		// Verify ConfigSchema is valid JSON
		if len(actionType.ConfigSchema) == 0 {
			t.Errorf("Action type %q has empty ConfigSchema", actionType.Type)
			continue
		}

		var schema map[string]interface{}
		if err := json.Unmarshal(actionType.ConfigSchema, &schema); err != nil {
			t.Errorf("Action type %q has invalid JSON schema: %v", actionType.Type, err)
			continue
		}

		t.Logf("✓ Action type %q schema validated (%d bytes, category: %s)",
			actionType.Type, len(actionType.ConfigSchema), actionType.Category)
	}
}

// Test_ActionTypeSchemas_ValidJSON verifies all embedded schemas are valid JSON
// that can be parsed without errors.
func Test_ActionTypeSchemas_ValidJSON(t *testing.T) {
	t.Parallel()

	actionTypes := referenceapi.GetActionTypes()

	for _, actionType := range actionTypes {
		t.Run(actionType.Type, func(t *testing.T) {
			var schema interface{}
			if err := json.Unmarshal(actionType.ConfigSchema, &schema); err != nil {
				t.Fatalf("Failed to parse schema for %q: %v", actionType.Type, err)
			}

			// Verify we can re-marshal it (round-trip test)
			reMarshaled, err := json.Marshal(schema)
			if err != nil {
				t.Fatalf("Failed to re-marshal schema for %q: %v", actionType.Type, err)
			}

			if len(reMarshaled) == 0 {
				t.Errorf("Re-marshaled schema for %q is empty", actionType.Type)
			}
		})
	}
}

// Test_ActionTypeSchemas_CategoryConsistency verifies action types are
// grouped into expected categories.
func Test_ActionTypeSchemas_CategoryConsistency(t *testing.T) {
	t.Parallel()

	actionTypes := referenceapi.GetActionTypes()

	expectedCategories := map[string][]string{
		"communication": {"create_alert", "send_email", "send_notification"},
		"inventory":     {"allocate_inventory"},
		"approval":      {"seek_approval"},
		"data":          {"update_field"},
	}

	for category, expectedTypes := range expectedCategories {
		var foundTypes []string
		for _, actionType := range actionTypes {
			if actionType.Category == category {
				foundTypes = append(foundTypes, actionType.Type)
			}
		}

		if len(foundTypes) != len(expectedTypes) {
			t.Errorf("Category %q: expected %d types, got %d", category, len(expectedTypes), len(foundTypes))
		}

		for _, expectedType := range expectedTypes {
			found := false
			for _, foundType := range foundTypes {
				if foundType == expectedType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Category %q missing expected type %q", category, expectedType)
			}
		}
	}
}

// Test_ActionTypeSchemas_AsyncFlags verifies async flags are set correctly.
func Test_ActionTypeSchemas_AsyncFlags(t *testing.T) {
	t.Parallel()

	actionTypes := referenceapi.GetActionTypes()

	// Map of action types to their expected async status
	expectedAsync := map[string]bool{
		"allocate_inventory": true,  // Async - inventory operations can be slow
		"create_alert":       false, // Sync - quick database insert
		"seek_approval":      true,  // Async - requires external approval
		"send_email":         true,  // Async - external email service
		"send_notification":  false, // Sync - in-app notification
		"update_field":       false, // Sync - direct database update
	}

	for _, actionType := range actionTypes {
		expected, ok := expectedAsync[actionType.Type]
		if !ok {
			t.Errorf("No async expectation defined for action type %q", actionType.Type)
			continue
		}

		if actionType.IsAsync != expected {
			t.Errorf("Action type %q: expected IsAsync=%v, got %v",
				actionType.Type, expected, actionType.IsAsync)
		}
	}
}
