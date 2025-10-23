package formdataregistry_test

import (
	"testing"

	"github.com/timmaaaz/ichor/app/domain/assets/assetapp"
	"github.com/timmaaaz/ichor/app/domain/core/userapp"
	"github.com/timmaaaz/ichor/app/sdk/formdataregistry"
)

func TestGetRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		model    interface{}
		expected []string
	}{
		{
			name:  "AssetNewModel",
			model: assetapp.NewAsset{},
			expected: []string{
				"valid_asset_id",
				"asset_condition_id",
				"serial_number",
			},
		},
		{
			name:  "UserNewModel",
			model: userapp.NewUser{},
			expected: []string{
				"username",
				"first_name",
				"last_name",
				"email",
				"birthday",
				"roles",
				"system_roles",
				"password",
				"password_confirm",
			},
		},
		{
			name:     "AssetUpdateModel",
			model:    assetapp.UpdateAsset{},
			expected: []string{}, // UpdateAsset has no required fields (all are omitempty)
		},
		{
			name:     "NilModel",
			model:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formdataregistry.GetRequiredFields(tt.model)

			// Check if the result contains all expected fields
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d required fields, got %d. Expected: %v, Got: %v",
					len(tt.expected), len(result), tt.expected, result)
				return
			}

			// Convert result to map for easier checking
			resultMap := make(map[string]bool)
			for _, field := range result {
				resultMap[field] = true
			}

			// Check each expected field is present
			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("Expected field %s not found in result: %v", expected, result)
				}
			}
		})
	}
}

func TestGetRequiredFields_EdgeCases(t *testing.T) {
	t.Run("PointerToStruct", func(t *testing.T) {
		model := &assetapp.NewAsset{}
		result := formdataregistry.GetRequiredFields(model)

		if len(result) != 3 {
			t.Errorf("Expected 3 required fields for pointer to NewAsset, got %d: %v", len(result), result)
		}
	})

	t.Run("NonStructType", func(t *testing.T) {
		model := "not a struct"
		result := formdataregistry.GetRequiredFields(model)

		if result != nil {
			t.Errorf("Expected nil for non-struct type, got %v", result)
		}
	})

	t.Run("EmptyStruct", func(t *testing.T) {
		type EmptyStruct struct{}
		model := EmptyStruct{}
		result := formdataregistry.GetRequiredFields(model)

		if len(result) != 0 {
			t.Errorf("Expected 0 required fields for empty struct, got %d: %v", len(result), result)
		}
	})
}