package productbus

import "fmt"

// InventoryType classifies a product's role in the supply chain.
// Nullable — customers without manufacturing do not need to set this.
type inventoryTypeSet struct {
	RawMaterial  InventoryType
	Component    InventoryType
	Consumable   InventoryType
	WIP          InventoryType
	FinishedGood InventoryType
}

// InventoryTypes is the set of allowed inventory type values.
var InventoryTypes = inventoryTypeSet{
	RawMaterial:  newInventoryType("raw_material"),
	Component:    newInventoryType("component"),
	Consumable:   newInventoryType("consumable"),
	WIP:          newInventoryType("wip"),
	FinishedGood: newInventoryType("finished_good"),
}

var inventoryTypeMap = make(map[string]InventoryType)

// InventoryType represents the product's role in the supply chain.
type InventoryType struct {
	name string
}

func newInventoryType(s string) InventoryType {
	it := InventoryType{s}
	inventoryTypeMap[s] = it
	return it
}

// String returns the string representation of the InventoryType.
func (it InventoryType) String() string {
	return it.name
}

// Equal returns true if the two InventoryTypes are equal.
func (it InventoryType) Equal(it2 InventoryType) bool {
	return it.name == it2.name
}

// MarshalText implements encoding.TextMarshaler.
func (it InventoryType) MarshalText() ([]byte, error) {
	return []byte(it.name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (it *InventoryType) UnmarshalText(data []byte) error {
	parsed, err := ParseInventoryType(string(data))
	if err != nil {
		return err
	}
	*it = parsed
	return nil
}

// ParseInventoryType parses a string into an InventoryType.
func ParseInventoryType(value string) (InventoryType, error) {
	it, exists := inventoryTypeMap[value]
	if !exists {
		return InventoryType{}, fmt.Errorf("invalid inventory type %q", value)
	}
	return it, nil
}
