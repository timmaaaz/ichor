package dbtest

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
)

// SalesSeed holds the results of seeding sales data.
type SalesSeed struct {
	OrderIDs         uuid.UUIDs
	OrderLineItemIDs uuid.UUIDs
}

func seedSales(ctx context.Context, busDomain BusDomain, foundation FoundationSeed, geoHR GeographyHRSeed, products ProductsSeed) (SalesSeed, error) {
	count := 5

	strIDs := make([]uuid.UUID, 0, len(geoHR.Streets))
	for _, s := range geoHR.Streets {
		strIDs = append(strIDs, s.ID)
	}

	contactInfoIDs := make([]uuid.UUID, 0, len(geoHR.ContactInfos))
	for _, ci := range geoHR.ContactInfos {
		contactInfoIDs = append(contactInfoIDs, ci.ID)
	}

	adminIDs := uuid.UUIDs{foundation.Admins[0].ID}

	customers, err := customersbus.TestSeedCustomersHistorical(ctx, count, 180, strIDs, contactInfoIDs, adminIDs, busDomain.Customers)
	if err != nil {
		return SalesSeed{}, fmt.Errorf("seeding customers : %w", err)
	}
	customerIDs := make([]uuid.UUID, 0, len(customers))
	for _, c := range customers {
		customerIDs = append(customerIDs, c.ID)
	}

	// Create order fulfillment statuses matching seed.sql
	ofls := make([]orderfulfillmentstatusbus.OrderFulfillmentStatus, 0, len(seedmodels.OrderFulfillmentStatusData))
	for _, data := range seedmodels.OrderFulfillmentStatusData {
		ofl, err := busDomain.OrderFulfillmentStatus.Create(ctx, orderfulfillmentstatusbus.NewOrderFulfillmentStatus{
			Name:        data.Name,
			Description: data.Description,
		})
		if err != nil {
			return SalesSeed{}, fmt.Errorf("seeding order fulfillment status %s: %w", data.Name, err)
		}
		ofls = append(ofls, ofl)
	}
	oflIDs := make([]uuid.UUID, 0, len(ofls))
	for _, ofl := range ofls {
		oflIDs = append(oflIDs, ofl.ID)
	}

	currencyIDs := make(uuid.UUIDs, len(foundation.Currencies))
	for i, c := range foundation.Currencies {
		currencyIDs[i] = c.ID
	}

	userIDs := make([]uuid.UUID, 0, len(foundation.Admins))
	for _, a := range foundation.Admins {
		userIDs = append(userIDs, a.ID)
	}

	// Use weighted random distribution for frontend demo (better heatmap visualization)
	orders, err := ordersbus.TestSeedOrdersFrontendWeighted(ctx, 200, 90, adminIDs, customerIDs, oflIDs, currencyIDs, busDomain.Order)
	if err != nil {
		return SalesSeed{}, fmt.Errorf("seeding Orders: %w", err)
	}
	orderIDs := make([]uuid.UUID, 0, len(orders))
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}

	// Create line item fulfillment statuses matching seed.sql
	lineItemFulfillmentStatusData := []struct {
		name        string
		description string
	}{
		{"ALLOCATED", "Line item has been allocated"},
		{"CANCELLED", "Line item has been cancelled"},
		{"PACKED", "Line item has been packed"},
		{"PENDING", "Line item is pending"},
		{"PICKED", "Line item has been picked"},
		{"PARTIALLY_PICKED", "Line item has been partially picked"},
		{"BACKORDERED", "Line item quantity is on backorder"},
		{"PENDING_REVIEW", "Line item requires supervisor review before proceeding"},
		{"SHIPPED", "Line item has been shipped"},
	}
	olStatuses := make([]lineitemfulfillmentstatusbus.LineItemFulfillmentStatus, 0, len(lineItemFulfillmentStatusData))
	for _, data := range lineItemFulfillmentStatusData {
		ols, err := busDomain.LineItemFulfillmentStatus.Create(ctx, lineitemfulfillmentstatusbus.NewLineItemFulfillmentStatus{
			Name:        data.name,
			Description: data.description,
		})
		if err != nil {
			return SalesSeed{}, fmt.Errorf("seeding line item fulfillment status %s: %w", data.name, err)
		}
		olStatuses = append(olStatuses, ols)
	}
	olStatusIDs := make([]uuid.UUID, 0, len(olStatuses))
	for _, ols := range olStatuses {
		olStatusIDs = append(olStatusIDs, ols.ID)
	}

	// Create map of order IDs to their created dates for historical line items
	orderDates := make(map[uuid.UUID]time.Time)
	for _, order := range orders {
		orderDates[order.ID] = order.CreatedDate
	}

	productIDs := make([]uuid.UUID, 0, len(products.Products))
	for _, p := range products.Products {
		productIDs = append(productIDs, p.ProductID)
	}

	lineItems, err := orderlineitemsbus.TestSeedOrderLineItemsHistorical(ctx, count, orderDates, orderIDs, productIDs, olStatusIDs, userIDs, busDomain.OrderLineItem)
	if err != nil {
		return SalesSeed{}, fmt.Errorf("seeding Order Line Items: %w", err)
	}

	lineItemIDs := make(uuid.UUIDs, len(lineItems))
	for i, li := range lineItems {
		lineItemIDs[i] = li.ID
	}

	return SalesSeed{
		OrderIDs:         uuid.UUIDs(orderIDs),
		OrderLineItemIDs: lineItemIDs,
	}, nil
}
