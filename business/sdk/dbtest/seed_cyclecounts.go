package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
)

func seedCycleCounts(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, products ProductsSeed, inventory InventorySeed) error {
	adminIDs := make(uuid.UUIDs, len(foundation.Admins))
	for i, a := range foundation.Admins {
		adminIDs[i] = a.ID
	}

	productIDs := make(uuid.UUIDs, len(products.Products))
	for i, p := range products.Products {
		productIDs[i] = p.ProductID
	}

	locationIDs := make(uuid.UUIDs, len(inventory.InventoryLocations))
	for i, loc := range inventory.InventoryLocations {
		locationIDs[i] = loc.LocationID
	}

	// Seed 3 cycle count sessions
	sessions, err := cyclecountsessionbus.TestSeedCycleCountSessions(ctx, 3, adminIDs, busDomain.CycleCountSession)
	if err != nil {
		return fmt.Errorf("seeding cycle count sessions: %w", err)
	}

	sessionIDs := make(uuid.UUIDs, len(sessions))
	for i, s := range sessions {
		sessionIDs[i] = s.ID
	}

	// Seed cycle count items (5 per session = 15 total)
	_, err = cyclecountitembus.TestSeedCycleCountItems(ctx, 15, sessionIDs, productIDs, locationIDs, busDomain.CycleCountItem)
	if err != nil {
		return fmt.Errorf("seeding cycle count items: %w", err)
	}

	return nil
}
