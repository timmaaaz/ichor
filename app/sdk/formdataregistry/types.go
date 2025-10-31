package formdataregistry

// EntityOperation defines the type of operation to perform on an entity.
type EntityOperation string

const (
	// OperationCreate indicates a new record should be created.
	OperationCreate EntityOperation = "create"

	// OperationUpdate indicates an existing record should be updated.
	OperationUpdate EntityOperation = "update"
)

// String returns the string representation of the operation.
func (e EntityOperation) String() string {
	return string(e)
}

// IsValid returns true if the operation is a known type.
func (e EntityOperation) IsValid() bool {
	switch e {
	case OperationCreate, OperationUpdate:
		return true
	default:
		return false
	}
}
