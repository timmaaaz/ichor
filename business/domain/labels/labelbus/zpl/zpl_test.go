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

func Test_Pick_Snapshot(t *testing.T) {
	got := zpl.Pick(zpl.PickData{
		OrderNumber:   "SO-7777",
		CustomerName:  "Acme Industries LLC",
		ProductName:   "Gadget B Mark II",
		SKU:           "SKU-042",
		UPC:           "098765432105",
		LotNumber:     strPtr("LOT-99"),
		SerialNumbers: []string{"SN-001", "SN-002", "SN-003"},
		Quantity:      3,
		LocationCode:  "STG-A02",
	})
	want := "^XA\n" +
		"^FO30,20^A0N,28,28^FDOrder: SO-7777^FS\n" +
		"^FO30,52^A0N,22,22^FDAcme Industries LLC^FS\n" +
		"^FO30,82^A0N,28,28^FDGadget B Mark II^FS\n" +
		"^FO30,117^A0N,22,22^FDSKU: SKU-042^FS\n" +
		"^FO30,149^BCN,50,Y,N,N^FD098765432105^FS\n" +
		"^FO30,229^A0N,22,22^FDLOT: LOT-99  QTY: 3^FS\n" +
		"^FO30,259^A0N,20,20^FDS/N: SN-001, SN-002, SN-003^FS\n" +
		"^FO30,287^A0N,22,22^FDFrom: STG-A02^FS\n" +
		"^XZ"
	if got != want {
		t.Fatalf("pick snapshot drift.\nwant:\n%q\ngot:\n%q\n", want, got)
	}
}
