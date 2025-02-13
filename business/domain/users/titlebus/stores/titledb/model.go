package titledb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/users/titlebus"
)

type title struct {
	ID          uuid.UUID `db:"title_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

func toDBTitle(t titlebus.Title) title {
	return title{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
	}
}

func toBusTitle(dbT title) titlebus.Title {
	return titlebus.Title{
		ID:          dbT.ID,
		Name:        dbT.Name,
		Description: dbT.Description,
	}
}

func toBusTitles(dbTs []title) []titlebus.Title {
	titles := make([]titlebus.Title, len(dbTs))
	for i, t := range dbTs {
		titles[i] = toBusTitle(t)
	}
	return titles
}
