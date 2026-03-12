package productuombus

import "errors"

// Set of error variables for CRUD operations.
var (
	ErrNotFound            = errors.New("product uom not found")
	ErrUniqueEntry         = errors.New("product uom entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)
