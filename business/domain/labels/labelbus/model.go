package labelbus

import (
	"time"

	"github.com/google/uuid"
)

// Label types — matches catalog spec §3.3.
const (
	TypeLocation  = "location"
	TypeContainer = "container"
	TypeLot       = "lot"
	TypeSerial    = "serial"
	TypeProduct   = "product"
	TypeReceiving = "receiving"
	TypePick      = "pick"
)

// LabelCatalog represents one stable printable label definition.
type LabelCatalog struct {
	ID          uuid.UUID
	Code        string
	Type        string
	EntityRef   string
	PayloadJSON string
	CreatedDate time.Time
}

// NewLabelCatalog is what callers provide to create a new entry.
type NewLabelCatalog struct {
	Code        string
	Type        string
	EntityRef   string
	PayloadJSON string
}

// UpdateLabelCatalog carries optional patch fields.
type UpdateLabelCatalog struct {
	Code        *string
	Type        *string
	EntityRef   *string
	PayloadJSON *string
}
