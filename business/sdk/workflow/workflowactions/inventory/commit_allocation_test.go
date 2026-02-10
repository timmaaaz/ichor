package inventory_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func Test_CommitAllocation(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_CommitAllocation")

	sd, err := insertCommitAllocationSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.Handler = inventory.NewCommitAllocationHandler(log, db.DB, db.BusDomain.InventoryItem)

	unitest.Run(t, commitAllocationTests(db.BusDomain, db.DB, sd), "commitAllocation")
}

// =============================================================================

type commitAllocationSeedData struct {
	unitest.SeedData
	Handler            *inventory.CommitAllocationHandler
	Products           []productbus.Product
	InventoryItems     []inventoryitembus.InventoryItem
	InventoryLocations []inventorylocationbus.InventoryLocation
	ExecutionContext   workflow.ActionExecutionContext
}

func insertCommitAllocationSeedData(busDomain dbtest.BusDomain) (commitAllocationSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	adminIDs := make([]uuid.UUID, len(admins))
	for i, admin := range admins {
		adminIDs[i] = admin.ID
	}

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding brand : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding product category : %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 5, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding product : %w", err)
	}

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, busDomain.Zones)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 10, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationIDs[i] = il.LocationID
	}

	inventoryItems, err := seedTestInventoryItems(ctx, productIDs, inventoryLocationIDs, *busDomain.InventoryItem)
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("seeding inventory items : %w", err)
	}

	// Pre-reserve 20 units on the first item for commit tests.
	reservedQty := 20
	updatedItem, err := busDomain.InventoryItem.Update(ctx, inventoryItems[0], inventoryitembus.UpdateInventoryItem{
		ReservedQuantity: &reservedQty,
	})
	if err != nil {
		return commitAllocationSeedData{}, fmt.Errorf("pre-reserving inventory : %w", err)
	}
	inventoryItems[0] = updatedItem

	ruleID := uuid.New()
	execContext := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "orders",
		EventType:     "on_update",
		UserID:        adminIDs[0],
		RuleID:        &ruleID,
		RuleName:      "Test Commit Allocation Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return commitAllocationSeedData{
		SeedData: unitest.SeedData{
			Admins:             []unitest.User{{User: admins[0]}},
			Products:           products,
			InventoryLocations: inventoryLocations,
			InventoryItems:     inventoryItems,
		},
		Products:           products,
		InventoryItems:     inventoryItems,
		InventoryLocations: inventoryLocations,
		ExecutionContext:   execContext,
	}, nil
}

// =============================================================================

func commitAllocationTests(busDomain dbtest.BusDomain, db *sqlx.DB, sd commitAllocationSeedData) []unitest.Table {
	return []unitest.Table{
		validateCommitAllocationConfig(sd),
		executeBasicCommit(busDomain, sd),
		executeOverCommit(sd),
	}
}

func validateCommitAllocationConfig(sd commitAllocationSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_config",
		ExpResp: "product_id is required",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_id": "` + uuid.New().String() + `", "quantity": 10}`)
			err := sd.Handler.Validate(config)
			if err != nil {
				return err.Error()
			}
			return nil
		},
		CmpFunc: func(got any, exp any) string {
			gotStr := fmt.Sprintf("%v", got)
			expStr := fmt.Sprintf("%v", exp)
			if gotStr != expStr {
				return fmt.Sprintf("got %v, want %v", gotStr, expStr)
			}
			return ""
		},
	}
}

func executeBasicCommit(busDomain dbtest.BusDomain, sd commitAllocationSeedData) unitest.Table {
	if len(sd.Products) == 0 || len(sd.InventoryItems) == 0 {
		return unitest.Table{
			Name:    "execute_basic_commit_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name: "execute_basic_commit",
		ExpResp: map[string]int{
			"reserved":  10, // 20 - 10 = 10
			"allocated": 10, // 0 + 10 = 10
		},
		ExcFunc: func(ctx context.Context) any {
			item := sd.InventoryItems[0]

			config := inventory.CommitAllocationConfig{
				ProductID:  item.ProductID.String(),
				LocationID: item.LocationID.String(),
				Quantity:   10,
			}
			configJSON, _ := json.Marshal(config)

			_, err := sd.Handler.Execute(ctx, configJSON, sd.ExecutionContext)
			if err != nil {
				return err
			}

			// Verify DB state.
			updatedItem, err := busDomain.InventoryItem.QueryByID(ctx, item.ID)
			if err != nil {
				return err
			}

			return map[string]int{
				"reserved":  updatedItem.ReservedQuantity,
				"allocated": updatedItem.AllocatedQuantity,
			}
		},
		CmpFunc: func(got any, exp any) string {
			gotMap, ok := got.(map[string]int)
			if !ok {
				return fmt.Sprintf("expected map[string]int, got %T", got)
			}
			expMap := exp.(map[string]int)

			if gotMap["reserved"] != expMap["reserved"] {
				return fmt.Sprintf("reserved: got %d, want %d", gotMap["reserved"], expMap["reserved"])
			}
			if gotMap["allocated"] != expMap["allocated"] {
				return fmt.Sprintf("allocated: got %d, want %d", gotMap["allocated"], expMap["allocated"])
			}
			return ""
		},
	}
}

func executeOverCommit(sd commitAllocationSeedData) unitest.Table {
	if len(sd.Products) < 2 || len(sd.InventoryItems) < 2 {
		return unitest.Table{
			Name:    "execute_over_commit_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "execute_over_commit",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			// Second item has reserved_quantity=0, try to commit 50 -> should error.
			item := sd.InventoryItems[1]

			config := inventory.CommitAllocationConfig{
				ProductID:  item.ProductID.String(),
				LocationID: item.LocationID.String(),
				Quantity:   50,
			}
			configJSON, _ := json.Marshal(config)

			_, err := sd.Handler.Execute(ctx, configJSON, sd.ExecutionContext)
			if err == nil {
				return fmt.Errorf("expected error but got nil")
			}

			// Should contain "insufficient reserved" in the error.
			return true
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}
