package orderlineitemsapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
)

func parseFilter(qp QueryParams) (orderlineitemsbus.QueryFilter, error) {
	var filter orderlineitemsbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.OrderID != "" {
		orderID, err := uuid.Parse(qp.OrderID)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("order_id", err)
		}
		filter.OrderID = &orderID
	}

	if qp.ProductID != "" {
		productID, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("product_id", err)
		}
		filter.ProductID = &productID
	}

	if qp.Quantity != "" {
		quantity, err := strconv.Atoi(qp.Quantity)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("quantity", err)
		}
		filter.Quantity = &quantity
	}

	if qp.Description != "" {
		filter.Description = &qp.Description
	}

	if qp.UnitPrice != "" {
		filter.UnitPrice = &qp.UnitPrice
	}

	if qp.Discount != "" {
		filter.Discount = &qp.Discount
	}

	if qp.DiscountType != "" {
		filter.DiscountType = &qp.DiscountType
	}

	if qp.LineTotal != "" {
		filter.LineTotal = &qp.LineTotal
	}

	if qp.LineItemFulfillmentStatusesID != "" {
		lineItemFulfillmentStatusesID, err := uuid.Parse(qp.LineItemFulfillmentStatusesID)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("line_item_fulfillment_statuses_id", err)
		}
		filter.LineItemFulfillmentStatusesID = &lineItemFulfillmentStatusesID
	}

	if qp.CreatedBy != "" {
		createdBy, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("created_by", err)
		}
		filter.CreatedBy = &createdBy
	}

	if qp.UpdatedBy != "" {
		updatedBy, err := uuid.Parse(qp.UpdatedBy)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("updated_by", err)
		}
		filter.UpdatedBy = &updatedBy
	}

	if qp.StartCreatedDate != "" {
		startCreatedDate, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("start_created_date", err)
		}
		filter.StartCreatedDate = &startCreatedDate
	}

	if qp.EndCreatedDate != "" {
		endCreatedDate, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &endCreatedDate
	}

	if qp.StartUpdatedDate != "" {
		startUpdatedDate, err := time.Parse(time.RFC3339, qp.StartUpdatedDate)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("start_updated_date", err)
		}
		filter.StartUpdatedDate = &startUpdatedDate
	}

	if qp.EndUpdatedDate != "" {
		endUpdatedDate, err := time.Parse(time.RFC3339, qp.EndUpdatedDate)
		if err != nil {
			return orderlineitemsbus.QueryFilter{}, errs.NewFieldsError("end_updated_date", err)
		}
		filter.EndUpdatedDate = &endUpdatedDate
	}

	return filter, nil
}
