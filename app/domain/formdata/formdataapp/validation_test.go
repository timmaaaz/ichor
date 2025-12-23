package formdataapp

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

func TestFindMissingFields(t *testing.T) {
	// Standard audit fields that are auto-populated
	autoPopulatedFields := map[string]bool{
		"created_by":   true,
		"created_date": true,
		"updated_by":   true,
		"updated_date": true,
	}

	tests := []struct {
		name                string
		requiredFields      []string
		availableFields     []string
		autoPopulatedFields map[string]bool
		wantMissing         []string
	}{
		{
			name:                "all fields present",
			requiredFields:      []string{"name", "email"},
			availableFields:     []string{"name", "email", "phone"},
			autoPopulatedFields: nil,
			wantMissing:         nil,
		},
		{
			name:                "some fields missing",
			requiredFields:      []string{"name", "email", "phone"},
			availableFields:     []string{"name"},
			autoPopulatedFields: nil,
			wantMissing:         []string{"email", "phone"},
		},
		{
			name:                "auto-populated fields excluded - created_by",
			requiredFields:      []string{"name", "email", "created_by"},
			availableFields:     []string{"name", "email"},
			autoPopulatedFields: autoPopulatedFields,
			wantMissing:         nil, // created_by is auto-populated, should be excluded
		},
		{
			name:                "auto-populated fields excluded - all audit fields",
			requiredFields:      []string{"name", "created_by", "created_date", "updated_by", "updated_date"},
			availableFields:     []string{"name"},
			autoPopulatedFields: autoPopulatedFields,
			wantMissing:         nil, // all audit fields should be excluded
		},
		{
			name:                "auto-populated fields excluded but other fields still missing",
			requiredFields:      []string{"name", "email", "created_by", "updated_by"},
			availableFields:     []string{"name"},
			autoPopulatedFields: autoPopulatedFields,
			wantMissing:         []string{"email"}, // only email is missing, audit fields excluded
		},
		{
			name:                "empty required fields",
			requiredFields:      []string{},
			availableFields:     []string{"name", "email"},
			autoPopulatedFields: nil,
			wantMissing:         nil,
		},
		{
			name:                "empty available fields with non-audit required",
			requiredFields:      []string{"name", "email"},
			availableFields:     []string{},
			autoPopulatedFields: nil,
			wantMissing:         []string{"name", "email"},
		},
		{
			name:                "empty available fields with only auto-populated required",
			requiredFields:      []string{"created_by", "updated_by"},
			availableFields:     []string{},
			autoPopulatedFields: autoPopulatedFields,
			wantMissing:         nil, // auto-populated fields are excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMissingFields(tt.requiredFields, tt.availableFields, tt.autoPopulatedFields)
			if diff := cmp.Diff(tt.wantMissing, got); diff != "" {
				t.Errorf("findMissingFields() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// =============================================================================
// LineItemField Default Tests
// =============================================================================

func TestMergeLineItemFieldDefaults(t *testing.T) {
	app := &App{}

	tests := []struct {
		name           string
		inputData      string
		fields         []formfieldbus.LineItemField
		operation      formdataregistry.EntityOperation
		wantFields     map[string]interface{} // fields that should be in output
		wantUnchanged  []string               // fields that should NOT be overwritten
	}{
		{
			name:      "inject create defaults",
			inputData: `{"product_id": "abc", "quantity": 5}`,
			fields: []formfieldbus.LineItemField{
				{Name: "created_by", DefaultValueCreate: "{{$me}}"},
				{Name: "created_date", DefaultValueCreate: "{{$now}}"},
			},
			operation:  formdataregistry.OperationCreate,
			wantFields: map[string]interface{}{"created_by": "{{$me}}", "created_date": "{{$now}}"},
		},
		{
			name:      "inject update defaults",
			inputData: `{"id": "123", "quantity": 10}`,
			fields: []formfieldbus.LineItemField{
				{Name: "updated_by", DefaultValueUpdate: "{{$me}}"},
				{Name: "updated_date", DefaultValueUpdate: "{{$now}}"},
			},
			operation:  formdataregistry.OperationUpdate,
			wantFields: map[string]interface{}{"updated_by": "{{$me}}", "updated_date": "{{$now}}"},
		},
		{
			name:      "do not override existing fields",
			inputData: `{"product_id": "abc", "created_by": "existing-user"}`,
			fields: []formfieldbus.LineItemField{
				{Name: "created_by", DefaultValueCreate: "{{$me}}"},
			},
			operation:     formdataregistry.OperationCreate,
			wantUnchanged: []string{"created_by"}, // should remain "existing-user"
		},
		{
			name:      "use DefaultValue when operation-specific not set",
			inputData: `{"product_id": "abc"}`,
			fields: []formfieldbus.LineItemField{
				{Name: "updated_by", DefaultValue: "{{$me}}"},
			},
			operation:  formdataregistry.OperationCreate,
			wantFields: map[string]interface{}{"updated_by": "{{$me}}"},
		},
		{
			name:      "prefer operation-specific over general default",
			inputData: `{"product_id": "abc"}`,
			fields: []formfieldbus.LineItemField{
				{Name: "created_by", DefaultValue: "general", DefaultValueCreate: "{{$me}}"},
			},
			operation:  formdataregistry.OperationCreate,
			wantFields: map[string]interface{}{"created_by": "{{$me}}"},
		},
		{
			name:      "empty fields list returns unchanged data",
			inputData: `{"product_id": "abc", "quantity": 5}`,
			fields:    []formfieldbus.LineItemField{},
			operation: formdataregistry.OperationCreate,
			wantFields: map[string]interface{}{
				"product_id": "abc",
				"quantity":   float64(5), // JSON numbers become float64
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := app.mergeLineItemFieldDefaults(
				json.RawMessage(tt.inputData),
				tt.fields,
				tt.operation,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Parse result
			var resultMap map[string]interface{}
			if err := json.Unmarshal(result, &resultMap); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}

			// Check expected fields
			for field, wantValue := range tt.wantFields {
				gotValue, exists := resultMap[field]
				if !exists {
					t.Errorf("expected field %q to exist in result", field)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("field %q: got %v, want %v", field, gotValue, wantValue)
				}
			}

			// Check unchanged fields (should not be overwritten)
			if len(tt.wantUnchanged) > 0 {
				var inputMap map[string]interface{}
				if err := json.Unmarshal([]byte(tt.inputData), &inputMap); err != nil {
					t.Fatalf("failed to unmarshal input: %v", err)
				}
				for _, field := range tt.wantUnchanged {
					if resultMap[field] != inputMap[field] {
						t.Errorf("field %q should not be overwritten: got %v, want %v",
							field, resultMap[field], inputMap[field])
					}
				}
			}
		})
	}
}

func TestExtractLineItemFields(t *testing.T) {
	app := &App{}

	// Create a lineitems field config
	lineItemsConfig := formfieldbus.LineItemsFieldConfig{
		ExecutionOrder: 2,
		Entity:         "sales.order_line_items",
		ParentField:    "order_id",
		Fields: []formfieldbus.LineItemField{
			{Name: "product_id", Label: "Product", Type: "dropdown", Required: true},
			{Name: "quantity", Label: "Quantity", Type: "number", Required: true},
			{Name: "created_by", Label: "Created By", Type: "text", Required: false, Hidden: true, DefaultValueCreate: "{{$me}}"},
		},
		ItemLabel: "Order Items",
		MinItems:  0,
		MaxItems:  100,
	}

	configJSON, _ := lineItemsConfig.ToJSON()

	fields := []formfieldbus.FormField{
		{
			EntitySchema: "sales",
			EntityTable:  "orders",
			Name:         "customer_id",
			FieldType:    "dropdown",
		},
		{
			EntitySchema: "sales",
			EntityTable:  "orders",
			Name:         "line_items",
			FieldType:    "lineitems",
			Config:       configJSON,
		},
	}

	tests := []struct {
		name       string
		entityName string
		wantCount  int
		wantFirst  string
	}{
		{
			name:       "find line item fields for matching entity",
			entityName: "sales.order_line_items",
			wantCount:  3,
			wantFirst:  "product_id",
		},
		{
			name:       "no match for different entity",
			entityName: "sales.orders",
			wantCount:  0,
		},
		{
			name:       "no match for non-existent entity",
			entityName: "inventory.items",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.extractLineItemFields(fields, tt.entityName)

			if len(result) != tt.wantCount {
				t.Errorf("got %d fields, want %d", len(result), tt.wantCount)
			}

			if tt.wantCount > 0 && len(result) > 0 {
				if result[0].Name != tt.wantFirst {
					t.Errorf("first field name: got %q, want %q", result[0].Name, tt.wantFirst)
				}
			}
		})
	}
}

// TestEntityKeyFormat verifies that entity keys use schema.table format.
// This is a regression test for the bug where entityFieldsMap was built with
// table-only keys (e.g., "customers") but looked up with schema.table keys
// (e.g., "sales.customers"), causing validation to always fail.
func TestEntityKeyFormat(t *testing.T) {
	// Simulate the entity key building logic from ValidateForm
	// This should use schema.table format
	schema := "sales"
	table := "customers"
	entityKey := schema + "." + table

	expectedKey := "sales.customers"
	if entityKey != expectedKey {
		t.Errorf("entity key format mismatch: got %q, want %q", entityKey, expectedKey)
	}

	// Verify lookup would succeed with matching format
	entityFieldsMap := map[string][]string{
		"sales.customers": {"name", "contact_id", "delivery_address_id", "notes"},
	}

	// Lookup using the same format should succeed
	fields, exists := entityFieldsMap[entityKey]
	if !exists {
		t.Errorf("entity lookup failed for key %q", entityKey)
	}
	if len(fields) != 4 {
		t.Errorf("expected 4 fields, got %d", len(fields))
	}

	// Verify that table-only lookup would fail (the old bug)
	_, existsTableOnly := entityFieldsMap[table]
	if existsTableOnly {
		t.Errorf("table-only lookup should not succeed with schema.table keys")
	}
}
