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
		Type:    labelbus.TypeProduct,
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
		Type:    labelbus.TypeProduct,
		Payload: raw,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected payload under cap to validate, got: %v", err)
	}
}

func Test_NewLabel_Validate_CodeTooLong(t *testing.T) {
	req := labelapp.NewLabel{
		Code: "WAREHOUSE-RECEIVING-DOCK-12A", // 28 chars, schema-allowed but not renderable
		Type: labelbus.TypeLocation,
	}
	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error for 28-char code, got nil")
	}
}

func Test_NewLabel_Validate_CodeAtRenderableCap(t *testing.T) {
	req := labelapp.NewLabel{
		Code: "STG-A01-B12C", // 12 chars, the BY4/812-dot upper bound
		Type: labelbus.TypeLocation,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected 12-char code to validate, got: %v", err)
	}
}

func Test_RenderPrintRequest_Validate_CodeTooLong(t *testing.T) {
	req := labelapp.RenderPrintRequest{
		Type: labelbus.TypeLocation,
		Code: "WAREHOUSE-RECEIVING-DOCK-12A",
	}
	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error for 28-char code, got nil")
	}
}

func Test_NewLabel_Validate_CodeOneOverCap(t *testing.T) {
	req := labelapp.NewLabel{
		Code: "STG-A01-B12CD", // 13 chars, one over the renderable limit
		Type: labelbus.TypeLocation,
	}
	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error for 13-char code, got nil")
	}
}

func Test_UpdateLabel_Validate_CodeTooLong(t *testing.T) {
	code := "WAREHOUSE-RECEIVING-DOCK-12A" // 28 chars
	req := labelapp.UpdateLabel{Code: &code}
	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error for 28-char code, got nil")
	}
}
