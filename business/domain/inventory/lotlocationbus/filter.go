package lotlocationbus

import "github.com/google/uuid"

type QueryFilter struct {
	ID         *uuid.UUID
	LotID      *uuid.UUID
	LocationID *uuid.UUID
}
