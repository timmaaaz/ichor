package convert_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/foundation/convert"
)

func Test_Convert(t *testing.T) {
	type testSrc struct {
		Name   *string
		ID     *uuid.UUID
		Status *int
	}

	type testDest struct {
		Name   string
		ID     uuid.UUID
		Status int
	}

	u := uuid.New()
	i := 15
	src := testSrc{
		Name:   dbtest.StringPointer("test"),
		ID:     &u,
		Status: &i,
	}

	dst := testDest{}

	err := convert.PopulateSameTypes(src, &dst)
	if err != nil {
		t.Errorf("Failed to convert with err: %v", err)
	}

	if dst.Name != *src.Name || dst.ID != u || dst.Status != *src.Status {
		t.Error("Destination values did not match source values")
	}
}

func Test_ConvertMissingValue(t *testing.T) {
	type testSrc struct {
		Name   *string
		ID     *uuid.UUID
		Status *int
	}

	type testDest struct {
		Name   string
		ID     uuid.UUID
		Status int
	}

	i := 15
	src := testSrc{
		Name:   dbtest.StringPointer("test"),
		Status: &i,
	}

	dst := testDest{}

	err := convert.PopulateSameTypes(src, &dst)
	if err != nil {
		t.Errorf("Failed to convert with err: %v", err)
	}

	if dst.Name != *src.Name || dst.Status != *src.Status {
		t.Error("Destination values did not match source values")
	}
}

func Test_ConvertFromStrings(t *testing.T) {
	type testSrc struct {
		Name   *string
		ID     *string
		Status *string
	}

	type testDest struct {
		Name   string
		ID     uuid.UUID
		Status int
	}

	i := "15"
	u := uuid.NewString()
	name := "name"

	src := testSrc{
		Name:   &name,
		ID:     &u,
		Status: &i,
	}

	dest := testDest{}

	err := convert.PopulateTypesFromStrings(src, &dest)
	if err != nil {
		t.Errorf("Failed to convert with err: %v", err)
	}

	if dest.Name != *src.Name || dest.ID != uuid.MustParse(u) || dest.Status != 15 {
		t.Error("Destination values did not match source values")
	}
}

func Test_ConvertFromStringsToPointers(t *testing.T) {
	type testDest struct {
		Name   *string
		ID     *uuid.UUID
		Status *int
	}

	type testSrc struct {
		Name   *string
		ID     *string
		Status *string
	}

	id := uuid.New()
	src := testSrc{
		Name:   dbtest.StringPointer("test"),
		ID:     dbtest.StringPointer(id.String()),
		Status: dbtest.StringPointer("15"),
	}

	dest := testDest{}

	err := convert.PopulateTypesFromStrings(src, &dest)
	if err != nil {
		t.Errorf("Failed to convert with err: %v", err)
	}

	if *dest.Name != *src.Name || *dest.ID != id || *dest.Status != 15 {
		t.Error("Destination values did not match source values")
	}
}

type testSrc struct {
	Name   *string
	ID     *uuid.UUID
	Status *int
}

type testDest struct {
	Name   string
	ID     uuid.UUID
	Status int
}

var src = testSrc{
	Name:   dbtest.StringPointer("test"),
	ID:     dbtest.UUIDPointer(uuid.New()),
	Status: dbtest.IntPointer(15),
}

var dst = testDest{}

var result = testDest{}

func BenchmarkAssignment(t *testing.B) {
	for i := 0; i < t.N; i++ {
		manualAssignment(src, &dst)
	}
	result = dst
}

func BenchmarkConversion(t *testing.B) {
	for i := 0; i < t.N; i++ {
		convert.PopulateSameTypes(src, dst)
	}

	result = dst
}

func manualAssignment(src testSrc, dst *testDest) {
	if src.ID != nil {
		dst.ID = *src.ID
	}

	if src.Name != nil {
		dst.Name = *src.Name
	}

	if src.Status != nil {
		dst.Status = *src.Status
	}
}
