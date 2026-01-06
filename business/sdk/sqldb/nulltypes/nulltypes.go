package nulltypes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NullRawMessage is a nullable json.RawMessage for database scanning.
// It implements sql.Scanner to handle NULL values from LEFT JOINs and nullable columns.
type NullRawMessage struct {
	Data  json.RawMessage
	Valid bool
}

// Scan implements the sql.Scanner interface for NullRawMessage.
func (n *NullRawMessage) Scan(value interface{}) error {
	if value == nil {
		n.Data, n.Valid = nil, false
		return nil
	}
	n.Valid = true
	switch v := value.(type) {
	case []byte:
		n.Data = json.RawMessage(v)
	case string:
		n.Data = json.RawMessage(v)
	default:
		return fmt.Errorf("unsupported type for NullRawMessage: %T", value)
	}
	return nil
}

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

const layout string = "2006-01-02 15:04:05.9999999"

// =============================================================================
// Pointers to sql types
// =============================================================================

// ToNullInt64 converts an int pointer to a sql.NullInt32.
func ToNullInt64(value *int) sql.NullInt64 {
	var tmp int64
	valid := false
	if value != nil {
		tmp = int64(*value)
		valid = true
	}
	return sql.NullInt64{
		Int64: tmp,
		Valid: valid,
	}
}

// ToNullString converts a string pointer into sql.NullString.
func ToNullString(value *string) sql.NullString {
	var tmp string
	valid := false
	if value != nil {
		tmp = *value
		valid = true
	}
	return sql.NullString{
		String: tmp,
		Valid:  valid,
	}
}

// ToNullBool converts a bool pointer into a sql.NullBool.
func ToNullBool(value *bool) sql.NullBool {
	var tmp bool
	valid := false
	if value != nil {
		tmp = *value
		valid = true
	}
	return sql.NullBool{
		Bool:  tmp,
		Valid: valid,
	}
}

// ToNullTime converts a time.Time pointer to a sql.NullTime.
func ToNullTime(value *time.Time) sql.NullTime {
	var tmp time.Time
	valid := false
	if value != nil {
		tmp = *value
		valid = true
	}
	return sql.NullTime{
		Time:  tmp,
		Valid: valid,
	}
}

func ToNullTimeOnly(value time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  value,
		Valid: true,
	}
}

// ToNullFloat64 converts a float64 pointer to a sql.NullFloat64.
func ToNullFloat64(value *float64) sql.NullFloat64 {
	var tmp float64
	valid := false
	if value != nil {
		tmp = *value
		valid = true
	}
	return sql.NullFloat64{
		Float64: tmp,
		Valid:   valid,
	}
}

// =============================================================================
// SQL types to pointers
// =============================================================================

func BoolPtr(b sql.NullBool) *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

func Int64Ptr(i sql.NullInt64) *int {
	if !i.Valid {
		return nil
	}
	tmp := int(i.Int64)
	return &tmp
}

func Int32Ptr(i sql.NullInt32) *int {
	if !i.Valid {
		return nil
	}
	tmp := int(i.Int32)
	return &tmp
}

func Int16Ptr(i sql.NullInt16) *int {
	if !i.Valid {
		return nil
	}
	tmp := int(i.Int16)
	return &tmp
}

func Float64Ptr(f sql.NullFloat64) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

func StringPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func TimePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// =============================================================================
// SQL Data Migration Support
// =============================================================================

func FormatNULLString(input *string) string {
	if input != nil {
		tmp := strings.ReplaceAll(*input, "'", "''")
		// tmp = regexp.MustCompile(`(?i)ATTN:`).ReplaceAllString(tmp, "ATTN\\:")
		return fmt.Sprintf("'%s'", tmp)
	}
	return "NULL"
}

func FormatNULLInt(input *int) string {
	if input != nil {
		return fmt.Sprintf("%d", *input)
	}
	return "NULL"
}

func FormatNULLTime(input *time.Time) string {
	if input != nil {
		return fmt.Sprintf("'%s'", input.Format(layout))
	}
	return "NULL"
}

func FormatNULLTimeToUTC(input sql.NullTime, location string) string {
	loc, err := time.LoadLocation(location)
	if err != nil {
		fmt.Println("format null time to utc: %w", err)
		os.Exit(1)
	}
	if input.Valid {
		stringTime, err := time.ParseInLocation(layout, input.Time.Format(layout), loc)

		if err != nil {
			fmt.Println("format null time to utc: %w", err)
			os.Exit(1)
		}

		return fmt.Sprintf("'%s'", stringTime.UTC().Format(layout))
	}
	return "NULL"
}

func FormatNULLBool(input sql.NullBool) string {
	if input.Valid {
		if input.Bool {
			return "1"
		}
		return "0"
	}
	return "NULL"
}

// Note: probably not the best place for this one
func FormatBool(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
func FormatNULLMoney(input sql.NullString) string {
	if input.Valid {
		return input.String
	}
	return "NULL"
}
