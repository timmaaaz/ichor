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
