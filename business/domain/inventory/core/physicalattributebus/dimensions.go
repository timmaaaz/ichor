package physicalattributebus

import (
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Dimension represents a physical attribute as a floating point number
type Dimension struct {
	value float64
}

const PRECISION = 4

// ParseDimension parses a float64 into a Dimension.
func ParseDimension(value string) (Dimension, error) {
	if err := ValidateDimensionFormat(value); err != nil {
		return Dimension{}, fmt.Errorf("invalid dimension format: %v", err)
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return Dimension{}, fmt.Errorf("error parsing dimension value: %w", err)
	}

	return Dimension{value: v}, nil
}

func MustParseDimension(value string) Dimension {
	dim, err := ParseDimension(value)
	if err != nil {
		panic(fmt.Sprintf("failed to parse dimension: %v", err))
	}
	return dim
}

// ParseDimensionFormat parses a float into a Dimension, truncates at 4 decimal places
func NewDimension(value float64) Dimension {
	return Dimension{value: toFixed(value, PRECISION)}
}

func ValidateDimensionFormat(value string) error {
	regex := regexp.MustCompile(`^\d*(\.\d{1,4})?$`)
	if !regex.MatchString(value) {
		return fmt.Errorf("invalid format. Expected a number with up to any amount of digits before the decimal point and exactly 4 digits after the decimal point")
	}
	return nil
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}
func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// Value returns the value of the Dimension.
func (d Dimension) Value() float64 {
	return toFixed(d.value, PRECISION)
}

// ValuePtr returns a pointer to the value of the Dimension.
func (d Dimension) ValuePtr() *float64 {
	v := toFixed(d.value, PRECISION)
	return &v
}

func (d Dimension) ToPtr() *Dimension {
	return &d
}

func (d Dimension) String() string {

	s := strconv.FormatFloat(d.value, 'f', PRECISION, 64)
	s = strings.TrimRight(s, "0")

	return s
}

// Equal returns true if the Dimension is equal to the given Dimension.
func (m Dimension) Equal(m2 Dimension) bool {
	return m.value == m2.value
}

func (d Dimension) ToNullFloat64() sql.NullFloat64 {
	if math.IsNaN(d.value) {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: toFixed(d.value, PRECISION), Valid: true}
}
