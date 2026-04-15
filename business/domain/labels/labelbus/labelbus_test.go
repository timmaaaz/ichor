package labelbus_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func Test_NewBusiness(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	bus := labelbus.NewBusiness(log, delegate.New(log), nil, nil)
	if bus == nil {
		t.Fatalf("NewBusiness returned nil")
	}
}

func Test_Render_DispatchesByType(t *testing.T) {
	tests := []struct {
		name string
		lc   labelbus.LabelCatalog
		want string
	}{
		{"tote", labelbus.LabelCatalog{Code: "TOTE-001", Type: labelbus.TypeTote}, "TOTE-001"},
		{"location", labelbus.LabelCatalog{Code: "STG-A01", Type: labelbus.TypeLocation}, "STG-A01"},
		{"receiving", labelbus.LabelCatalog{
			Code: "RX-1", Type: labelbus.TypeReceiving,
			PayloadJSON: `{"productName":"Widget","sku":"SKU-1","upc":"012345678905","lotNumber":null,"expiryDate":null,"quantity":10,"poNumber":"PO-1"}`,
		}, "PO: PO-1"},
		{"pick", labelbus.LabelCatalog{
			Code: "PK-1", Type: labelbus.TypePick,
			PayloadJSON: `{"orderNumber":"SO-1","customerName":"Acme","productName":"Gadget","sku":"SKU-2","upc":"098765432105","lotNumber":null,"serialNumbers":[],"quantity":5,"locationCode":"STG-A02"}`,
		}, "Order: SO-1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := labelbus.Render(tc.lc)
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if !strings.Contains(string(out), tc.want) {
				t.Fatalf("expected %q in output, got: %s", tc.want, out)
			}
		})
	}
}

func Test_Render_UnknownType(t *testing.T) {
	_, err := labelbus.Render(labelbus.LabelCatalog{Code: "X", Type: "nonsense"})
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
}

func Test_Render_BadPayload(t *testing.T) {
	_, err := labelbus.Render(labelbus.LabelCatalog{
		Code: "RX", Type: labelbus.TypeReceiving, PayloadJSON: `{not json`,
	})
	if err == nil {
		t.Fatal("expected unmarshal error, got nil")
	}
}
