package restrictedcolumnbus

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/permissions/testing"
)

// TestSeedRestrictedColumns is a helper method for testing.
func TestSeedRestrictedColumns(ctx context.Context, api *Business) ([]RestrictedColumn, error) {
	restrictedColumns := make([]RestrictedColumn, len(testing.RestrictedColumns))

	for i, rcMap := range testing.RestrictedColumns {
		nrc, err := testing.MapToStruct[NewRestrictedColumn](rcMap)
		if err != nil {
			return nil, fmt.Errorf("mapping to struct at idx %d: %w", i, err)
		}

		rc, err := api.Create(ctx, nrc)
		if err != nil {
			return nil, fmt.Errorf("seeding restricted column: idx: %d, : %w", i, err)
		}

		restrictedColumns[i] = rc
	}

	return restrictedColumns, nil
}
