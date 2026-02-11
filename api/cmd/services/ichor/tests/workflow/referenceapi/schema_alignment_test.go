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

	actionTypes := referenceapi.GetActionTypes(nil)

	if len(actionTypes) == 0 {
		t.Fatal("Expected action types to be returned, got empty slice")
	}

	// Verify we have all 17 expected action types in alphabetical order
	expectedTypes := []string{
		"allocate_inventory",
		"check_inventory",
		"check_reorder_point",
		"commit_allocation",
		"create_alert",
		"create_entity",
		"delay",
		"evaluate_condition",
		"log_audit_entry",
		"lookup_entity",
		"release_reservation",
		"reserve_inventory",
		"seek_approval",
		"send_email",
		"send_notification",
		"transition_status",
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

		// Verify output ports are present
		if len(actionType.OutputPorts) == 0 {
			t.Errorf("Action type %q has no output ports", actionType.Type)
			continue
		}

		// Verify at least one port is marked as default
		hasDefault := false
		for _, port := range actionType.OutputPorts {
			if port.IsDefault {
				hasDefault = true
				break
			}
		}
		if !hasDefault {
			t.Errorf("Action type %q has no default output port", actionType.Type)
		}

		t.Logf("Action type %q schema validated (%d bytes, category: %s, ports: %d)",
			actionType.Type, len(actionType.ConfigSchema), actionType.Category, len(actionType.OutputPorts))
	}
}

// Test_ActionTypeSchemas_ValidJSON verifies all embedded schemas are valid JSON
// that can be parsed without errors.
func Test_ActionTypeSchemas_ValidJSON(t *testing.T) {
	t.Parallel()

	actionTypes := referenceapi.GetActionTypes(nil)

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

	actionTypes := referenceapi.GetActionTypes(nil)

	expectedCategories := map[string][]string{
		"communication": {"create_alert", "send_email", "send_notification"},
		"inventory":     {"allocate_inventory", "check_inventory", "check_reorder_point", "commit_allocation", "release_reservation", "reserve_inventory"},
		"approval":      {"seek_approval"},
		"data":          {"create_entity", "log_audit_entry", "lookup_entity", "transition_status", "update_field"},
		"control":       {"delay", "evaluate_condition"},
	}

	for category, expectedTypes := range expectedCategories {
		var foundTypes []string
		for _, actionType := range actionTypes {
			if actionType.Category == category {
				foundTypes = append(foundTypes, actionType.Type)
			}
		}

		if len(foundTypes) != len(expectedTypes) {
			t.Errorf("Category %q: expected %d types, got %d (found: %v)", category, len(expectedTypes), len(foundTypes), foundTypes)
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

	actionTypes := referenceapi.GetActionTypes(nil)

	// Map of action types to their expected async status
	expectedAsync := map[string]bool{
		"allocate_inventory": true,
		"check_inventory":    false,
		"check_reorder_point": false,
		"commit_allocation":  false,
		"create_alert":       false,
		"create_entity":      false,
		"delay":              false,
		"evaluate_condition": false,
		"log_audit_entry":    false,
		"lookup_entity":      false,
		"release_reservation": false,
		"reserve_inventory":  false,
		"seek_approval":      true,
		"send_email":         true,
		"send_notification":  false,
		"transition_status":  false,
		"update_field":       false,
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
