package inventory_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func Test_RejectInventoryAdjustment(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_RejectInventoryAdjustment")

	sd, err := insertAdjustmentSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("seeding: %v", err)
	}

	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})

	sd.RejectHandler = inventory.NewRejectInventoryAdjustmentHandler(log, db.BusDomain.InventoryAdjustment)

	unitest.Run(t, rejectAdjustmentValidateTests(sd), "validate")
	unitest.Run(t, rejectAdjustmentExecuteTests(sd), "execute")
}

func rejectAdjustmentValidateTests(sd adjustmentSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "missing_adjustment_id",
			ExpResp: "adjustment_id is required",
			ExcFunc: func(ctx context.Context) any {
				err := sd.RejectHandler.Validate(json.RawMessage(`{"rejection_reason":"bad count"}`))
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
			Name:    "missing_rejection_reason",
			ExpResp: "rejection_reason is required",
			ExcFunc: func(ctx context.Context) any {
				err := sd.RejectHandler.Validate(json.RawMessage(`{"adjustment_id":"` + uuid.New().String() + `"}`))
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
				return sd.RejectHandler.Validate(json.RawMessage(`{"adjustment_id":"` + uuid.New().String() + `","rejection_reason":"bad count"}`))
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

func rejectAdjustmentExecuteTests(sd adjustmentSeedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "reject_pending",
			ExpResp: "rejected",
			ExcFunc: func(ctx context.Context) any {
				cfg, _ := json.Marshal(inventory.RejectInventoryAdjustmentConfig{
					AdjustmentID:    sd.PendingAdj.InventoryAdjustmentID.String(),
					RejectionReason: "count is wrong",
				})
				result, err := sd.RejectHandler.Execute(ctx, cfg, sd.ExecutionContext)
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
				cfg, _ := json.Marshal(inventory.RejectInventoryAdjustmentConfig{
					AdjustmentID:    sd.ApprovedAdj.InventoryAdjustmentID.String(),
					RejectionReason: "cannot reject approved",
				})
				result, err := sd.RejectHandler.Execute(ctx, cfg, sd.ExecutionContext)
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
				cfg, _ := json.Marshal(inventory.RejectInventoryAdjustmentConfig{
					AdjustmentID:    sd.RejectedAdj.InventoryAdjustmentID.String(),
					RejectionReason: "already done",
				})
				result, err := sd.RejectHandler.Execute(ctx, cfg, sd.ExecutionContext)
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
				cfg, _ := json.Marshal(inventory.RejectInventoryAdjustmentConfig{
					AdjustmentID:    uuid.New().String(),
					RejectionReason: "irrelevant",
				})
				result, err := sd.RejectHandler.Execute(ctx, cfg, sd.ExecutionContext)
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
