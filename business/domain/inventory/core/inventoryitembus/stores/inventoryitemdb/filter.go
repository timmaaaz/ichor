package inventoryitemdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/inventoryitembus"
)

func applyFilter(filter inventoryitembus.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ItemID != nil {
		data["item_id"] = *filter.ItemID
		wc = append(wc, "item_id = :item_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.LocationID != nil {
		data["location_id"] = *filter.LocationID
		wc = append(wc, "location_id = :location_id")
	}
	if filter.ReservedQuantity != nil {
		data["reserved_quantity"] = *filter.ReservedQuantity
		wc = append(wc, "reserved_quantity = :reserved_quantity")
	}

	if filter.AllocatedQuantity != nil {
		data["allocated_quantity"] = *filter.AllocatedQuantity
		wc = append(wc, "allocated_quantity = :allocated_quantity")
	}

	if filter.MinimumStock != nil {
		data["minimum_stock"] = *filter.MinimumStock
		wc = append(wc, "minimum_stock = :minimum_stock")
	}

	if filter.MaximumStock != nil {
		data["maximum_stock"] = *filter.MaximumStock
		wc = append(wc, "maximum_stock = :maximum_stock")
	}

	if filter.ReorderPoint != nil {
		data["reorder_point"] = *filter.ReorderPoint
		wc = append(wc, "reorder_point = :reorder_point")
	}

	if filter.EconomicOrderQuantity != nil {
		data["economic_order_quantity"] = *filter.EconomicOrderQuantity
		wc = append(wc, "economic_order_quantity = :economic_order_quantity")
	}

	if filter.SafetyStock != nil {
		data["safety_stock"] = *filter.SafetyStock
		wc = append(wc, "safety_stock = :safety_stock")
	}

	if filter.AvgDailyUsage != nil {
		data["avg_daily_usage"] = *filter.AvgDailyUsage
		wc = append(wc, "avg_daily_usage = :avg_daily_usage")
	}

	if filter.Quantity != nil {
		data["quantity"] = *filter.Quantity
		wc = append(wc, "quantity = :quantity")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "updated_date = :updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}

}
