package zpl_test

import (
	"testing"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/zpl"
)

func strPtr(s string) *string { return &s }

func Test_Receiving_Snapshot(t *testing.T) {
	got := zpl.Receiving(zpl.ReceivingData{
		ProductName: "Widget Assembly Type A",
		SKU:         "SKU-001",
		UPC:         "012345678905",
		LotNumber:   strPtr("LOT-42"),
		ExpiryDate:  strPtr("2027-01-15"),
		Quantity:    100,
		PONumber:    "PO-12345",
	})
	want := "^XA\n" +
		"^FO30,30^A0N,36,36^FDWidget Assembly Type A^FS\n" +
		"^FO30,75^A0N,24,24^FDSKU: SKU-001^FS\n" +
		"^FO30,115^BCN,60,Y,N,N^FD012345678905^FS\n" +
		"^FO30,210^A0N,24,24^FDLOT: LOT-42  EXP: 2027-01-15^FS\n" +
		"^FO30,245^A0N,24,24^FDQTY: 100  PO: PO-12345^FS\n" +
		"^XZ"
	if got != want {
		t.Fatalf("receiving snapshot drift.\nwant:\n%q\ngot:\n%q\n", want, got)
	}
}
