package labelbus

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/zpl"
)

// Render dispatches to the appropriate ZPL template based on label Type.
// For location/container, uses only lc.Code. For product, unmarshals
// lc.PayloadJSON into ProductData (JSON tags match TS field names).
//
// All user-supplied string fields are passed through zpl.Sanitize to
// strip ZPL command-prefix characters (^ and ~) before rendering.
// Without this, a payload containing `^FS^XZ^XA...` could terminate
// a data field and inject arbitrary commands into the print stream.
func Render(lc LabelCatalog) ([]byte, error) {
	switch lc.Type {
	case TypeLocation:
		return []byte(zpl.Location(zpl.LocationData{Code: zpl.Sanitize(lc.Code)})), nil
	case TypeContainer:
		return []byte(zpl.Tote(zpl.ToteData{Code: zpl.Sanitize(lc.Code)})), nil
	case TypeProduct:
		var d zpl.ProductData
		if err := json.Unmarshal([]byte(lc.PayloadJSON), &d); err != nil {
			return nil, fmt.Errorf("unmarshal product: %w", err)
		}
		d.ProductName = zpl.Sanitize(d.ProductName)
		d.SKU = zpl.Sanitize(d.SKU)
		d.UPC = zpl.Sanitize(d.UPC)
		d.LotNumber = zpl.SanitizePtr(d.LotNumber)
		return []byte(zpl.Product(d)), nil
	default:
		return nil, fmt.Errorf("unknown label type: %q", lc.Type)
	}
}
