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
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
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

func Test_ApproveInventoryAdjustment(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ApproveInventoryAdjustment")

	sd, err := insertAdjustmentSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("seeding: %v", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.ApproveHandler = inventory.NewApproveInventoryAdjustmentHandler(log, db.BusDomain.InventoryAdjustment)

	unitest.Run(t, approveAdjustmentValidateTests(sd), "validate")
	unitest.Run(t, approveAdjustmentExecuteTests(sd), "execute")
}

// =============================================================================
// Shared seed data for adjustment approval tests

type adjustmentSeedData struct {
	unitest.SeedData
	ApproveHandler  *inventory.ApproveInventoryAdjustmentHandler
	RejectHandler   *inventory.RejectInventoryAdjustmentHandler
	PendingAdj      inventoryadjustmentbus.InventoryAdjustment
	ApprovedAdj     inventoryadjustmentbus.InventoryAdjustment
	RejectedAdj     inventoryadjustmentbus.InventoryAdjustment
	ExecutionContext workflow.ActionExecutionContext
}

func insertAdjustmentSeedData(busDomain dbtest.BusDomain) (adjustmentSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	adminIDs := make([]uuid.UUID, len(admins))
	for i, a := range admins {
		adminIDs[i] = a.ID
	}

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("querying regions: %w", err)
	}

	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding cities: %w", err)
	}

	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding streets: %w", err)
	}

	streetIDs := make([]uuid.UUID, 0, len(streets))
	for _, s := range streets {
		streetIDs = append(streetIDs, s.ID)
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}

	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 2, contactIDs, busDomain.Brand)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding brands: %w", err)
	}

	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, busDomain.ProductCategory)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}

	productCategoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		productCategoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 3, brandIDs, productCategoryIDs, busDomain.Product)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding products: %w", err)
	}

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ProductID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, adminIDs[0], streetIDs, busDomain.Warehouse)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}

	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 3, warehouseIDs, busDomain.Zones)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding zones: %w", err)
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 5, warehouseIDs, zones, busDomain.InventoryLocation)
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}

	locationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, l := range inventoryLocations {
		locationIDs[i] = l.LocationID
	}

	adjustedByIDs := uuid.UUIDs{adminIDs[0]}

	// Seed a pending adjustment.
	pendingAdj, err := busDomain.InventoryAdjustment.Create(ctx, inventoryadjustmentbus.NewInventoryAdjustment{
		ProductID:      productIDs[0],
		LocationID:     locationIDs[0],
		AdjustedBy:     adminIDs[0],
		QuantityChange: 10,
		ReasonCode:     "cycle_count",
		Notes:          "test pending",
		AdjustmentDate: time.Now(),
	})
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("creating pending adjustment: %w", err)
	}

	// Seed an already-approved adjustment by calling Approve on it.
	preApproveAdj, err := busDomain.InventoryAdjustment.Create(ctx, inventoryadjustmentbus.NewInventoryAdjustment{
		ProductID:      productIDs[1%len(productIDs)],
		LocationID:     locationIDs[1%len(locationIDs)],
		AdjustedBy:     adminIDs[0],
		QuantityChange: 5,
		ReasonCode:     "data_entry_error",
		Notes:          "pre-approved",
		AdjustmentDate: time.Now(),
	})
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("creating pre-approve adjustment: %w", err)
	}
	approvedAdj, err := busDomain.InventoryAdjustment.Approve(ctx, preApproveAdj, adminIDs[0], "looks good")
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("approving adjustment: %w", err)
	}

	// Seed an already-rejected adjustment.
	preRejectAdj, err := busDomain.InventoryAdjustment.Create(ctx, inventoryadjustmentbus.NewInventoryAdjustment{
		ProductID:      productIDs[2%len(productIDs)],
		LocationID:     locationIDs[2%len(locationIDs)],
		AdjustedBy:     adminIDs[0],
		QuantityChange: -3,
		ReasonCode:     "damaged",
		Notes:          "pre-rejected",
		AdjustmentDate: time.Now(),
	})
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("creating pre-reject adjustment: %w", err)
	}
	rejectedAdj, err := busDomain.InventoryAdjustment.Reject(ctx, preRejectAdj, adminIDs[0], "invalid count")
	if err != nil {
		return adjustmentSeedData{}, fmt.Errorf("rejecting adjustment: %w", err)
	}

	_ = adjustedByIDs

	ruleID := uuid.New()
	execCtx := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "inventory.inventory_adjustments",
		EventType:     "on_create",
		UserID:        adminIDs[0],
		RuleID:        &ruleID,
		RuleName:      "Test Adjustment Approval Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return adjustmentSeedData{
		SeedData: unitest.SeedData{
			Admins: []unitest.User{{User: admins[0]}},
		},
		PendingAdj:      pendingAdj,
		ApprovedAdj:     approvedAdj,
		RejectedAdj:     rejectedAdj,
		ExecutionContext: execCtx,
	}, nil
}

// =============================================================================
// Approve tests

func approveAdjustmentValidateTests(sd adjustmentSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "missing_adjustment_id",
			ExpResp: "adjustment_id is required",
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
			Name:    "invalid_uuid",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := sd.ApproveHandler.Validate(json.RawMessage(`{"adjustment_id":"not-a-uuid"}`))
				return err != nil
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
				return sd.ApproveHandler.Validate(json.RawMessage(`{"adjustment_id":"` + uuid.New().String() + `"}`))
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

func approveAdjustmentExecuteTests(sd adjustmentSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "approve_pending",
			ExpResp: "approved",
			ExcFunc: func(ctx context.Context) any {
				cfg, _ := json.Marshal(inventory.ApproveInventoryAdjustmentConfig{
					AdjustmentID:   sd.PendingAdj.InventoryAdjustmentID.String(),
					ApprovalReason: "looks good",
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
				cfg, _ := json.Marshal(inventory.ApproveInventoryAdjustmentConfig{
					AdjustmentID: sd.ApprovedAdj.InventoryAdjustmentID.String(),
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
				cfg, _ := json.Marshal(inventory.ApproveInventoryAdjustmentConfig{
					AdjustmentID: sd.RejectedAdj.InventoryAdjustmentID.String(),
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
				cfg, _ := json.Marshal(inventory.ApproveInventoryAdjustmentConfig{
					AdjustmentID: uuid.New().String(),
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
