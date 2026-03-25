package procurement_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/procurement"
)

func TestCreatePurchaseOrder_Validate(t *testing.T) {
	handler := procurement.NewCreatePurchaseOrderHandler(nil, nil, nil, nil, nil)

	validStatusID := uuid.New().String()
	validWarehouseID := uuid.New().String()
	validLocationID := uuid.New().String()
	validCurrencyID := uuid.New().String()
	validProductID := uuid.New().String()
	validLineItemStatusID := uuid.New().String()

	validConfig := procurement.CreatePurchaseOrderConfig{
		PurchaseOrderStatusID: validStatusID,
		DeliveryWarehouseID:   validWarehouseID,
		DeliveryLocationID:    validLocationID,
		CurrencyID:            validCurrencyID,
		LineItems: []procurement.CreatePOLineItemConfig{
			{
				ProductID:        validProductID,
				QuantityOrdered:  10,
				LineItemStatusID: validLineItemStatusID,
			},
		},
	}

	// copyLineItems returns a deep copy of the line items slice so mutations
	// in one test case don't bleed into subsequent cases.
	copyLineItems := func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
		if len(c.LineItems) > 0 {
			cp := make([]procurement.CreatePOLineItemConfig, len(c.LineItems))
			copy(cp, c.LineItems)
			c.LineItems = cp
		}
		return c
	}

	tests := []struct {
		name      string
		modify    func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid config",
			modify:  func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig { return c },
			wantErr: false,
		},
		{
			name: "missing purchase_order_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.PurchaseOrderStatusID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "purchase_order_status_id is required",
		},
		{
			name: "invalid purchase_order_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.PurchaseOrderStatusID = "not-a-uuid"
				return c
			},
			wantErr:   true,
			errSubstr: "invalid purchase_order_status_id",
		},
		{
			name: "missing delivery_warehouse_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.DeliveryWarehouseID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "delivery_warehouse_id is required",
		},
		{
			name: "missing delivery_location_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.DeliveryLocationID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "delivery_location_id is required",
		},
		{
			name: "missing currency_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.CurrencyID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "currency_id is required",
		},
		{
			name: "no line items when source_from_event is false",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.LineItems = nil
				return c
			},
			wantErr:   true,
			errSubstr: "at least one line item is required",
		},
		{
			name: "source_from_event without default_line_item_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.SourceFromEvent = true
				c.LineItems = nil
				return c
			},
			wantErr:   true,
			errSubstr: "default_line_item_status_id is required",
		},
		{
			name: "source_from_event with valid default_line_item_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.SourceFromEvent = true
				c.DefaultLineItemStatusID = validLineItemStatusID
				c.LineItems = nil
				return c
			},
			wantErr: false,
		},
		{
			name: "line item missing product_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c = copyLineItems(c)
				c.LineItems[0].ProductID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "product_id is required",
		},
		{
			name: "line item zero quantity",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c = copyLineItems(c)
				c.LineItems[0].QuantityOrdered = 0
				return c
			},
			wantErr:   true,
			errSubstr: "quantity_ordered must be greater than 0",
		},
		{
			name: "line item negative quantity",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c = copyLineItems(c)
				c.LineItems[0].QuantityOrdered = -5
				return c
			},
			wantErr:   true,
			errSubstr: "quantity_ordered must be greater than 0",
		},
		{
			name: "line item missing line_item_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c = copyLineItems(c)
				c.LineItems[0].LineItemStatusID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "line_item_status_id is required",
		},
		{
			name:      "invalid json",
			modify:    nil, // special case
			wantErr:   true,
			errSubstr: "invalid configuration format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configBytes json.RawMessage
			if tt.modify == nil {
				configBytes = json.RawMessage(`{invalid`)
			} else {
				cfg := tt.modify(validConfig)
				data, _ := json.Marshal(cfg)
				configBytes = data
			}

			err := handler.Validate(configBytes)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errSubstr)
				}
			}
		})
	}
}
