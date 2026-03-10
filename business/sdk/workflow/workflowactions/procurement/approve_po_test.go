package procurement_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/procurement"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func Test_ApprovePurchaseOrder(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ApprovePurchaseOrder")

	sd, err := insertPOSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("seeding: %v", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.ApproveHandler = procurement.NewApprovePurchaseOrderHandler(log, db.BusDomain.PurchaseOrder)

	unitest.Run(t, approvePOValidateTests(sd), "validate")
	unitest.Run(t, approvePOExecuteTests(sd), "execute")
}

// =============================================================================
// Shared seed data for PO approval tests

type poSeedData struct {
	unitest.SeedData
	ApproveHandler  *procurement.ApprovePurchaseOrderHandler
	RejectHandler   *procurement.RejectPurchaseOrderHandler
	PendingPO       purchaseorderbus.PurchaseOrder
	ApprovedPO      purchaseorderbus.PurchaseOrder
	RejectedPO      purchaseorderbus.PurchaseOrder
	ExecutionContext workflow.ActionExecutionContext
}

func insertPOSeedData(busDomain dbtest.BusDomain) (poSeedData, error) {
	ctx := context.Background()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return poSeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	adminIDs := uuid.UUIDs{admins[0].ID}

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return poSeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, 0, len(regions))
	for _, r := range regions {
		regionIDs = append(regionIDs, r.ID)
	}

	cities, err := citybus.TestSeedCities(ctx, 3, regionIDs, busDomain.City)
	if err != nil {
		return poSeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	cityIDs := make([]uuid.UUID, 0, len(cities))
	for _, c := range cities {
		cityIDs = append(cityIDs, c.ID)
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, busDomain.Street)
	if err != nil {
		return poSeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	streetIDs := make(uuid.UUIDs, len(streets))
	for i, s := range streets {
		streetIDs[i] = s.ID
	}

	tzs, err := busDomain.Timezone.QueryAll(ctx)
	if err != nil {
		return poSeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, 0, len(tzs))
	for _, tz := range tzs {
		tzIDs = append(tzIDs, tz.ID)
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 3, streetIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return poSeedData{}, fmt.Errorf("seeding contact infos: %w", err)
	}
	contactIDs := make(uuid.UUIDs, len(contactInfos))
	for i, c := range contactInfos {
		contactIDs[i] = c.ID
	}

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 3, contactIDs, busDomain.Supplier)
	if err != nil {
		return poSeedData{}, fmt.Errorf("seeding suppliers: %w", err)
	}
	supplierIDs := make(uuid.UUIDs, len(suppliers))
	for i, s := range suppliers {
		supplierIDs[i] = s.SupplierID
	}

	statuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 3, busDomain.PurchaseOrderStatus)
	if err != nil {
		return poSeedData{}, fmt.Errorf("seeding PO statuses: %w", err)
	}
	statusIDs := make(uuid.UUIDs, len(statuses))
	for i, s := range statuses {
		statusIDs[i] = s.ID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 2, admins[0].ID, streetIDs, busDomain.Warehouse)
	if err != nil {
		return poSeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}
	warehouseIDs := make(uuid.UUIDs, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	currencies, err := currencybus.TestSeedCurrencies(ctx, 2, busDomain.Currency)
	if err != nil {
		return poSeedData{}, fmt.Errorf("seeding currencies: %w", err)
	}
	currencyIDs := make(uuid.UUIDs, len(currencies))
	for i, c := range currencies {
		currencyIDs[i] = c.ID
	}

	now := time.Now().UTC()

	// Create a pending (unapproved, unrejected) PO.
	pendingPO, err := busDomain.PurchaseOrder.Create(ctx, purchaseorderbus.NewPurchaseOrder{
		OrderNumber:           "PO-TEST-PENDING-001",
		SupplierID:            supplierIDs[0],
		PurchaseOrderStatusID: statusIDs[0],
		DeliveryWarehouseID:   warehouseIDs[0],
		DeliveryLocationID:    uuid.Nil,
		DeliveryStreetID:      streetIDs[0],
		OrderDate:             now,
		ExpectedDeliveryDate:  now.Add(14 * 24 * time.Hour),
		Subtotal:              1000.00,
		TaxAmount:             80.00,
		ShippingCost:          50.00,
		TotalAmount:           1130.00,
		CurrencyID:            currencyIDs[0],
		RequestedBy:           admins[0].ID,
		Notes:                 "pending test PO",
		CreatedBy:             admins[0].ID,
	})
	if err != nil {
		return poSeedData{}, fmt.Errorf("creating pending PO: %w", err)
	}

	// Create and approve a PO.
	preApprovePO, err := busDomain.PurchaseOrder.Create(ctx, purchaseorderbus.NewPurchaseOrder{
		OrderNumber:           "PO-TEST-APPROVED-002",
		SupplierID:            supplierIDs[1%len(supplierIDs)],
		PurchaseOrderStatusID: statusIDs[1%len(statusIDs)],
		DeliveryWarehouseID:   warehouseIDs[0],
		DeliveryLocationID:    uuid.Nil,
		DeliveryStreetID:      streetIDs[1%len(streetIDs)],
		OrderDate:             now,
		ExpectedDeliveryDate:  now.Add(14 * 24 * time.Hour),
		Subtotal:              2000.00,
		TaxAmount:             160.00,
		ShippingCost:          50.00,
		TotalAmount:           2210.00,
		CurrencyID:            currencyIDs[0],
		RequestedBy:           admins[0].ID,
		Notes:                 "to-be-approved test PO",
		CreatedBy:             admins[0].ID,
	})
	if err != nil {
		return poSeedData{}, fmt.Errorf("creating pre-approve PO: %w", err)
	}
	approvedPO, err := busDomain.PurchaseOrder.Approve(ctx, preApprovePO, admins[0].ID, "looks good")
	if err != nil {
		return poSeedData{}, fmt.Errorf("approving PO: %w", err)
	}

	// Create and reject a PO.
	preRejectPO, err := busDomain.PurchaseOrder.Create(ctx, purchaseorderbus.NewPurchaseOrder{
		OrderNumber:           "PO-TEST-REJECTED-003",
		SupplierID:            supplierIDs[2%len(supplierIDs)],
		PurchaseOrderStatusID: statusIDs[2%len(statusIDs)],
		DeliveryWarehouseID:   warehouseIDs[0],
		DeliveryLocationID:    uuid.Nil,
		DeliveryStreetID:      streetIDs[2%len(streetIDs)],
		OrderDate:             now,
		ExpectedDeliveryDate:  now.Add(14 * 24 * time.Hour),
		Subtotal:              3000.00,
		TaxAmount:             240.00,
		ShippingCost:          50.00,
		TotalAmount:           3290.00,
		CurrencyID:            currencyIDs[0],
		RequestedBy:           admins[0].ID,
		Notes:                 "to-be-rejected test PO",
		CreatedBy:             admins[0].ID,
	})
	if err != nil {
		return poSeedData{}, fmt.Errorf("creating pre-reject PO: %w", err)
	}
	rejectedPO, err := busDomain.PurchaseOrder.Reject(ctx, preRejectPO, admins[0].ID, "wrong supplier")
	if err != nil {
		return poSeedData{}, fmt.Errorf("rejecting PO: %w", err)
	}

	_ = adminIDs

	ruleID := uuid.New()
	execCtx := workflow.ActionExecutionContext{
		EntityID:      uuid.New(),
		EntityName:    "procurement.purchase_orders",
		EventType:     "on_create",
		UserID:        admins[0].ID,
		RuleID:        &ruleID,
		RuleName:      "Test PO Approval Rule",
		ExecutionID:   uuid.New(),
		Timestamp:     time.Now().UTC(),
		TriggerSource: workflow.TriggerSourceAutomation,
	}

	return poSeedData{
		SeedData: unitest.SeedData{
			Admins: []unitest.User{{User: admins[0]}},
		},
		PendingPO:       pendingPO,
		ApprovedPO:      approvedPO,
		RejectedPO:      rejectedPO,
		ExecutionContext: execCtx,
	}, nil
}

// =============================================================================
// Approve PO tests

func approvePOValidateTests(sd poSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "missing_purchase_order_id",
			ExpResp: "purchase_order_id is required",
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
				err := sd.ApproveHandler.Validate(json.RawMessage(`{"purchase_order_id":"not-a-uuid"}`))
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
				return sd.ApproveHandler.Validate(json.RawMessage(`{"purchase_order_id":"` + uuid.New().String() + `"}`))
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

func approvePOExecuteTests(sd poSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "approve_pending",
			ExpResp: "approved",
			ExcFunc: func(ctx context.Context) any {
				cfg, _ := json.Marshal(procurement.ApprovePurchaseOrderConfig{
					PurchaseOrderID: sd.PendingPO.ID.String(),
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
				cfg, _ := json.Marshal(procurement.ApprovePurchaseOrderConfig{
					PurchaseOrderID: sd.ApprovedPO.ID.String(),
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
				cfg, _ := json.Marshal(procurement.ApprovePurchaseOrderConfig{
					PurchaseOrderID: sd.RejectedPO.ID.String(),
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
				cfg, _ := json.Marshal(procurement.ApprovePurchaseOrderConfig{
					PurchaseOrderID: uuid.New().String(),
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
