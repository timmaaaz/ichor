package tagbus

import "github.com/google/uuid"

type Tag struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type NewTag struct {
	Name        string
	Description string
}

type UpdateTag struct {
	Name        *string
	Description *string
}
