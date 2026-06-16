package inventory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ReleaseToPickingConfig holds the config for the release_to_picking handler.
type ReleaseToPickingConfig struct {
	// OrderID is the customer order to release. Use "{{entity_id}}" when the action
	// is wired to a button (the order is the page entity) or a static UUID when wired
	// to a workflow rule that targets a specific order.
	OrderID string `json:"order_id"`
}

// ReleaseToPickingHandler releases a customer order to picking: it transitions the
// order's fulfillment status from PENDING/PROCESSING to PICKING and fans the order's
// line items out into inventory.pick_tasks rows (directed-work model). It is
// self-contained and invokable identically by a button or a workflow rule.
//
// Execute returns map[string]any with key "output" (string) and one of:
//   - "released"           — order transitioned to PICKING and pick tasks created
//   - "invalid_status"     — order is not PENDING/PROCESSING (no writes)
//   - "not_found"          — order not found
//   - "no_line_items"      — order has no line items
//   - "insufficient_stock" — released with coverage gaps, or nothing pickable (no transition)
//   - "failure"            — unexpected error
type ReleaseToPickingHandler struct {
	log                       *logger.Logger
	db                        *sqlx.DB
	ordersBus                 *ordersbus.Business
	orderLineItemsBus         *orderlineitemsbus.Business
	pickTaskBus               *picktaskbus.Business
	inventoryItemBus          *inventoryitembus.Business
	orderFulfillmentStatusBus *orderfulfillmentstatusbus.Business
}

// NewReleaseToPickingHandler creates a new release_to_picking handler.
func NewReleaseToPickingHandler(
	log *logger.Logger,
	db *sqlx.DB,
	ordersBus *ordersbus.Business,
	orderLineItemsBus *orderlineitemsbus.Business,
	pickTaskBus *picktaskbus.Business,
	inventoryItemBus *inventoryitembus.Business,
	orderFulfillmentStatusBus *orderfulfillmentstatusbus.Business,
) *ReleaseToPickingHandler {
	return &ReleaseToPickingHandler{
		log:                       log,
		db:                        db,
		ordersBus:                 ordersBus,
		orderLineItemsBus:         orderLineItemsBus,
		pickTaskBus:               pickTaskBus,
		inventoryItemBus:          inventoryItemBus,
		orderFulfillmentStatusBus: orderFulfillmentStatusBus,
	}
}

// GetType returns the action type.
func (h *ReleaseToPickingHandler) GetType() string { return "release_to_picking" }

// IsAsync returns false — release completes inline.
func (h *ReleaseToPickingHandler) IsAsync() bool { return false }

// SupportsManualExecution returns true.
func (h *ReleaseToPickingHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description.
func (h *ReleaseToPickingHandler) GetDescription() string {
	return "Release a customer order to picking: transition to PICKING status and generate pick tasks for its line items"
}

// Validate validates the release_to_picking configuration.
func (h *ReleaseToPickingHandler) Validate(config json.RawMessage) error {
	var cfg ReleaseToPickingConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.OrderID == "" {
		return fmt.Errorf("order_id is required")
	}
	// A templated order_id (e.g. "{{entity_id}}") is resolved at runtime; only a
	// static value must parse as a UUID here.
	if !strings.Contains(cfg.OrderID, "{{") {
		if _, err := uuid.Parse(cfg.OrderID); err != nil {
			return fmt.Errorf("invalid order_id: %w", err)
		}
	}
	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *ReleaseToPickingHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "released", Description: "Order transitioned to PICKING and pick tasks created", IsDefault: true},
		{Name: "invalid_status", Description: "Order is not PENDING/PROCESSING — cannot release"},
		{Name: "not_found", Description: "Order not found"},
		{Name: "no_line_items", Description: "Order has no line items"},
		{Name: "insufficient_stock", Description: "Released with coverage gaps, or nothing pickable"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *ReleaseToPickingHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{
			EntityName: "sales.orders",
			EventType:  "on_update",
			Fields:     []string{"order_fulfillment_status_id"},
			// The PICKING status UUID is resolved at runtime (by name), so the produced
			// value is not statically knowable — mark it Indeterminate.
			Changes: []workflow.ProducedChange{
				{FieldName: "order_fulfillment_status_id", Operator: workflow.OperatorChangedTo, Indeterminate: true},
			},
		},
		{
			EntityName: "inventory.pick_tasks",
			EventType:  "on_create",
		},
	}
}

// pickPlanEntry is one line item's resolved FEFO pick suggestion.
type pickPlanEntry struct {
	lineID     uuid.UUID
	productID  uuid.UUID
	locationID uuid.UUID
	qty        int
}

// Execute releases a customer order to picking.
func (h *ReleaseToPickingHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	// Step 1: nil-guard buses.
	if h.db == nil || h.ordersBus == nil || h.orderLineItemsBus == nil || h.pickTaskBus == nil || h.inventoryItemBus == nil || h.orderFulfillmentStatusBus == nil {
		return map[string]any{"output": "failure", "error": "release_to_picking dependencies not configured"}, nil
	}

	var cfg ReleaseToPickingConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	// Step 2: resolve the order id. Default to the execution-context entity (button
	// path). A static config UUID overrides; a templated value (unresolved "{{...}}")
	// is ignored in favor of the entity id.
	orderID := execCtx.EntityID
	if cfg.OrderID != "" && !strings.Contains(cfg.OrderID, "{{") {
		parsed, err := uuid.Parse(cfg.OrderID)
		if err != nil {
			return map[string]any{"output": "failure", "error": "invalid order_id"}, nil
		}
		orderID = parsed
	}
	if orderID == uuid.Nil {
		return map[string]any{"output": "failure", "error": "no order id"}, nil
	}

	// Step 3: resolve fulfillment-status UUIDs by name. A missing status is a real
	// misconfiguration (the picking pipeline cannot function) — return an error.
	pendingID, err := h.resolveStatusID(ctx, "PENDING")
	if err != nil {
		return nil, fmt.Errorf("resolve PENDING status: %w", err)
	}
	processingID, err := h.resolveStatusID(ctx, "PROCESSING")
	if err != nil {
		return nil, fmt.Errorf("resolve PROCESSING status: %w", err)
	}
	pickingID, err := h.resolveStatusID(ctx, "PICKING")
	if err != nil {
		return nil, fmt.Errorf("resolve PICKING status: %w", err)
	}

	// Step 4: fetch the order.
	order, err := h.ordersBus.QueryByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, ordersbus.ErrNotFound) {
			return map[string]any{"output": "not_found", "order_id": orderID.String()}, nil
		}
		return nil, fmt.Errorf("query order: %w", err)
	}

	// Step 5: validate current status — only PENDING/PROCESSING may be released.
	if order.FulfillmentStatusID != pendingID && order.FulfillmentStatusID != processingID {
		return map[string]any{
			"output":          "invalid_status",
			"order_id":        orderID.String(),
			"current_status":  order.FulfillmentStatusID.String(),
		}, nil
	}

	// Single-page read cap shared by the line-item and existing-task reads below. Real
	// orders stay far under this, but we log if a release ever hits the cap so silent
	// truncation is visible rather than producing under-released lines or incomplete
	// idempotency dedup.
	pg := page.MustParse("1", "1000")

	// Step 6: fetch line items.
	lineItems, err := h.orderLineItemsBus.Query(ctx,
		orderlineitemsbus.QueryFilter{OrderID: &orderID},
		orderlineitemsbus.DefaultOrderBy,
		pg,
	)
	if err != nil {
		return nil, fmt.Errorf("query order line items: %w", err)
	}
	if len(lineItems) == 0 {
		return map[string]any{"output": "no_line_items", "order_id": orderID.String()}, nil
	}
	if len(lineItems) >= pg.RowsPerPage() {
		h.log.Warn(ctx, "release_to_picking: order line items hit the page cap; lines beyond the cap are not released this run",
			"order_id", orderID.String(), "cap", pg.RowsPerPage())
	}

	// Step 7: idempotency pre-read — skip lines that already have a non-terminal pick task.
	existingTasks, err := h.pickTaskBus.Query(ctx,
		picktaskbus.QueryFilter{SalesOrderID: &orderID},
		picktaskbus.DefaultOrderBy,
		pg,
	)
	if err != nil {
		return nil, fmt.Errorf("query existing pick tasks: %w", err)
	}
	if len(existingTasks) >= pg.RowsPerPage() {
		h.log.Warn(ctx, "release_to_picking: existing pick tasks hit the page cap; idempotency dedup may be incomplete and could create duplicate tasks",
			"order_id", orderID.String(), "cap", pg.RowsPerPage())
	}
	activeLines := make(map[uuid.UUID]bool)
	for _, t := range existingTasks {
		if t.Status.Equal(picktaskbus.Statuses.Pending) || t.Status.Equal(picktaskbus.Statuses.InProgress) {
			activeLines[t.SalesOrderLineItemID] = true
		}
	}

	// Step 8: build a FEFO pick plan (read-only suggestion; no inventory mutation).
	var plan []pickPlanEntry
	insufficient := false
	for _, line := range lineItems {
		if activeLines[line.ID] {
			continue
		}
		remaining := line.Quantity - line.PickedQuantity
		if remaining <= 0 {
			continue
		}

		buckets, err := h.inventoryItemBus.QueryAvailableForAllocation(ctx, line.ProductID, nil, nil, "fefo", 100)
		if err != nil {
			return nil, fmt.Errorf("query available inventory: %w", err)
		}

		for _, b := range buckets {
			if remaining <= 0 {
				break
			}
			avail := b.Quantity - b.AllocatedQuantity
			if avail <= 0 {
				continue
			}
			take := avail
			if take > remaining {
				take = remaining
			}
			plan = append(plan, pickPlanEntry{
				lineID:     line.ID,
				productID:  line.ProductID,
				locationID: b.LocationID,
				qty:        take,
			})
			remaining -= take
		}

		if remaining > 0 {
			insufficient = true
		}
	}

	// Step 9: nothing pickable → do not transition the order.
	if len(plan) == 0 {
		return map[string]any{"output": "insufficient_stock", "order_id": orderID.String(), "tasks_created": 0}, nil
	}

	// Step 10: single transaction — flip status and create the pick tasks.
	tx, err := h.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	txOrdersBus, err := h.ordersBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("tx orders bus: %w", err)
	}
	txPickTaskBus, err := h.pickTaskBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("tx pick task bus: %w", err)
	}

	if _, err := txOrdersBus.Update(ctx, order, ordersbus.UpdateOrder{
		FulfillmentStatusID: &pickingID,
		UpdatedBy:           &execCtx.UserID,
	}); err != nil {
		return nil, fmt.Errorf("transition order to PICKING: %w", err)
	}

	for _, p := range plan {
		if _, err := txPickTaskBus.Create(ctx, picktaskbus.NewPickTask{
			SalesOrderID:         orderID,
			SalesOrderLineItemID: p.lineID,
			ProductID:            p.productID,
			LocationID:           p.locationID,
			QuantityToPick:       p.qty,
			CreatedBy:            execCtx.UserID,
		}); err != nil {
			return nil, fmt.Errorf("create pick task: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	output := "released"
	if insufficient {
		output = "insufficient_stock"
	}

	h.log.Info(ctx, "release_to_picking: order released",
		"order_id", orderID, "tasks_created", len(plan), "insufficient", insufficient)

	return map[string]any{
		"output":        output,
		"order_id":      orderID.String(),
		"tasks_created": len(plan),
		"transitioned":  true,
	}, nil
}

// resolveStatusID resolves an order fulfillment status UUID by name.
func (h *ReleaseToPickingHandler) resolveStatusID(ctx context.Context, name string) (uuid.UUID, error) {
	statuses, err := h.orderFulfillmentStatusBus.Query(ctx,
		orderfulfillmentstatusbus.QueryFilter{Name: &name},
		orderfulfillmentstatusbus.DefaultOrderBy,
		page.MustParse("1", "1"),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("query: %w", err)
	}
	if len(statuses) == 0 {
		return uuid.Nil, fmt.Errorf("order fulfillment status %q not found", name)
	}
	return statuses[0].ID, nil
}
