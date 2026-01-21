package ordersdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
)

// TODO: Switch these over to use string.builder?

func applyFilter(filter ordersbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Number != nil {
		data["number"] = *filter.Number
		wc = append(wc, "number ILIKE :number")
	}

	if filter.CustomerID != nil {
		data["customer_id"] = *filter.CustomerID
		wc = append(wc, "customer_id = :customer_id")
	}

	if filter.FulfillmentStatusID != nil {
		data["fulfillment_status_id"] = *filter.FulfillmentStatusID
		wc = append(wc, "order_fulfillment_status_id = :fulfillment_status_id")
	}

	if filter.BillingAddressID != nil {
		data["billing_address_id"] = *filter.BillingAddressID
		wc = append(wc, "billing_address_id = :billing_address_id")
	}

	if filter.ShippingAddressID != nil {
		data["shipping_address_id"] = *filter.ShippingAddressID
		wc = append(wc, "shipping_address_id = :shipping_address_id")
	}

	if filter.CurrencyID != nil {
		data["currency_id"] = *filter.CurrencyID
		wc = append(wc, "currency_id = :currency_id")
	}

	if filter.PaymentTermID != nil {
		data["payment_term_id"] = *filter.PaymentTermID
		wc = append(wc, "payment_term_id = :payment_term_id")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
	}

	if filter.UpdatedBy != nil {
		data["updated_by"] = *filter.UpdatedBy
		wc = append(wc, "updated_by = :updated_by")
	}

	if filter.StartDueDate != nil {
		data["start_due_date"] = *filter.StartDueDate
		wc = append(wc, "due_date >= :start_due_date")
	}

	if filter.EndDueDate != nil {
		data["end_due_date"] = *filter.EndDueDate
		wc = append(wc, "due_date <= :end_due_date")
	}

	if filter.StartOrderDate != nil {
		data["start_order_date"] = *filter.StartOrderDate
		wc = append(wc, "order_date >= :start_order_date")
	}

	if filter.EndOrderDate != nil {
		data["end_order_date"] = *filter.EndOrderDate
		wc = append(wc, "order_date <= :end_order_date")
	}

	if filter.StartCreatedDate != nil {
		data["start_created_date"] = *filter.StartCreatedDate
		wc = append(wc, "created_date >= :start_created_date")
	}

	if filter.EndCreatedDate != nil {
		data["end_created_date"] = *filter.EndCreatedDate
		wc = append(wc, "created_date <= :end_created_date")
	}

	if filter.StartUpdatedDate != nil {
		data["start_updated_date"] = *filter.StartUpdatedDate
		wc = append(wc, "updated_date >= :start_updated_date")
	}

	if filter.EndUpdatedDate != nil {
		data["end_updated_date"] = *filter.EndUpdatedDate
		wc = append(wc, "updated_date <= :end_updated_date")
	}

	// Monetary range filters
	if filter.MinSubtotal != nil {
		data["min_subtotal"] = *filter.MinSubtotal
		wc = append(wc, "subtotal::numeric >= :min_subtotal::numeric")
	}

	if filter.MaxSubtotal != nil {
		data["max_subtotal"] = *filter.MaxSubtotal
		wc = append(wc, "subtotal::numeric <= :max_subtotal::numeric")
	}

	if filter.MinTaxAmount != nil {
		data["min_tax_amount"] = *filter.MinTaxAmount
		wc = append(wc, "tax_amount::numeric >= :min_tax_amount::numeric")
	}

	if filter.MaxTaxAmount != nil {
		data["max_tax_amount"] = *filter.MaxTaxAmount
		wc = append(wc, "tax_amount::numeric <= :max_tax_amount::numeric")
	}

	if filter.MinShippingCost != nil {
		data["min_shipping_cost"] = *filter.MinShippingCost
		wc = append(wc, "shipping_cost::numeric >= :min_shipping_cost::numeric")
	}

	if filter.MaxShippingCost != nil {
		data["max_shipping_cost"] = *filter.MaxShippingCost
		wc = append(wc, "shipping_cost::numeric <= :max_shipping_cost::numeric")
	}

	if filter.MinTotalAmount != nil {
		data["min_total_amount"] = *filter.MinTotalAmount
		wc = append(wc, "total_amount::numeric >= :min_total_amount::numeric")
	}

	if filter.MaxTotalAmount != nil {
		data["max_total_amount"] = *filter.MaxTotalAmount
		wc = append(wc, "total_amount::numeric <= :max_total_amount::numeric")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
