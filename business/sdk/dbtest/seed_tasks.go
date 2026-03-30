package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
)

func seedTasks(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, products ProductsSeed, inventory InventorySeed, sales SalesSeed) error {
	productIDs := make(uuid.UUIDs, len(products.Products))
	for i, p := range products.Products {
		productIDs[i] = p.ProductID
	}

	locationIDs := make(uuid.UUIDs, len(inventory.InventoryLocations))
	for i, loc := range inventory.InventoryLocations {
		locationIDs[i] = loc.LocationID
	}

	adminIDs := make(uuid.UUIDs, len(foundation.Admins))
	for i, a := range foundation.Admins {
		adminIDs[i] = a.ID
	}

	// Seed 15 put-away tasks for frontend
	_, err := putawaytaskbus.TestSeedPutAwayTasks(ctx, 15, productIDs, locationIDs, adminIDs, busDomain.PutAwayTask)
	if err != nil {
		return fmt.Errorf("seeding put-away tasks: %w", err)
	}

	// Seed 15 pick tasks for frontend
	_, err = picktaskbus.TestSeedPickTasks(ctx, 15, sales.OrderIDs, sales.OrderLineItemIDs, productIDs, locationIDs, adminIDs, busDomain.PickTask)
	if err != nil {
		return fmt.Errorf("seeding pick tasks: %w", err)
	}

	return nil
}
