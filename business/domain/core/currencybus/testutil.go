package currencybus

import (
	"context"
	"fmt"
)

// TestSeedCurrencies creates test currencies for testing purposes.
func TestSeedCurrencies(ctx context.Context, n int, api *Business) ([]Currency, error) {
	currencies := make([]Currency, n)

	for i := 0; i < n; i++ {
		currency, err := api.Create(ctx, NewCurrency{
			Code:          fmt.Sprintf("TS%d", i),
			Name:          fmt.Sprintf("Test Currency %d", i),
			Symbol:        fmt.Sprintf("$%d", i),
			Locale:        "en-US",
			DecimalPlaces: 2,
			IsActive:      true,
			SortOrder:     100 + i,
		})
		if err != nil {
			return nil, fmt.Errorf("creating currency: %w", err)
		}

		currencies[i] = currency
	}

	return currencies, nil
}
