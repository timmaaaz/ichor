package types

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type RoundedFloat struct {
	Value float64
}

func NewRoundedFloat(v float64) RoundedFloat {
	return RoundedFloat{ToFixed(v, 4)}
}

func ParseRoundedFloat(v string) (RoundedFloat, error) {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return RoundedFloat{}, fmt.Errorf("parsing error: %v", err)
	}

	return NewRoundedFloat(f), nil
}

func MustParseRoundedFloat(v string) RoundedFloat {
	rf, err := ParseRoundedFloat(v)
	if err != nil {
		panic(err)
	}
	return rf
}

func (rf RoundedFloat) Ptr() *RoundedFloat {
	return &rf
}

func (rf RoundedFloat) String() string {
	s := strconv.FormatFloat(rf.Value, 'f', 4, 64)

	return strings.TrimRight(s, "0")
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// DBValue returns the value of the Interval for database storage.
func (rf RoundedFloat) DBValue() sql.NullString {
	if rf.String() == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: rf.String(), Valid: true}
}
