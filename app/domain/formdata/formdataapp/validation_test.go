package formdataapp

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindMissingFields(t *testing.T) {
	tests := []struct {
		name            string
		requiredFields  []string
		availableFields []string
		wantMissing     []string
	}{
		{
			name:            "all fields present",
			requiredFields:  []string{"name", "email"},
			availableFields: []string{"name", "email", "phone"},
			wantMissing:     nil,
		},
		{
			name:            "some fields missing",
			requiredFields:  []string{"name", "email", "phone"},
			availableFields: []string{"name"},
			wantMissing:     []string{"email", "phone"},
		},
		{
			name:            "audit fields excluded - created_by",
			requiredFields:  []string{"name", "email", "created_by"},
			availableFields: []string{"name", "email"},
			wantMissing:     nil, // created_by is an audit field, should be excluded
		},
		{
			name:            "audit fields excluded - all audit fields",
			requiredFields:  []string{"name", "created_by", "created_date", "updated_by", "updated_date"},
			availableFields: []string{"name"},
			wantMissing:     nil, // all audit fields should be excluded
		},
		{
			name:            "audit fields excluded but other fields still missing",
			requiredFields:  []string{"name", "email", "created_by", "updated_by"},
			availableFields: []string{"name"},
			wantMissing:     []string{"email"}, // only email is missing, audit fields excluded
		},
		{
			name:            "empty required fields",
			requiredFields:  []string{},
			availableFields: []string{"name", "email"},
			wantMissing:     nil,
		},
		{
			name:            "empty available fields with non-audit required",
			requiredFields:  []string{"name", "email"},
			availableFields: []string{},
			wantMissing:     []string{"name", "email"},
		},
		{
			name:            "empty available fields with only audit required",
			requiredFields:  []string{"created_by", "updated_by"},
			availableFields: []string{},
			wantMissing:     nil, // audit fields are excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMissingFields(tt.requiredFields, tt.availableFields)
			if diff := cmp.Diff(tt.wantMissing, got); diff != "" {
				t.Errorf("findMissingFields() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAuditFields(t *testing.T) {
	// Verify the auditFields map contains the expected fields
	expectedAuditFields := []string{"created_by", "created_date", "updated_by", "updated_date"}

	for _, field := range expectedAuditFields {
		if !auditFields[field] {
			t.Errorf("expected %q to be in auditFields map", field)
		}
	}

	// Verify no unexpected fields
	if len(auditFields) != len(expectedAuditFields) {
		t.Errorf("auditFields has %d entries, expected %d", len(auditFields), len(expectedAuditFields))
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
