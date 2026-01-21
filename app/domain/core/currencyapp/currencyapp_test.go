package currencyapp_test

import (
	"context"
	"testing"

	"github.com/timmaaaz/ichor/app/domain/core/currencyapp"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func Test_QueryByCode(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CurrencyApp_QueryByCode")

	ctx := context.Background()

	// Seed currencies with known codes (codes will be "TS0", "TS1", "TS2")
	currencies, err := currencybus.TestSeedCurrencies(ctx, 3, db.BusDomain.Currency)
	if err != nil {
		t.Fatalf("seeding currencies: %s", err)
	}

	app := currencyapp.NewApp(db.BusDomain.Currency)

	// Test: QueryByCode - success case
	t.Run("queryByCode-success", func(t *testing.T) {
		expectedCurrency := currencies[0]

		id, err := app.QueryByCode(ctx, expectedCurrency.Code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if id != expectedCurrency.ID {
			t.Errorf("expected ID %v, got %v", expectedCurrency.ID, id)
		}
	})

	// Test: QueryByCode - success with different currency
	t.Run("queryByCode-success-second", func(t *testing.T) {
		expectedCurrency := currencies[1]

		id, err := app.QueryByCode(ctx, expectedCurrency.Code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if id != expectedCurrency.ID {
			t.Errorf("expected ID %v, got %v", expectedCurrency.ID, id)
		}
	})

	// Test: QueryByCode - not found case
	t.Run("queryByCode-notFound", func(t *testing.T) {
		_, err := app.QueryByCode(ctx, "NONEXISTENT")
		if err == nil {
			t.Error("expected error for non-existent code, got nil")
		}
	})
}
