package countrybus

import (
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID     *uuid.UUID
	Number *int
	Name   *string
	Alpha2 *string
	Alpha3 *string
}
