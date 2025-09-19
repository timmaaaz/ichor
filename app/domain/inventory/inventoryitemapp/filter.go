package inventoryitemapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (inventoryitembus.QueryFilter, error) {
	var filter inventoryitembus.QueryFilter

	if qp.ItemID != "" {
		id, err := uuid.Parse(qp.ItemID)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("item_id", err)
		}
		filter.ItemID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &id
	}

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("location_id", err)
		}
		filter.LocationID = &id
	}

	if qp.Quantity != "" {
		quantity, err := strconv.Atoi(qp.Quantity)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("quantity", err)
		}
		filter.Quantity = &quantity
	}

	if qp.ReservedQuantity != "" {
		quantity, err := strconv.Atoi(qp.ReservedQuantity)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("reserved_quantity", err)
		}
		filter.ReservedQuantity = &quantity
	}

	if qp.AllocatedQuantity != "" {
		quantity, err := strconv.Atoi(qp.AllocatedQuantity)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("allocated_quantity", err)
		}
		filter.AllocatedQuantity = &quantity
	}

	if qp.MinimumStock != "" {
		quantity, err := strconv.Atoi(qp.MinimumStock)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("minimum_stock", err)
		}
		filter.MinimumStock = &quantity
	}

	if qp.MaximumStock != "" {
		quantity, err := strconv.Atoi(qp.MaximumStock)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("maximum_stock", err)
		}
		filter.MaximumStock = &quantity
	}

	if qp.AvgDailyUsage != "" {
		quantity, err := strconv.Atoi(qp.AvgDailyUsage)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("avg_daily_usage", err)
		}
		filter.AvgDailyUsage = &quantity
	}

	if qp.ReorderPoint != "" {
		quantity, err := strconv.Atoi(qp.ReorderPoint)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("reorder_point", err)
		}
		filter.ReorderPoint = &quantity
	}

	if qp.EconomicOrderQuantity != "" {
		quantity, err := strconv.Atoi(qp.EconomicOrderQuantity)

		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("economic_order_quantity", err)
		}
		filter.EconomicOrderQuantity = &quantity
	}

	if qp.SafetyStock != "" {
		quantity, err := strconv.Atoi(qp.SafetyStock)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("safety_stock", err)
		}
		filter.SafetyStock = &quantity
	}

	if qp.CreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &date
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return inventoryitembus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &t
	}

	return filter, nil
}
