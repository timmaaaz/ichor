package restrictedcolumnbus

import (
	"context"
	"fmt"
)

// TestNewRestrictedColumns is a helper method for testing.
func TestNewRestrictedColumns() []NewRestrictedColumn {
	newRestrictedColumns := make([]NewRestrictedColumn, 3)

	newRestrictedColumns[0] = NewRestrictedColumn{
		TableName:  "valid_assets",
		ColumnName: "est_price",
	}
	newRestrictedColumns[1] = NewRestrictedColumn{
		TableName:  "valid_assets",
		ColumnName: "name",
	}
	newRestrictedColumns[2] = NewRestrictedColumn{
		TableName:  "valid_assets",
		ColumnName: "price",
	}

	return newRestrictedColumns
}

// TestSeedRestrictedColumns is a helper method for testing.
func TestSeedRestrictedColumns(ctx context.Context, api *Business) ([]RestrictedColumn, error) {
	newRestrictedColumns := TestNewRestrictedColumns()
	restrictedColumns := make([]RestrictedColumn, 3)

	for i, nrc := range newRestrictedColumns {
		rc, err := api.Create(ctx, nrc)
		if err != nil {
			return nil, fmt.Errorf("seeding restricted column: idx: %d, : %w", i, err)
		}

		restrictedColumns[i] = rc
	}

	return restrictedColumns, nil
}
