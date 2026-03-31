package dbtest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
)

// TasksSeed holds the results of seeding task data for downstream consumers.
type TasksSeed struct {
	PutAwayTasks []putawaytaskbus.PutAwayTask
	PickTasks    []picktaskbus.PickTask
}

// seedTasks seeds put-away and pick tasks for the frontend demo database.
func seedTasks(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, products ProductsSeed, inventory InventorySeed, sales SalesSeed) (TasksSeed, error) {
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
	putAwayTasks, err := putawaytaskbus.TestSeedPutAwayTasks(ctx, 15, productIDs, locationIDs, adminIDs, busDomain.PutAwayTask)
	if err != nil {
		return TasksSeed{}, fmt.Errorf("seeding put-away tasks: %w", err)
	}

	// Seed 15 pick tasks for frontend
	pickTasks, err := picktaskbus.TestSeedPickTasks(ctx, 15, sales.OrderIDs, sales.OrderLineItemIDs, productIDs, locationIDs, adminIDs, busDomain.PickTask)
	if err != nil {
		return TasksSeed{}, fmt.Errorf("seeding pick tasks: %w", err)
	}

	return TasksSeed{
		PutAwayTasks: putAwayTasks,
		PickTasks:    pickTasks,
	}, nil
}
