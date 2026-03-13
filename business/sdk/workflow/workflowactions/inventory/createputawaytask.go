package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// CreatePutAwayTaskConfig defines the configuration for the create_put_away_task action.
type CreatePutAwayTaskConfig struct {
	// SourceFromPO: when true, resolves product_id from trigger context's supplier_product_id.
	// When false, uses ProductID from config.
	SourceFromPO bool `json:"source_from_po"`

	// ProductID is a static product UUID. Used only when SourceFromPO is false.
	ProductID string `json:"product_id,omitempty"`

	// LocationStrategy determines how the destination location is resolved.
	// "po_delivery" — use the PO's delivery_location_id (requires PO lookup via RawData["purchase_order_id"])
	// "static"      — use LocationID field below
	LocationStrategy string `json:"location_strategy"`

	// LocationID is a static location UUID. Used when LocationStrategy is "static".
	LocationID string `json:"location_id,omitempty"`

	// ReferenceNumber is a template string. Supports {{variable}} substitution from RawData.
	// Example: "PO-RCV-{{purchase_order_id}}"
	// Defaults to "PO-{purchase_order_id}" when empty and purchase_order_id is in RawData.
	ReferenceNumber string `json:"reference_number,omitempty"`
}

// CreatePutAwayTaskHandler creates a put-away task for received inventory.
// It is triggered by on_update events on purchase_order_line_items when quantity_received increases.
//
// Execute returns map[string]any with key "output" (string) and one of:
//   - "created"  — task created; also includes "task_id" string
//   - "created" + "skipped": true — delta <= 0, no task needed
//   - "no_location"      — po_delivery strategy but PO has no delivery_location_id
//   - "product_not_found" — supplier_product_id lookup failed
//   - "failure"           — unexpected error
type CreatePutAwayTaskHandler struct {
	log                *logger.Logger
	putAwayTaskBus     *putawaytaskbus.Business
	supplierProductBus *supplierproductbus.Business
	purchaseOrderBus   *purchaseorderbus.Business
}

// NewCreatePutAwayTaskHandler creates a new CreatePutAwayTaskHandler.
func NewCreatePutAwayTaskHandler(
	log *logger.Logger,
	putAwayTaskBus *putawaytaskbus.Business,
	supplierProductBus *supplierproductbus.Business,
	purchaseOrderBus *purchaseorderbus.Business,
) *CreatePutAwayTaskHandler {
	return &CreatePutAwayTaskHandler{
		log:                log,
		putAwayTaskBus:     putAwayTaskBus,
		supplierProductBus: supplierProductBus,
		purchaseOrderBus:   purchaseOrderBus,
	}
}

// GetType returns the action type identifier.
func (h *CreatePutAwayTaskHandler) GetType() string {
	return "create_put_away_task"
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *CreatePutAwayTaskHandler) GetDescription() string {
	return "Creates a put-away task directing floor workers to shelve received goods"
}

// SupportsManualExecution returns true — this action can be triggered manually.
func (h *CreatePutAwayTaskHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false — this action completes inline.
func (h *CreatePutAwayTaskHandler) IsAsync() bool {
	return false
}

// Validate validates the action configuration.
func (h *CreatePutAwayTaskHandler) Validate(config json.RawMessage) error {
	var cfg CreatePutAwayTaskConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if cfg.LocationStrategy != "po_delivery" && cfg.LocationStrategy != "static" {
		return fmt.Errorf("location_strategy must be 'po_delivery' or 'static', got %q", cfg.LocationStrategy)
	}

	if !cfg.SourceFromPO {
		if cfg.ProductID == "" {
			return fmt.Errorf("product_id is required when source_from_po is false")
		}
		if _, err := uuid.Parse(cfg.ProductID); err != nil {
			return fmt.Errorf("invalid product_id: %w", err)
		}
	}

	if cfg.LocationStrategy == "static" {
		if cfg.LocationID == "" {
			return fmt.Errorf("location_id is required when location_strategy is 'static'")
		}
		if _, err := uuid.Parse(cfg.LocationID); err != nil {
			return fmt.Errorf("invalid location_id: %w", err)
		}
	}

	return nil
}

// Execute creates a put-away task based on the config and execution context.
// Must return map[string]any with an "output" key for edge routing.
func (h *CreatePutAwayTaskHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg CreatePutAwayTaskConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	// Step 1: Compute quantity delta from FieldChanges.
	// JSON deserializes numbers as float64 — use explicit conversion.
	delta := 0
	if fc, ok := execCtx.FieldChanges["quantity_received"]; ok {
		oldF, _ := toFloat64(fc.OldValue)
		newF, _ := toFloat64(fc.NewValue)
		delta = int(newF - oldF)
	} else if raw, ok := execCtx.RawData["quantity_received"]; ok {
		if v, ok2 := toFloat64(raw); ok2 {
			delta = int(v)
			h.log.Warn(ctx, "create_put_away_task: FieldChanges not present; falling back to RawData quantity_received",
				"entity_id", execCtx.EntityID)
		}
	}

	if delta <= 0 {
		h.log.Info(ctx, "create_put_away_task: delta <= 0, skipping task creation",
			"delta", delta, "entity_id", execCtx.EntityID)
		return map[string]any{"output": "created", "skipped": true, "reason": "delta <= 0"}, nil
	}

	// Step 2: Resolve product_id.
	var productID uuid.UUID
	if cfg.SourceFromPO {
		spIDStr, _ := execCtx.RawData["supplier_product_id"].(string)
		spID, err := uuid.Parse(spIDStr)
		if err != nil {
			return map[string]any{"output": "product_not_found", "error": "invalid or missing supplier_product_id in RawData"}, nil
		}
		sp, err := h.supplierProductBus.QueryByID(ctx, spID)
		if err != nil {
			h.log.Info(ctx, "create_put_away_task: supplier product lookup failed",
				"supplier_product_id", spIDStr, "error", err)
			return map[string]any{"output": "product_not_found", "error": err.Error()}, nil
		}
		productID = sp.ProductID
	} else {
		productID, _ = uuid.Parse(cfg.ProductID)
	}

	// Step 3: Resolve location_id.
	var locationID uuid.UUID
	if cfg.LocationStrategy == "po_delivery" {
		poIDStr, _ := execCtx.RawData["purchase_order_id"].(string)
		poID, err := uuid.Parse(poIDStr)
		if err != nil {
			return map[string]any{"output": "no_location", "error": "invalid or missing purchase_order_id in RawData"}, nil
		}
		po, err := h.purchaseOrderBus.QueryByID(ctx, poID)
		if err != nil {
			return map[string]any{"output": "no_location", "error": err.Error()}, nil
		}
		if po.DeliveryLocationID == uuid.Nil {
			h.log.Info(ctx, "create_put_away_task: PO has no delivery_location_id",
				"purchase_order_id", poIDStr)
			return map[string]any{"output": "no_location"}, nil
		}
		locationID = po.DeliveryLocationID
	} else {
		locationID, _ = uuid.Parse(cfg.LocationID)
	}

	// Step 4: Resolve reference number (supports {{variable}} template substitution).
	refNum := cfg.ReferenceNumber
	if refNum == "" {
		if poIDStr, ok := execCtx.RawData["purchase_order_id"].(string); ok {
			refNum = "PO-" + poIDStr
		}
	} else {
		refNum = resolveTemplate(refNum, execCtx.RawData)
	}

	// Step 5: Create the task.
	task, err := h.putAwayTaskBus.Create(ctx, putawaytaskbus.NewPutAwayTask{
		ProductID:       productID,
		LocationID:      locationID,
		Quantity:        delta,
		ReferenceNumber: refNum,
		CreatedBy:       execCtx.UserID,
	})
	if err != nil {
		h.log.Error(ctx, "create_put_away_task: failed to create task", "error", err)
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	h.log.Info(ctx, "create_put_away_task: task created",
		"task_id", task.ID, "product_id", productID, "location_id", locationID, "quantity", delta)

	return map[string]any{
		"output":  "created",
		"task_id": task.ID.String(),
	}, nil
}

// resolveTemplate replaces {{key}} placeholders in tmpl with values from data.
func resolveTemplate(tmpl string, data map[string]any) string {
	result := tmpl
	for k, v := range data {
		result = strings.ReplaceAll(result, "{{"+k+"}}", fmt.Sprintf("%v", v))
	}
	return result
}

// toFloat64 converts any numeric value to float64.
// JSON deserializes numbers as float64, but RawData may contain int types in test code.
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
