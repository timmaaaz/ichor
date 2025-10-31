package purchaseorderlineitemapp

import (
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(purchaseorderlineitembus.OrderByID, order.ASC)

var orderByFields = map[string]string{
	"id":                   purchaseorderlineitembus.OrderByID,
	"purchaseOrderId":      purchaseorderlineitembus.OrderByPurchaseOrderID,
	"supplierProductId":    purchaseorderlineitembus.OrderBySupplierProductID,
	"quantityOrdered":      purchaseorderlineitembus.OrderByQuantityOrdered,
	"quantityReceived":     purchaseorderlineitembus.OrderByQuantityReceived,
	"lineTotal":            purchaseorderlineitembus.OrderByLineTotal,
	"expectedDeliveryDate": purchaseorderlineitembus.OrderByExpectedDeliveryDate,
	"actualDeliveryDate":   purchaseorderlineitembus.OrderByActualDeliveryDate,
	"createdDate":          purchaseorderlineitembus.OrderByCreatedDate,
	"updatedDate":          purchaseorderlineitembus.OrderByUpdatedDate,
}
