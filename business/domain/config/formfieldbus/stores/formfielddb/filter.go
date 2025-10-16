package formfielddb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

func applyFilter(filter formfieldbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.FormID != nil {
		data["form_id"] = *filter.FormID
		wc = append(wc, "form_id = :form_id")
	}

	if filter.Name != nil {
		data["name"] = *filter.Name
		wc = append(wc, "name = :name")
	}

	if filter.FieldType != nil {
		data["field_type"] = *filter.FieldType
		wc = append(wc, "field_type = :field_type")
	}

	if filter.Required != nil {
		data["required"] = *filter.Required
		wc = append(wc, "required = :required")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}