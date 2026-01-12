package ordersapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
)

func parseQueryParams(r *http.Request) (ordersapp.QueryParams, error) {
	values := r.URL.Query()

	filter := ordersapp.QueryParams{
		Page:                values.Get("page"),
		Rows:                values.Get("rows"),
		OrderBy:             values.Get("orderBy"),
		ID:                  values.Get("id"),
		Number:              values.Get("number"),
		CustomerID:          values.Get("customer_id"),
		FulfillmentStatusID: values.Get("fulfillment_status_id"),
		BillingAddressID:    values.Get("billing_address_id"),
		ShippingAddressID:   values.Get("shipping_address_id"),
		Currency:            values.Get("currency"),
		PaymentTerms:        values.Get("payment_terms"),
		CreatedBy:           values.Get("created_by"),
		UpdatedBy:           values.Get("updated_by"),
		StartDueDate:        values.Get("start_due_date"),
		EndDueDate:          values.Get("end_due_date"),
		StartOrderDate:      values.Get("start_order_date"),
		EndOrderDate:        values.Get("end_order_date"),
		StartCreatedDate:    values.Get("start_created_date"),
		EndCreatedDate:      values.Get("end_created_date"),
		StartUpdatedDate:    values.Get("start_updated_date"),
		EndUpdatedDate:      values.Get("end_updated_date"),
		// Monetary range filters
		MinSubtotal:     values.Get("min_subtotal"),
		MaxSubtotal:     values.Get("max_subtotal"),
		MinTaxAmount:    values.Get("min_tax_amount"),
		MaxTaxAmount:    values.Get("max_tax_amount"),
		MinShippingCost: values.Get("min_shipping_cost"),
		MaxShippingCost: values.Get("max_shipping_cost"),
		MinTotalAmount:  values.Get("min_total_amount"),
		MaxTotalAmount:  values.Get("max_total_amount"),
	}

	return filter, nil
}
