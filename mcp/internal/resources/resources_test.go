package resources

import (
	"testing"
)

func TestParseDBResourceURI(t *testing.T) {
	tests := []struct {
		uri        string
		wantSchema string
		wantTable  string
		wantErr    bool
	}{
		{"config://db/core/users", "core", "users", false},
		{"config://db/inventory/locations", "inventory", "locations", false},
		{"config://db/", "", "", true},
		{"config://db/core/", "", "", true},
		{"config://db/core", "", "", true},
		{"config://other", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			schema, table, err := parseDBResourceURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDBResourceURI(%q) error = %v, wantErr %v", tt.uri, err, tt.wantErr)
			}
			if schema != tt.wantSchema {
				t.Errorf("schema = %q, want %q", schema, tt.wantSchema)
			}
			if table != tt.wantTable {
				t.Errorf("table = %q, want %q", table, tt.wantTable)
			}
		})
	}
}

func TestParseEnumResourceURI(t *testing.T) {
	tests := []struct {
		uri        string
		wantSchema string
		wantName   string
		wantErr    bool
	}{
		{"config://enums/core/role_type", "core", "role_type", false},
		{"config://enums/workflow/trigger_type", "workflow", "trigger_type", false},
		{"config://enums/", "", "", true},
		{"config://enums/core/", "", "", true},
		{"config://enums/core", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			schema, name, err := parseEnumResourceURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseEnumResourceURI(%q) error = %v, wantErr %v", tt.uri, err, tt.wantErr)
			}
			if schema != tt.wantSchema {
				t.Errorf("schema = %q, want %q", schema, tt.wantSchema)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
		})
	}
}
