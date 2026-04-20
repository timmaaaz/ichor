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
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
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

func Test_ApproveTransferOrder(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ApproveTransferOrder")

	sd, err := insertTransferOrderSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("seeding: %v", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.ApproveHandler = inventory.NewApproveTransferOrderHandler(log, db.BusDomain.TransferOrder)

	unitest.Run(t, approveTransferOrderValidateTests(sd), "validate")
	unitest.Run(t, approveTransferOrderExecuteTests(sd), "execute")
}

// =============================================================================
// Shared seed data for transfer order approval tests

type transferOrderSeedData struct {
	unitest.SeedData
	ApproveHandler  *inventory.ApproveTransferOrderHandler
	RejectHandler   *inventory.RejectTransferOrderHandler
	PendingTO       transferorderbus.TransferOrder
	ApprovedTO      transferorderbus.TransferOrder
	RejectedTO      transferorderbus.TransferOrder
	ExecutionContext workflow.ActionExecutionContext
}

func insertTransferOrderSeedData(busDomain dbtest.BusDomain) (transferOrderSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	adminIDs := make([]uuid.UUID, len(admins))
	for i, a := range admins {
		adminIDs[i] = a.ID
	}

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("querying regions: %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding cities: %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding streets: %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding brands: %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 3, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding products: %w", err)
	}

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 3, warehouseIDs, busDomain.Zones)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding zones: %w", err)
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 6, warehouseIDs, zones, busDomain.InventoryLocation)
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}

	locationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, l := range inventoryLocations {
		locationIDs[i] = l.LocationID
	}

	// Create a pending transfer order.
	pendingTO, err := busDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
		ProductID:      productIDs[0],
		FromLocationID: locationIDs[0],
		ToLocationID:   locationIDs[1],
		RequestedByID:  adminIDs[0],
		Quantity:       10,
		Status:         "pending",
		TransferDate:   time.Now(),
	})
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("creating pending transfer order: %w", err)
	}

	// Create an already-approved transfer order.
	preApproveTO, err := busDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
		ProductID:      productIDs[1%len(productIDs)],
		FromLocationID: locationIDs[2%len(locationIDs)],
		ToLocationID:   locationIDs[3%len(locationIDs)],
		RequestedByID:  adminIDs[0],
		Quantity:       5,
		Status:         "pending",
		TransferDate:   time.Now(),
	})
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("creating pre-approve TO: %w", err)
	}
	approvedTO, err := busDomain.TransferOrder.Approve(ctx, preApproveTO, adminIDs[0], "looks good")
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("approving TO: %w", err)
	}

	// Create an already-rejected transfer order.
	preRejectTO, err := busDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
		ProductID:      productIDs[2%len(productIDs)],
		FromLocationID: locationIDs[4%len(locationIDs)],
		ToLocationID:   locationIDs[5%len(locationIDs)],
		RequestedByID:  adminIDs[0],
		Quantity:       7,
		Status:         "pending",
		TransferDate:   time.Now(),
	})
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("creating pre-reject TO: %w", err)
	}
	rejectedTO, err := busDomain.TransferOrder.Reject(ctx, preRejectTO, adminIDs[0], "wrong location")
	if err != nil {
		return transferOrderSeedData{}, fmt.Errorf("rejecting TO: %w", err)
	}

	ruleID := uuid.New()
	execCtx := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "inventory.transfer_orders",
		EventType:     "on_create",
		UserID:        adminIDs[0],
		RuleID:        &ruleID,
		RuleName:      "Test Transfer Order Approval Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return transferOrderSeedData{
		SeedData: unitest.SeedData{
			Admins: []unitest.User{{User: admins[0]}},
		},
		PendingTO:       pendingTO,
		ApprovedTO:      approvedTO,
		RejectedTO:      rejectedTO,
		ExecutionContext: execCtx,
	}, nil
}

// =============================================================================
// Approve transfer order tests

func approveTransferOrderValidateTests(sd transferOrderSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "missing_transfer_order_id",
			ExpResp: "transfer_order_id is required",
			ExcFunc: func(ctx context.Context) any {
				err := sd.ApproveHandler.Validate(json.RawMessage(`{}`))
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
		},
		{
			Name:    "valid_config",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				return sd.ApproveHandler.Validate(json.RawMessage(`{"transfer_order_id":"` + uuid.New().String() + `"}`))
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
	}
}

func approveTransferOrderExecuteTests(sd transferOrderSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "approve_pending",
			ExpResp: "approved",
			ExcFunc: func(ctx context.Context) any {
				cfg, _ := json.Marshal(inventory.ApproveTransferOrderConfig{
					TransferOrderID: sd.PendingTO.TransferID.String(),
					ApprovalReason:  "looks good",
				})
				result, err := sd.ApproveHandler.Execute(ctx, cfg, sd.ExecutionContext)
				if err != nil {
					return err
				}
				return result.(map[string]any)["output"]
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "already_approved",
			ExpResp: "already_approved",
			ExcFunc: func(ctx context.Context) any {
				cfg, _ := json.Marshal(inventory.ApproveTransferOrderConfig{
					TransferOrderID: sd.ApprovedTO.TransferID.String(),
				})
				result, err := sd.ApproveHandler.Execute(ctx, cfg, sd.ExecutionContext)
				if err != nil {
					return err
				}
				return result.(map[string]any)["output"]
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "already_rejected",
			ExpResp: "already_rejected",
			ExcFunc: func(ctx context.Context) any {
				cfg, _ := json.Marshal(inventory.ApproveTransferOrderConfig{
					TransferOrderID: sd.RejectedTO.TransferID.String(),
				})
				result, err := sd.ApproveHandler.Execute(ctx, cfg, sd.ExecutionContext)
				if err != nil {
					return err
				}
				return result.(map[string]any)["output"]
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
		{
			Name:    "not_found",
			ExpResp: "not_found",
			ExcFunc: func(ctx context.Context) any {
				cfg, _ := json.Marshal(inventory.ApproveTransferOrderConfig{
					TransferOrderID: uuid.New().String(),
				})
				result, err := sd.ApproveHandler.Execute(ctx, cfg, sd.ExecutionContext)
				if err != nil {
					return err
				}
				return result.(map[string]any)["output"]
			},
			CmpFunc: func(got, exp any) string {
				if got != exp {
					return fmt.Sprintf("got %v, want %v", got, exp)
				}
				return ""
			},
		},
	}
}
