package labelapp_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/app/domain/labels/labelapp"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
)

func Test_RenderPrintRequest_Validate_PayloadTooLarge(t *testing.T) {
	// Build a JSON payload > 64KB.
	big := strings.Repeat("a", 70*1024)
	raw, err := json.Marshal(map[string]string{"junk": big})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	req := labelapp.RenderPrintRequest{
		Type:    labelbus.TypeReceiving,
		Payload: raw,
	}
	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error for oversize payload, got nil")
	}
}

func Test_RenderPrintRequest_Validate_PayloadAtCap(t *testing.T) {
	// Payload just under the cap should validate.
	big := strings.Repeat("a", 60*1024)
	raw, err := json.Marshal(map[string]string{"junk": big})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	req := labelapp.RenderPrintRequest{
		Type:    labelbus.TypeReceiving,
		Payload: raw,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected payload under cap to validate, got: %v", err)
	}
}
