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
	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productbus"
	"github.com/timmaaaz/ichor/business/domain/movement/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/warehouse/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
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

// AllocationResult represents the result of an allocation
type AllocationResult struct {
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
	InventoryItemID   uuid.UUID  `json:"inventory_item_id"`
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
	Result         *AllocationResult
}

// Database models for allocation tracking
type allocationResult struct {
	ID             string    `db:"id"`
	IdempotencyKey string    `db:"idempotency_key"`
	AllocationData []byte    `db:"allocation_data"`
	CreatedAt      time.Time `db:"created_at"`
}

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
) *AllocateInventoryHandler {
	return &AllocateInventoryHandler{
		log:              log,
		db:               db,
		queueClient:      queueClient,
		inventoryItemBus: inventoryItemBus,
		locationBus:      locationBus,
		transactionBus:   transactionBus,
		productBus:       productBus,
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

// Execute queues the allocation request for async processing
func (h *AllocateInventoryHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg AllocateInventoryConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return QueuedAllocationResponse{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// Generate idempotency key based on execution context
	idempotencyKey := fmt.Sprintf("%s_%s_%s", execContext.ExecutionID, execContext.RuleID, h.GetType())

	// Check if this allocation was already processed (idempotency)
	existing, err := h.checkIdempotency(ctx, idempotencyKey)
	if err != nil {
		return QueuedAllocationResponse{}, fmt.Errorf("idempotency check failed: %w", err)
	}
	if existing != nil {
		h.log.Info(ctx, "Allocation already processed, returning existing result",
			"idempotency_key", idempotencyKey,
			"allocation_id", existing.AllocationID)
		// Return error for already processed - caller should use GetResult
		return QueuedAllocationResponse{}, fmt.Errorf("allocation already processed with key: %s", idempotencyKey)
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

	// Queue to RabbitMQ for async processing
	queueType := rabbitmq.QueueTypeInventory
	message := &rabbitmq.Message{
		ID:           request.ID,
		Type:         "inventory_allocation",
		EntityName:   execContext.EntityName,
		EntityID:     execContext.EntityID,
		EventType:    "allocate_inventory",
		Payload:      h.requestToPayload(request),
		Priority:     uint8(request.Priority),
		Attempts:     0,
		MaxAttempts:  request.MaxRetries,
		CreatedAt:    request.CreatedAt,
		ScheduledFor: request.CreatedAt,
		UserID:       execContext.UserID,
	}

	if err := h.queueClient.Publish(ctx, queueType, message); err != nil {
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

// GetResult retrieves the result of a previously processed allocation
func (h *AllocateInventoryHandler) GetResult(ctx context.Context, idempotencyKey string) (*AllocationResult, error) {
	return h.checkIdempotency(ctx, idempotencyKey)
}

// ProcessAllocation handles the actual allocation logic (called by queue consumer)
func (h *AllocateInventoryHandler) ProcessAllocation(ctx context.Context, request AllocationRequest) (*AllocationResult, error) {
	startTime := time.Now()
	idempotencyKey := fmt.Sprintf("%s_%s_%s", request.ExecutionID, request.Context.RuleID, h.GetType())

	// Double-check idempotency in case of race conditions
	existing, err := h.checkIdempotency(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("idempotency check failed: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// Start transaction with appropriate isolation level
	tx, err := h.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, // Balance between consistency and performance
	})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result := &AllocationResult{
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

	// Store result for idempotency
	if err := h.storeAllocationResult(ctx, tx, result); err != nil {
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
	// Create transactional business instances
	txItemBus, err := h.inventoryItemBus.NewWithTx(tx)
	if err != nil {
		return nil, &FailedItem{
			ProductID:    item.ProductID,
			Reason:       "transaction_setup_failed",
			ErrorMessage: err.Error(),
		}
	}

	// Query available inventory
	filter := inventoryitembus.QueryFilter{
		ProductID:  &item.ProductID,
		LocationID: item.LocationID,
		// Need custom filter for available quantity
	}

	// Order by created_date for FIFO/LIFO
	orderBy := inventoryitembus.DefaultOrderBy
	if config.AllocationStrategy == "lifo" {
		orderBy = order.NewBy("created_date", order.DESC)
	}

	items, err := txItemBus.Query(ctx, filter, orderBy, page.MustParse("1", "10"))
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

	// Allocate from available inventory
	remaining := item.Quantity
	totalAllocated := 0
	var allocatedItem *AllocatedItem

	for _, invItem := range items {
		if remaining <= 0 {
			break
		}

		// Calculate available quantity
		available := invItem.Quantity - invItem.ReservedQuantity - invItem.AllocatedQuantity
		if available <= 0 {
			continue
		}

		toAllocate := min(remaining, available)

		// Update inventory item based on allocation mode
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

		// Create inventory transaction record
		txTransBus, err := h.transactionBus.NewWithTx(tx)
		if err != nil {
			h.log.Error(ctx, "Failed to create transaction bus", "error", err)
		} else {
			transaction := inventorytransactionbus.NewInventoryTransaction{
				ProductID:       item.ProductID,
				LocationID:      invItem.LocationID,
				UserID:          execContext.UserID,
				TransactionType: config.AllocationMode,
				Quantity:        -toAllocate,
				ReferenceNumber: config.ReferenceID,
			}

			if _, err := txTransBus.Create(ctx, transaction); err != nil {
				h.log.Error(ctx, "Failed to create transaction record", "error", err)
				// Decide if this should fail the allocation or just log
			}
		}

		// Track allocation
		allocatedItem = &AllocatedItem{
			ProductID:         item.ProductID,
			LocationID:        invItem.LocationID,
			RequestedQuantity: item.Quantity,
			AllocatedQuantity: toAllocate,
			InventoryItemID:   invItem.ItemID,
			AllocationMode:    config.AllocationMode,
		}

		if config.AllocationMode == "reserve" {
			expiresAt := time.Now().Add(time.Duration(config.ReservationHours) * time.Hour)
			allocatedItem.ExpiresAt = &expiresAt
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

	if allocatedItem == nil {
		return nil, &FailedItem{
			ProductID:         item.ProductID,
			RequestedQuantity: item.Quantity,
			AvailableQuantity: 0,
			Reason:            "no_allocation",
			ErrorMessage:      "Unable to allocate any inventory",
		}
	}

	allocatedItem.AllocatedQuantity = totalAllocated
	return allocatedItem, nil
}

// buildInventoryQuery builds the query based on allocation strategy
func (h *AllocateInventoryHandler) buildInventoryQuery(strategy string, item AllocationItem) string {
	baseQuery := `
		SELECT id, product_id, location_id, quantity, reserved_quantity, allocated_quantity, created_date
		FROM inventory_items
		WHERE product_id = :product_id
		AND (quantity - reserved_quantity - allocated_quantity) > 0`

	if item.WarehouseID != nil {
		baseQuery += ` AND location_id IN (SELECT id FROM inventory_locations WHERE warehouse_id = :warehouse_id)`
	}
	if item.LocationID != nil {
		baseQuery = `
		SELECT id, product_id, location_id, quantity, reserved_quantity, allocated_quantity, created_date
		FROM inventory_items
		WHERE product_id = :product_id
		AND location_id = :location_id
		AND (quantity - reserved_quantity - allocated_quantity) > 0`
	}

	// Add ordering based on strategy
	switch strategy {
	case "fifo":
		baseQuery += " ORDER BY created_date ASC"
	case "lifo":
		baseQuery += " ORDER BY created_date DESC"
	case "nearest_expiry":
		// Would need to join with lot_trackings table
		baseQuery += " ORDER BY created_date ASC" // Simplified for now
	case "lowest_cost":
		// Would need to join with product_costs table
		baseQuery += " ORDER BY created_date ASC" // Simplified for now
	default:
		baseQuery += " ORDER BY created_date ASC"
	}

	baseQuery += " LIMIT 10 FOR UPDATE" // Lock rows for update
	return baseQuery
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

func (h *AllocateInventoryHandler) checkIdempotency(ctx context.Context, key string) (*AllocationResult, error) {
	data := struct {
		IdempotencyKey string `db:"idempotency_key"`
	}{
		IdempotencyKey: key,
	}

	const q = `
		SELECT id, idempotency_key, allocation_data, created_at
		FROM allocation_results 
		WHERE idempotency_key = :idempotency_key`

	var dbResult allocationResult
	if err := sqldb.NamedQueryStruct(ctx, h.log, h.db, q, data, &dbResult); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var result AllocationResult
	if err := json.Unmarshal(dbResult.AllocationData, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *AllocateInventoryHandler) storeAllocationResult(ctx context.Context, tx *sqlx.Tx, result *AllocationResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	storeData := struct {
		ID             string    `db:"id"`
		IdempotencyKey string    `db:"idempotency_key"`
		AllocationData []byte    `db:"allocation_data"`
		CreatedAt      time.Time `db:"created_at"`
	}{
		ID:             result.AllocationID.String(),
		IdempotencyKey: result.IdempotencyKey,
		AllocationData: data,
		CreatedAt:      result.CreatedAt,
	}

	const q = `
		INSERT INTO allocation_results (id, idempotency_key, allocation_data, created_at)
		VALUES (:id, :idempotency_key, :allocation_data, :created_at)
		ON CONFLICT (idempotency_key) DO NOTHING`

	return sqldb.NamedExecContext(ctx, h.log, tx, q, storeData)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
