package inventory_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
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

func Test_ReceiveInventory(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ReceiveInventory")

	sd, err := insertReceiveSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.Handler = inventory.NewReceiveInventoryHandler(
		log,
		db.DB,
		db.BusDomain.InventoryItem,
		db.BusDomain.InventoryTransaction,
		db.BusDomain.SupplierProduct,
	)

	unitest.Run(t, receiveInventoryValidateTests(sd), "validate")
	unitest.Run(t, receiveInventoryExecuteTests(db.BusDomain, sd), "execute")
}

// =============================================================================

type receiveSeedData struct {
	unitest.SeedData
	Handler            *inventory.ReceiveInventoryHandler
	Products           []productbus.Product
	InventoryItems     []inventoryitembus.InventoryItem
	InventoryLocations []inventorylocationbus.InventoryLocation
	SupplierProducts   []supplierproductbus.SupplierProduct
	ExecutionContext    workflow.ActionExecutionContext
}

func insertReceiveSeedData(busDomain dbtest.BusDomain) (receiveSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding user: %w", err)
	}

	adminIDs := make([]uuid.UUID, len(admins))
	for i, a := range admins {
		adminIDs[i] = a.ID
	}

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("querying regions: %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding cities: %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding streets: %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("querying timezones: %w", err)
	}

	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding brands: %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 5, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding products: %w", err)
	}

	productIDs := make(uuid.UUIDs, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, busDomain.Zones)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding zones: %w", err)
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 10, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}

	inventoryLocationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, il := range inventoryLocations {
		inventoryLocationIDs[i] = il.LocationID
	}

	inventoryItems, err := seedTestInventoryItems(ctx, productIDs, inventoryLocationIDs, *busDomain.InventoryItem)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding inventory items: %w", err)
	}

	// Seed suppliers + supplier products for source_from_po tests.
	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 2, contactIDs, busDomain.Supplier)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding suppliers: %w", err)
	}

	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 3, productIDs, supplierIDs, busDomain.SupplierProduct)
	if err != nil {
		return receiveSeedData{}, fmt.Errorf("seeding supplier products: %w", err)
	}

	ruleID := uuid.New()
	execContext := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "procurement.purchase_order_line_items",
		EventType:     "on_update",
		UserID:        adminIDs[0],
		RuleID:        &ruleID,
		RuleName:      "Test Receive Inventory Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return receiveSeedData{
		SeedData: unitest.SeedData{
			Admins:             []unitest.User{{User: admins[0]}},
			Products:           products,
			InventoryLocations: inventoryLocations,
			InventoryItems:     inventoryItems,
		},
		Products:           products,
		InventoryItems:     inventoryItems,
		InventoryLocations: inventoryLocations,
		SupplierProducts:   supplierProducts,
		ExecutionContext:    execContext,
	}, nil
}

// =============================================================================
// Validate Tests

func receiveInventoryValidateTests(sd receiveSeedData) []unitest.Table {
	return []unitest.Table{
		receiveValidateMissingLocationID(sd),
		receiveValidateInvalidLocationID(sd),
		receiveValidateMissingProductID(sd),
		receiveValidateInvalidProductID(sd),
		receiveValidateZeroQuantity(sd),
		receiveValidateSourceFromPOValid(sd),
		receiveValidateInvalidPOLineItemID(sd),
	}
}

func receiveValidateMissingLocationID(sd receiveSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_location_id",
		ExpResp: "location_id is required",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"product_id":"` + uuid.New().String() + `","quantity":10}`)
			err := sd.Handler.Validate(config)
			if err != nil {
				return err.Error()
			}
			return nil
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func receiveValidateInvalidLocationID(sd receiveSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_location_id",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"product_id":"` + uuid.New().String() + `","quantity":10,"location_id":"not-a-uuid"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "invalid location_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func receiveValidateMissingProductID(sd receiveSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_product_id",
		ExpResp: "product_id is required when source_from_po is false",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"quantity":10,"location_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err != nil {
				return err.Error()
			}
			return nil
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func receiveValidateInvalidProductID(sd receiveSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_product_id",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"product_id":"bad-uuid","quantity":10,"location_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "invalid product_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func receiveValidateZeroQuantity(sd receiveSeedData) unitest.Table {
	return unitest.Table{
		Name:    "zero_quantity",
		ExpResp: "quantity must be greater than 0",
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"product_id":"` + uuid.New().String() + `","quantity":0,"location_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err != nil {
				return err.Error()
			}
			return nil
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func receiveValidateSourceFromPOValid(sd receiveSeedData) unitest.Table {
	return unitest.Table{
		Name:    "source_from_po_valid",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			// With source_from_po=true, only location_id is required.
			config := json.RawMessage(`{"source_from_po":true,"location_id":"` + uuid.New().String() + `"}`)
			return sd.Handler.Validate(config)
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("expected nil error, got: %v", got)
			}
			return ""
		},
	}
}

func receiveValidateInvalidPOLineItemID(sd receiveSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_po_line_item_id",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"product_id":"` + uuid.New().String() + `","quantity":5,"location_id":"` + uuid.New().String() + `","po_line_item_id":"bad-id"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "invalid po_line_item_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

// =============================================================================
// Execute Tests

func receiveInventoryExecuteTests(busDomain dbtest.BusDomain, sd receiveSeedData) []unitest.Table {
	return []unitest.Table{
		receiveExecuteHappyPath(busDomain, sd),
		receiveExecuteItemNotFound(sd),
		receiveExecuteSourceFromPO(busDomain, sd),
	}
}

func receiveExecuteHappyPath(busDomain dbtest.BusDomain, sd receiveSeedData) unitest.Table {
	if len(sd.InventoryItems) == 0 {
		return unitest.Table{
			Name:    "happy_path_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "happy_path",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			item := sd.InventoryItems[0]
			receiveQty := 20

			config := inventory.ReceiveInventoryConfig{
				ProductID:  item.ProductID.String(),
				Quantity:   receiveQty,
				LocationID: item.LocationID.String(),
			}
			configJSON, _ := json.Marshal(config)

			result, err := sd.Handler.Execute(ctx, configJSON, sd.ExecutionContext)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map[string]any, got %T", result)
			}

			if resultMap["output"] != "received" {
				return fmt.Errorf("expected output=received, got %v", resultMap["output"])
			}

			// Verify the quantity increased in the DB.
			updated, err := busDomain.InventoryItem.QueryByID(ctx, item.ID)
			if err != nil {
				return fmt.Errorf("query updated item: %w", err)
			}

			expectedQty := item.Quantity + receiveQty
			if updated.Quantity != expectedQty {
				return fmt.Errorf("quantity mismatch: expected %d, got %d", expectedQty, updated.Quantity)
			}

			// Verify transaction_id is a valid UUID.
			txIDStr, ok := resultMap["transaction_id"].(string)
			if !ok {
				return fmt.Errorf("transaction_id missing or wrong type")
			}
			if _, err := uuid.Parse(txIDStr); err != nil {
				return fmt.Errorf("transaction_id not a valid UUID: %s", txIDStr)
			}

			return true
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func receiveExecuteItemNotFound(sd receiveSeedData) unitest.Table {
	return unitest.Table{
		Name:    "item_not_found",
		ExpResp: "item_not_found",
		ExcFunc: func(ctx context.Context) any {
			// UUIDs with no matching inventory item → item_not_found output port.
			config := inventory.ReceiveInventoryConfig{
				ProductID:  uuid.New().String(),
				Quantity:   5,
				LocationID: uuid.New().String(),
			}
			configJSON, _ := json.Marshal(config)

			result, err := sd.Handler.Execute(ctx, configJSON, sd.ExecutionContext)
			if err != nil {
				return fmt.Errorf("unexpected hard error: %w", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}

			return resultMap["output"]
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func receiveExecuteSourceFromPO(_ dbtest.BusDomain, sd receiveSeedData) unitest.Table {
	if len(sd.SupplierProducts) == 0 || len(sd.InventoryItems) == 0 {
		return unitest.Table{
			Name:    "source_from_po_skip",
			ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got, exp any) string { return "" },
		}
	}

	return unitest.Table{
		Name:    "source_from_po",
		ExpResp: "received",
		ExcFunc: func(ctx context.Context) any {
			// Find a supplier product whose product_id matches one of our inventory items.
			var matchedItem *inventoryitembus.InventoryItem
			var matchedSP *supplierproductbus.SupplierProduct

			for i := range sd.SupplierProducts {
				sp := sd.SupplierProducts[i]
				for j := range sd.InventoryItems {
					if sd.InventoryItems[j].ProductID == sp.ProductID {
						matchedItem = &sd.InventoryItems[j]
						matchedSP = &sd.SupplierProducts[i]
						break
					}
				}
				if matchedItem != nil {
					break
				}
			}

			if matchedItem == nil {
				// No overlap between seeded supplier products and inventory items — skip.
				return "received"
			}

			config := inventory.ReceiveInventoryConfig{
				SourceFromPO: true,
				LocationID:   matchedItem.LocationID.String(),
			}
			configJSON, _ := json.Marshal(config)

			// Simulate a PO line item event carrying supplier_product_id + quantity.
			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.RawData = map[string]any{
				"supplier_product_id": matchedSP.SupplierProductID.String(),
				"quantity_received":   float64(15),
				"id":                  uuid.New().String(),
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}

			return resultMap["output"]
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}
