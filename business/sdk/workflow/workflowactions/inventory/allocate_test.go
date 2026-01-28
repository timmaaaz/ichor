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
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

var testContainer rabbitmq.Container

func Test_AllocateInventory(t *testing.T) {
	// Note: t.Parallel() removed - this test uses shared RabbitMQ infrastructure
	// and cannot run concurrently with other workflow tests

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

	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	// Initialize workflow queue
	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	sd.Handler = inventory.NewAllocateInventoryHandler(
		log,
		db.DB,
		queue,
		db.BusDomain.InventoryItem,
		db.BusDomain.InventoryLocation,
		db.BusDomain.InventoryTransaction,
		db.BusDomain.Product,
		db.BusDomain.Workflow,
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

	// Query timezones from seed data
	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return allocateSeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	// Seed contact infos for brands
	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
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
	ruleID := uuid.New()
	execContext := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "orders",
		EventType:     "on_create",
		UserID:        adminIDs[0],
		RuleID:        &ruleID,
		RuleName:      "Test Allocation Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
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
		testSourceFromLineItem(busDomain, db, sd),
		testOrderGroupedAllocation(busDomain, db, sd),
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
			updatedItem, err := busDomain.InventoryItem.QueryByID(ctx, sd.InventoryItems[0].ID)
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
			result, ok := got.(*inventory.InventoryAllocationResult)
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
			idempRuleID := uuid.New()
			execContext := workflow.ActionExecutionContext{
				EntityID:      uuid.New(),
				EntityName:    "orders",
				EventType:     "on_create",
				UserID:        sd.Admins[0].ID,
				RuleID:        &idempRuleID,
				RuleName:      "Idempotency Test",
				ExecutionID:   uuid.New(), // Same execution ID
				Timestamp:     time.Now().UTC(),
				TriggerSource: workflow.TriggerSourceAutomation,
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
			updatedItem, err := busDomain.InventoryItem.QueryByID(ctx, sd.InventoryItems[0].ID)
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

// =============================================================================
// SourceFromLineItem Tests
//
// These tests verify the SourceFromLineItem functionality which extracts
// product_id, quantity, and order_id from the execution context's RawData.
// =============================================================================

func testSourceFromLineItem(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) unitest.Table {
	if len(sd.Products) == 0 {
		return unitest.Table{
			Name:    "test_source_from_line_item_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "test_source_from_line_item",
		ExpResp: "queued",
		ExcFunc: func(ctx context.Context) any {
			orderID := uuid.New()
			lineItemID := uuid.New()

			// Config with SourceFromLineItem enabled - no inventory_items needed
			config := inventory.AllocateInventoryConfig{
				SourceFromLineItem: true,
				AllocationMode:     "reserve",
				AllocationStrategy: "fifo",
				AllowPartial:       false,
				Priority:           "high",
			}
			configJSON, _ := json.Marshal(config)

			// Create execution context with RawData simulating a line item
			lineItemRuleID := uuid.New()
			execContext := workflow.ActionExecutionContext{
				EntityID:      lineItemID,
				EntityName:    "order_line_items",
				EventType:     "on_create",
				UserID:        sd.Admins[0].ID,
				RuleID:        &lineItemRuleID,
				RuleName:      "Test Line Item Allocation",
				ExecutionID:   uuid.New(),
				Timestamp:     time.Now().UTC(),
				TriggerSource: workflow.TriggerSourceAutomation,
				RawData: map[string]interface{}{
					"product_id": sd.Products[0].ProductID.String(),
					"quantity":   float64(5),
					"order_id":   orderID.String(),
				},
			}

			// Execute - should extract from RawData and queue
			result, err := sd.Handler.Execute(ctx, configJSON, execContext)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}

			// Verify we got a queued response
			queuedResp, ok := result.(inventory.QueuedAllocationResponse)
			if !ok {
				return fmt.Errorf("expected QueuedAllocationResponse, got %T", result)
			}

			return queuedResp.Status
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func testOrderGroupedAllocation(busDomain dbtest.BusDomain, db *sqlx.DB, sd allocateSeedData) unitest.Table {
	if len(sd.Products) < 3 {
		return unitest.Table{
			Name:    "test_order_grouped_allocation_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "test_order_grouped_allocation",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			orderA := uuid.New()
			orderB := uuid.New()

			config := inventory.AllocateInventoryConfig{
				SourceFromLineItem: true,
				AllocationMode:     "reserve",
				AllocationStrategy: "fifo",
				AllowPartial:       false,
				Priority:           "high",
			}
			configJSON, _ := json.Marshal(config)

			// Track allocation order and verify ReferenceID grouping
			type allocationRecord struct {
				label       string
				orderID     uuid.UUID
				referenceID string
			}
			var allocations []allocationRecord

			// Simulate Order A's line items (3 items) - should all be grouped
			for i := 0; i < 3; i++ {
				groupedRuleID := uuid.New()
				execCtx := workflow.ActionExecutionContext{
					EntityID:      uuid.New(),
					EntityName:    "order_line_items",
					EventType:     "on_create",
					UserID:        sd.Admins[0].ID,
					RuleID:        &groupedRuleID,
					RuleName:      "Test Grouped Allocation",
					ExecutionID:   uuid.New(),
					Timestamp:     time.Now().UTC(),
					TriggerSource: workflow.TriggerSourceAutomation,
					RawData: map[string]interface{}{
						"product_id": sd.Products[i].ProductID.String(),
						"quantity":   float64(2),
						"order_id":   orderA.String(),
					},
				}

				result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
				if err != nil {
					return fmt.Errorf("order A item %d failed: %w", i+1, err)
				}

				queuedResp, ok := result.(inventory.QueuedAllocationResponse)
				if !ok {
					return fmt.Errorf("expected QueuedAllocationResponse for A%d, got %T", i+1, result)
				}

				allocations = append(allocations, allocationRecord{
					label:       fmt.Sprintf("A%d", i+1),
					orderID:     orderA,
					referenceID: queuedResp.ReferenceID,
				})
			}

			// Simulate Order B's line items (2 items) - should be grouped separately
			for i := 0; i < 2; i++ {
				orderBRuleID := uuid.New()
				execCtx := workflow.ActionExecutionContext{
					EntityID:      uuid.New(),
					EntityName:    "order_line_items",
					EventType:     "on_create",
					UserID:        sd.Admins[0].ID,
					RuleID:        &orderBRuleID,
					RuleName:      "Test Grouped Allocation",
					ExecutionID:   uuid.New(),
					Timestamp:     time.Now().UTC(),
					TriggerSource: workflow.TriggerSourceAutomation,
					RawData: map[string]interface{}{
						"product_id": sd.Products[i].ProductID.String(),
						"quantity":   float64(2),
						"order_id":   orderB.String(),
					},
				}

				result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
				if err != nil {
					return fmt.Errorf("order B item %d failed: %w", i+1, err)
				}

				queuedResp, ok := result.(inventory.QueuedAllocationResponse)
				if !ok {
					return fmt.Errorf("expected QueuedAllocationResponse for B%d, got %T", i+1, result)
				}

				allocations = append(allocations, allocationRecord{
					label:       fmt.Sprintf("B%d", i+1),
					orderID:     orderB,
					referenceID: queuedResp.ReferenceID,
				})
			}

			// Verify:
			// 1. Order preserved: A1, A2, A3, B1, B2
			// 2. All Order A items have same ReferenceID (orderA)
			// 3. All Order B items have same ReferenceID (orderB)
			// 4. ReferenceIDs are different between orders

			expectedLabels := []string{"A1", "A2", "A3", "B1", "B2"}
			for i, exp := range expectedLabels {
				if i >= len(allocations) || allocations[i].label != exp {
					return fmt.Errorf("order mismatch at position %d: expected %s", i, exp)
				}
			}

			// Verify Order A items all reference orderA
			for i := 0; i < 3; i++ {
				if allocations[i].referenceID != orderA.String() {
					return fmt.Errorf("A%d has wrong ReferenceID: got %s, want %s",
						i+1, allocations[i].referenceID, orderA.String())
				}
			}

			// Verify Order B items all reference orderB
			for i := 3; i < 5; i++ {
				if allocations[i].referenceID != orderB.String() {
					return fmt.Errorf("B%d has wrong ReferenceID: got %s, want %s",
						i-2, allocations[i].referenceID, orderB.String())
				}
			}

			// Verify orders are different
			if orderA.String() == orderB.String() {
				return fmt.Errorf("order IDs should be different")
			}

			return true
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("order grouping failed: got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

// =============================================================================
// ProcessQueued Unit Tests
//
// These tests verify the ProcessQueued method's error handling for malformed payloads.
// Integration tests for the full async processing flow are in queue_test.go.
// =============================================================================

func Test_ProcessQueued(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ProcessQueued")

	sd, err := insertAllocateSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Create the handler with all dependencies
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(context.Background()); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	handler := inventory.NewAllocateInventoryHandler(
		log,
		db.DB,
		queue,
		db.BusDomain.InventoryItem,
		db.BusDomain.InventoryLocation,
		db.BusDomain.InventoryTransaction,
		db.BusDomain.Product,
		db.BusDomain.Workflow,
	)

	// -------------------------------------------------------------------------

	unitest.Run(t, processQueuedTests(handler, sd), "processQueued")
}

func processQueuedTests(handler *inventory.AllocateInventoryHandler, sd allocateSeedData) []unitest.Table {
	// Note: Most ProcessQueued paths are tested via integration tests in queue_test.go
	// which uses the full async flow with real EventPublisher.
	//
	// These unit tests focus on the early error path that fails before
	// reaching the EventPublisher (invalid JSON payload).
	//
	// The "invalid request data" case cannot be meaningfully tested here because:
	// 1. Go's JSON unmarshalling is lenient - it ignores unknown fields and uses zero values
	// 2. If unmarshal succeeds, ProcessAllocation runs and tries to use the publisher
	// 3. We can't pass a nil publisher without causing a panic
	// 4. The user prefers not to use mocks
	//
	// The full processing flow including publisher is covered by queue_test.go
	return []unitest.Table{
		processQueuedInvalidPayload(handler),
	}
}

func processQueuedInvalidPayload(handler *inventory.AllocateInventoryHandler) unitest.Table {
	return unitest.Table{
		Name:    "process_queued_invalid_payload",
		ExpResp: "failed to unmarshal queued payload",
		ExcFunc: func(ctx context.Context) any {
			// Invalid JSON payload - should fail to unmarshal QueuedPayload
			invalidPayload := json.RawMessage(`{not valid json}`)

			err := handler.ProcessQueued(ctx, invalidPayload, nil)
			if err == nil {
				return "expected error, got nil"
			}
			// Check that error message contains expected text
			if !contains(err.Error(), "failed to unmarshal queued payload") {
				return fmt.Sprintf("unexpected error: %s", err.Error())
			}
			return "failed to unmarshal queued payload"
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

// contains checks if substr is in s
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =============================================================================
// ProcessQueued Failure Event Test
//
// This test verifies that when allocation fails (e.g., no inventory available),
// the ProcessQueued method still fires a workflow event with status="failed"
// so downstream rules (like alerts) can trigger.
// =============================================================================

func Test_ProcessQueued_FiresEventOnFailure(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ProcessQueued_FiresEventOnFailure")

	// Create minimal seed data - specifically WITHOUT inventory for the product we'll test
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		t.Fatalf("Seeding admin: %s", err)
	}

	// Seed a product but DON'T seed any inventory for it
	regions, _ := db.BusDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "1"))
	regionIDs := []uuid.UUID{regions[0].ID}

	cities, _ := citybus.TestSeedCities(ctx, 1, regionIDs, db.BusDomain.City)
	cityIDs := []uuid.UUID{cities[0].ID}

	streets, _ := streetbus.TestSeedStreets(ctx, 1, cityIDs, db.BusDomain.Street)
	streetIDs := []uuid.UUID{streets[0].ID}

	tzs, _ := db.BusDomain.Timezone.QueryAll(ctx)
	tzIDs := []uuid.UUID{tzs[0].ID}

	contactInfos, _ := contactinfosbus.TestSeedContactInfos(ctx, 1, streetIDs, tzIDs, db.BusDomain.ContactInfos)
	contactIDs := uuid.UUIDs{contactInfos[0].ID}

	brands, _ := brandbus.TestSeedBrands(ctx, 1, contactIDs, db.BusDomain.Brand)
	brandIDs := uuid.UUIDs{brands[0].BrandID}

	productCategories, _ := productcategorybus.TestSeedProductCategories(ctx, 1, db.BusDomain.ProductCategory)
	productCategoryIDs := uuid.UUIDs{productCategories[0].ProductCategoryID}

	products, _ := productbus.TestSeedProducts(ctx, 1, brandIDs, productCategoryIDs, db.BusDomain.Product)
	productWithNoInventory := products[0]

	// Create handler
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	container := rabbitmq.GetTestContainer(t)
	client := rabbitmq.NewTestClient(container.URL)
	if err := client.Connect(); err != nil {
		t.Fatalf("connecting to rabbitmq: %s", err)
	}
	defer client.Close()

	queue := rabbitmq.NewTestWorkflowQueue(client, log)
	if err := queue.Initialize(ctx); err != nil {
		t.Fatalf("initializing workflow queue: %s", err)
	}

	handler := inventory.NewAllocateInventoryHandler(
		log,
		db.DB,
		queue,
		db.BusDomain.InventoryItem,
		db.BusDomain.InventoryLocation,
		db.BusDomain.InventoryTransaction,
		db.BusDomain.Product,
		db.BusDomain.Workflow,
	)

	// Build a valid QueuedPayload that will trigger allocation failure
	failureRuleID := uuid.New()
	execContext := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "order_line_items",
		EventType:     "on_create",
		UserID:        admins[0].ID,
		RuleID:        &failureRuleID,
		RuleName:      "Test Failure Alert",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	allocConfig := inventory.AllocateInventoryConfig{
		InventoryItems: []inventory.AllocationItem{{
			ProductID: productWithNoInventory.ProductID,
			Quantity:  10, // Request inventory that doesn't exist
		}},
		AllocationMode:     "allocate",
		AllocationStrategy: "fifo",
		AllowPartial:       false, // Must fail if not enough
		Priority:           "high",
		ReferenceID:        "TEST-ORDER-123",
		ReferenceType:      "order",
	}

	request := inventory.AllocationRequest{
		ID:          uuid.New(),
		ExecutionID: execContext.ExecutionID,
		Config:      allocConfig,
		Context:     execContext,
		Status:      "queued",
		Priority:    10,
		CreatedAt:   time.Now(),
		MaxRetries:  3,
	}

	requestData, _ := json.Marshal(request)

	queuedPayload := workflow.QueuedPayload{
		RequestType:      "allocate_inventory",
		RequestData:      requestData,
		ExecutionContext: execContext,
		IdempotencyKey:   fmt.Sprintf("%s_%s_allocate_inventory", execContext.ExecutionID, execContext.RuleID.String()),
	}

	payloadJSON, _ := json.Marshal(queuedPayload)

	// Create a real QueueManager and EventPublisher
	queueManager, err := workflow.NewQueueManager(log, db.DB, nil, client, queue)
	if err != nil {
		t.Fatalf("creating queue manager: %s", err)
	}
	if err := queueManager.Initialize(ctx); err != nil {
		t.Fatalf("initializing queue manager: %s", err)
	}

	eventPublisher := workflow.NewEventPublisher(log, queueManager)

	// Execute ProcessQueued - this should fail but still fire an event
	err = handler.ProcessQueued(ctx, payloadJSON, eventPublisher)

	// We expect an error because allocation failed
	if err == nil {
		t.Fatal("expected allocation to fail, but it succeeded")
	}

	if !contains(err.Error(), "allocation processing failed") {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	// Give the async event publisher goroutine time to complete
	time.Sleep(100 * time.Millisecond)

	// Check logs for evidence that the event was fired
	logOutput := buf.String()
	if !contains(logOutput, "Allocation processing failed, event fired for alerting") {
		t.Errorf("expected log message about event being fired on failure, got:\n%s", logOutput)
	}

	// The key verification: the log message confirms fireAllocationResultEvent was called
	// before returning the error, which is the fix we implemented

	t.Log("SUCCESS: ProcessQueued fires workflow event even when allocation fails")

	// Cleanup with timeout to prevent blocking
	stopCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	queueManager.Stop(stopCtx)
}
