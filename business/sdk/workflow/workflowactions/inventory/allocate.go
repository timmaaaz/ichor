package inventory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// TODO: Add to AllocationStrategy options:
// - "nearest_location"     // Minimize shipping distance
// - "lowest_cost"          // Choose cheapest warehouse
// - "nearest_expiry"       // FEFO for perishables
// - "load_balancing"       // Distribute across warehouses
// - "priority_zone"        // VIP customers get premium stock

// AllocateInventoryConfig represents the configuration for inventory allocation
type AllocateInventoryConfig struct {
	InventoryItems     []AllocationItem `json:"inventory_items"`
	AllocationMode     string           `json:"allocation_mode"`     // 'reserve' or 'allocate'
	AllocationStrategy string           `json:"allocation_strategy"` // 'fifo', 'lifo', 'nearest_expiry', 'lowest_cost'
	AllowPartial       bool             `json:"allow_partial"`
	ReservationHours   int              `json:"reservation_duration_hours,omitempty"`
	Priority           string           `json:"priority"` // 'low', 'medium', 'high', 'critical'
	TimeoutMs          int              `json:"timeout_ms,omitempty"`
	ReferenceID        string           `json:"reference_id,omitempty"`   // Order ID, etc.
	ReferenceType      string           `json:"reference_type,omitempty"` // 'order', 'transfer', etc.
}

// AllocationItem represents a single item to allocate
type AllocationItem struct {
	ProductID   uuid.UUID  `json:"product_id"`
	Quantity    int        `json:"quantity"`
	WarehouseID *uuid.UUID `json:"warehouse_id,omitempty"`
	LocationID  *uuid.UUID `json:"location_id,omitempty"`
}

// AllocationRequest is queued to RabbitMQ for async processing
type AllocationRequest struct {
	ID          uuid.UUID                       `json:"id"`
	ExecutionID uuid.UUID                       `json:"execution_id"`
	Config      AllocateInventoryConfig         `json:"config"`
	Context     workflow.ActionExecutionContext `json:"context"`
	Status      string                          `json:"status"`
	Priority    int                             `json:"priority"`
	CreatedAt   time.Time                       `json:"created_at"`
	RetryCount  int                             `json:"retry_count"`
	MaxRetries  int                             `json:"max_retries"`
}

// InventoryAllocationResult represents the result of an inventory allocation
type InventoryAllocationResult struct {
	AllocationID    uuid.UUID       `json:"allocation_id"`
	Status          string          `json:"status"` // 'success', 'partial', 'failed'
	AllocatedItems  []AllocatedItem `json:"allocated_items"`
	FailedItems     []FailedItem    `json:"failed_items"`
	TotalRequested  int             `json:"total_requested"`
	TotalAllocated  int             `json:"total_allocated"`
	ExecutionTimeMs int64           `json:"execution_time_ms"`
	IdempotencyKey  string          `json:"idempotency_key"`
	Warnings        []string        `json:"warnings"`
	CreatedAt       time.Time       `json:"created_at"`
	CompletedAt     time.Time       `json:"completed_at"`
}

// AllocatedItem represents a successfully allocated item
type AllocatedItem struct {
	ProductID         uuid.UUID  `json:"product_id"`
	LocationID        uuid.UUID  `json:"location_id"`
	RequestedQuantity int        `json:"requested_quantity"`
	AllocatedQuantity int        `json:"allocated_quantity"`
	InventoryID       uuid.UUID  `json:"inventory_item_id"`
	AllocationMode    string     `json:"allocation_mode"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"` // For reservations
}

// FailedItem represents an item that couldn't be allocated
type FailedItem struct {
	ProductID         uuid.UUID `json:"product_id"`
	RequestedQuantity int       `json:"requested_quantity"`
	AvailableQuantity int       `json:"available_quantity"`
	Reason            string    `json:"reason"`
	ErrorMessage      string    `json:"error_message"`
}

func (e ErrAlreadyProcessed) Error() string {
	return fmt.Sprintf("allocation already processed with key: %s", e.IdempotencyKey)
}

// QueuedAllocationResponse represents the immediate response when allocation is queued
type QueuedAllocationResponse struct {
	AllocationID   uuid.UUID `json:"allocation_id"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	Priority       int       `json:"priority"`
	Message        string    `json:"message"`
}

// ErrAlreadyProcessed is returned when an allocation has already been processed
type ErrAlreadyProcessed struct {
	IdempotencyKey string
	Result         *InventoryAllocationResult
}

// Database models for allocation tracking
// type allocationResult struct {
// 	ID             string    `db:"id"`
// 	IdempotencyKey string    `db:"idempotency_key"`
// 	AllocationData []byte    `db:"allocation_data"`
// 	CreatedAt      time.Time `db:"created_at"`
// }

type inventoryItemLock struct {
	ID                string    `db:"id"`
	ProductID         string    `db:"product_id"`
	LocationID        string    `db:"location_id"`
	Quantity          int       `db:"quantity"`
	ReservedQuantity  int       `db:"reserved_quantity"`
	AllocatedQuantity int       `db:"allocated_quantity"`
	CreatedDate       time.Time `db:"created_date"`
}

// AllocateInventoryHandler handles allocate_inventory actions
type AllocateInventoryHandler struct {
	log              *logger.Logger
	db               *sqlx.DB
	queueClient      *rabbitmq.WorkflowQueue
	inventoryItemBus *inventoryitembus.Business
	locationBus      *inventorylocationbus.Business
	transactionBus   *inventorytransactionbus.Business
	productBus       *productbus.Business
	workflowBus      *workflow.Business
}

// NewAllocateInventoryHandler creates a new allocate inventory handler
func NewAllocateInventoryHandler(
	log *logger.Logger,
	db *sqlx.DB,
	queueClient *rabbitmq.WorkflowQueue,
	inventoryItemBus *inventoryitembus.Business,
	locationBus *inventorylocationbus.Business,
	transactionBus *inventorytransactionbus.Business,
	productBus *productbus.Business,
	workflowBus *workflow.Business,
) *AllocateInventoryHandler {
	return &AllocateInventoryHandler{
		log:              log,
		db:               db,
		queueClient:      queueClient,
		inventoryItemBus: inventoryItemBus,
		locationBus:      locationBus,
		transactionBus:   transactionBus,
		productBus:       productBus,
		workflowBus:      workflowBus,
	}
}

// GetType returns the action type
func (h *AllocateInventoryHandler) GetType() string {
	return "allocate_inventory"
}

// Validate validates the allocation configuration
func (h *AllocateInventoryHandler) Validate(config json.RawMessage) error {
	var cfg AllocateInventoryConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.InventoryItems) == 0 {
		return errors.New("inventory_items list is required and must not be empty")
	}

	validStrategies := map[string]bool{
		"fifo": true, "lifo": true, "nearest_expiry": true, "lowest_cost": true,
	}
	if !validStrategies[cfg.AllocationStrategy] {
		return fmt.Errorf("invalid allocation_strategy: %s", cfg.AllocationStrategy)
	}

	validModes := map[string]bool{"reserve": true, "allocate": true}
	if !validModes[cfg.AllocationMode] {
		return fmt.Errorf("invalid allocation_mode: %s", cfg.AllocationMode)
	}

	validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	if !validPriorities[cfg.Priority] {
		return fmt.Errorf("invalid priority: %s", cfg.Priority)
	}

	// Validate items
	for i, item := range cfg.InventoryItems {
		if item.ProductID == uuid.Nil {
			return fmt.Errorf("item %d: product_id is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be greater than 0", i)
		}
	}

	if cfg.AllocationMode == "reserve" && cfg.ReservationHours <= 0 {
		cfg.ReservationHours = 24 // Default to 24 hours
	}

	return nil
}

// Execute queues the allocation request for async processing.
// Returns QueuedAllocationResponse with tracking info.
func (h *AllocateInventoryHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg AllocateInventoryConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return QueuedAllocationResponse{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// Generate idempotency key based on execution context
	idempotencyKey := fmt.Sprintf("%s_%s_%s", execContext.ExecutionID, execContext.RuleID, h.GetType())

	// Check if this allocation was already processed (idempotency)
	existing, idempotencyResult, err := h.workflowBus.QueryAllocationResultByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return QueuedAllocationResponse{}, fmt.Errorf("execute: %w", err)
	}

	switch idempotencyResult {
	case workflow.IdempotencyNotFound:
		// Good - no existing allocation, we can proceed with processing
	case workflow.IdempotencyExists:
		h.log.Info(ctx, "Allocation already processed, returning existing result",
			"idempotency_key", idempotencyKey,
			"allocation_id", existing.ID)
		// Return error for already processed - caller should use GetResult
		return QueuedAllocationResponse{}, fmt.Errorf("allocation already processed with key: %s, allocation_id: %s", idempotencyKey, existing.ID)
	}

	// Create allocation request
	request := AllocationRequest{
		ID:          uuid.New(),
		ExecutionID: execContext.ExecutionID,
		Config:      cfg,
		Context:     execContext,
		Status:      "queued",
		Priority:    h.calculatePriority(cfg.Priority),
		CreatedAt:   time.Now(),
		RetryCount:  0,
		MaxRetries:  3,
	}

	// Serialize the request for the queued payload
	requestData, err := json.Marshal(request)
	if err != nil {
		return QueuedAllocationResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create standard QueuedPayload wrapper for async processing
	queuedPayload := workflow.QueuedPayload{
		RequestType:      h.GetType(), // "allocate_inventory"
		RequestData:      requestData,
		ExecutionContext: execContext,
		IdempotencyKey:   idempotencyKey,
	}

	payloadData, err := json.Marshal(queuedPayload)
	if err != nil {
		return QueuedAllocationResponse{}, fmt.Errorf("failed to marshal queued payload: %w", err)
	}

	// Convert to map for RabbitMQ message payload
	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadData, &payloadMap); err != nil {
		return QueuedAllocationResponse{}, fmt.Errorf("failed to convert payload to map: %w", err)
	}

	// Queue to RabbitMQ for async processing using generic async_action type
	message := &rabbitmq.Message{
		ID:           request.ID,
		Type:         "async_action", // Generic type for all async actions
		EntityName:   execContext.EntityName,
		EntityID:     execContext.EntityID,
		EventType:    h.GetType(), // "allocate_inventory"
		Payload:      payloadMap,
		Priority:     uint8(request.Priority),
		Attempts:     0,
		MaxAttempts:  request.MaxRetries,
		CreatedAt:    request.CreatedAt,
		ScheduledFor: request.CreatedAt,
		UserID:       execContext.UserID,
	}

	if err := h.queueClient.Publish(ctx, rabbitmq.QueueTypeInventory, message); err != nil {
		return QueuedAllocationResponse{}, fmt.Errorf("failed to queue allocation request: %w", err)
	}

	// Return immediate response with tracking info
	return QueuedAllocationResponse{
		AllocationID:   request.ID,
		Status:         "queued",
		IdempotencyKey: idempotencyKey,
		Priority:       request.Priority,
		Message:        "Allocation request queued for processing",
	}, nil
}

// ProcessQueued implements AsyncActionHandler for processing queued allocation requests.
// This is called by the queue manager when an async_action message is dequeued.
func (h *AllocateInventoryHandler) ProcessQueued(ctx context.Context, payload json.RawMessage, publisher *workflow.EventPublisher) error {
	// Deserialize the standard QueuedPayload wrapper
	var queuedPayload workflow.QueuedPayload
	if err := json.Unmarshal(payload, &queuedPayload); err != nil {
		return fmt.Errorf("failed to unmarshal queued payload: %w", err)
	}

	// Deserialize the action-specific AllocationRequest from RequestData
	var request AllocationRequest
	if err := json.Unmarshal(queuedPayload.RequestData, &request); err != nil {
		return fmt.Errorf("failed to unmarshal allocation request: %w", err)
	}

	// Process the allocation using existing logic
	result, err := h.ProcessAllocation(ctx, request)
	if err != nil {
		return fmt.Errorf("allocation processing failed: %w", err)
	}

	// Fire workflow event for downstream rule processing
	h.fireAllocationResultEvent(ctx, result, request, publisher)

	h.log.Info(ctx, "Async allocation completed",
		"allocation_id", result.AllocationID,
		"status", result.Status,
		"total_allocated", result.TotalAllocated)

	return nil
}

// fireAllocationResultEvent fires a workflow event for downstream rule processing.
// This is called internally by ProcessQueued to trigger rules like "on allocation success -> update status".
func (h *AllocateInventoryHandler) fireAllocationResultEvent(ctx context.Context, result *InventoryAllocationResult, request AllocationRequest, publisher *workflow.EventPublisher) {
	event := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "allocation_results",
		EntityID:   result.AllocationID,
		Timestamp:  result.CompletedAt,
		UserID:     request.Context.UserID,
		RawData: map[string]interface{}{
			"status":            result.Status,
			"reference_id":      request.Config.ReferenceID,
			"reference_type":    request.Config.ReferenceType,
			"total_allocated":   result.TotalAllocated,
			"total_requested":   result.TotalRequested,
			"allocated_items":   result.AllocatedItems,
			"failed_items":      result.FailedItems,
			"idempotency_key":   result.IdempotencyKey,
			"execution_time_ms": result.ExecutionTimeMs,
		},
	}

	publisher.PublishCustomEvent(ctx, event)
}

// ProcessAllocation handles the actual allocation logic (called by queue consumer)
func (h *AllocateInventoryHandler) ProcessAllocation(ctx context.Context, request AllocationRequest) (*InventoryAllocationResult, error) {
	startTime := time.Now()
	idempotencyKey := fmt.Sprintf("%s_%s_%s", request.ExecutionID, request.Context.RuleID, h.GetType())

	// Double-check idempotency in case of race conditions
	existing, idempotencyResult, err := h.workflowBus.QueryAllocationResultByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("idempotency check failed: %w", err)
	}

	switch idempotencyResult {
	case workflow.IdempotencyExists:
		// Allocation was already processed, return the existing result
		var cachedResult InventoryAllocationResult // <-- Use renamed struct
		if err := json.Unmarshal(existing.AllocationData, &cachedResult); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cached result: %w", err)
		}
		return &cachedResult, nil
	case workflow.IdempotencyNotFound:
		// Good - proceed with allocation
	}

	// Start transaction with appropriate isolation level
	tx, err := h.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, // Balance between consistency and performance
	})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result := &InventoryAllocationResult{
		AllocationID:   request.ID,
		Status:         "processing",
		AllocatedItems: []AllocatedItem{},
		FailedItems:    []FailedItem{},
		IdempotencyKey: idempotencyKey,
		CreatedAt:      request.CreatedAt,
		Warnings:       []string{},
	}

	// Process each item
	for _, item := range request.Config.InventoryItems {
		allocated, failed := h.allocateItem(ctx, tx, item, request.Config, request.Context)

		if allocated != nil {
			result.AllocatedItems = append(result.AllocatedItems, *allocated)
			result.TotalAllocated += allocated.AllocatedQuantity
		}
		if failed != nil {
			result.FailedItems = append(result.FailedItems, *failed)
			if !request.Config.AllowPartial {
				// Rollback if partial allocation not allowed
				return nil, fmt.Errorf("allocation failed for product %s: %s",
					failed.ProductID, failed.ErrorMessage)
			}
		}
		result.TotalRequested += item.Quantity
	}

	// Determine final status
	if len(result.FailedItems) == 0 && result.TotalAllocated == result.TotalRequested {
		result.Status = "success"
	} else if len(result.AllocatedItems) > 0 {
		result.Status = "partial"
	} else {
		result.Status = "failed"
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	txWorkflowBus, err := h.workflowBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactional workflow bus: %w", err)
	}

	_, err = txWorkflowBus.CreateAllocationResult(ctx, workflow.NewAllocationResult{
		IdempotencyKey: idempotencyKey,
		AllocationData: data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store allocation result: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	result.CompletedAt = time.Now()
	result.ExecutionTimeMs = time.Since(startTime).Milliseconds()

	h.log.Info(ctx, "Allocation completed",
		"allocation_id", result.AllocationID,
		"status", result.Status,
		"total_allocated", result.TotalAllocated,
		"execution_time_ms", result.ExecutionTimeMs)

	return result, nil
}

// allocateItem allocates a single item using the specified strategy
func (h *AllocateInventoryHandler) allocateItem(
	ctx context.Context,
	tx *sqlx.Tx,
	item AllocationItem,
	config AllocateInventoryConfig,
	execContext workflow.ActionExecutionContext,
) (*AllocatedItem, *FailedItem) {
	// Create transactional business instance
	txItemBus, err := h.inventoryItemBus.NewWithTx(tx)
	if err != nil {
		return nil, &FailedItem{
			ProductID:    item.ProductID,
			Reason:       "transaction_setup_failed",
			ErrorMessage: err.Error(),
		}
	}

	// Use the specialized query method for allocation
	items, err := txItemBus.QueryAvailableForAllocation(
		ctx,
		item.ProductID,
		item.LocationID,  // Optional: specific location
		item.WarehouseID, // Optional: specific warehouse
		config.AllocationStrategy,
		10, // Process in batches to avoid locking too many rows
	)

	/* TODO: Advanced Allocation Strategy Requirements
	 *
	 * For nearest_expiry:
	 *   - Fetch lot tracking data along with inventory
	 *   - Prioritize lots nearing expiration
	 *   - May need to split allocation across multiple lots
	 *
	 * For lowest_cost:
	 *   - Factor in warehouse-specific costs (storage, labor, shipping)
	 *   - Calculate total landed cost including shipping to customer
	 *   - May need real-time shipping rate calculations
	 *
	 * For nearest_location:
	 *   - Get customer delivery address from order/context
	 *   - Calculate distances from available warehouses
	 *   - Consider shipping zones and transit times
	 *   - Cache distance calculations for performance
	 *
	 * For load_balancing:
	 *   - Query current warehouse utilization levels
	 *   - Factor in pending picks and current workload
	 *   - Distribute evenly across warehouses with capacity
	 *
	 * For priority_zone:
	 *   - Check customer tier/priority level
	 *   - Route VIP orders to pick locations vs reserve
	 *   - Consider zone-specific SLAs and handling requirements
	 *
	 * Additional considerations:
	 *   - Multi-warehouse allocation for large orders
	 *   - Minimum order quantities per warehouse
	 *   - Shipping cutoff times by warehouse
	 *   - Weekend/holiday warehouse schedules
	 */

	if err != nil {
		return nil, &FailedItem{
			ProductID:         item.ProductID,
			RequestedQuantity: item.Quantity,
			Reason:            "query_failed",
			ErrorMessage:      err.Error(),
		}
	}

	if len(items) == 0 {
		return nil, &FailedItem{
			ProductID:         item.ProductID,
			RequestedQuantity: item.Quantity,
			AvailableQuantity: 0,
			Reason:            "insufficient_inventory",
			ErrorMessage:      "No inventory available",
		}
	}

	// Rest of allocation logic remains the same...
	remaining := item.Quantity
	totalAllocated := 0
	var allocatedItem *AllocatedItem

	// Process available inventory using selected strategy
	for _, invItem := range items {
		if remaining <= 0 {
			break
		}

		available := invItem.Quantity - invItem.ReservedQuantity - invItem.AllocatedQuantity
		if available <= 0 {
			continue
		}

		toAllocate := min(remaining, available)

		// Update inventory based on allocation mode
		var update inventoryitembus.UpdateInventoryItem
		if config.AllocationMode == "reserve" {
			newReserved := invItem.ReservedQuantity + toAllocate
			update.ReservedQuantity = &newReserved
		} else {
			newAllocated := invItem.AllocatedQuantity + toAllocate
			update.AllocatedQuantity = &newAllocated
		}

		_, err := txItemBus.Update(ctx, invItem, update)
		if err != nil {
			return nil, &FailedItem{
				ProductID:         item.ProductID,
				RequestedQuantity: item.Quantity,
				Reason:            "update_failed",
				ErrorMessage:      err.Error(),
			}
		}

		// Track allocation details
		if allocatedItem == nil {
			allocatedItem = &AllocatedItem{
				ProductID:         item.ProductID,
				LocationID:        invItem.LocationID,
				RequestedQuantity: item.Quantity,
				AllocatedQuantity: 0,
				InventoryID:       invItem.ID,
				AllocationMode:    config.AllocationMode,
			}

			if config.AllocationMode == "reserve" {
				expiresAt := time.Now().Add(time.Duration(config.ReservationHours) * time.Hour)
				allocatedItem.ExpiresAt = &expiresAt
			}
		}

		remaining -= toAllocate
		totalAllocated += toAllocate
	}

	// Check if we allocated enough
	if remaining > 0 && !config.AllowPartial {
		return nil, &FailedItem{
			ProductID:         item.ProductID,
			RequestedQuantity: item.Quantity,
			AvailableQuantity: item.Quantity - remaining,
			Reason:            "insufficient_inventory",
			ErrorMessage:      fmt.Sprintf("Only %d available, %d requested", item.Quantity-remaining, item.Quantity),
		}
	}

	if allocatedItem != nil {
		allocatedItem.AllocatedQuantity = totalAllocated
	}

	return allocatedItem, nil
}

// Helper functions
func (h *AllocateInventoryHandler) calculatePriority(priority string) int {
	priorities := map[string]int{
		"low":      1,
		"medium":   5,
		"high":     10,
		"critical": 20,
	}
	if p, ok := priorities[priority]; ok {
		return p
	}
	return 5
}

func (h *AllocateInventoryHandler) requestToPayload(request AllocationRequest) map[string]interface{} {
	data, _ := json.Marshal(request)
	var payload map[string]interface{}
	json.Unmarshal(data, &payload)
	return payload
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
