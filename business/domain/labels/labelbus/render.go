package labelbus

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/zpl"
)

// Render dispatches to the appropriate ZPL template based on label Type.
// For location/tote, uses only lc.Code. For receiving/pick, unmarshals
// lc.PayloadJSON into the typed data struct (JSON tags match TS field
// names in zpl/types.go so frontend payloads pass through unchanged).
func Render(lc LabelCatalog) ([]byte, error) {
	switch lc.Type {
	case TypeLocation:
		return []byte(zpl.Location(zpl.LocationData{Code: lc.Code})), nil
	case TypeTote:
		return []byte(zpl.Tote(zpl.ToteData{Code: lc.Code})), nil
	case TypeReceiving:
		var d zpl.ReceivingData
		if err := json.Unmarshal([]byte(lc.PayloadJSON), &d); err != nil {
			return nil, fmt.Errorf("unmarshal receiving: %w", err)
		}
		return []byte(zpl.Receiving(d)), nil
	case TypePick:
		var d zpl.PickData
		if err := json.Unmarshal([]byte(lc.PayloadJSON), &d); err != nil {
			return nil, fmt.Errorf("unmarshal pick: %w", err)
		}
		return []byte(zpl.Pick(d)), nil
	default:
		return nil, fmt.Errorf("unknown label type: %q", lc.Type)
	}
}
