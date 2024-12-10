package titlebus

import "github.com/google/uuid"

type Title struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type NewTitle struct {
	Name        string
	Description string
}

type UpdateTitle struct {
	Name        *string
	Description *string
}
