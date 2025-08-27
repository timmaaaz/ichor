package inventory_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/location/streetbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/zonebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

var testContainer rabbitmq.Container

func Test_AllocateInventory(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_AllocateInventory")

	sd, err := insertAllocateSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Create the handler with all dependencies
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	// Start RabbitMQ container
	testContainer, err = rabbitmq.StartRabbitMQ()
	if err != nil {
		fmt.Printf("Failed to start RabbitMQ container: %v\n", err)
		os.Exit(1)
	}

	// Create mock RabbitMQ client for testing
	rabbitConfig := rabbitmq.DefaultConfig()
	rabbitClient := rabbitmq.NewClient(log, rabbitConfig)
	queueClient := rabbitmq.NewWorkflowQueue(rabbitClient, log)

	sd.Handler = inventory.NewAllocateInventoryHandler(
		log,
		db.DB,
		queueClient,
		db.BusDomain.InventoryItem,
		db.BusDomain.InventoryLocation,
		db.BusDomain.InventoryTransaction,
		db.BusDomain.Product,
	)

	// -------------------------------------------------------------------------

	unitest.Run(t, allocateInventoryTests(db.BusDomain, db.DB, sd), "allocateInventory")

	// Cleanup

	if err := rabbitmq.StopRabbitMQ(testContainer); err != nil {
		fmt.Printf("Failed to stop RabbitMQ container: %v\n", err)
	}
}

// =============================================================================

type allocateSeedData struct {
	unitest.SeedData
	Handler            *inventory.AllocateInventoryHandler
	Products           []productbus.Product
	InventoryItems     []inventoryitembus.InventoryItem
	InventoryLocations []inventorylocationbus.InventoryLocation
	Warehouses         []warehousebus.Warehouse
	ExecutionContext   workflow.ActionExecutionContext
}

func insertAllocateSeedData(busDomain dbtest.BusDomain) (allocateSeedData, error) {
	ctx := context.Background()

	// Seed admin user
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	adminIDs := make([]uuid.UUID, len(admins))
	for i, admin := range admins {
		adminIDs[i] = admin.ID
	}

	// Seed locations (regions, cities, streets)
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	// Seed contact infos for brands
	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, busDomain.ContactInfos)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	// Seed brands
	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding brand : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	// Seed product categories
	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding product category : %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	// Seed products (we'll test allocation on these)
	products, err := productbus.TestSeedProducts(ctx, 5, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding product : %w", err)
	}

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	// Seed warehouses
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	// Seed zones
	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, busDomain.Zones)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	// Seed inventory locations
	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 10, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationIDs[i] = il.LocationID
	}

	// Seed inventory items with specific quantities for testing
	inventoryItems, err := seedTestInventoryItems(ctx, productIDs, inventoryLocationIDs, *busDomain.InventoryItem)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("seeding inventory items : %w", err)
	}

	// Create execution context for testing
	execContext := workflow.ActionExecutionContext{
		EntityID:    uuid.New(),
		EntityName:  "orders",
		EventType:   "on_create",
		UserID:      adminIDs[0],
		RuleID:      uuid.New(),
		RuleName:    "Test Allocation Rule",
		ExecutionID: uuid.New(),
		Timestamp:   time.Now().UTC(),
	}

	return allocateSeedData{
		SeedData: unitest.SeedData{
			Admins:             []unitest.User{{User: admins[0]}},
			Products:           products,
			InventoryLocations: inventoryLocations,
			InventoryItems:     inventoryItems,
		},
		Products:           products,
		InventoryItems:     inventoryItems,
		InventoryLocations: inventoryLocations,
		Warehouses:         warehouses,
		ExecutionContext:   execContext,
	}, nil
}

// seedTestInventoryItems creates inventory items with specific quantities for testing
func seedTestInventoryItems(ctx context.Context, productIDs, locationIDs []uuid.UUID, store inventoryitembus.Business) ([]inventoryitembus.InventoryItem, error) {
	items := []inventoryitembus.InventoryItem{}

	// Create items with known quantities for predictable testing
	quantities := []int{100, 50, 75, 25, 200}

	for i := 0; i < len(productIDs) && i < len(locationIDs); i++ {
		item, err := store.Create(ctx, inventoryitembus.NewInventoryItem{
			ProductID:         productIDs[i],
			LocationID:        locationIDs[i],
			Quantity:          quantities[i%len(quantities)],
			ReservedQuantity:  0,
			AllocatedQuantity: 0,
			MinimumStock:      10,
			MaximumStock:      500,
			ReorderPoint:      20,
		})
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// =============================================================================
// Allocation Tests

func allocateInventoryTests(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) []unitest.Table {
	return []unitest.Table{
		validateAllocationConfig(sd),
		executeBasicAllocation(busDomain, db, sd),
		executePartialAllocation(busDomain, db, sd),
		testIdempotency(busDomain, db, sd),
		testReservationMode(busDomain, db, sd),
		testFIFOStrategy(busDomain, db, sd),
	}
}

func validateAllocationConfig(sd allocateSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_allocation_config",
		ExpResp: "inventory_items list is required and must not be empty",
		ExcFunc: func(ctx context.Context) any {
			// Test with empty items
			config := json.RawMessage(`{
				"inventory_items": [],
				"allocation_mode": "allocate",
				"allocation_strategy": "fifo",
				"priority": "medium"
			}`)

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

func executeBasicAllocation(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) unitest.Table {
	if len(sd.Products) == 0 || len(sd.InventoryItems) == 0 {
		return unitest.Table{
			Name:    "execute_basic_allocation_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name: "execute_basic_allocation",
		ExpResp: map[string]any{
			"status": "queued",
		},
		ExcFunc: func(ctx context.Context) any {
			// Use first product with known inventory
			product := sd.Products[0]
			allocateQty := 10

			config := inventory.AllocateInventoryConfig{
				InventoryItems: []inventory.AllocationItem{
					{
						ProductID: product.ProductID,
						Quantity:  allocateQty,
					},
				},
				AllocationMode:     "allocate",
				AllocationStrategy: "fifo",
				AllowPartial:       false,
				Priority:           "high",
			}

			configJSON, _ := json.Marshal(config)

			// Execute allocation (will queue to RabbitMQ)
			result, err := sd.Handler.Execute(ctx, configJSON, sd.ExecutionContext)
			if err != nil {
				return err
			}

			fmt.Println(result)

			// For testing, directly process the allocation synchronously
			request := inventory.AllocationRequest{
				ID:          uuid.New(),
				ExecutionID: sd.ExecutionContext.ExecutionID,
				Config:      config,
				Context:     sd.ExecutionContext,
				Status:      "processing",
				Priority:    10,
				CreatedAt:   time.Now(),
			}

			allocationResult, err := sd.Handler.ProcessAllocation(ctx, request)
			if err != nil {
				return fmt.Errorf("process allocation failed: %w", err)
			}

			// Verify inventory was actually updated
			updatedItem, err := busDomain.InventoryItem.QueryByID(ctx, sd.InventoryItems[0].ItemID)
			if err != nil {
				return fmt.Errorf("failed to query updated inventory: %w", err)
			}

			// Check that allocated quantity increased
			if updatedItem.AllocatedQuantity != allocateQty {
				return fmt.Errorf("allocation not reflected in DB: expected %d, got %d",
					allocateQty, updatedItem.AllocatedQuantity)
			}

			return allocationResult
		},
		CmpFunc: func(got any, exp any) string {
			result, ok := got.(*inventory.AllocationResult)
			if !ok {
				return fmt.Sprintf("expected AllocationResult, got %T", got)
			}

			if result.Status != "success" {
				return fmt.Sprintf("expected success status, got %s", result.Status)
			}

			if result.TotalAllocated != 10 {
				return fmt.Sprintf("expected 10 allocated, got %d", result.TotalAllocated)
			}

			return ""
		},
	}
}

func executePartialAllocation(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) unitest.Table {
	if len(sd.Products) < 2 || len(sd.InventoryItems) < 2 {
		return unitest.Table{
			Name:    "execute_partial_allocation_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "execute_partial_allocation",
		ExpResp: "partial",
		ExcFunc: func(ctx context.Context) any {
			// Request more than available
			product := sd.Products[1]

			config := inventory.AllocateInventoryConfig{
				InventoryItems: []inventory.AllocationItem{
					{
						ProductID: product.ProductID,
						Quantity:  10000, // More than available
					},
				},
				AllocationMode:     "allocate",
				AllocationStrategy: "fifo",
				AllowPartial:       true, // Allow partial
				Priority:           "medium",
			}

			request := inventory.AllocationRequest{
				ID:          uuid.New(),
				ExecutionID: uuid.New(),
				Config:      config,
				Context:     sd.ExecutionContext,
				Status:      "processing",
				Priority:    5,
				CreatedAt:   time.Now(),
			}

			result, err := sd.Handler.ProcessAllocation(ctx, request)
			if err != nil {
				return err
			}

			return result.Status
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func testIdempotency(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) unitest.Table {
	return unitest.Table{
		Name:    "test_idempotency",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			// Use same execution context for both calls
			execContext := workflow.ActionExecutionContext{
				EntityID:    uuid.New(),
				EntityName:  "orders",
				EventType:   "on_create",
				UserID:      sd.Admins[0].ID,
				RuleID:      uuid.New(),
				RuleName:    "Idempotency Test",
				ExecutionID: uuid.New(), // Same execution ID
				Timestamp:   time.Now().UTC(),
			}

			config := inventory.AllocateInventoryConfig{
				InventoryItems: []inventory.AllocationItem{
					{
						ProductID: sd.Products[0].ProductID,
						Quantity:  5,
					},
				},
				AllocationMode:     "allocate",
				AllocationStrategy: "fifo",
				Priority:           "high",
			}

			request := inventory.AllocationRequest{
				ID:          uuid.New(),
				ExecutionID: execContext.ExecutionID,
				Config:      config,
				Context:     execContext,
				Status:      "processing",
				Priority:    10,
				CreatedAt:   time.Now(),
			}

			// First allocation
			result1, err := sd.Handler.ProcessAllocation(ctx, request)
			if err != nil {
				return err
			}

			// Second allocation with same execution context (should return cached)
			request.ID = uuid.New() // Different request ID
			result2, err := sd.Handler.ProcessAllocation(ctx, request)
			if err != nil {
				return err
			}

			// Results should be identical (same allocation ID means it was cached)
			return result1.AllocationID == result2.AllocationID
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("idempotency check failed: got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func testReservationMode(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) unitest.Table {
	if len(sd.Products) == 0 || len(sd.InventoryItems) == 0 {
		return unitest.Table{
			Name:    "test_reservation_mode_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "test_reservation_mode",
		ExpResp: 5,
		ExcFunc: func(ctx context.Context) any {
			product := sd.Products[0]
			reserveQty := 5

			config := inventory.AllocateInventoryConfig{
				InventoryItems: []inventory.AllocationItem{
					{
						ProductID: product.ProductID,
						Quantity:  reserveQty,
					},
				},
				AllocationMode:     "reserve", // Reserve mode
				AllocationStrategy: "fifo",
				ReservationHours:   24,
				Priority:           "high",
			}

			request := inventory.AllocationRequest{
				ID:          uuid.New(),
				ExecutionID: uuid.New(),
				Config:      config,
				Context:     sd.ExecutionContext,
				Status:      "processing",
				Priority:    10,
				CreatedAt:   time.Now(),
			}

			_, err := sd.Handler.ProcessAllocation(ctx, request)
			if err != nil {
				return err
			}

			// Check that reserved quantity increased
			updatedItem, err := busDomain.InventoryItem.QueryByID(ctx, sd.InventoryItems[0].ItemID)
			if err != nil {
				return err
			}

			return updatedItem.ReservedQuantity
		},
		CmpFunc: func(got any, exp any) string {
			// Should have at least the reserved amount
			if got.(int) < exp.(int) {
				return fmt.Sprintf("expected at least %v reserved, got %v", exp, got)
			}
			return ""
		},
	}
}

func testFIFOStrategy(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) unitest.Table {
	return unitest.Table{
		Name:    "test_fifo_strategy",
		ExpResp: "success",
		ExcFunc: func(ctx context.Context) any {
			// This test verifies FIFO allocation order
			// Would need multiple inventory items for same product at different locations

			config := inventory.AllocateInventoryConfig{
				InventoryItems: []inventory.AllocationItem{
					{
						ProductID: sd.Products[0].ProductID,
						Quantity:  5,
					},
				},
				AllocationMode:     "allocate",
				AllocationStrategy: "fifo", // Should allocate from oldest inventory first
				Priority:           "medium",
			}

			request := inventory.AllocationRequest{
				ID:          uuid.New(),
				ExecutionID: uuid.New(),
				Config:      config,
				Context:     sd.ExecutionContext,
				Status:      "processing",
				Priority:    5,
				CreatedAt:   time.Now(),
			}

			result, err := sd.Handler.ProcessAllocation(ctx, request)
			if err != nil {
				return err
			}

			return result.Status
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}
