package procurement_test

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
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/procurement"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func TestCreatePurchaseOrder_Validate(t *testing.T) {
	handler := procurement.NewCreatePurchaseOrderHandler(nil, nil, nil, nil, nil)

	validStatusID := uuid.New().String()
	validWarehouseID := uuid.New().String()
	validLocationID := uuid.New().String()
	validCurrencyID := uuid.New().String()
	validProductID := uuid.New().String()
	validLineItemStatusID := uuid.New().String()

	validConfig := procurement.CreatePurchaseOrderConfig{
		PurchaseOrderStatusID: validStatusID,
		DeliveryWarehouseID:   validWarehouseID,
		DeliveryLocationID:    validLocationID,
		CurrencyID:            validCurrencyID,
		LineItems: []procurement.CreatePOLineItemConfig{
			{
				ProductID:        validProductID,
				QuantityOrdered:  10,
				LineItemStatusID: validLineItemStatusID,
			},
		},
	}

	// copyLineItems returns a deep copy of the line items slice so mutations
	// in one test case don't bleed into subsequent cases.
	copyLineItems := func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
		if len(c.LineItems) > 0 {
			cp := make([]procurement.CreatePOLineItemConfig, len(c.LineItems))
			copy(cp, c.LineItems)
			c.LineItems = cp
		}
		return c
	}

	tests := []struct {
		name      string
		modify    func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid config",
			modify:  func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig { return c },
			wantErr: false,
		},
		{
			name: "missing purchase_order_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.PurchaseOrderStatusID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "purchase_order_status_id is required",
		},
		{
			name: "invalid purchase_order_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.PurchaseOrderStatusID = "not-a-uuid"
				return c
			},
			wantErr:   true,
			errSubstr: "invalid purchase_order_status_id",
		},
		{
			name: "missing delivery_warehouse_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.DeliveryWarehouseID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "delivery_warehouse_id is required",
		},
		{
			name: "missing delivery_location_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.DeliveryLocationID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "delivery_location_id is required",
		},
		{
			name: "missing currency_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.CurrencyID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "currency_id is required",
		},
		{
			name: "no line items when source_from_event is false",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.LineItems = nil
				return c
			},
			wantErr:   true,
			errSubstr: "at least one line item is required",
		},
		{
			name: "source_from_event without default_line_item_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.SourceFromEvent = true
				c.LineItems = nil
				return c
			},
			wantErr:   true,
			errSubstr: "default_line_item_status_id is required",
		},
		{
			name: "source_from_event with valid default_line_item_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.SourceFromEvent = true
				c.DefaultLineItemStatusID = validLineItemStatusID
				c.LineItems = nil
				return c
			},
			wantErr: false,
		},
		{
			name: "line item missing product_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c = copyLineItems(c)
				c.LineItems[0].ProductID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "product_id is required",
		},
		{
			name: "line item zero quantity",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c = copyLineItems(c)
				c.LineItems[0].QuantityOrdered = 0
				return c
			},
			wantErr:   true,
			errSubstr: "quantity_ordered must be greater than 0",
		},
		{
			name: "line item negative quantity",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c = copyLineItems(c)
				c.LineItems[0].QuantityOrdered = -5
				return c
			},
			wantErr:   true,
			errSubstr: "quantity_ordered must be greater than 0",
		},
		{
			name: "line item missing line_item_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c = copyLineItems(c)
				c.LineItems[0].LineItemStatusID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "line_item_status_id is required",
		},
		{
			name:      "invalid json",
			modify:    nil, // special case
			wantErr:   true,
			errSubstr: "invalid configuration format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configBytes json.RawMessage
			if tt.modify == nil {
				configBytes = json.RawMessage(`{invalid`)
			} else {
				cfg := tt.modify(validConfig)
				data, _ := json.Marshal(cfg)
				configBytes = data
			}

			err := handler.Validate(configBytes)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errSubstr)
				}
			}
		})
	}
}

func TestCreatePurchaseOrder_Metadata(t *testing.T) {
	handler := procurement.NewCreatePurchaseOrderHandler(nil, nil, nil, nil, nil)

	t.Run("GetType", func(t *testing.T) {
		if got := handler.GetType(); got != "create_purchase_order" {
			t.Fatalf("expected create_purchase_order, got %s", got)
		}
	})

	t.Run("SupportsManualExecution", func(t *testing.T) {
		if !handler.SupportsManualExecution() {
			t.Fatal("expected true")
		}
	})

	t.Run("IsAsync", func(t *testing.T) {
		if handler.IsAsync() {
			t.Fatal("expected false")
		}
	})

	t.Run("GetDescription", func(t *testing.T) {
		desc := handler.GetDescription()
		if desc == "" {
			t.Fatal("expected non-empty description")
		}
	})

	t.Run("GetOutputPorts", func(t *testing.T) {
		ports := handler.GetOutputPorts()
		if len(ports) != 3 {
			t.Fatalf("expected 3 output ports, got %d", len(ports))
		}
		portNames := make(map[string]bool)
		for _, p := range ports {
			portNames[p.Name] = true
		}
		for _, expected := range []string{"created", "no_supplier_found", "failure"} {
			if !portNames[expected] {
				t.Fatalf("missing output port: %s", expected)
			}
		}
	})

	t.Run("GetEntityModifications", func(t *testing.T) {
		mods := handler.GetEntityModifications(nil)
		if len(mods) != 2 {
			t.Fatalf("expected 2 entity modifications, got %d", len(mods))
		}
	})
}

// =============================================================================
// Execute tests — require real Postgres with seeded procurement data.
// =============================================================================

type createPOSeedData struct {
	Handler          *procurement.CreatePurchaseOrderHandler
	Admin            userbus.User
	SupplierProducts []supplierproductbus.SupplierProduct
	POStatuses       []purchaseorderstatusbus.PurchaseOrderStatus
	LIStatuses       []purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus
	Warehouses       []warehousebus.Warehouse
	Currencies       []currencybus.Currency
	ExecCtx          workflow.ActionExecutionContext
}

func insertCreatePOSeedData(db *dbtest.Database) (createPOSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	regions, err := db.BusDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, db.BusDomain.City)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	cityIDs := make([]uuid.UUID, len(cities))
	for i, c := range cities {
		cityIDs[i] = c.ID
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, db.BusDomain.Street)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	streetIDs := make(uuid.UUIDs, len(streets))
	for i, s := range streets {
		streetIDs[i] = s.ID
	}

	tzs, err := db.BusDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, db.BusDomain.ContactInfos)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}
	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 2, contactIDs, db.BusDomain.Supplier)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding suppliers: %w", err)
	}
	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, db.BusDomain.Brand)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding brands: %w", err)
	}
	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, db.BusDomain.ProductCategory)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}
	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 3, brandIDs, productCategoryIDs, db.BusDomain.Product)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding products: %w", err)
	}
	productIDs := make(uuid.UUIDs, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 3, productIDs, supplierIDs, db.BusDomain.SupplierProduct)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding supplier products: %w", err)
	}

	poStatuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 2, db.BusDomain.PurchaseOrderStatus)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding PO statuses: %w", err)
	}

	liStatuses, err := purchaseorderlineitemstatusbus.TestSeedPurchaseOrderLineItemStatuses(ctx, 2, db.BusDomain.PurchaseOrderLineItemStatus)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding LI statuses: %w", err)
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, admins[0].ID, streetIDs, db.BusDomain.Warehouse)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}

	currencies, err := currencybus.TestSeedCurrencies(ctx, 2, db.BusDomain.Currency)
	if err != nil {
		return createPOSeedData{}, fmt.Errorf("seeding currencies: %w", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	handler := procurement.NewCreatePurchaseOrderHandler(
		log,
		db.DB,
		db.BusDomain.PurchaseOrder,
		db.BusDomain.PurchaseOrderLineItem,
		db.BusDomain.SupplierProduct,
	)

	ruleID := uuid.New()
	execCtx := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "procurement.purchase_orders",
		EventType:     "on_create",
		UserID:        admins[0].ID,
		RuleID:        &ruleID,
		RuleName:      "Test Create PO Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return createPOSeedData{
		Handler:          handler,
		Admin:            admins[0],
		SupplierProducts: supplierProducts,
		POStatuses:       poStatuses,
		LIStatuses:       liStatuses,
		Warehouses:       warehouses,
		Currencies:       currencies,
		ExecCtx:          execCtx,
	}, nil
}

func Test_CreatePurchaseOrder_Execute(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_CreatePO_Execute")

	sd, err := insertCreatePOSeedData(db)
	if err != nil {
		t.Fatalf("seeding: %v", err)
	}

	// Use the first supplier product's ProductID — guaranteed to have a supplier mapping.
	knownProductID := sd.SupplierProducts[0].ProductID

	t.Run("happy_path", func(t *testing.T) {
		cfg := procurement.CreatePurchaseOrderConfig{
			PurchaseOrderStatusID: sd.POStatuses[0].ID.String(),
			DeliveryWarehouseID:   sd.Warehouses[0].ID.String(),
			DeliveryLocationID:    uuid.Nil.String(),
			CurrencyID:            sd.Currencies[0].ID.String(),
			LineItems: []procurement.CreatePOLineItemConfig{
				{
					ProductID:        knownProductID.String(),
					QuantityOrdered:  5,
					LineItemStatusID: sd.LIStatuses[0].ID.String(),
				},
			},
		}
		configJSON, _ := json.Marshal(cfg)

		result, err := sd.Handler.Execute(context.Background(), configJSON, sd.ExecCtx)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", result)
		}

		if resultMap["output"] != "created" {
			t.Fatalf("expected output=created, got %v", resultMap["output"])
		}
		if resultMap["purchase_order_id"] == nil || resultMap["purchase_order_id"] == "" {
			t.Fatal("expected non-empty purchase_order_id")
		}
		lineItemIDs, ok := resultMap["line_item_ids"].([]string)
		if !ok || len(lineItemIDs) != 1 {
			t.Fatalf("expected 1 line_item_id, got %v", resultMap["line_item_ids"])
		}
	})

	t.Run("no_supplier_found", func(t *testing.T) {
		cfg := procurement.CreatePurchaseOrderConfig{
			PurchaseOrderStatusID: sd.POStatuses[0].ID.String(),
			DeliveryWarehouseID:   sd.Warehouses[0].ID.String(),
			DeliveryLocationID:    uuid.Nil.String(),
			CurrencyID:            sd.Currencies[0].ID.String(),
			LineItems: []procurement.CreatePOLineItemConfig{
				{
					ProductID:        uuid.New().String(), // No supplier product for this UUID.
					QuantityOrdered:  5,
					LineItemStatusID: sd.LIStatuses[0].ID.String(),
				},
			},
		}
		configJSON, _ := json.Marshal(cfg)

		result, err := sd.Handler.Execute(context.Background(), configJSON, sd.ExecCtx)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", result)
		}
		if resultMap["output"] != "no_supplier_found" {
			t.Fatalf("expected output=no_supplier_found, got %v", resultMap["output"])
		}
	})

	// =========================================================================
	// extractFromEvent tests — exercise unexported extractFromEvent through Execute.
	// =========================================================================

	t.Run("extract_from_event_with_quantity", func(t *testing.T) {
		cfg := procurement.CreatePurchaseOrderConfig{
			PurchaseOrderStatusID:   sd.POStatuses[0].ID.String(),
			DeliveryWarehouseID:     sd.Warehouses[0].ID.String(),
			DeliveryLocationID:      uuid.Nil.String(),
			CurrencyID:              sd.Currencies[0].ID.String(),
			SourceFromEvent:         true,
			DefaultLineItemStatusID: sd.LIStatuses[0].ID.String(),
		}
		configJSON, _ := json.Marshal(cfg)

		ruleID := uuid.New()
		eventCtx := workflow.ActionExecutionContext{
			EntityID:      uuid.New(),
			EntityName:    "inventory.inventory_items",
			EventType:     "on_update",
			UserID:        sd.Admin.ID,
			RuleID:        &ruleID,
			RuleName:      "Test Reorder Rule",
			ExecutionID:   uuid.New(),
			Timestamp:     time.Now().UTC(),
			TriggerSource: workflow.TriggerSourceAutomation,
			RawData: map[string]interface{}{
				"product_id": knownProductID.String(),
				"quantity":   float64(10),
			},
		}

		result, err := sd.Handler.Execute(context.Background(), configJSON, eventCtx)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", result)
		}
		if resultMap["output"] != "created" {
			t.Fatalf("expected output=created, got %v (error: %v)", resultMap["output"], resultMap["error"])
		}
	})

	t.Run("extract_from_event_with_reorder_quantity", func(t *testing.T) {
		cfg := procurement.CreatePurchaseOrderConfig{
			PurchaseOrderStatusID:   sd.POStatuses[0].ID.String(),
			DeliveryWarehouseID:     sd.Warehouses[0].ID.String(),
			DeliveryLocationID:      uuid.Nil.String(),
			CurrencyID:              sd.Currencies[0].ID.String(),
			SourceFromEvent:         true,
			DefaultLineItemStatusID: sd.LIStatuses[0].ID.String(),
		}
		configJSON, _ := json.Marshal(cfg)

		ruleID := uuid.New()
		eventCtx := workflow.ActionExecutionContext{
			EntityID:      uuid.New(),
			EntityName:    "inventory.inventory_items",
			EventType:     "on_update",
			UserID:        sd.Admin.ID,
			RuleID:        &ruleID,
			RuleName:      "Test Reorder Rule",
			ExecutionID:   uuid.New(),
			Timestamp:     time.Now().UTC(),
			TriggerSource: workflow.TriggerSourceAutomation,
			RawData: map[string]interface{}{
				"product_id":       knownProductID.String(),
				"reorder_quantity": float64(20),
			},
		}

		result, err := sd.Handler.Execute(context.Background(), configJSON, eventCtx)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", result)
		}
		if resultMap["output"] != "created" {
			t.Fatalf("expected output=created, got %v (error: %v)", resultMap["output"], resultMap["error"])
		}
	})

	t.Run("extract_from_event_missing_product_id", func(t *testing.T) {
		cfg := procurement.CreatePurchaseOrderConfig{
			PurchaseOrderStatusID:   sd.POStatuses[0].ID.String(),
			DeliveryWarehouseID:     sd.Warehouses[0].ID.String(),
			DeliveryLocationID:      uuid.Nil.String(),
			CurrencyID:              sd.Currencies[0].ID.String(),
			SourceFromEvent:         true,
			DefaultLineItemStatusID: sd.LIStatuses[0].ID.String(),
		}
		configJSON, _ := json.Marshal(cfg)

		ruleID := uuid.New()
		eventCtx := workflow.ActionExecutionContext{
			EntityID:      uuid.New(),
			EntityName:    "inventory.inventory_items",
			EventType:     "on_update",
			UserID:        sd.Admin.ID,
			RuleID:        &ruleID,
			RuleName:      "Test Reorder Rule",
			ExecutionID:   uuid.New(),
			Timestamp:     time.Now().UTC(),
			TriggerSource: workflow.TriggerSourceAutomation,
			RawData: map[string]interface{}{
				"quantity": float64(10),
			},
		}

		result, err := sd.Handler.Execute(context.Background(), configJSON, eventCtx)
		if err != nil {
			t.Fatalf("unexpected hard error: %s", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", result)
		}
		if resultMap["output"] != "failure" {
			t.Fatalf("expected output=failure, got %v", resultMap["output"])
		}
	})

	t.Run("extract_from_event_missing_quantity", func(t *testing.T) {
		cfg := procurement.CreatePurchaseOrderConfig{
			PurchaseOrderStatusID:   sd.POStatuses[0].ID.String(),
			DeliveryWarehouseID:     sd.Warehouses[0].ID.String(),
			DeliveryLocationID:      uuid.Nil.String(),
			CurrencyID:              sd.Currencies[0].ID.String(),
			SourceFromEvent:         true,
			DefaultLineItemStatusID: sd.LIStatuses[0].ID.String(),
		}
		configJSON, _ := json.Marshal(cfg)

		ruleID := uuid.New()
		eventCtx := workflow.ActionExecutionContext{
			EntityID:      uuid.New(),
			EntityName:    "inventory.inventory_items",
			EventType:     "on_update",
			UserID:        sd.Admin.ID,
			RuleID:        &ruleID,
			RuleName:      "Test Reorder Rule",
			ExecutionID:   uuid.New(),
			Timestamp:     time.Now().UTC(),
			TriggerSource: workflow.TriggerSourceAutomation,
			RawData: map[string]interface{}{
				"product_id": knownProductID.String(),
			},
		}

		result, err := sd.Handler.Execute(context.Background(), configJSON, eventCtx)
		if err != nil {
			t.Fatalf("unexpected hard error: %s", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", result)
		}
		if resultMap["output"] != "failure" {
			t.Fatalf("expected output=failure, got %v", resultMap["output"])
		}
	})

	t.Run("extract_from_event_zero_quantity", func(t *testing.T) {
		cfg := procurement.CreatePurchaseOrderConfig{
			PurchaseOrderStatusID:   sd.POStatuses[0].ID.String(),
			DeliveryWarehouseID:     sd.Warehouses[0].ID.String(),
			DeliveryLocationID:      uuid.Nil.String(),
			CurrencyID:              sd.Currencies[0].ID.String(),
			SourceFromEvent:         true,
			DefaultLineItemStatusID: sd.LIStatuses[0].ID.String(),
		}
		configJSON, _ := json.Marshal(cfg)

		ruleID := uuid.New()
		eventCtx := workflow.ActionExecutionContext{
			EntityID:      uuid.New(),
			EntityName:    "inventory.inventory_items",
			EventType:     "on_update",
			UserID:        sd.Admin.ID,
			RuleID:        &ruleID,
			RuleName:      "Test Reorder Rule",
			ExecutionID:   uuid.New(),
			Timestamp:     time.Now().UTC(),
			TriggerSource: workflow.TriggerSourceAutomation,
			RawData: map[string]interface{}{
				"product_id": knownProductID.String(),
				"quantity":   float64(0),
			},
		}

		result, err := sd.Handler.Execute(context.Background(), configJSON, eventCtx)
		if err != nil {
			t.Fatalf("unexpected hard error: %s", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", result)
		}
		if resultMap["output"] != "failure" {
			t.Fatalf("expected output=failure, got %v", resultMap["output"])
		}
	})
}
