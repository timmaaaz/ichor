package physicalattributebus

import "github.com/timmaaaz/ichor/business/sdk/order"

var DefaultOrderBy = order.NewBy(OrderByProductID, order.ASC)

const (
	OrderByAttributeID         = "id"
	OrderByProductID           = "product_id"
	OrderByLength              = "length"
	OrderByWidth               = "width"
	OrderByHeight              = "height"
	OrderByWeight              = "weight"
	OrderByWeightUnit          = "weight_unit"
	OrderByColor               = "color"
	OrderBySize                = "size"
	OrderByMaterial            = "material"
	OrderByStorageRequirements = "storage_requirements"
	OrderByHazmatClass         = "hazmat_class"
	OrderByShelfLifeDays       = "shelf_life_days"
	OrderByCreatedDate         = "created_date"
	OrderByUpdatedDate         = "updated_date"
)
