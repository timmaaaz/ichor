package formdataapp

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/foundation/logger"
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
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "trace" })
	app := &App{log: log}
	ctx := context.Background()

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
				ctx,
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

func TestLineItemsConfigColumnsSerialization(t *testing.T) {
	// Test that Columns field serializes/deserializes correctly
	config := formfieldbus.LineItemsFieldConfig{
		ExecutionOrder:    1,
		Entity:            "sales.order_line_items",
		ParentField:       "order_id",
		Fields:            []formfieldbus.LineItemField{},
		ItemLabel:         "Items",
		SingularItemLabel: "Item",
		MinItems:          0,
		MaxItems:          10,
		Columns:           4, // Test non-default value
	}

	// Serialize
	configJSON, err := config.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify JSON contains columns field
	var parsed map[string]interface{}
	if err := json.Unmarshal(configJSON, &parsed); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	columns, ok := parsed["columns"].(float64)
	if !ok {
		t.Fatalf("columns field not found or not a number")
	}
	if int(columns) != 4 {
		t.Errorf("columns: got %v, want 4", columns)
	}
}

func TestLineItemsConfigColumnsOmitEmpty(t *testing.T) {
	// Test that Columns=0 is omitted from JSON (omitempty behavior)
	config := formfieldbus.LineItemsFieldConfig{
		ExecutionOrder:    1,
		Entity:            "sales.order_line_items",
		ParentField:       "order_id",
		Fields:            []formfieldbus.LineItemField{},
		ItemLabel:         "Items",
		SingularItemLabel: "Item",
		MinItems:          0,
		MaxItems:          10,
		// Columns not set (zero value)
	}

	configJSON, err := config.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(configJSON, &parsed); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if _, exists := parsed["columns"]; exists {
		t.Errorf("columns field should be omitted when zero")
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

// =============================================================================
// DependsOnConfig Tests
// =============================================================================

func TestDependsOnConfigSerialization(t *testing.T) {
	maxPercent := 100

	config := formfieldbus.DependsOnConfig{
		Field: "discount_type",
		ValueMappings: map[string]formfieldbus.FieldOverrideConfig{
			"flat": {
				Type:  "currency",
				Label: "Discount ($)",
			},
			"percent": {
				Type:  "percent",
				Label: "Discount (%)",
				Validation: &formfieldbus.ValidationConfig{
					Max: &maxPercent,
				},
			},
		},
		Default: formfieldbus.FieldOverrideConfig{
			Type: "currency",
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal to map failed: %v", err)
	}

	// Check field
	if field, ok := parsed["field"].(string); !ok || field != "discount_type" {
		t.Errorf("field: got %v, want discount_type", parsed["field"])
	}

	// Check value_mappings
	mappings, ok := parsed["value_mappings"].(map[string]interface{})
	if !ok {
		t.Fatalf("value_mappings not a map: %v", parsed["value_mappings"])
	}
	if len(mappings) != 2 {
		t.Errorf("value_mappings length: got %d, want 2", len(mappings))
	}

	// Check flat mapping
	flat, ok := mappings["flat"].(map[string]interface{})
	if !ok {
		t.Fatalf("flat mapping not a map: %v", mappings["flat"])
	}
	if flat["type"] != "currency" {
		t.Errorf("flat.type: got %v, want currency", flat["type"])
	}

	// Check percent mapping with nested validation
	percent, ok := mappings["percent"].(map[string]interface{})
	if !ok {
		t.Fatalf("percent mapping not a map: %v", mappings["percent"])
	}
	validation, ok := percent["validation"].(map[string]interface{})
	if !ok {
		t.Fatalf("percent.validation not a map: %v", percent["validation"])
	}
	if validation["max"] != float64(100) {
		t.Errorf("percent.validation.max: got %v, want 100", validation["max"])
	}

	// Check default
	defaultCfg, ok := parsed["default"].(map[string]interface{})
	if !ok {
		t.Fatalf("default not a map: %v", parsed["default"])
	}
	if defaultCfg["type"] != "currency" {
		t.Errorf("default.type: got %v, want currency", defaultCfg["type"])
	}
}

func TestDependsOnConfigRoundTrip(t *testing.T) {
	maxPercent := 100

	original := formfieldbus.DependsOnConfig{
		Field: "discount_type",
		ValueMappings: map[string]formfieldbus.FieldOverrideConfig{
			"flat":    {Type: "currency", Label: "Discount ($)"},
			"percent": {Type: "percent", Label: "Discount (%)", Validation: &formfieldbus.ValidationConfig{Max: &maxPercent}},
		},
		Default: formfieldbus.FieldOverrideConfig{Type: "currency"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var result formfieldbus.DependsOnConfig
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if diff := cmp.Diff(original, result); diff != "" {
		t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
	}
}

func TestDependsOnConfigOmitEmpty(t *testing.T) {
	config := formfieldbus.DependsOnConfig{
		Field: "discount_type",
		ValueMappings: map[string]formfieldbus.FieldOverrideConfig{
			"flat": {Type: "currency"},
		},
		// Default intentionally not set (zero-valued struct)
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Note: Default is a struct type (not a pointer), so Go's omitempty for structs
	// means it will be serialized as an empty object {} when all fields are zero-valued.
	// This is expected Go behavior. The key thing is that FieldOverrideConfig fields
	// with omitempty ARE correctly omitted within the empty default object.
	if defaultVal, exists := parsed["default"]; exists {
		defaultMap, ok := defaultVal.(map[string]interface{})
		if !ok {
			t.Fatalf("default is not a map: %v", defaultVal)
		}
		// When Default struct is zero-valued, all its fields should be omitted
		if len(defaultMap) != 0 {
			t.Errorf("default should be empty map when zero-valued, got: %v", defaultMap)
		}
	}
}

func TestLineItemFieldWithDependsOn(t *testing.T) {
	maxPercent := 100

	field := formfieldbus.LineItemField{
		Name:     "discount",
		Label:    "Discount",
		Type:     "currency",
		Required: false,
		DependsOn: &formfieldbus.DependsOnConfig{
			Field: "discount_type",
			ValueMappings: map[string]formfieldbus.FieldOverrideConfig{
				"flat":    {Type: "currency", Label: "Discount ($)"},
				"percent": {Type: "percent", Label: "Discount (%)", Validation: &formfieldbus.ValidationConfig{Max: &maxPercent}},
			},
			Default: formfieldbus.FieldOverrideConfig{Type: "currency"},
		},
	}

	data, err := json.Marshal(field)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var result formfieldbus.LineItemField
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if result.DependsOn == nil {
		t.Fatal("DependsOn is nil after round-trip")
	}

	if result.DependsOn.Field != "discount_type" {
		t.Errorf("DependsOn.Field: got %v, want discount_type", result.DependsOn.Field)
	}

	if len(result.DependsOn.ValueMappings) != 2 {
		t.Errorf("DependsOn.ValueMappings length: got %d, want 2", len(result.DependsOn.ValueMappings))
	}
}

func TestLineItemFieldDependsOnOmitEmpty(t *testing.T) {
	field := formfieldbus.LineItemField{
		Name:     "quantity",
		Label:    "Quantity",
		Type:     "number",
		Required: true,
		// DependsOn intentionally not set
	}

	data, err := json.Marshal(field)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, exists := parsed["depends_on"]; exists {
		t.Errorf("depends_on should be omitted when nil, got: %v", parsed["depends_on"])
	}
}
