// Package pickingapp provides the application layer for warehouse picking operations.
// All pick/short-pick operations are wrapped in a single DB transaction spanning
// multiple business domains (inventory, order line items, orders).
package pickingapp

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/timmaaaz/ichor/app/domain/sales/orderlineitemsapp"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// App manages the set of app layer API functions for picking operations.
type App struct {
	log                          *logger.Logger
	db                           *sqlx.DB
	ordersBus                    *ordersbus.Business
	orderLineItemsBus             *orderlineitemsbus.Business
	inventoryItemBus             *inventoryitembus.Business
	inventoryTransactionBus      *inventorytransactionbus.Business
	orderFulfillmentStatusBus    *orderfulfillmentstatusbus.Business
	lineItemFulfillmentStatusBus *lineitemfulfillmentstatusbus.Business
}

// NewApp constructs a picking app API for use.
func NewApp(
	log *logger.Logger,
	db *sqlx.DB,
	ordersBus *ordersbus.Business,
	orderLineItemsBus *orderlineitemsbus.Business,
	inventoryItemBus *inventoryitembus.Business,
	inventoryTransactionBus *inventorytransactionbus.Business,
	orderFulfillmentStatusBus *orderfulfillmentstatusbus.Business,
	lineItemFulfillmentStatusBus *lineitemfulfillmentstatusbus.Business,
) *App {
	return &App{
		log:                          log,
		db:                           db,
		ordersBus:                    ordersBus,
		orderLineItemsBus:             orderLineItemsBus,
		inventoryItemBus:             inventoryItemBus,
		inventoryTransactionBus:      inventoryTransactionBus,
		orderFulfillmentStatusBus:    orderFulfillmentStatusBus,
		lineItemFulfillmentStatusBus: lineItemFulfillmentStatusBus,
	}
}

// PickQuantity records a full or partial pick for an order line item atomically.
func (a *App) PickQuantity(ctx context.Context, lineItemID uuid.UUID, req PickQuantityRequest) (orderlineitemsapp.OrderLineItem, error) {
	quantity, err := strconv.Atoi(req.Quantity)
	if err != nil || quantity <= 0 {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "quantity must be a positive integer")
	}

	pickedBy, err := uuid.Parse(req.PickedBy)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "invalid picked_by uuid")
	}

	locationID, err := uuid.Parse(req.LocationID)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "invalid location_id uuid")
	}

	// Fetch line item.
	lineItem, err := a.orderLineItemsBus.QueryByID(ctx, lineItemID)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.NewError(err)
	}

	// Fetch parent order.
	order, err := a.ordersBus.QueryByID(ctx, lineItem.OrderID)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.NewError(err)
	}

	// Resolve status IDs outside the transaction (read-only).
	pickingStatusID, err := a.resolveOrderFulfillmentStatusID(ctx, "PICKING")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve PICKING status: %s", err)
	}
	packingStatusID, err := a.resolveOrderFulfillmentStatusID(ctx, "PACKING")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve PACKING status: %s", err)
	}
	pickedStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "PICKED")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve PICKED status: %s", err)
	}
	cancelledLineStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "CANCELLED")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve CANCELLED line status: %s", err)
	}

	// Validate order is in PICKING status.
	if order.FulfillmentStatusID != pickingStatusID {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "order must be in PICKING status to pick items")
	}

	// Validate quantity.
	remaining := lineItem.Quantity - lineItem.PickedQuantity
	if quantity > remaining {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "quantity exceeds remaining pick quantity (%d remaining)", remaining)
	}

	// Begin transaction.
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "begin tx: %s", err)
	}
	defer tx.Rollback()

	// Create tx-bound buses.
	txInventoryItemBus, err := a.inventoryItemBus.NewWithTx(tx)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "tx inventory item bus: %s", err)
	}
	txInventoryTransactionBus, err := a.inventoryTransactionBus.NewWithTx(tx)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "tx inventory transaction bus: %s", err)
	}
	txOrderLineItemsBus, err := a.orderLineItemsBus.NewWithTx(tx)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "tx order line items bus: %s", err)
	}
	txOrdersBus, err := a.ordersBus.NewWithTx(tx)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "tx orders bus: %s", err)
	}

	// Lock and fetch inventory item for the pick location (FOR UPDATE).
	invItems, err := txInventoryItemBus.QueryAvailableForAllocation(ctx, lineItem.ProductID, &locationID, nil, "fefo", 1)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "query inventory: %s", err)
	}
	if len(invItems) == 0 {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "insufficient stock at specified location")
	}
	invItem := invItems[0]

	availableQty := invItem.Quantity - invItem.AllocatedQuantity
	if quantity > availableQty {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "insufficient stock at specified location")
	}

	// Decrement inventory quantity and allocated quantity.
	newQty := invItem.Quantity - quantity
	newAllocQty := invItem.AllocatedQuantity - quantity
	if newAllocQty < 0 {
		newAllocQty = 0
	}
	if _, err := txInventoryItemBus.Update(ctx, invItem, inventoryitembus.UpdateInventoryItem{
		Quantity:          &newQty,
		AllocatedQuantity: &newAllocQty,
	}); err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "update inventory item: %s", err)
	}

	// Create inventory transaction (negative quantity = outbound pick).
	negQty := -quantity
	if _, err := txInventoryTransactionBus.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       lineItem.ProductID,
		LocationID:      locationID,
		UserID:          pickedBy,
		TransactionType: "pick",
		Quantity:        negQty,
		ReferenceNumber: order.Number,
		TransactionDate: time.Now().UTC(),
	}); err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "create inventory transaction: %s", err)
	}

	// Update line item: increment picked_quantity and set status to PICKED.
	newPickedQty := lineItem.PickedQuantity + quantity
	updatedLineItem, err := txOrderLineItemsBus.Update(ctx, lineItem, orderlineitemsbus.UpdateOrderLineItem{
		PickedQuantity:                &newPickedQty,
		LineItemFulfillmentStatusesID: &pickedStatusID,
		UpdatedBy:                     &pickedBy,
	})
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "update order line item: %s", err)
	}

	// Check if all line items are now in a terminal state → advance order to PACKING.
	allItems, err := txOrderLineItemsBus.Query(ctx, orderlineitemsbus.QueryFilter{OrderID: &lineItem.OrderID}, orderlineitemsbus.DefaultOrderBy, page.MustParse("1", "1000"))
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "query order line items: %s", err)
	}

	if allItemsPickingComplete(allItems, pickedStatusID, cancelledLineStatusID) {
		if _, err := txOrdersBus.Update(ctx, order, ordersbus.UpdateOrder{
			FulfillmentStatusID: &packingStatusID,
			UpdatedBy:           &pickedBy,
		}); err != nil {
			return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "advance order to PACKING: %s", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "commit tx: %s", err)
	}

	return orderlineitemsapp.ToAppOrderLineItem(updatedLineItem), nil
}

// ShortPick records a short-pick event for an order line item.
func (a *App) ShortPick(ctx context.Context, lineItemID uuid.UUID, req ShortPickRequest) (orderlineitemsapp.OrderLineItem, error) {
	pickedQty, err := strconv.Atoi(req.PickedQuantity)
	if err != nil || pickedQty < 0 {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "picked_quantity must be a non-negative integer")
	}

	pickedBy, err := uuid.Parse(req.PickedBy)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "invalid picked_by uuid")
	}

	// LocationID is optional for backorder (no inventory touched); required for all other types.
	var locationID uuid.UUID
	if req.LocationID != "" {
		locationID, err = uuid.Parse(req.LocationID)
		if err != nil {
			return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "invalid location_id uuid")
		}
	} else if req.ShortPickType != "backorder" {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "location_id is required for %s picks", req.ShortPickType)
	}

	// Validate substitute fields for "substitute" type.
	if req.ShortPickType == "substitute" {
		if req.SubstituteProductID == nil || *req.SubstituteProductID == "" {
			return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "substitute_product_id required for substitute picks")
		}
	}

	// Fetch line item.
	lineItem, err := a.orderLineItemsBus.QueryByID(ctx, lineItemID)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.NewError(err)
	}

	// Fetch parent order.
	order, err := a.ordersBus.QueryByID(ctx, lineItem.OrderID)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.NewError(err)
	}

	// Resolve status IDs outside the transaction.
	pickingStatusID, err := a.resolveOrderFulfillmentStatusID(ctx, "PICKING")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve PICKING status: %s", err)
	}
	packingStatusID, err := a.resolveOrderFulfillmentStatusID(ctx, "PACKING")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve PACKING status: %s", err)
	}
	partiallyPickedStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "PARTIALLY_PICKED")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve PARTIALLY_PICKED status: %s", err)
	}
	backorderedStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "BACKORDERED")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve BACKORDERED status: %s", err)
	}
	pendingReviewStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "PENDING_REVIEW")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve PENDING_REVIEW status: %s", err)
	}
	pickedStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "PICKED")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve PICKED status: %s", err)
	}
	cancelledLineStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "CANCELLED")
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "resolve CANCELLED line status: %s", err)
	}

	// Validate order is in PICKING status.
	if order.FulfillmentStatusID != pickingStatusID {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "order must be in PICKING status")
	}

	// Begin transaction.
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "begin tx: %s", err)
	}
	defer tx.Rollback()

	txInventoryItemBus, err := a.inventoryItemBus.NewWithTx(tx)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "tx inventory item bus: %s", err)
	}
	txInventoryTransactionBus, err := a.inventoryTransactionBus.NewWithTx(tx)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "tx inventory transaction bus: %s", err)
	}
	txOrderLineItemsBus, err := a.orderLineItemsBus.NewWithTx(tx)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "tx order line items bus: %s", err)
	}
	txOrdersBus, err := a.ordersBus.NewWithTx(tx)
	if err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "tx orders bus: %s", err)
	}

	var updatedLineItem orderlineitemsbus.OrderLineItem
	shouldAdvanceOrder := false

	switch req.ShortPickType {
	case "partial":
		backorderedQty := lineItem.Quantity - pickedQty
		updatedLineItem, shouldAdvanceOrder, err = a.doPartialOrBackorderPick(
			ctx, txInventoryItemBus, txInventoryTransactionBus, txOrderLineItemsBus,
			lineItem, order, pickedQty, backorderedQty, partiallyPickedStatusID,
			locationID, pickedBy, req.ShortPickReason,
		)
		if err != nil {
			return orderlineitemsapp.OrderLineItem{}, err
		}

	case "backorder":
		updatedLineItem, shouldAdvanceOrder, err = a.doPartialOrBackorderPick(
			ctx, txInventoryItemBus, txInventoryTransactionBus, txOrderLineItemsBus,
			lineItem, order, 0, lineItem.Quantity, backorderedStatusID,
			locationID, pickedBy, req.ShortPickReason,
		)
		if err != nil {
			return orderlineitemsapp.OrderLineItem{}, err
		}

	case "substitute":
		subProductID, err := uuid.Parse(*req.SubstituteProductID)
		if err != nil {
			return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "invalid substitute_product_id")
		}
		subQty := pickedQty
		if req.SubstituteQuantity != nil {
			subQty, err = strconv.Atoi(*req.SubstituteQuantity)
			if err != nil || subQty <= 0 {
				return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.InvalidArgument, "substitute_quantity must be a positive integer")
			}
		}
		updatedLineItem, shouldAdvanceOrder, err = a.doSubstitutePick(
			ctx, txInventoryItemBus, txInventoryTransactionBus, txOrderLineItemsBus,
			lineItem, order, subProductID, subQty, backorderedStatusID, pickedStatusID,
			locationID, pickedBy, req.ShortPickReason,
		)
		if err != nil {
			return orderlineitemsapp.OrderLineItem{}, err
		}

	case "skip":
		// Record reason, set status to PENDING_REVIEW; do NOT advance order.
		reason := req.ShortPickReason
		updatedLineItem, err = txOrderLineItemsBus.Update(ctx, lineItem, orderlineitemsbus.UpdateOrderLineItem{
			LineItemFulfillmentStatusesID: &pendingReviewStatusID,
			ShortPickReason:               &reason,
			UpdatedBy:                     &pickedBy,
		})
		if err != nil {
			return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "update order line item: %s", err)
		}
		// Do not advance order — supervisor must resolve PENDING_REVIEW items first.
		shouldAdvanceOrder = false
	}

	if shouldAdvanceOrder {
		allItems, err := txOrderLineItemsBus.Query(ctx, orderlineitemsbus.QueryFilter{OrderID: &lineItem.OrderID}, orderlineitemsbus.DefaultOrderBy, page.MustParse("1", "1000"))
		if err != nil {
			return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "query order line items: %s", err)
		}
		if allItemsShortPickComplete(allItems, pickedStatusID, partiallyPickedStatusID, backorderedStatusID, cancelledLineStatusID) {
			if _, err := txOrdersBus.Update(ctx, order, ordersbus.UpdateOrder{
				FulfillmentStatusID: &packingStatusID,
				UpdatedBy:           &pickedBy,
			}); err != nil {
				return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "advance order to PACKING: %s", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return orderlineitemsapp.OrderLineItem{}, errs.Newf(errs.Internal, "commit tx: %s", err)
	}

	return orderlineitemsapp.ToAppOrderLineItem(updatedLineItem), nil
}

// CompletePacking advances an order from PACKING to READY_TO_SHIP.
func (a *App) CompletePacking(ctx context.Context, orderID uuid.UUID, req CompletePackingRequest) (ordersapp.Order, error) {
	packedBy, err := uuid.Parse(req.PackedBy)
	if err != nil {
		return ordersapp.Order{}, errs.Newf(errs.InvalidArgument, "invalid packed_by uuid")
	}

	order, err := a.ordersBus.QueryByID(ctx, orderID)
	if err != nil {
		return ordersapp.Order{}, errs.NewError(err)
	}

	packingStatusID, err := a.resolveOrderFulfillmentStatusID(ctx, "PACKING")
	if err != nil {
		return ordersapp.Order{}, errs.Newf(errs.Internal, "resolve PACKING status: %s", err)
	}
	if order.FulfillmentStatusID != packingStatusID {
		return ordersapp.Order{}, errs.Newf(errs.InvalidArgument, "order must be in PACKING status")
	}

	readyToShipStatusID, err := a.resolveOrderFulfillmentStatusID(ctx, "READY_TO_SHIP")
	if err != nil {
		return ordersapp.Order{}, errs.Newf(errs.Internal, "resolve READY_TO_SHIP status: %s", err)
	}

	pickedStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "PICKED")
	if err != nil {
		return ordersapp.Order{}, errs.Newf(errs.Internal, "resolve PICKED status: %s", err)
	}
	partiallyPickedStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "PARTIALLY_PICKED")
	if err != nil {
		return ordersapp.Order{}, errs.Newf(errs.Internal, "resolve PARTIALLY_PICKED status: %s", err)
	}
	backorderedStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "BACKORDERED")
	if err != nil {
		return ordersapp.Order{}, errs.Newf(errs.Internal, "resolve BACKORDERED status: %s", err)
	}
	cancelledStatusID, err := a.resolveLineItemFulfillmentStatusID(ctx, "CANCELLED")
	if err != nil {
		return ordersapp.Order{}, errs.Newf(errs.Internal, "resolve CANCELLED status: %s", err)
	}

	// Verify all line items are in a terminal picking state.
	lineItems, err := a.orderLineItemsBus.Query(ctx, orderlineitemsbus.QueryFilter{OrderID: &orderID}, orderlineitemsbus.DefaultOrderBy, page.MustParse("1", "1000"))
	if err != nil {
		return ordersapp.Order{}, errs.Newf(errs.Internal, "query line items: %s", err)
	}
	terminalStatuses := map[uuid.UUID]bool{
		pickedStatusID:          true,
		partiallyPickedStatusID: true,
		backorderedStatusID:     true,
		cancelledStatusID:       true,
	}
	for _, item := range lineItems {
		if !terminalStatuses[item.LineItemFulfillmentStatusesID] {
			return ordersapp.Order{}, errs.Newf(errs.InvalidArgument, "all line items must be picked before completing packing")
		}
	}

	updatedOrder, err := a.ordersBus.Update(ctx, order, ordersbus.UpdateOrder{
		FulfillmentStatusID: &readyToShipStatusID,
		UpdatedBy:           &packedBy,
	})
	if err != nil {
		return ordersapp.Order{}, errs.Newf(errs.Internal, "update order: %s", err)
	}

	return ordersapp.ToAppOrder(updatedOrder), nil
}

// =============================================================================
// helpers

func (a *App) resolveOrderFulfillmentStatusID(ctx context.Context, name string) (uuid.UUID, error) {
	statuses, err := a.orderFulfillmentStatusBus.Query(ctx,
		orderfulfillmentstatusbus.QueryFilter{Name: &name},
		orderfulfillmentstatusbus.DefaultOrderBy,
		page.MustParse("1", "1"),
	)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("query: %w", err)
	}
	if len(statuses) == 0 {
		return uuid.UUID{}, fmt.Errorf("order fulfillment status %q not found", name)
	}
	return statuses[0].ID, nil
}

func (a *App) resolveLineItemFulfillmentStatusID(ctx context.Context, name string) (uuid.UUID, error) {
	statuses, err := a.lineItemFulfillmentStatusBus.Query(ctx,
		lineitemfulfillmentstatusbus.QueryFilter{Name: &name},
		lineitemfulfillmentstatusbus.DefaultOrderBy,
		page.MustParse("1", "1"),
	)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("query: %w", err)
	}
	if len(statuses) == 0 {
		return uuid.UUID{}, fmt.Errorf("line item fulfillment status %q not found", name)
	}
	return statuses[0].ID, nil
}

// doPartialOrBackorderPick handles the "partial" and "backorder" short-pick types.
func (a *App) doPartialOrBackorderPick(
	ctx context.Context,
	txInventoryItemBus *inventoryitembus.Business,
	txInventoryTransactionBus *inventorytransactionbus.Business,
	txOrderLineItemsBus *orderlineitemsbus.Business,
	lineItem orderlineitemsbus.OrderLineItem,
	order ordersbus.Order,
	pickedQty int,
	backorderedQty int,
	lineItemStatusID uuid.UUID,
	locationID uuid.UUID,
	pickedBy uuid.UUID,
	reason string,
) (orderlineitemsbus.OrderLineItem, bool, error) {
	// Only create an inventory transaction if we actually picked something.
	if pickedQty > 0 {
		invItems, err := txInventoryItemBus.QueryAvailableForAllocation(ctx, lineItem.ProductID, &locationID, nil, "fefo", 1)
		if err != nil {
			return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "query inventory: %s", err)
		}
		if len(invItems) == 0 || (invItems[0].Quantity-invItems[0].AllocatedQuantity) < pickedQty {
			return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.InvalidArgument, "insufficient stock at specified location")
		}
		invItem := invItems[0]
		newQty := invItem.Quantity - pickedQty
		newAllocQty := invItem.AllocatedQuantity - pickedQty
		if newAllocQty < 0 {
			newAllocQty = 0
		}
		if _, err := txInventoryItemBus.Update(ctx, invItem, inventoryitembus.UpdateInventoryItem{
			Quantity:          &newQty,
			AllocatedQuantity: &newAllocQty,
		}); err != nil {
			return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "update inventory item: %s", err)
		}

		negQty := -pickedQty
		if _, err := txInventoryTransactionBus.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
			ProductID:       lineItem.ProductID,
			LocationID:      locationID,
			UserID:          pickedBy,
			TransactionType: "pick",
			Quantity:        negQty,
			ReferenceNumber: order.Number,
			TransactionDate: time.Now().UTC(),
		}); err != nil {
			return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "create inventory transaction: %s", err)
		}
	}

	newPickedQty := lineItem.PickedQuantity + pickedQty
	updatedItem, err := txOrderLineItemsBus.Update(ctx, lineItem, orderlineitemsbus.UpdateOrderLineItem{
		PickedQuantity:                &newPickedQty,
		BackorderedQuantity:           &backorderedQty,
		LineItemFulfillmentStatusesID: &lineItemStatusID,
		ShortPickReason:               &reason,
		UpdatedBy:                     &pickedBy,
	})
	if err != nil {
		return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "update order line item: %s", err)
	}

	return updatedItem, true, nil
}

// doSubstitutePick handles the "substitute" short-pick type.
func (a *App) doSubstitutePick(
	ctx context.Context,
	txInventoryItemBus *inventoryitembus.Business,
	txInventoryTransactionBus *inventorytransactionbus.Business,
	txOrderLineItemsBus *orderlineitemsbus.Business,
	lineItem orderlineitemsbus.OrderLineItem,
	order ordersbus.Order,
	subProductID uuid.UUID,
	subQty int,
	backorderedStatusID uuid.UUID,
	pickedStatusID uuid.UUID,
	locationID uuid.UUID,
	pickedBy uuid.UUID,
	reason string,
) (orderlineitemsbus.OrderLineItem, bool, error) {
	// Mark original line item as BACKORDERED.
	zero := 0
	orderedQty := lineItem.Quantity
	if _, err := txOrderLineItemsBus.Update(ctx, lineItem, orderlineitemsbus.UpdateOrderLineItem{
		PickedQuantity:                &zero,
		BackorderedQuantity:           &orderedQty,
		LineItemFulfillmentStatusesID: &backorderedStatusID,
		ShortPickReason:               &reason,
		UpdatedBy:                     &pickedBy,
	}); err != nil {
		return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "backorder original line item: %s", err)
	}

	// Check substitute inventory.
	invItems, err := txInventoryItemBus.QueryAvailableForAllocation(ctx, subProductID, &locationID, nil, "fefo", 1)
	if err != nil {
		return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "query substitute inventory: %s", err)
	}
	if len(invItems) == 0 || (invItems[0].Quantity-invItems[0].AllocatedQuantity) < subQty {
		return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.InvalidArgument, "insufficient stock for substitute product")
	}
	invItem := invItems[0]
	newQty := invItem.Quantity - subQty
	newAllocQty := invItem.AllocatedQuantity - subQty
	if newAllocQty < 0 {
		newAllocQty = 0
	}
	if _, err := txInventoryItemBus.Update(ctx, invItem, inventoryitembus.UpdateInventoryItem{
		Quantity:          &newQty,
		AllocatedQuantity: &newAllocQty,
	}); err != nil {
		return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "update substitute inventory: %s", err)
	}

	// Create inventory transaction for substitute pick.
	negQty := -subQty
	if _, err := txInventoryTransactionBus.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       subProductID,
		LocationID:      locationID,
		UserID:          pickedBy,
		TransactionType: "pick",
		Quantity:        negQty,
		ReferenceNumber: order.Number,
		TransactionDate: time.Now().UTC(),
	}); err != nil {
		return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "create substitute inventory transaction: %s", err)
	}

	// Compute line total for the substitute: unitPrice * subQty - discount.
	unitDec, _ := decimal.NewFromString(lineItem.UnitPrice.Value())
	discDec, _ := decimal.NewFromString(lineItem.Discount.Value())
	subLineTotal, _ := types.ParseMoney(unitDec.Mul(decimal.NewFromInt(int64(subQty))).Sub(discDec).StringFixed(2))

	// Create new line item for substitute product.
	subLineItem, err := txOrderLineItemsBus.Create(ctx, orderlineitemsbus.NewOrderLineItem{
		OrderID:                       lineItem.OrderID,
		ProductID:                     subProductID,
		Description:                   fmt.Sprintf("Substitute for %s", lineItem.Description),
		Quantity:                      subQty,
		UnitPrice:                     lineItem.UnitPrice,
		Discount:                      lineItem.Discount,
		DiscountType:                  lineItem.DiscountType,
		LineTotal:                     subLineTotal,
		LineItemFulfillmentStatusesID: pickedStatusID,
		CreatedBy:                     pickedBy,
	})
	if err != nil {
		return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "create substitute line item: %s", err)
	}

	// Update substitute line item's picked_quantity.
	updatedSub, err := txOrderLineItemsBus.Update(ctx, subLineItem, orderlineitemsbus.UpdateOrderLineItem{
		PickedQuantity: &subQty,
		UpdatedBy:      &pickedBy,
	})
	if err != nil {
		return orderlineitemsbus.OrderLineItem{}, false, errs.Newf(errs.Internal, "update substitute line item picked qty: %s", err)
	}

	return updatedSub, true, nil
}

// allItemsPickingComplete returns true if all order line items are PICKED or CANCELLED.
// Returns false for empty item lists to prevent spurious order advancement.
func allItemsPickingComplete(items []orderlineitemsbus.OrderLineItem, pickedStatusID, cancelledStatusID uuid.UUID) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if item.LineItemFulfillmentStatusesID != pickedStatusID && item.LineItemFulfillmentStatusesID != cancelledStatusID {
			return false
		}
	}
	return true
}

// allItemsShortPickComplete returns true when all items are in a terminal picking state
// (PICKED, PARTIALLY_PICKED, BACKORDERED, or CANCELLED).
// Returns false for empty item lists to prevent spurious order advancement.
func allItemsShortPickComplete(items []orderlineitemsbus.OrderLineItem, pickedID, partiallyPickedID, backorderedID, cancelledID uuid.UUID) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		s := item.LineItemFulfillmentStatusesID
		if s != pickedID && s != partiallyPickedID && s != backorderedID && s != cancelledID {
			return false
		}
	}
	return true
}
