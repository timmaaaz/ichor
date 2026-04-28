package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
	"github.com/timmaaaz/ichor/business/domain/hr/reportstobus"
	"github.com/timmaaaz/ichor/business/domain/hr/titlebus"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// FoundationSeed holds the results of seeding foundational users and currencies.
type FoundationSeed struct {
	Admins        []userbus.User
	Reporters     []userbus.User
	Bosses        []userbus.User
	USDCurrencyID uuid.UUID
	Currencies    []currencybus.Currency
}

func seedFoundation(ctx context.Context, busDomain BusDomain) (FoundationSeed, error) {
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("seeding user : %w", err)
	}

	// Extra users for hierarchy
	reporters, err := userbus.TestSeedUsersWithNoFKs(ctx, 20, userbus.Roles.User, busDomain.User)
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("seeding reporter : %w", err)
	}

	reporterIDs := make([]uuid.UUID, len(reporters))
	for i, r := range reporters {
		reporterIDs[i] = r.ID
	}

	// Assign warehouse zones to 2 reporters for zone-sliced multi-worker pick
	// coverage. reporters[0] gets STG-A/STG-B; reporters[1] gets STG-C/PCK.
	// Other reporters keep the empty default. Failures here surface loudly:
	// a missing reporters[0]/[1] panics rather than silently skipping.
	zonesA := []string{"STG-A", "STG-B"}
	zonesB := []string{"STG-C", "PCK"}
	updated0, err := busDomain.User.Update(ctx, reporters[0], userbus.UpdateUser{AssignedZones: &zonesA})
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("assign zones to reporter[0]: %w", err)
	}
	reporters[0] = updated0
	updated1, err := busDomain.User.Update(ctx, reporters[1], userbus.UpdateUser{AssignedZones: &zonesB})
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("assign zones to reporter[1]: %w", err)
	}
	reporters[1] = updated1

	bosses, err := userbus.TestSeedUsersWithNoFKs(ctx, 10, userbus.Roles.User, busDomain.User)
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("seeding reporter : %w", err)
	}

	bossIDs := make([]uuid.UUID, len(bosses))
	for i, b := range bosses {
		bossIDs[i] = b.ID
	}

	_, err = reportstobus.TestSeedReportsTo(ctx, 30, reporterIDs, bossIDs, busDomain.ReportsTo)
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("seeding reportsto : %w", err)
	}

	_, err = commentbus.TestSeedUserApprovalCommentHistorical(ctx, 10, 90, reporterIDs[:5], reporterIDs[5:], busDomain.UserApprovalComment)
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("seeding approval comments : %w", err)
	}

	_, err = titlebus.TestSeedTitles(ctx, 10, busDomain.Title)
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("seeding fulfillment statues : %w", err)
	}

	// Query for USD currency (seeded in seed.sql) for product costs.
	// All product prices are stored in USD - conversion happens at display time.
	usdCode := "USD"
	usdCurrencies, err := busDomain.Currency.Query(ctx, currencybus.QueryFilter{Code: &usdCode}, currencybus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("querying USD currency: %w", err)
	}
	if len(usdCurrencies) == 0 {
		return FoundationSeed{}, fmt.Errorf("USD currency not found - ensure seed.sql has run")
	}
	usdCurrencyID := usdCurrencies[0].ID

	// Seed test currencies for orders (variety in demo data).
	currencies, err := currencybus.TestSeedCurrencies(ctx, 5, busDomain.Currency)
	if err != nil {
		return FoundationSeed{}, fmt.Errorf("seeding currencies: %w", err)
	}

	return FoundationSeed{
		Admins:        admins,
		Reporters:     reporters,
		Bosses:        bosses,
		USDCurrencyID: usdCurrencyID,
		Currencies:    currencies,
	}, nil
}
