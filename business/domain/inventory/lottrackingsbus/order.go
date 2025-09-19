package lottrackingsbus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByLotID, order.ASC)

const (
	OrderByLotID             = "id"
	OrderBySupplierProductID = "supplier_product_id"
	OrderByLotNumber         = "lot_number"
	OrderByManufactureDate   = "manufacture_date"
	OrderByExpirationDate    = "expiration_date"
	OrderByRecievedDate      = "received_date"
	OrderByQuantity          = "quantity"
	OrderByQualityStatus     = "quality_status"
	OrderByCreatedDate       = "created_date"
	OrderByUpdatedDate       = "updated_date"
)
