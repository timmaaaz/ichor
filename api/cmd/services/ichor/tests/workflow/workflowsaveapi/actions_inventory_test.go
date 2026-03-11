// Package workflowsaveapi_test contains standalone integration tests for inventory
// and procurement workflow actions using a real Temporal container.
//
// These tests spin up a minimal Temporal worker registered with ONLY the handler
// under test (plus core workflows). They do NOT use InitWorkflowInfra because
// the inventory/procurement handlers require bus dependencies that the shared
// infra helper does not wire up.
//
// Tests use the same Temporal test container as other workflow tests (shared via
// foundationtemporal.GetTestContainer) for efficient reuse.
package workflowsaveapi_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/procurement"
	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// =============================================================================
// receive_inventory Action Tests
// =============================================================================

// TestReceiveInventoryAction verifies the receive_inventory handler end-to-end:
//  1. Seeds an inventory item with a known product + location
//  2. Creates a workflow rule with a receive_inventory action containing a static config
//  3. Fires a trigger event
//  4. Polls for an inbound inventory transaction created as a side effect
//  5. Verifies the inventory item quantity increased
func TestReceiveInventoryAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ReceiveInventoryAction")
	ctx := context.Background()

	// ── Seed prerequisites ────────────────────────────────────────────────────

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		t.Fatalf("seeding admin users: %v", err)
	}
	createdBy := admins[0].ID

	// Regions → cities → streets for warehouse address chain.
	regions, err := db.BusDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "3"))
	if err != nil || len(regions) == 0 {
		t.Fatalf("querying seeded regions: %v", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 2, regionIDs, db.BusDomain.City)
	if err != nil {
		t.Fatalf("seeding cities: %v", err)
	}
	cityIDs := make([]uuid.UUID, len(cities))
	for i, c := range cities {
		cityIDs[i] = c.ID
	}

	streets, err := streetbus.TestSeedStreets(ctx, 2, cityIDs, db.BusDomain.Street)
	if err != nil {
		t.Fatalf("seeding streets: %v", err)
	}
	streetIDs := make([]uuid.UUID, len(streets))
	for i, s := range streets {
		streetIDs[i] = s.ID
	}

	// Warehouse → zone → inventory location.
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 1, createdBy, streetIDs, db.BusDomain.Warehouse)
	if err != nil {
		t.Fatalf("seeding warehouses: %v", err)
	}
	warehouseIDs := []uuid.UUID{warehouses[0].ID}

	zones, err := zonebus.TestSeedZone(ctx, 1, warehouseIDs, db.BusDomain.Zones)
	if err != nil {
		t.Fatalf("seeding zones: %v", err)
	}
	zoneIDs := []uuid.UUID{zones[0].ZoneID}

	locations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 1, warehouseIDs, zoneIDs, db.BusDomain.InventoryLocation)
	if err != nil {
		t.Fatalf("seeding inventory locations: %v", err)
	}
	locationID := locations[0].LocationID

	// Products: query seeded products from migrations.
	products, err := db.BusDomain.Product.Query(ctx, productbus.QueryFilter{}, productbus.DefaultOrderBy, page.MustParse("1", "2"))
	if err != nil {
		t.Fatalf("querying seeded products: %v", err)
	}
	if len(products) == 0 {
		t.Skip("requires seeded products — run database migrations and seeding first")
	}
	productID := products[0].ProductID

	// Inventory item at known product + location.
	items, err := inventoryitembus.TestSeedInventoryItems(ctx, 1, []uuid.UUID{locationID}, []uuid.UUID{productID}, db.BusDomain.InventoryItem)
	if err != nil {
		t.Fatalf("seeding inventory items: %v", err)
	}
	item := items[0]
	initialQty := item.Quantity

	// ── Set up minimal Temporal worker ────────────────────────────────────────

	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
	if err != nil {
		t.Fatalf("connecting to Temporal: %v", err)
	}

	taskQueue := "test-workflow-receive-inv-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflowtemporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(workflowtemporal.ExecuteBranchUntilConvergence)

	registry := workflow.NewActionRegistry()
	registry.Register(inventory.NewReceiveInventoryHandler(
		db.Log,
		db.DB,
		db.BusDomain.InventoryItem,
		db.BusDomain.InventoryTransaction,
		db.BusDomain.SupplierProduct,
	))

	activities := &workflowtemporal.Activities{
		Registry:      registry,
		AsyncRegistry: workflowtemporal.NewAsyncRegistry(),
	}
	w.RegisterActivity(activities)

	if err := w.Start(); err != nil {
		tc.Close()
		t.Fatalf("starting Temporal worker: %v", err)
	}
	t.Cleanup(func() { w.Stop(); tc.Close() })

	// ── Set up workflow infra ─────────────────────────────────────────────────

	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
	edgeStore := edgedb.NewStore(db.Log, db.DB)
	triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
	if err := triggerProcessor.Initialize(ctx); err != nil {
		t.Fatalf("initializing trigger processor: %v", err)
	}
	workflowTrigger := workflowtemporal.NewWorkflowTrigger(db.Log, tc, triggerProcessor, edgeStore, workflowStore).
		WithTaskQueue(taskQueue)

	// ── Seed workflow rule + action + edge ────────────────────────────────────

	customerEntity, err := workflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("querying customers entity: %v", err)
	}
	entityType, err := workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %v", err)
	}
	triggerTypeCreate, err := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %v", err)
	}

	rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "ReceiveInventory-Test-" + uuid.New().String()[:8],
		Description:   "Integration test for receive_inventory action",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerTypeCreate.ID,
		IsActive:      true,
		CreatedBy:     createdBy,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	// receive_inventory config: static product_id + location_id, receive 50 units.
	receiveQty := 50
	actionConfig := map[string]any{
		"product_id":  productID.String(),
		"location_id": locationID.String(),
		"quantity":    receiveQty,
		"notes":       "integration test receipt",
	}
	configBytes, _ := json.Marshal(actionConfig)

	recvTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "Receive Inventory Template",
		Description:   "Template for receive_inventory test",
		ActionType:    "receive_inventory",
		DefaultConfig: json.RawMessage(configBytes),
		CreatedBy:     createdBy,
	})
	if err != nil {
		t.Fatalf("creating receive_inventory template: %v", err)
	}

	action, err := workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Receive 50 units",
		Description:      "Receive 50 units into inventory",
		ActionConfig:     json.RawMessage(configBytes),
		IsActive:         true,
		TemplateID:       &recvTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating receive_inventory action: %v", err)
	}

	_, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	// Refresh trigger processor so the new rule is matched.
	if err := triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing trigger processor: %v", err)
	}

	// ── Fire trigger event ────────────────────────────────────────────────────

	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"test": true},
		UserID:     createdBy,
	}
	if err := workflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger event: %v", err)
	}

	// ── Poll for inventory transaction created by receive_inventory ───────────

	txType := "inbound"
	var txRecord inventorytransactionbus.InventoryTransaction
	for i := 0; i < 30; i++ {
		txs, err := db.BusDomain.InventoryTransaction.Query(
			ctx,
			inventorytransactionbus.QueryFilter{
				ProductID:       &productID,
				LocationID:      &locationID,
				TransactionType: &txType,
			},
			inventorytransactionbus.DefaultOrderBy,
			page.MustParse("1", "10"),
		)
		if err != nil {
			t.Fatalf("polling inventory transactions: %v", err)
		}
		if len(txs) > 0 {
			txRecord = txs[0]
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if txRecord.InventoryTransactionID == uuid.Nil {
		t.Fatal("timeout: no inbound inventory transaction found after 15s — receive_inventory may have failed")
	}

	t.Logf("inventory transaction created: id=%s quantity=%d", txRecord.InventoryTransactionID, txRecord.Quantity)

	// Verify transaction fields.
	if txRecord.Quantity != receiveQty {
		t.Errorf("expected transaction quantity %d, got %d", receiveQty, txRecord.Quantity)
	}
	if txRecord.ProductID != productID {
		t.Errorf("expected product_id %s, got %s", productID, txRecord.ProductID)
	}
	if txRecord.LocationID != locationID {
		t.Errorf("expected location_id %s, got %s", locationID, txRecord.LocationID)
	}

	// Verify the inventory item quantity increased.
	updatedItems, err := db.BusDomain.InventoryItem.Query(
		ctx,
		inventoryitembus.QueryFilter{
			ProductID:  &productID,
			LocationID: &locationID,
		},
		inventoryitembus.DefaultOrderBy,
		page.MustParse("1", "1"),
	)
	if err != nil || len(updatedItems) == 0 {
		t.Fatalf("querying updated inventory item: %v", err)
	}
	updatedQty := updatedItems[0].Quantity
	expectedQty := initialQty + receiveQty
	if updatedQty != expectedQty {
		t.Errorf("expected inventory quantity %d (initial %d + received %d), got %d", expectedQty, initialQty, receiveQty, updatedQty)
	}

	t.Logf("SUCCESS: receive_inventory verified — initial=%d received=%d final=%d tx=%s",
		initialQty, receiveQty, updatedQty, txRecord.InventoryTransactionID)
}

// =============================================================================
// create_purchase_order Action Tests
// =============================================================================

// TestCreatePurchaseOrderAction verifies the create_purchase_order handler end-to-end:
//  1. Seeds suppliers, supplier products, PO statuses, warehouses, currencies
//  2. Creates a workflow rule with a create_purchase_order action (explicit line items)
//  3. Fires a trigger event
//  4. Polls for a purchase order created as a side effect
//  5. Verifies the PO was created with the expected supplier and line item count
func TestCreatePurchaseOrderAction(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CreatePurchaseOrderAction")
	ctx := context.Background()

	// ── Seed prerequisites ────────────────────────────────────────────────────

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		t.Fatalf("seeding admin users: %v", err)
	}
	createdBy := admins[0].ID

	// Regions → cities → streets for supplier contact info and warehouse address.
	regions, err := db.BusDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "3"))
	if err != nil || len(regions) == 0 {
		t.Fatalf("querying seeded regions: %v", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 2, regionIDs, db.BusDomain.City)
	if err != nil {
		t.Fatalf("seeding cities: %v", err)
	}
	cityIDs := make([]uuid.UUID, len(cities))
	for i, c := range cities {
		cityIDs[i] = c.ID
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, db.BusDomain.Street)
	if err != nil {
		t.Fatalf("seeding streets: %v", err)
	}
	streetIDs := make([]uuid.UUID, len(streets))
	for i, s := range streets {
		streetIDs[i] = s.ID
	}

	// Timezones for supplier contact info.
	tzs, err := db.BusDomain.Timezone.QueryAll(ctx)
	if err != nil || len(tzs) == 0 {
		t.Fatalf("querying timezones: %v", err)
	}
	tzIDs := make([]uuid.UUID, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	// Contact infos for suppliers.
	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 2, streetIDs, tzIDs, db.BusDomain.ContactInfos)
	if err != nil {
		t.Fatalf("seeding contact infos: %v", err)
	}
	contactInfoIDs := make([]uuid.UUID, len(contactInfos))
	for i, ci := range contactInfos {
		contactInfoIDs[i] = ci.ID
	}

	// Suppliers.
	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 1, contactInfoIDs, db.BusDomain.Supplier)
	if err != nil {
		t.Fatalf("seeding suppliers: %v", err)
	}
	supplierID := suppliers[0].SupplierID

	// Products: query seeded products from migrations.
	products, err := db.BusDomain.Product.Query(ctx, productbus.QueryFilter{}, productbus.DefaultOrderBy, page.MustParse("1", "2"))
	if err != nil {
		t.Fatalf("querying seeded products: %v", err)
	}
	if len(products) == 0 {
		t.Skip("requires seeded products — run database migrations and seeding first")
	}
	productID := products[0].ProductID

	// Supplier products: link the supplier to the product.
	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 1, []uuid.UUID{productID}, []uuid.UUID{supplierID}, db.BusDomain.SupplierProduct)
	if err != nil {
		t.Fatalf("seeding supplier products: %v", err)
	}
	supplierProductID := supplierProducts[0].SupplierProductID

	// Warehouses.
	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 1, createdBy, streetIDs, db.BusDomain.Warehouse)
	if err != nil {
		t.Fatalf("seeding warehouses: %v", err)
	}
	warehouseID := warehouses[0].ID

	// Warehouse → zone → inventory location for delivery_location_id.
	warehouseIDs := []uuid.UUID{warehouseID}
	zones, err := zonebus.TestSeedZone(ctx, 1, warehouseIDs, db.BusDomain.Zones)
	if err != nil {
		t.Fatalf("seeding zones: %v", err)
	}
	zoneIDs := []uuid.UUID{zones[0].ZoneID}

	locations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 1, warehouseIDs, zoneIDs, db.BusDomain.InventoryLocation)
	if err != nil {
		t.Fatalf("seeding inventory locations: %v", err)
	}
	locationID := locations[0].LocationID

	// PO statuses.
	poStatuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 1, db.BusDomain.PurchaseOrderStatus)
	if err != nil {
		t.Fatalf("seeding purchase order statuses: %v", err)
	}
	poStatusID := poStatuses[0].ID

	// Line item statuses.
	lineItemStatuses, err := purchaseorderlineitemstatusbus.TestSeedPurchaseOrderLineItemStatuses(ctx, 1, db.BusDomain.PurchaseOrderLineItemStatus)
	if err != nil {
		t.Fatalf("seeding purchase order line item statuses: %v", err)
	}
	lineItemStatusID := lineItemStatuses[0].ID

	// Currencies.
	currencies, err := currencybus.TestSeedCurrencies(ctx, 1, db.BusDomain.Currency)
	if err != nil {
		t.Fatalf("seeding currencies: %v", err)
	}
	currencyID := currencies[0].ID

	// ── Set up minimal Temporal worker ────────────────────────────────────────

	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
	if err != nil {
		t.Fatalf("connecting to Temporal: %v", err)
	}

	taskQueue := "test-workflow-create-po-" + t.Name()
	w := worker.New(tc, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflowtemporal.ExecuteGraphWorkflow)
	w.RegisterWorkflow(workflowtemporal.ExecuteBranchUntilConvergence)

	registry := workflow.NewActionRegistry()
	registry.Register(procurement.NewCreatePurchaseOrderHandler(
		db.Log,
		db.DB,
		db.BusDomain.PurchaseOrder,
		db.BusDomain.PurchaseOrderLineItem,
		db.BusDomain.SupplierProduct,
	))

	activities := &workflowtemporal.Activities{
		Registry:      registry,
		AsyncRegistry: workflowtemporal.NewAsyncRegistry(),
	}
	w.RegisterActivity(activities)

	if err := w.Start(); err != nil {
		tc.Close()
		t.Fatalf("starting Temporal worker: %v", err)
	}
	t.Cleanup(func() { w.Stop(); tc.Close() })

	// ── Set up workflow infra ─────────────────────────────────────────────────

	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
	edgeStore := edgedb.NewStore(db.Log, db.DB)
	triggerProcessor := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
	if err := triggerProcessor.Initialize(ctx); err != nil {
		t.Fatalf("initializing trigger processor: %v", err)
	}
	workflowTrigger := workflowtemporal.NewWorkflowTrigger(db.Log, tc, triggerProcessor, edgeStore, workflowStore).
		WithTaskQueue(taskQueue)

	// ── Seed workflow rule + action + edge ────────────────────────────────────

	customerEntity, err := workflowBus.QueryEntityByName(ctx, "customers")
	if err != nil {
		t.Fatalf("querying customers entity: %v", err)
	}
	entityType, err := workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %v", err)
	}
	triggerTypeCreate, err := workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %v", err)
	}

	rule, err := workflowBus.CreateRule(ctx, workflow.NewAutomationRule{
		Name:          "CreatePO-Test-" + uuid.New().String()[:8],
		Description:   "Integration test for create_purchase_order action",
		EntityID:      customerEntity.ID,
		EntityTypeID:  entityType.ID,
		TriggerTypeID: triggerTypeCreate.ID,
		IsActive:      true,
		CreatedBy:     createdBy,
	})
	if err != nil {
		t.Fatalf("creating rule: %v", err)
	}

	// create_purchase_order config: explicit supplier and one line item.
	actionConfig := map[string]any{
		"supplier_id":               supplierID.String(),
		"purchase_order_status_id":  poStatusID.String(),
		"delivery_warehouse_id":     warehouseID.String(),
		"delivery_location_id":      locationID.String(),
		"currency_id":               currencyID.String(),
		"expected_delivery_days":    7,
		"notes":                     "integration test PO",
		"line_items": []map[string]any{
			{
				"product_id":          productID.String(),
				"supplier_product_id": supplierProductID.String(),
				"quantity_ordered":    10,
				"unit_cost":           25.00,
				"line_item_status_id": lineItemStatusID.String(),
				"notes":               "test line item",
			},
		},
	}
	configBytes, _ := json.Marshal(actionConfig)

	poTemplate, err := workflowBus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "Create PO Template",
		Description:   "Template for create_purchase_order test",
		ActionType:    "create_purchase_order",
		DefaultConfig: json.RawMessage(configBytes),
		CreatedBy:     createdBy,
	})
	if err != nil {
		t.Fatalf("creating create_purchase_order template: %v", err)
	}

	action, err := workflowBus.CreateRuleAction(ctx, workflow.NewRuleAction{
		AutomationRuleID: rule.ID,
		Name:             "Create Test PO",
		Description:      "Creates a purchase order for integration test",
		ActionConfig:     json.RawMessage(configBytes),
		IsActive:         true,
		TemplateID:       &poTemplate.ID,
	})
	if err != nil {
		t.Fatalf("creating create_purchase_order action: %v", err)
	}

	_, err = workflowBus.CreateActionEdge(ctx, workflow.NewActionEdge{
		RuleID:         rule.ID,
		SourceActionID: nil,
		TargetActionID: action.ID,
		EdgeType:       "start",
		EdgeOrder:      0,
	})
	if err != nil {
		t.Fatalf("creating start edge: %v", err)
	}

	// Refresh trigger processor so the new rule is matched.
	if err := triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing trigger processor: %v", err)
	}

	// ── Fire trigger event ────────────────────────────────────────────────────

	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "customers",
		EntityID:   uuid.New(),
		Timestamp:  time.Now(),
		RawData:    map[string]any{"test": true},
		UserID:     createdBy,
	}
	if err := workflowTrigger.OnEntityEvent(ctx, event); err != nil {
		t.Fatalf("firing trigger event: %v", err)
	}

	// ── Poll for the purchase order created by create_purchase_order ──────────

	var po purchaseorderbus.PurchaseOrder
	for i := 0; i < 30; i++ {
		pos, err := db.BusDomain.PurchaseOrder.Query(
			ctx,
			purchaseorderbus.QueryFilter{
				SupplierID:          &supplierID,
				DeliveryWarehouseID: &warehouseID,
			},
			purchaseorderbus.DefaultOrderBy,
			page.MustParse("1", "10"),
		)
		if err != nil {
			t.Fatalf("polling purchase orders: %v", err)
		}
		if len(pos) > 0 {
			po = pos[0]
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if po.ID == uuid.Nil {
		t.Fatal("timeout: no purchase order found after 15s — create_purchase_order may have failed")
	}

	t.Logf("purchase order created: id=%s order_number=%s supplier=%s", po.ID, po.OrderNumber, po.SupplierID)

	// Verify PO fields.
	if po.SupplierID != supplierID {
		t.Errorf("expected supplier_id %s, got %s", supplierID, po.SupplierID)
	}
	if po.DeliveryWarehouseID != warehouseID {
		t.Errorf("expected delivery_warehouse_id %s, got %s", warehouseID, po.DeliveryWarehouseID)
	}
	if po.PurchaseOrderStatusID != poStatusID {
		t.Errorf("expected purchase_order_status_id %s, got %s", poStatusID, po.PurchaseOrderStatusID)
	}
	if po.CurrencyID != currencyID {
		t.Errorf("expected currency_id %s, got %s", currencyID, po.CurrencyID)
	}
	if po.OrderNumber == "" {
		t.Error("expected non-empty order number")
	}

	// Verify at least one line item was created for the PO.
	lineItems, err := db.BusDomain.PurchaseOrderLineItem.QueryByPurchaseOrderID(ctx, po.ID)
	if err != nil {
		t.Fatalf("querying line items: %v", err)
	}
	if len(lineItems) == 0 {
		t.Fatal("expected at least 1 line item to be created by create_purchase_order action")
	}
	t.Logf("line items created: count=%d first_id=%s quantity=%d", len(lineItems), lineItems[0].ID, lineItems[0].QuantityOrdered)

	t.Logf("SUCCESS: create_purchase_order verified — po_id=%s order_number=%s supplier=%s",
		po.ID, po.OrderNumber, po.SupplierID)
}
