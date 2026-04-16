package zpl_test

import (
	"testing"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/zpl"
)

func Test_Sanitize(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain text passes through", "Widget A", "Widget A"},
		{"comma preserved", "Acme, Inc.", "Acme, Inc."},
		{"caret stripped", "Evil^FS^XZ^XA^FD", "EvilFSXZXAFD"},
		{"tilde stripped", "~JSNormal", "JSNormal"},
		{"both stripped", "^FS~HI", "FSHI"},
		{"unicode preserved", "Café Öl ✓", "Café Öl ✓"},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := zpl.Sanitize(tc.in)
			if got != tc.want {
				t.Fatalf("Sanitize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func Test_SanitizePtr_Nil(t *testing.T) {
	if got := zpl.SanitizePtr(nil); got != nil {
		t.Fatalf("SanitizePtr(nil) = %v, want nil", got)
	}
}

func Test_SanitizePtr_Value(t *testing.T) {
	in := "LOT-^FS"
	got := zpl.SanitizePtr(&in)
	if got == nil {
		t.Fatal("SanitizePtr returned nil for non-nil input")
	}
	if *got != "LOT-FS" {
		t.Fatalf("SanitizePtr: want %q, got %q", "LOT-FS", *got)
	}
}
