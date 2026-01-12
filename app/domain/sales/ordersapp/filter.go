package ordersapp

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
)

func parseFilter(qp QueryParams) (ordersbus.QueryFilter, error) {
	var filter ordersbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.Number != "" {
		filter.Number = &qp.Number
	}

	if qp.CustomerID != "" {
		id, err := uuid.Parse(qp.CustomerID)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("customer_id", err)
		}
		filter.CustomerID = &id
	}

	if qp.FulfillmentStatusID != "" {
		id, err := uuid.Parse(qp.FulfillmentStatusID)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("fulfillment_status_id", err)
		}
		filter.FulfillmentStatusID = &id
	}

	if qp.BillingAddressID != "" {
		id, err := uuid.Parse(qp.BillingAddressID)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("billing_address_id", err)
		}
		filter.BillingAddressID = &id
	}

	if qp.ShippingAddressID != "" {
		id, err := uuid.Parse(qp.ShippingAddressID)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("shipping_address_id", err)
		}
		filter.ShippingAddressID = &id
	}

	if qp.Currency != "" {
		filter.Currency = &qp.Currency
	}

	if qp.PaymentTerms != "" {
		filter.PaymentTerms = &qp.PaymentTerms
	}

	if qp.CreatedBy != "" {
		id, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("created_by", err)
		}
		filter.CreatedBy = &id
	}

	if qp.UpdatedBy != "" {
		id, err := uuid.Parse(qp.UpdatedBy)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("updated_by", err)
		}
		filter.UpdatedBy = &id
	}

	if qp.StartDueDate != "" {
		startDueDate, err := time.Parse(time.RFC3339, qp.StartDueDate)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("start_due_date", err)
		}
		filter.StartDueDate = &startDueDate
	}

	if qp.EndDueDate != "" {
		endDueDate, err := time.Parse(time.RFC3339, qp.EndDueDate)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("end_due_date", err)
		}
		filter.EndDueDate = &endDueDate
	}

	if qp.StartOrderDate != "" {
		startOrderDate, err := time.Parse(time.RFC3339, qp.StartOrderDate)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("start_order_date", err)
		}
		filter.StartOrderDate = &startOrderDate
	}

	if qp.EndOrderDate != "" {
		endOrderDate, err := time.Parse(time.RFC3339, qp.EndOrderDate)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("end_order_date", err)
		}
		filter.EndOrderDate = &endOrderDate
	}

	if qp.StartCreatedDate != "" {
		startCreatedDate, err := time.Parse(time.RFC3339, qp.StartCreatedDate)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("start_created_date", err)
		}
		filter.StartCreatedDate = &startCreatedDate
	}

	if qp.EndCreatedDate != "" {
		endCreatedDate, err := time.Parse(time.RFC3339, qp.EndCreatedDate)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("end_created_date", err)
		}
		filter.EndCreatedDate = &endCreatedDate
	}

	if qp.StartUpdatedDate != "" {
		startUpdatedDate, err := time.Parse(time.RFC3339, qp.StartUpdatedDate)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("start_updated_date", err)
		}
		filter.StartUpdatedDate = &startUpdatedDate
	}

	if qp.EndUpdatedDate != "" {
		endUpdatedDate, err := time.Parse(time.RFC3339, qp.EndUpdatedDate)
		if err != nil {
			return ordersbus.QueryFilter{}, errs.NewFieldsError("end_updated_date", err)
		}
		filter.EndUpdatedDate = &endUpdatedDate
	}

	// Monetary range filters - passed as strings for numeric comparison
	if qp.MinSubtotal != "" {
		filter.MinSubtotal = &qp.MinSubtotal
	}
	if qp.MaxSubtotal != "" {
		filter.MaxSubtotal = &qp.MaxSubtotal
	}
	if qp.MinTaxAmount != "" {
		filter.MinTaxAmount = &qp.MinTaxAmount
	}
	if qp.MaxTaxAmount != "" {
		filter.MaxTaxAmount = &qp.MaxTaxAmount
	}
	if qp.MinShippingCost != "" {
		filter.MinShippingCost = &qp.MinShippingCost
	}
	if qp.MaxShippingCost != "" {
		filter.MaxShippingCost = &qp.MaxShippingCost
	}
	if qp.MinTotalAmount != "" {
		filter.MinTotalAmount = &qp.MinTotalAmount
	}
	if qp.MaxTotalAmount != "" {
		filter.MaxTotalAmount = &qp.MaxTotalAmount
	}

	return filter, nil
}
