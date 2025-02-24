package restrictedcolumnbus

import (
	"context"
	"fmt"
)

// TestNewRestrictedColumns is a helper method for testing.
func TestNewRestrictedColumns(n int) []NewRestrictedColumn {
	newRestrictedColumns := make([]NewRestrictedColumn, n)

	for i := 0; i < n; i++ {
		nrc := NewRestrictedColumn{
			TableName:  fmt.Sprintf("TableName%d", i),
			ColumnName: fmt.Sprintf("ColumnName%d", i),
		}

		newRestrictedColumns[i] = nrc
	}

	return newRestrictedColumns
}

// TestSeedRestrictedColumns is a helper method for testing.
func TestSeedRestrictedColumns(ctx context.Context, n int, api *Business) ([]RestrictedColumn, error) {
	newRestrictedColumns := TestNewRestrictedColumns(n)
	restrictedColumns := make([]RestrictedColumn, n)

	for i, nrc := range newRestrictedColumns {
		rc, err := api.Create(ctx, nrc)
		if err != nil {
			return nil, fmt.Errorf("seeding restricted column: idx: %d, : %w", i, err)
		}

		restrictedColumns[i] = rc
	}

	return restrictedColumns, nil
}
