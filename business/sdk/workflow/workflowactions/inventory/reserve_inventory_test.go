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

func Test_ReserveInventory(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ReserveInventory")

	sd, err := insertReserveInventorySeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.Handler = inventory.NewReserveInventoryHandler(log, db.DB, db.BusDomain.InventoryItem, db.BusDomain.Workflow)

	unitest.Run(t, reserveInventoryTests(db.BusDomain, db.DB, sd), "reserveInventory")
}

// =============================================================================

type reserveInventorySeedData struct {
	unitest.SeedData
	Handler            *inventory.ReserveInventoryHandler
	Products           []productbus.Product
	InventoryItems     []inventoryitembus.InventoryItem
	InventoryLocations []inventorylocationbus.InventoryLocation
	Warehouses         []warehousebus.Warehouse
	ExecutionContext   workflow.ActionExecutionContext
}

func insertReserveInventorySeedData(busDomain dbtest.BusDomain) (reserveInventorySeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	adminIDs := make([]uuid.UUID, len(admins))
	for i, admin := range admins {
		adminIDs[i] = admin.ID
	}

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding brand : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding product category : %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 5, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding product : %w", err)
	}

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, busDomain.Zones)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 10, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationIDs[i] = il.LocationID
	}

	inventoryItems, err := seedTestInventoryItems(ctx, productIDs, inventoryLocationIDs, *busDomain.InventoryItem)
	if err != nil {
		return reserveInventorySeedData{}, fmt.Errorf("seeding inventory items : %w", err)
	}

	ruleID := uuid.New()
	execContext := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "orders",
		EventType:     "on_create",
		UserID:        adminIDs[0],
		RuleID:        &ruleID,
		RuleName:      "Test Reserve Inventory Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return reserveInventorySeedData{
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

// =============================================================================

func reserveInventoryTests(busDomain dbtest.BusDomain, db *sqlx.DB, sd reserveInventorySeedData) []unitest.Table {
	return []unitest.Table{
		validateReserveConfig(sd),
		executeBasicReserve(busDomain, sd),
		executePartialReserve(sd),
		executeInsufficientStrict(sd),
		testReserveIdempotency(sd),
		executeReserveSourceFromLineItem(sd),
	}
}

func validateReserveConfig(sd reserveInventorySeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_config",
		ExpResp: "product_id is required when source_from_line_item is false",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"quantity": 10, "allocation_strategy": "fifo"}`)
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

func executeBasicReserve(busDomain dbtest.BusDomain, sd reserveInventorySeedData) unitest.Table {
	if len(sd.Products) == 0 || len(sd.InventoryItems) == 0 {
		return unitest.Table{
			Name:    "execute_basic_reserve_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "execute_basic_reserve",
		ExpResp: 10,
		ExcFunc: func(ctx context.Context) any {
			// Reserve 10 from first product (qty=100).
			config := inventory.ReserveInventoryConfig{
				ProductID:              sd.Products[0].ProductID.String(),
				Quantity:               10,
				AllocationStrategy:     "fifo",
				ReservationDurationHrs: 24,
			}
			configJSON, _ := json.Marshal(config)

			// Use unique execution context to avoid idempotency collision.
			reserveRuleID := uuid.New()
			execCtx := workflow.ActionExecutionContext{
				EntityID:      uuid.New(),
				EntityName:    "orders",
				EventType:     "on_create",
				UserID:        sd.Admins[0].ID,
				RuleID:        &reserveRuleID,
				RuleName:      "Basic Reserve Test",
				ExecutionID:   uuid.New(),
				Timestamp:     time.Now().UTC(),
				TriggerSource: workflow.TriggerSourceAutomation,
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return err
			}

			reserveResult, ok := result.(inventory.ReserveInventoryResult)
			if !ok {
				return fmt.Errorf("expected ReserveInventoryResult, got %T", result)
			}

			if reserveResult.Status != "success" {
				return fmt.Errorf("expected success, got %s", reserveResult.Status)
			}

			// Verify DB state.
			updatedItem, err := busDomain.InventoryItem.QueryByID(ctx, sd.InventoryItems[0].ID)
			if err != nil {
				return err
			}

			return updatedItem.ReservedQuantity
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executePartialReserve(sd reserveInventorySeedData) unitest.Table {
	if len(sd.Products) < 2 || len(sd.InventoryItems) < 2 {
		return unitest.Table{
			Name:    "execute_partial_reserve_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "execute_partial_reserve",
		ExpResp: "partial",
		ExcFunc: func(ctx context.Context) any {
			// Second product has qty=50, try to reserve 10000 with allow_partial=true.
			config := inventory.ReserveInventoryConfig{
				ProductID:              sd.Products[1].ProductID.String(),
				Quantity:               10000,
				AllocationStrategy:     "fifo",
				AllowPartial:           true,
				ReservationDurationHrs: 24,
			}
			configJSON, _ := json.Marshal(config)

			partialRuleID := uuid.New()
			execCtx := workflow.ActionExecutionContext{
				EntityID:      uuid.New(),
				EntityName:    "orders",
				EventType:     "on_create",
				UserID:        sd.Admins[0].ID,
				RuleID:        &partialRuleID,
				RuleName:      "Partial Reserve Test",
				ExecutionID:   uuid.New(),
				Timestamp:     time.Now().UTC(),
				TriggerSource: workflow.TriggerSourceAutomation,
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return err
			}

			reserveResult, ok := result.(inventory.ReserveInventoryResult)
			if !ok {
				return fmt.Errorf("expected ReserveInventoryResult, got %T", result)
			}

			return reserveResult.Status
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeInsufficientStrict(sd reserveInventorySeedData) unitest.Table {
	if len(sd.Products) < 3 || len(sd.InventoryItems) < 3 {
		return unitest.Table{
			Name:    "execute_insufficient_strict_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "execute_insufficient_strict",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			// Third product has qty=75, reserve 10000 with allow_partial=false -> error.
			config := inventory.ReserveInventoryConfig{
				ProductID:              sd.Products[2].ProductID.String(),
				Quantity:               10000,
				AllocationStrategy:     "fifo",
				AllowPartial:           false,
				ReservationDurationHrs: 24,
			}
			configJSON, _ := json.Marshal(config)

			strictRuleID := uuid.New()
			execCtx := workflow.ActionExecutionContext{
				EntityID:      uuid.New(),
				EntityName:    "orders",
				EventType:     "on_create",
				UserID:        sd.Admins[0].ID,
				RuleID:        &strictRuleID,
				RuleName:      "Strict Reserve Test",
				ExecutionID:   uuid.New(),
				Timestamp:     time.Now().UTC(),
				TriggerSource: workflow.TriggerSourceAutomation,
			}

			_, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err == nil {
				return fmt.Errorf("expected error but got nil")
			}

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

func testReserveIdempotency(sd reserveInventorySeedData) unitest.Table {
	if len(sd.Products) == 0 || len(sd.InventoryItems) == 0 {
		return unitest.Table{
			Name:    "test_idempotency_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "test_idempotency",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := inventory.ReserveInventoryConfig{
				ProductID:              sd.Products[0].ProductID.String(),
				Quantity:               5,
				AllocationStrategy:     "fifo",
				ReservationDurationHrs: 24,
			}
			configJSON, _ := json.Marshal(config)

			// Same execution context for both calls.
			idempRuleID := uuid.New()
			execCtx := workflow.ActionExecutionContext{
				EntityID:      uuid.New(),
				EntityName:    "orders",
				EventType:     "on_create",
				UserID:        sd.Admins[0].ID,
				RuleID:        &idempRuleID,
				RuleName:      "Idempotency Reserve Test",
				ExecutionID:   uuid.New(), // Same execution ID for both calls.
				Timestamp:     time.Now().UTC(),
				TriggerSource: workflow.TriggerSourceAutomation,
			}

			// First call.
			result1, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return err
			}

			// Second call with same execution context -> should return cached.
			result2, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return err
			}

			r1, ok1 := result1.(inventory.ReserveInventoryResult)
			r2, ok2 := result2.(inventory.ReserveInventoryResult)
			if !ok1 || !ok2 {
				return fmt.Errorf("unexpected types: %T, %T", result1, result2)
			}

			// Same reservation ID means cached result was returned.
			return r1.ReservationID == r2.ReservationID
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("idempotency check failed: got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeReserveSourceFromLineItem(sd reserveInventorySeedData) unitest.Table {
	if len(sd.Products) < 4 || len(sd.InventoryItems) < 4 {
		return unitest.Table{
			Name:    "execute_source_from_line_item_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "execute_source_from_line_item",
		ExpResp: "success",
		ExcFunc: func(ctx context.Context) any {
			config := inventory.ReserveInventoryConfig{
				SourceFromLineItem:     true,
				AllocationStrategy:     "fifo",
				ReservationDurationHrs: 24,
			}
			configJSON, _ := json.Marshal(config)

			lineItemRuleID := uuid.New()
			execCtx := workflow.ActionExecutionContext{
				EntityID:      uuid.New(),
				EntityName:    "order_line_items",
				EventType:     "on_create",
				UserID:        sd.Admins[0].ID,
				RuleID:        &lineItemRuleID,
				RuleName:      "Line Item Reserve Test",
				ExecutionID:   uuid.New(),
				Timestamp:     time.Now().UTC(),
				TriggerSource: workflow.TriggerSourceAutomation,
				RawData: map[string]interface{}{
					"product_id": sd.Products[3].ProductID.String(),
					"quantity":   float64(5),
					"order_id":   uuid.New().String(),
				},
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return err
			}

			reserveResult, ok := result.(inventory.ReserveInventoryResult)
			if !ok {
				return fmt.Errorf("expected ReserveInventoryResult, got %T", result)
			}

			return reserveResult.Status
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}
