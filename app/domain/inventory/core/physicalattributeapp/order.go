package physicalattributeapp

import (
	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("product_id", order.ASC)

var orderByFields = map[string]string{
	"id":                   physicalattributebus.OrderByAttributeID,
	"product_id":           physicalattributebus.OrderByProductID,
	"length":               physicalattributebus.OrderByLength,
	"width":                physicalattributebus.OrderByWidth,
	"height":               physicalattributebus.OrderByHeight,
	"weight":               physicalattributebus.OrderByWeight,
	"weight_unit":          physicalattributebus.OrderByWeightUnit,
	"color":                physicalattributebus.OrderByColor,
	"size":                 physicalattributebus.OrderBySize,
	"material":             physicalattributebus.OrderByMaterial,
	"storage_requirements": physicalattributebus.OrderByStorageRequirements,
	"hazmat_class":         physicalattributebus.OrderByHazmatClass,
	"shelf_life_days":      physicalattributebus.OrderByShelfLifeDays,
	"created_date":         physicalattributebus.OrderByCreatedDate,
	"updated_date":         physicalattributebus.OrderByUpdatedDate,
}
