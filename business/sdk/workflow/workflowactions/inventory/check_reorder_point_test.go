package inventory_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
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

func Test_CheckReorderPoint(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_CheckReorderPoint")

	sd, err := insertCheckReorderPointSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.Handler = inventory.NewCheckReorderPointHandler(log, db.BusDomain.InventoryItem)

	unitest.Run(t, checkReorderPointTests(db.BusDomain, sd), "checkReorderPoint")
}

// =============================================================================

type checkReorderPointSeedData struct {
	unitest.SeedData
	Handler            *inventory.CheckReorderPointHandler
	Products           []productbus.Product
	InventoryItems     []inventoryitembus.InventoryItem
	InventoryLocations []inventorylocationbus.InventoryLocation
	ExecutionContext   workflow.ActionExecutionContext
}

func insertCheckReorderPointSeedData(busDomain dbtest.BusDomain) (checkReorderPointSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding user : %w", err)
	}

	adminIDs := make([]uuid.UUID, len(admins))
	for i, admin := range admins {
		adminIDs[i] = admin.ID
	}

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("querying regions : %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding cities : %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding streets : %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("querying timezones : %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding contact info : %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding brand : %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding product category : %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 5, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding product : %w", err)
	}

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding warehouses : %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, busDomain.Zones)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding zones : %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 10, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding inventory locations : %w", err)
	}

	inventoryLocationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationIDs[i] = il.LocationID
	}

	inventoryItems, err := seedTestInventoryItems(ctx, productIDs, inventoryLocationIDs, *busDomain.InventoryItem)
	if err != nil {
		return checkReorderPointSeedData{}, fmt.Errorf("seeding inventory items : %w", err)
	}

	ruleID := uuid.New()
	execContext := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "orders",
		EventType:     "on_create",
		UserID:        adminIDs[0],
		RuleID:        &ruleID,
		RuleName:      "Test Reorder Point Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return checkReorderPointSeedData{
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

func checkReorderPointTests(busDomain dbtest.BusDomain, sd checkReorderPointSeedData) []unitest.Table {
	return []unitest.Table{
		validateCheckReorderPointConfig(sd),
		executeReorderPointBelowThreshold(busDomain, sd),
		executeReorderPointAboveThreshold(sd),
		executeReorderPointCustomThreshold(sd),
	}
}

func validateCheckReorderPointConfig(sd checkReorderPointSeedData) unitest.Table {
	return unitest.Table{
		Name:    "validate_config",
		ExpResp: "product_id is required when source_from_line_item is false",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{}`)
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

func executeReorderPointBelowThreshold(busDomain dbtest.BusDomain, sd checkReorderPointSeedData) unitest.Table {
	if len(sd.Products) == 0 || len(sd.InventoryItems) == 0 {
		return unitest.Table{
			Name:    "execute_below_reorder_point_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "execute_below_reorder_point",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			// Seed item has qty=100, ReorderPoint=20.
			// Pre-reserve 85 so available=15 < 20 -> needs reorder -> true_branch.
			item := sd.InventoryItems[0]
			reservedQty := 85
			_, err := busDomain.InventoryItem.Update(ctx, item, inventoryitembus.UpdateInventoryItem{
				ReservedQuantity: &reservedQty,
			})
			if err != nil {
				return fmt.Errorf("pre-reserve failed: %w", err)
			}

			config := inventory.CheckReorderPointConfig{
				ProductID: sd.Products[0].ProductID.String(),
			}
			configJSON, _ := json.Marshal(config)

			result, err := sd.Handler.Execute(ctx, configJSON, sd.ExecutionContext)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map[string]any, got %T", result)
			}

			// Result=true means needs reorder (below threshold).
			return resultMap["needs_reorder"].(bool)
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeReorderPointAboveThreshold(sd checkReorderPointSeedData) unitest.Table {
	if len(sd.Products) < 2 || len(sd.InventoryItems) < 2 {
		return unitest.Table{
			Name:    "execute_above_reorder_point_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "execute_above_reorder_point",
		ExpResp: false,
		ExcFunc: func(ctx context.Context) any {
			// Second product has qty=50, no reservations, ReorderPoint=20.
			// available=50 > 20 -> does NOT need reorder -> false_branch.
			config := inventory.CheckReorderPointConfig{
				ProductID: sd.Products[1].ProductID.String(),
			}
			configJSON, _ := json.Marshal(config)

			result, err := sd.Handler.Execute(ctx, configJSON, sd.ExecutionContext)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map[string]any, got %T", result)
			}

			// Result=false means does NOT need reorder (above threshold).
			return resultMap["needs_reorder"].(bool)
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func executeReorderPointCustomThreshold(sd checkReorderPointSeedData) unitest.Table {
	if len(sd.Products) < 3 || len(sd.InventoryItems) < 3 {
		return unitest.Table{
			Name:    "execute_custom_threshold_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got any, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "execute_custom_threshold",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			// Third product has qty=75, ReorderPoint=20.
			// Custom threshold of 150 overrides. available=75 < 150 -> needs reorder.
			customThreshold := 150
			config := inventory.CheckReorderPointConfig{
				ProductID:       sd.Products[2].ProductID.String(),
				CustomThreshold: &customThreshold,
			}
			configJSON, _ := json.Marshal(config)

			result, err := sd.Handler.Execute(ctx, configJSON, sd.ExecutionContext)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map[string]any, got %T", result)
			}

			return resultMap["needs_reorder"].(bool)
		},
		CmpFunc: func(got any, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}
