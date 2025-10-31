package purchaseorderapp

import (
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(purchaseorderbus.OrderByOrderDate, order.DESC)

var orderByFields = map[string]string{
	"id":                     purchaseorderbus.OrderByID,
	"orderNumber":            purchaseorderbus.OrderByOrderNumber,
	"supplierId":             purchaseorderbus.OrderBySupplierID,
	"orderDate":              purchaseorderbus.OrderByOrderDate,
	"expectedDeliveryDate":   purchaseorderbus.OrderByExpectedDeliveryDate,
	"actualDeliveryDate":     purchaseorderbus.OrderByActualDeliveryDate,
	"totalAmount":            purchaseorderbus.OrderByTotalAmount,
	"requestedBy":            purchaseorderbus.OrderByRequestedBy,
	"approvedBy":             purchaseorderbus.OrderByApprovedBy,
	"createdDate":            purchaseorderbus.OrderByCreatedDate,
	"updatedDate":            purchaseorderbus.OrderByUpdatedDate,
}
