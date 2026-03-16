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
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
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

func Test_CreatePutAwayTask(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_CreatePutAwayTask")

	sd, err := insertCreatePutAwayTaskSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.Handler = inventory.NewCreatePutAwayTaskHandler(
		log,
		db.BusDomain.PutAwayTask,
		db.BusDomain.SupplierProduct,
		db.BusDomain.PurchaseOrder,
	)

	unitest.Run(t, createPutAwayTaskValidateTests(sd), "validate")
	unitest.Run(t, createPutAwayTaskExecuteTests(db.BusDomain, sd), "execute")
}

// =============================================================================

type createPutAwayTaskSeedData struct {
	unitest.SeedData
	Handler            *inventory.CreatePutAwayTaskHandler
	Products           []productbus.Product
	InventoryLocations []inventorylocationbus.InventoryLocation
	SupplierProducts   []supplierproductbus.SupplierProduct
	POWithLocation     purchaseorderbus.PurchaseOrder // has DeliveryLocationID set
	POWithoutLocation  purchaseorderbus.PurchaseOrder // DeliveryLocationID = uuid.Nil
	LineItems          []purchaseorderlineitembus.PurchaseOrderLineItem
	ExecutionContext   workflow.ActionExecutionContext
}

func insertCreatePutAwayTaskSeedData(busDomain dbtest.BusDomain) (createPutAwayTaskSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding user: %w", err)
	}
	adminIDs := make([]uuid.UUID, len(admins))
	for i, a := range admins {
		adminIDs[i] = a.ID
	}

	// Geography chain (required for warehouses and inventory locations)
	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}
	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	// Products (required for put_away_tasks.product_id FK)
	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding brands: %w", err)
	}
	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}
	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 5, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding products: %w", err)
	}
	productIDs := make(uuid.UUIDs, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	// Inventory locations (required for put_away_tasks.location_id FK)
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}
	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 4, warehouseIDs, busDomain.Zones)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding zones: %w", err)
	}
	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ZoneID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 5, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}

	// Procurement entities (for PO-based location resolution and supplier_product -> product mapping)
	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 2, contactIDs, busDomain.Supplier)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding suppliers: %w", err)
	}
	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 3, productIDs, supplierIDs, busDomain.SupplierProduct)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding supplier products: %w", err)
	}

	// Purchase order setup — seed reference data (test DB is empty)
	poStatuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 3, busDomain.PurchaseOrderStatus)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding PO statuses: %w", err)
	}
	poStatusIDs := make(uuid.UUIDs, len(poStatuses))
	for i, s := range poStatuses {
		poStatusIDs[i] = s.ID
	}

	currencies, err := currencybus.TestSeedCurrencies(ctx, 2, busDomain.Currency)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding currencies: %w", err)
	}
	currencyIDs := make(uuid.UUIDs, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	// PO with delivery location set — used for happy-path po_delivery test
	poWithLocation, err := busDomain.PurchaseOrder.Create(ctx, purchaseorderbus.NewPurchaseOrder{
		OrderNumber:             "TEST-PO-WITH-LOC-001",
		SupplierID:              supplierIDs[0],
		PurchaseOrderStatusID:   poStatusIDs[0],
		DeliveryWarehouseID:     warehouseIDs[0],
		DeliveryLocationID:      inventoryLocations[0].LocationID, // set — key for po_delivery test
		DeliveryStreetID:        streetIDs[0],
		OrderDate:               time.Now().UTC(),
		ExpectedDeliveryDate:    time.Now().UTC().Add(time.Hour * 24 * 14),
		Subtotal:                1000.00,
		TaxAmount:               80.00,
		ShippingCost:            50.00,
		TotalAmount:             1130.00,
		CurrencyID:              currencyIDs[0],
		RequestedBy:             adminIDs[0],
		Notes:                   "Test PO with delivery location",
		SupplierReferenceNumber: "SUP-TEST-001",
		CreatedBy:               adminIDs[0],
	})
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("creating PO with location: %w", err)
	}

	// PO without delivery location — used for no_location output port test
	poWithoutLocation, err := busDomain.PurchaseOrder.Create(ctx, purchaseorderbus.NewPurchaseOrder{
		OrderNumber:             "TEST-PO-NO-LOC-002",
		SupplierID:              supplierIDs[0],
		PurchaseOrderStatusID:   poStatusIDs[0],
		DeliveryWarehouseID:     warehouseIDs[0],
		DeliveryLocationID:      uuid.Nil, // no location set
		DeliveryStreetID:        streetIDs[0],
		OrderDate:               time.Now().UTC(),
		ExpectedDeliveryDate:    time.Now().UTC().Add(time.Hour * 24 * 14),
		Subtotal:                500.00,
		TaxAmount:               40.00,
		ShippingCost:            25.00,
		TotalAmount:             565.00,
		CurrencyID:              currencyIDs[0],
		RequestedBy:             adminIDs[0],
		Notes:                   "Test PO without delivery location",
		SupplierReferenceNumber: "SUP-TEST-002",
		CreatedBy:               adminIDs[0],
	})
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("creating PO without location: %w", err)
	}

	// PO line items (for completeness — tests read from RawData directly, not from DB)
	liStatuses, err := purchaseorderlineitemstatusbus.TestSeedPurchaseOrderLineItemStatuses(ctx, 1, busDomain.PurchaseOrderLineItemStatus)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding line item statuses: %w", err)
	}
	liStatusIDs := make(uuid.UUIDs, len(liStatuses))
	for i, s := range liStatuses {
		liStatusIDs[i] = s.ID
	}

	spIDs := make(uuid.UUIDs, len(supplierProducts))
	for i, sp := range supplierProducts {
		spIDs[i] = sp.SupplierProductID
	}
	poIDs := uuid.UUIDs{poWithLocation.ID, poWithoutLocation.ID}

	lineItems, err := purchaseorderlineitembus.TestSeedPurchaseOrderLineItems(ctx, 4, poIDs, spIDs, liStatusIDs, adminIDs, busDomain.PurchaseOrderLineItem)
	if err != nil {
		return createPutAwayTaskSeedData{}, fmt.Errorf("seeding line items: %w", err)
	}

	ruleID := uuid.New()
	execContext := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "purchase_order_line_items",
		EventType:     "on_update",
		UserID:        adminIDs[0],
		RuleID:        &ruleID,
		RuleName:      "Test Create Put-Away Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return createPutAwayTaskSeedData{
		SeedData: unitest.SeedData{
			Admins:             []unitest.User{{User: admins[0]}},
			Products:           products,
			InventoryLocations: inventoryLocations,
		},
		Products:           products,
		InventoryLocations: inventoryLocations,
		SupplierProducts:   supplierProducts,
		POWithLocation:     poWithLocation,
		POWithoutLocation:  poWithoutLocation,
		LineItems:          lineItems,
		ExecutionContext:    execContext,
	}, nil
}

// =============================================================================
// Validate Tests

func createPutAwayTaskValidateTests(sd createPutAwayTaskSeedData) []unitest.Table {
	return []unitest.Table{
		putAwayValidateMissingLocationStrategy(sd),
		putAwayValidateInvalidLocationStrategy(sd),
		putAwayValidateMissingProductIDWhenStatic(sd),
		putAwayValidateInvalidProductID(sd),
		putAwayValidateMissingLocationIDWhenStatic(sd),
		putAwayValidateInvalidLocationID(sd),
		putAwayValidateSourceFromPOPODeliveryValid(sd),
	}
}

func putAwayValidateMissingLocationStrategy(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_location_strategy",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"source_from_po":false,"product_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "location_strategy")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateInvalidLocationStrategy(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_location_strategy",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"bad_value","product_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "location_strategy")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateMissingProductIDWhenStatic(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_product_id_when_static",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"static","location_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "product_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateInvalidProductID(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_product_id",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"static","product_id":"not-a-uuid","location_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "product_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateMissingLocationIDWhenStatic(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "missing_location_id_when_static",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"static","product_id":"` + uuid.New().String() + `"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "location_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateInvalidLocationID(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "invalid_location_id",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			config := json.RawMessage(`{"location_strategy":"static","product_id":"` + uuid.New().String() + `","location_id":"not-a-uuid"}`)
			err := sd.Handler.Validate(config)
			if err == nil {
				return false
			}
			return strings.Contains(err.Error(), "location_id")
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayValidateSourceFromPOPODeliveryValid(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "source_from_po_po_delivery_valid",
		ExpResp: nil,
		ExcFunc: func(ctx context.Context) any {
			// source_from_po=true with po_delivery requires no product_id or location_id
			config := json.RawMessage(`{"source_from_po":true,"location_strategy":"po_delivery"}`)
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

// =============================================================================
// Execute Tests

func createPutAwayTaskExecuteTests(busDomain dbtest.BusDomain, sd createPutAwayTaskSeedData) []unitest.Table {
	return []unitest.Table{
		putAwayExecuteHappyPathStatic(busDomain, sd),
		putAwayExecuteSourceFromPODelivery(busDomain, sd),
		putAwayExecuteNoLocationOnPO(sd),
		putAwayExecuteProductNotFound(sd),
		putAwayExecuteZeroDelta(sd),
		putAwayExecuteNegativeDelta(sd),
		putAwayExecuteTemplateReferenceNumber(busDomain, sd),
	}
}

func putAwayExecuteHappyPathStatic(busDomain dbtest.BusDomain, sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "happy_path_static",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			product := sd.Products[1]
			location := sd.InventoryLocations[1]
			delta := 10

			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     false,
				ProductID:        product.ProductID.String(),
				LocationStrategy: "static",
				LocationID:       location.LocationID.String(),
				ReferenceNumber:  "STATIC-TEST",
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {
					OldValue: float64(0),
					NewValue: float64(delta),
				},
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map[string]any, got %T", result)
			}

			if resultMap["output"] != "created" {
				return fmt.Errorf("expected output=created, got %v", resultMap["output"])
			}

			taskIDStr, ok := resultMap["task_id"].(string)
			if !ok {
				return fmt.Errorf("task_id missing or wrong type: %v", resultMap["task_id"])
			}
			taskID, err := uuid.Parse(taskIDStr)
			if err != nil {
				return fmt.Errorf("task_id not a valid UUID: %s", taskIDStr)
			}

			// Verify task exists in DB with correct values
			task, err := busDomain.PutAwayTask.QueryByID(ctx, taskID)
			if err != nil {
				return fmt.Errorf("querying created task: %w", err)
			}
			if task.ProductID != product.ProductID {
				return fmt.Errorf("product_id mismatch: expected %s, got %s", product.ProductID, task.ProductID)
			}
			if task.LocationID != location.LocationID {
				return fmt.Errorf("location_id mismatch: expected %s, got %s", location.LocationID, task.LocationID)
			}
			if task.Quantity != delta {
				return fmt.Errorf("quantity mismatch: expected %d, got %d", delta, task.Quantity)
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

func putAwayExecuteSourceFromPODelivery(busDomain dbtest.BusDomain, sd createPutAwayTaskSeedData) unitest.Table {
	if len(sd.SupplierProducts) == 0 {
		return unitest.Table{
			Name: "source_from_po_po_delivery_skip", ExpResp: "skipped",
			ExcFunc: func(ctx context.Context) any { return "skipped" },
			CmpFunc: func(got, exp any) string { return "" },
		}
	}
	return unitest.Table{
		Name:    "source_from_po_po_delivery",
		ExpResp: "created",
		ExcFunc: func(ctx context.Context) any {
			sp := sd.SupplierProducts[0]

			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     true,
				LocationStrategy: "po_delivery",
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(0), NewValue: float64(15)},
			}
			execCtx.RawData = map[string]any{
				"supplier_product_id": sp.SupplierProductID.String(),
				"purchase_order_id":   sd.POWithLocation.ID.String(),
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			// Verify location matches po.DeliveryLocationID
			if output := resultMap["output"]; output != "created" {
				return fmt.Errorf("expected output=created, got %v", output)
			}
			taskIDStr, ok := resultMap["task_id"].(string)
			if !ok {
				return fmt.Errorf("task_id missing or wrong type: %v", resultMap["task_id"])
			}
			taskID, _ := uuid.Parse(taskIDStr)
			task, err := busDomain.PutAwayTask.QueryByID(ctx, taskID)
			if err != nil {
				return fmt.Errorf("querying task: %w", err)
			}
			if task.LocationID != sd.POWithLocation.DeliveryLocationID {
				return fmt.Errorf("location_id should match PO.DeliveryLocationID: expected %s, got %s",
					sd.POWithLocation.DeliveryLocationID, task.LocationID)
			}
			if task.ProductID != sp.ProductID {
				return fmt.Errorf("product_id should match supplier product: expected %s, got %s",
					sp.ProductID, task.ProductID)
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

func putAwayExecuteNoLocationOnPO(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "no_location_on_po",
		ExpResp: "no_location",
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     true,
				LocationStrategy: "po_delivery",
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(0), NewValue: float64(5)},
			}
			execCtx.RawData = map[string]any{
				"supplier_product_id": sd.SupplierProducts[0].SupplierProductID.String(),
				"purchase_order_id":   sd.POWithoutLocation.ID.String(), // no DeliveryLocationID
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
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

func putAwayExecuteProductNotFound(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "product_not_found",
		ExpResp: "product_not_found",
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     true,
				LocationStrategy: "po_delivery",
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(0), NewValue: float64(5)},
			}
			execCtx.RawData = map[string]any{
				"supplier_product_id": uuid.New().String(), // unknown — lookup will fail
				"purchase_order_id":   sd.POWithLocation.ID.String(),
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
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

func putAwayExecuteZeroDelta(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "zero_delta",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     false,
				ProductID:        sd.Products[0].ProductID.String(),
				LocationStrategy: "static",
				LocationID:       sd.InventoryLocations[0].LocationID.String(),
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(10), NewValue: float64(10)}, // same = delta 0
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			return resultMap["output"] == "skipped"
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayExecuteNegativeDelta(sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "negative_delta",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     false,
				ProductID:        sd.Products[0].ProductID.String(),
				LocationStrategy: "static",
				LocationID:       sd.InventoryLocations[0].LocationID.String(),
			}
			configJSON, _ := json.Marshal(cfg)

			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(10), NewValue: float64(5)}, // correction: delta = -5
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			return resultMap["output"] == "skipped"
		},
		CmpFunc: func(got, exp any) string {
			if got != exp {
				return fmt.Sprintf("got %v, want %v", got, exp)
			}
			return ""
		},
	}
}

func putAwayExecuteTemplateReferenceNumber(busDomain dbtest.BusDomain, sd createPutAwayTaskSeedData) unitest.Table {
	return unitest.Table{
		Name:    "template_reference_number",
		ExpResp: true,
		ExcFunc: func(ctx context.Context) any {
			cfg := inventory.CreatePutAwayTaskConfig{
				SourceFromPO:     false,
				ProductID:        sd.Products[2].ProductID.String(),
				LocationStrategy: "static",
				LocationID:       sd.InventoryLocations[2].LocationID.String(),
				ReferenceNumber:  "PO-RCV-{{purchase_order_id}}",
			}
			configJSON, _ := json.Marshal(cfg)

			poID := sd.POWithLocation.ID.String()
			execCtx := sd.ExecutionContext
			execCtx.ExecutionID = uuid.New()
			execCtx.FieldChanges = map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(0), NewValue: float64(7)},
			}
			execCtx.RawData = map[string]any{
				"purchase_order_id": poID,
			}

			result, err := sd.Handler.Execute(ctx, configJSON, execCtx)
			if err != nil {
				return fmt.Errorf("execute failed: %w", err)
			}
			resultMap, ok := result.(map[string]any)
			if !ok {
				return fmt.Errorf("expected map, got %T", result)
			}
			if resultMap["output"] != "created" {
				return fmt.Errorf("expected output=created, got %v", resultMap["output"])
			}

			taskIDStr, ok := resultMap["task_id"].(string)
			if !ok {
				return fmt.Errorf("task_id missing or wrong type: %v", resultMap["task_id"])
			}
			taskID, _ := uuid.Parse(taskIDStr)
			task, err := busDomain.PutAwayTask.QueryByID(ctx, taskID)
			if err != nil {
				return fmt.Errorf("querying task: %w", err)
			}

			expected := "PO-RCV-" + poID
			if task.ReferenceNumber != expected {
				return fmt.Errorf("reference_number mismatch: expected %q, got %q", expected, task.ReferenceNumber)
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
