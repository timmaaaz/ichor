package formbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

// TestNewForms generates n NewForm structs for testing.
func TestNewForms(n int) []NewForm {
	newForms := make([]NewForm, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nf := NewForm{
			Name: fmt.Sprintf("Form%d", idx),
		}

		newForms[i] = nf
	}

	return newForms
}

// TestSeedForms seeds the database with n forms and returns them sorted by name.
func TestSeedForms(ctx context.Context, n int, api *Business) ([]Form, error) {
	newForms := TestNewForms(n)

	forms := make([]Form, len(newForms))

	for i, nf := range newForms {
		form, err := api.Create(ctx, nf)
		if err != nil {
			return nil, fmt.Errorf("seeding form: idx: %d : %w", i, err)
		}

		forms[i] = form
	}

	sort.Slice(forms, func(i, j int) bool {
		return forms[i].Name <= forms[j].Name
	})

	return forms, nil
}