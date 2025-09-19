package lottrackingsapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":                  lottrackingsbus.OrderByLotID,
	"supplier_product_id": lottrackingsbus.OrderBySupplierProductID,
	"lot_number":          lottrackingsbus.OrderByLotNumber,
	"manufacture_date":    lottrackingsbus.OrderByManufactureDate,
	"expiration_date":     lottrackingsbus.OrderByExpirationDate,
	"received_date":       lottrackingsbus.OrderByRecievedDate,
	"quantity":            lottrackingsbus.OrderByQuantity,
	"quality_status":      lottrackingsbus.OrderByQualityStatus,
	"created_date":        lottrackingsbus.OrderByCreatedDate,
	"updated_date":        lottrackingsbus.OrderByUpdatedDate,
}
