package nulltypes

import (
	"database/sql"

	"github.com/google/uuid"
)

func ToNullableUUID(u uuid.UUID) sql.NullString {
	if u == uuid.Nil {
		return sql.NullString{}
	}
	return sql.NullString{
		String: u.String(),
		Valid:  true,
	}
}

func FromNullableUUID(v sql.NullString) uuid.UUID {
	if !v.Valid {
		return uuid.Nil
	}
	u, err := uuid.Parse(v.String)
	if err != nil {
		panic(err)
	}
	return u
}
