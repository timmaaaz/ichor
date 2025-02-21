package rolebus

import (
	"context"
	"fmt"
)

// TestNewRoles is a helper method for testing.
func TestNewRoles(n int) []NewRole {
	newRoles := make([]NewRole, n)

	for i := 0; i < n; i++ {
		nr := NewRole{
			Name:        fmt.Sprintf("Name%d", i),
			Description: fmt.Sprintf("Description%d", i),
		}

		newRoles[i] = nr
	}

	return newRoles
}

// TestSeedRoles is a helper method for testing.
func TestSeedRoles(ctx context.Context, n int, api *Business) ([]Role, error) {
	newRoles := TestNewRoles(n)
	roles := make([]Role, n)

	for i, nr := range newRoles {
		r, err := api.Create(ctx, nr)
		if err != nil {
			return nil, fmt.Errorf("seeding role: idx: %d, : %w", i, err)
		}

		roles[i] = r
	}

	return roles, nil
}
