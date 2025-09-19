package zoneapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

func parseFilter(qp QueryParams) (zonebus.QueryFilter, error) {
	var filter zonebus.QueryFilter

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.Description != "" {
		filter.Description = &qp.Description
	}

	if qp.WarehouseID != "" {
		id, err := uuid.Parse(qp.WarehouseID)
		if err != nil {
			return zonebus.QueryFilter{}, err
		}
		filter.WarehouseID = &id
	}

	if qp.ZoneID != "" {
		id, err := uuid.Parse(qp.ZoneID)
		if err != nil {
			return zonebus.QueryFilter{}, err
		}
		filter.ZoneID = &id
	}

	if qp.UpdatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return zonebus.QueryFilter{}, err
		}
		filter.UpdatedDate = &date
	}

	if qp.CreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return zonebus.QueryFilter{}, err
		}
		filter.CreatedDate = &date
	}

	return filter, nil
}
