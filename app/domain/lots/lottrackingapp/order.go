package lottrackingapp

import (
	"github.com/timmaaaz/ichor/business/domain/lot/lottrackingbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("id", order.ASC)

var orderByFields = map[string]string{
	"id":                  lottrackingbus.OrderByLotID,
	"supplier_product_id": lottrackingbus.OrderBySupplierProductID,
	"lot_number":          lottrackingbus.OrderByLotNumber,
	"manufacture_date":    lottrackingbus.OrderByManufactureDate,
	"expiration_date":     lottrackingbus.OrderByExpirationDate,
	"received_date":       lottrackingbus.OrderByRecievedDate,
	"quantity":            lottrackingbus.OrderByQuantity,
	"quality_status":      lottrackingbus.OrderByQualityStatus,
	"created_date":        lottrackingbus.OrderByCreatedDate,
	"updated_date":        lottrackingbus.OrderByUpdatedDate,
}
