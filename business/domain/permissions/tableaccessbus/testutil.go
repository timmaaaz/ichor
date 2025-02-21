package tableaccessbus

import (
	"context"

	"github.com/google/uuid"
)

// TestNewTableAccesses creates a slice of NewTableAccess for testing purposes.
func TestNewTableAccesses(n int, roleID uuid.UUID, tables []string) []NewTableAccess {
	newTableAccesses := make([]NewTableAccess, n)

	for i := 0; i < n; i++ {
		nta := NewTableAccess{
			RoleID:    roleID,
			TableName: tables[i],
			CanCreate: true,
			CanRead:   true,
			CanUpdate: true,
			CanDelete: true,
		}

		newTableAccesses[i] = nta
	}

	return newTableAccesses
}

// TestSeedTableAccesses is a helper method for testing.
func TestSeedTableAccesses(ctx context.Context, n int, roleID uuid.UUID, tables []string, api *Business) ([]TableAccess, error) {
	newTableAccesses := TestNewTableAccesses(n, roleID, tables)
	tableAccesses := make([]TableAccess, n)

	for i, nta := range newTableAccesses {
		ta, err := api.Create(ctx, nta)
		if err != nil {
			return nil, err
		}

		tableAccesses[i] = ta
	}

	return tableAccesses, nil
}
