package supplierproductapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/timeutil"

	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus/types"
)

func parseFilter(qp QueryParams) (supplierproductbus.QueryFilter, error) {
	var filter supplierproductbus.QueryFilter

	if qp.SupplierProductID != "" {
		id, err := uuid.Parse(qp.SupplierProductID)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("supplier_product_id", err)
		}
		filter.SupplierProductID = &id
	}

	if qp.SupplierID != "" {
		id, err := uuid.Parse(qp.SupplierID)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("supplier_id", err)
		}
		filter.SupplierID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &id
	}
	if qp.SupplierPartNumber != "" {
		filter.SupplierPartNumber = &qp.SupplierPartNumber
	}

	if qp.MinOrderQuantity != "" {
		min, err := strconv.Atoi(qp.MinOrderQuantity)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("min_order_quantity", err)
		}
		filter.MinOrderQuantity = &min
	}

	if qp.MaxOrderQuantity != "" {
		max, err := strconv.Atoi(qp.MaxOrderQuantity)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("max_order_quantity", err)
		}
		filter.MaxOrderQuantity = &max
	}

	if qp.LeadTimeDays != "" {
		leadTime, err := strconv.Atoi(qp.LeadTimeDays)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("lead_time_days", err)
		}
		filter.LeadTimeDays = &leadTime
	}

	if qp.UnitCost != "" {
		cost, err := types.ParseMoney(qp.UnitCost)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("unit_cost", err)
		}

		filter.UnitCost = &cost
	}

	if qp.IsPrimarySupplier != "" {
		isPrimarySupplier, err := strconv.ParseBool(qp.IsPrimarySupplier)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("is_primary_supplier", err)
		}
		filter.IsPrimarySupplier = &isPrimarySupplier
	}

	if qp.CreatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.CreatedDate)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("created_date", err)
		}
		filter.CreatedDate = &date
	}

	if qp.UpdatedDate != "" {
		date, err := time.Parse(timeutil.FORMAT, qp.UpdatedDate)
		if err != nil {
			return supplierproductbus.QueryFilter{}, errs.NewFieldsError("updated_date", err)
		}
		filter.UpdatedDate = &date

	}

	return filter, nil
}
