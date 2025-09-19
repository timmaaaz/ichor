package rolebus

import (
	"context"
	"fmt"
)

// TestRole represents a role specified for the test.
func TestSeedRoles(ctx context.Context, n int, api *Business) ([]Role, error) {

	roles := make([]Role, n)
	for i := 0; i < n; i++ {
		role, err := api.Create(ctx, NewRole{
			Name:        fmt.Sprintf("Role%d", i),
			Description: fmt.Sprintf("Role%d Description", i),
		})
		if err != nil {
			return nil, fmt.Errorf("creating role : %w", err)
		}

		roles[i] = role
	}

	return roles, nil
}
