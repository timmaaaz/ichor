package data_test

import (
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
)

func TestIsValidColumnName(t *testing.T) {
	// Regex: ^[a-z][a-z0-9_]{0,62}$
	// Must start with lowercase letter, then 0-62 lowercase letters/digits/underscores.
	// Total max: 63 chars (PostgreSQL NAMEDATALEN limit).
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"valid simple", "status", true},
		{"valid with underscore", "order_number", true},
		{"valid with digits", "field2", true},
		{"valid single char", "a", true},
		{"valid 63 chars", "a" + repeat("b", 62), true},
		{"empty string", "", false},
		{"starts with digit", "1field", false},
		{"starts with underscore", "_field", false},
		{"uppercase", "Status", false},
		{"mixed case", "orderNumber", false},
		{"sql injection semicolon", "status; DROP TABLE users", false},
		{"sql injection comment", "status -- comment", false},
		{"contains space", "order number", false},
		{"contains dot", "table.column", false},
		{"unicode", "naïve", false},
		{"special chars", "col@name", false},
		{"too long 64 chars", "a" + repeat("b", 63), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := data.IsValidColumnName(tt.input)
			if got != tt.expect {
				t.Errorf("IsValidColumnName(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestIsValidTableName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		// Known valid entries from the whitelist
		{"valid sales.orders", "sales.orders", true},
		{"valid core.users", "core.users", true},
		{"valid inventory.inventory_items", "inventory.inventory_items", true},
		{"valid workflow.automation_rules", "workflow.automation_rules", true},
		{"valid procurement.purchase_orders", "procurement.purchase_orders", true},
		{"valid products.products", "products.products", true},
		{"valid assets.assets", "assets.assets", true},
		{"valid hr.offices", "hr.offices", true},
		{"valid geography.countries", "geography.countries", true},
		{"valid config.table_configs", "config.table_configs", true},
		// Unknown table names
		{"unknown table", "sales.nonexistent", false},
		{"unknown schema", "fake.users", false},
		{"empty string", "", false},
		{"no schema", "users", false},
		{"sql fragment", "sales.orders; DROP TABLE users", false},
		{"partial match", "sales.order", false},
		{"extra whitespace", " sales.orders ", false},
		{"uppercase", "Sales.Orders", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := data.IsValidTableName(tt.input)
			if got != tt.expect {
				t.Errorf("IsValidTableName(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestIsValidOperator(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		// All valid operators
		{"equals", "equals", true},
		{"not_equals", "not_equals", true},
		{"greater_than", "greater_than", true},
		{"less_than", "less_than", true},
		{"contains", "contains", true},
		{"is_null", "is_null", true},
		{"is_not_null", "is_not_null", true},
		{"in", "in", true},
		{"not_in", "not_in", true},
		// Invalid operators
		{"empty string", "", false},
		{"sql equals", "=", false},
		{"sql like", "LIKE", false},
		{"unknown", "between", false},
		{"uppercase valid", "EQUALS", false},
		{"space padding", " equals ", false},
		{"sql fragment", "1=1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := data.IsValidOperator(tt.input)
			if got != tt.expect {
				t.Errorf("IsValidOperator(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

// repeat returns a string of n copies of s.
func repeat(s string, n int) string {
	result := make([]byte, n*len(s))
	for i := 0; i < n; i++ {
		copy(result[i*len(s):], s)
	}
	return string(result)
}
