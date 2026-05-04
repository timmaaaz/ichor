package zpl_test

import (
	"testing"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/zpl"
)

func strPtr(s string) *string { return &s }

func Test_Location_Snapshot(t *testing.T) {
	got := zpl.Location(zpl.LocationData{Code: "STG-A02"})
	want := "^XA\n" +
		"^FO40,80^A0N,150,150^FDSTG-A02^FS\n" +
		"^FO40,300^BY4^BCN,250,Y,N,N^FDSTG-A02^FS\n" +
		"^XZ\n"
	if got != want {
		t.Fatalf("location snapshot drift.\nwant:\n%q\ngot:\n%q\n", want, got)
	}
}

func Test_Tote_Snapshot(t *testing.T) {
	got := zpl.Tote(zpl.ToteData{Code: "TOTE-007"})
	want := "^XA\n" +
		"^FO40,80^A0N,150,150^FDTOTE-007^FS\n" +
		"^FO40,300^BY4^BCN,250,Y,N,N^FDTOTE-007^FS\n" +
		"^XZ\n"
	if got != want {
		t.Fatalf("tote snapshot drift.\nwant:\n%q\ngot:\n%q\n", want, got)
	}
}

func Test_Product_Snapshot(t *testing.T) {
	got := zpl.Product(zpl.ProductData{
		ProductName: "Widget Assembly Type A",
		SKU:         "SKU-001",
		UPC:         "012345678905",
		LotNumber:   strPtr("LOT-42"),
	})
	want := "^XA\n" +
		"^FO40,60^A0N,60,60^FDWidget Assembly Type A^FS\n" +
		"^FO40,150^A0N,40,40^FDSKU: SKU-001^FS\n" +
		"^FO40,230^BY3^BCN,120,Y,N,N^FD012345678905^FS\n" +
		"^FO40,430^A0N,40,40^FDLOT: LOT-42^FS\n" +
		"^XZ\n"
	if got != want {
		t.Fatalf("product snapshot drift.\nwant:\n%q\ngot:\n%q\n", want, got)
	}
}

func Test_Product_Snapshot_NilLot(t *testing.T) {
	got := zpl.Product(zpl.ProductData{
		ProductName: "Widget Assembly Type A",
		SKU:         "SKU-001",
		UPC:         "012345678905",
		LotNumber:   nil,
	})
	want := "^XA\n" +
		"^FO40,60^A0N,60,60^FDWidget Assembly Type A^FS\n" +
		"^FO40,150^A0N,40,40^FDSKU: SKU-001^FS\n" +
		"^FO40,230^BY3^BCN,120,Y,N,N^FD012345678905^FS\n" +
		"^XZ\n"
	if got != want {
		t.Fatalf("product snapshot drift (nil lot).\nwant:\n%q\ngot:\n%q\n", want, got)
	}
}

// Test_Tote_Body_Identical_To_Location pins the deliberate Phase 0b
// invariant that Tote and Location produce byte-identical ZPL when
// given the same code. Per tote.go's package comment, the two
// templates diverge in Phase 1+ (lot-expiry/icon fields on totes);
// when that lands, this test must be deleted in the same commit
// that diverges them.
func Test_Tote_Body_Identical_To_Location(t *testing.T) {
	codes := []string{"STG-A02", "TOTE-007", "X", "12-CHAR-CODE"}
	for _, c := range codes {
		loc := zpl.Location(zpl.LocationData{Code: c})
		tote := zpl.Tote(zpl.ToteData{Code: c})
		if loc != tote {
			t.Fatalf("Tote/Location body drift for code %q.\nlocation:\n%q\ntote:\n%q\n",
				c, loc, tote)
		}
	}
}
