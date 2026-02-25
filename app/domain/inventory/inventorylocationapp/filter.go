package inventorylocationapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus/types"
)

func parseFilter(qp QueryParams) (inventorylocationbus.QueryFilter, error) {
	var filter inventorylocationbus.QueryFilter

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return inventorylocationbus.QueryFilter{}, errs.NewFieldsError("location_id", err)
		}
		filter.LocationID = &id
	}

	if qp.ZoneID != "" {
		id, err := uuid.Parse(qp.ZoneID)
		if err != nil {
			return inventorylocationbus.QueryFilter{}, errs.NewFieldsError("zone_id", err)
		}
		filter.ZoneID = &id
	}

	if qp.WarehouseID != "" {
		id, err := uuid.Parse(qp.WarehouseID)
		if err != nil {
			return inventorylocationbus.QueryFilter{}, errs.NewFieldsError("warehouse_id", err)
		}
		filter.WarehouseID = &id
	}

	if qp.Aisle != "" {
		filter.Aisle = &qp.Aisle
	}

	if qp.Shelf != "" {
		filter.Shelf = &qp.Shelf
	}

	if qp.Bin != "" {
		filter.Bin = &qp.Bin
	}

	if qp.Rack != "" {
		filter.Rack = &qp.Rack
	}

	if qp.LocationCode != "" {
		filter.LocationCode = &qp.LocationCode
	}

	if qp.IsPickLocation != "" {
		b, err := strconv.ParseBool(qp.IsPickLocation)
		if err != nil {
			return inventorylocationbus.QueryFilter{}, errs.NewFieldsError("is_pick_location", err)
		}
		filter.IsPickLocation = &b
	}

	if qp.IsReserveLocation != "" {
		b, err := strconv.ParseBool(qp.IsReserveLocation)
		if err != nil {
			return inventorylocationbus.QueryFilter{}, errs.NewFieldsError("is_reserve_location", err)
		}
		filter.IsReserveLocation = &b
	}

	if qp.MaxCapacity != "" {
		i, err := strconv.Atoi(qp.MaxCapacity)
		if err != nil {
			return inventorylocationbus.QueryFilter{}, errs.NewFieldsError("max_capacity", err)
		}
		filter.MaxCapacity = &i
	}

	if qp.CurrentUtilization != "" {
		i, err := types.ParseRoundedFloat(qp.CurrentUtilization)
		if err != nil {
			return inventorylocationbus.QueryFilter{}, errs.NewFieldsError("current_utilization", err)
		}
		filter.CurrentUtilization = &i
	}

	if qp.CreatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return inventorylocationbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &t
	}

	if qp.UpdatedDate != "" {
		t, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return inventorylocationbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &t
	}

	return filter, nil
}
