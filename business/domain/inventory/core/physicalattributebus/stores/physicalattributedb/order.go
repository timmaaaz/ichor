package physicalattributedb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	physicalattributebus.OrderByAttributeID:         "id",
	physicalattributebus.OrderByProductID:           "product_id",
	physicalattributebus.OrderByLength:              "length",
	physicalattributebus.OrderByWidth:               "width",
	physicalattributebus.OrderByHeight:              "height",
	physicalattributebus.OrderByWeight:              "weight",
	physicalattributebus.OrderByWeightUnit:          "weight_unit",
	physicalattributebus.OrderByColor:               "color",
	physicalattributebus.OrderBySize:                "size",
	physicalattributebus.OrderByMaterial:            "material",
	physicalattributebus.OrderByStorageRequirements: "storage_requirements",
	physicalattributebus.OrderByHazmatClass:         "hazmat_class",
	physicalattributebus.OrderByShelfLifeDays:       "shelf_life_days",
	physicalattributebus.OrderByCreatedDate:         "created_date",
	physicalattributebus.OrderByUpdatedDate:         "updated_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
