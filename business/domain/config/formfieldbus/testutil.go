package formfieldbus

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// TestNewFormFields generates n NewFormField structs for testing.
func TestNewFormFields(n int, formIDs []uuid.UUID, entities []workflow.Entity) []NewFormField {
	newFormFields := make([]NewFormField, n)

	idx := rand.Intn(10000)
	fieldTypes := []string{"text", "number", "email", "textarea", "select", "checkbox", "date"}

	for i := 0; i < n; i++ {
		idx++

		// Create a sample config
		config := map[string]interface{}{
			"placeholder": fmt.Sprintf("Enter field %d", idx),
			"maxLength":   100,
		}
		configJSON, _ := json.Marshal(config)

		nff := NewFormField{
			FormID:     formIDs[rand.Intn(len(formIDs))],
			EntityID:   entities[rand.Intn(len(entities))].ID,
			Name:       fmt.Sprintf("field_%d", idx),
			Label:      fmt.Sprintf("Field %d", idx),
			FieldType:  fieldTypes[rand.Intn(len(fieldTypes))],
			FieldOrder: i,
			Required:   rand.Intn(2) == 0, // Random true/false
			Config:     json.RawMessage(configJSON),
		}

		newFormFields[i] = nff
	}

	return newFormFields
}

// TestSeedFormFields seeds the database with n form fields and returns them sorted by name.
func TestSeedFormFields(ctx context.Context, n int, formIDs []uuid.UUID, api *Business, entities []workflow.Entity) ([]FormField, error) {
	newFormFields := TestNewFormFields(n, formIDs, entities)

	formFields := make([]FormField, len(newFormFields))

	for i, nff := range newFormFields {
		formField, err := api.Create(ctx, nff)
		if err != nil {
			return nil, fmt.Errorf("seeding form field: idx: %d : %w", i, err)
		}

		formFields[i] = formField
	}

	sort.Slice(formFields, func(i, j int) bool {
		return formFields[i].Name <= formFields[j].Name
	})

	return formFields, nil
}
