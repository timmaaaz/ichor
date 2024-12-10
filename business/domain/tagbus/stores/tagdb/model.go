package tagdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/tagbus"
)

type tag struct {
	ID          uuid.UUID `db:"tag_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

func toDBTag(bus tagbus.Tag) tag {
	return tag{
		ID:          bus.ID,
		Name:        bus.Name,
		Description: bus.Description,
	}
}

func toBusTag(dbTag tag) tagbus.Tag {
	return tagbus.Tag{
		ID:          dbTag.ID,
		Name:        dbTag.Name,
		Description: dbTag.Description,
	}
}

func toBusTags(dbTags []tag) []tagbus.Tag {
	busTags := make([]tagbus.Tag, len(dbTags))
	for i, dbTag := range dbTags {
		busTags[i] = toBusTag(dbTag)
	}
	return busTags
}
